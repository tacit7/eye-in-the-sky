package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/urielmaldonado/eye-in-the-sky/internal/window"
)

// Server represents the dashboard HTTP server
type Server struct {
	templates     *template.Template
	port          string
	windowManager *window.Manager
}

// Agent represents an agent for template rendering
type Agent struct {
	ID                     string
	Status                 string
	WorktreePath          string
	FeatureDescription    string
	CurrentTask           string
	LastActivityAt        time.Time
	LastActivityFormatted string
	StatusIcon            string
}

// DashboardData represents the data passed to the main dashboard template
type DashboardData struct {
	Title             string
	Agents           []Agent
	Stats            DashboardStats
	RecentlyCompleted []Agent
}

// DashboardStats represents summary statistics for the dashboard
type DashboardStats struct {
	TotalActive    int
	TotalToday     int
	AvgDuration    string
	InactiveCount  int
}

// NewServer creates a new dashboard server
func NewServer(port string) *Server {
	return &Server{
		port:          port,
		windowManager: window.NewManager(),
	}
}

// LoadTemplates loads and parses HTML templates
func (s *Server) LoadTemplates(templateDir string) error {
	patterns := []string{
		filepath.Join(templateDir, "*.html"),
	}

	templates, err := template.ParseGlob(patterns[0])
	if err != nil {
		return err
	}

	s.templates = templates
	return nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Static file handler
	fs := http.FileServer(http.Dir("web/static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route handlers
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/agent/", s.handleAgentDetail)

	// API routes for AJAX calls
	http.HandleFunc("/api/agents/", s.handleAPIAgents)

	// Window management API routes
	http.HandleFunc("/api/window/get-id", s.handleGetWindowID)
	http.HandleFunc("/api/window/bring-to-front", s.handleBringToFront)
	http.HandleFunc("/api/window/list", s.handleListWindows)

	log.Printf("Dashboard server starting on http://localhost:%s", s.port)
	return http.ListenAndServe(":"+s.port, nil)
}

// handleIndex serves the main dashboard page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Mock data for now - this will be replaced with actual database queries
	mockAgents := []Agent{
		{
			ID:                     "a3f7d2e1",
			Status:                 "active",
			WorktreePath:          "/Users/dev/project1",
			FeatureDescription:    "User authentication system",
			CurrentTask:           "Implementing JWT token validation",
			LastActivityAt:        time.Now().Add(-10 * time.Minute),
			LastActivityFormatted: "10m ago",
			StatusIcon:            "circle-fill",
		},
		{
			ID:                     "b8c9e4f2",
			Status:                 "working",
			WorktreePath:          "/Users/dev/project2",
			FeatureDescription:    "API endpoint development",
			CurrentTask:           "Creating user registration endpoint",
			LastActivityAt:        time.Now().Add(-5 * time.Minute),
			LastActivityFormatted: "5m ago",
			StatusIcon:            "play-circle-fill",
		},
		{
			ID:                     "c1d5a7b3",
			Status:                 "idle",
			WorktreePath:          "/Users/dev/project3",
			FeatureDescription:    "Frontend component library",
			CurrentTask:           "",
			LastActivityAt:        time.Now().Add(-2 * time.Hour),
			LastActivityFormatted: "2h ago",
			StatusIcon:            "pause-circle-fill",
		},
	}

	data := DashboardData{
		Title:  "Agent Dashboard",
		Agents: mockAgents,
		Stats: DashboardStats{
			TotalActive:   len(mockAgents),
			TotalToday:    5,
			AvgDuration:   "2h 15m",
			InactiveCount: 1,
		},
		RecentlyCompleted: []Agent{},
	}

	if err := s.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleAgentDetail serves the agent detail page
func (s *Server) handleAgentDetail(w http.ResponseWriter, r *http.Request) {
	// Extract agent ID from URL path
	agentID := r.URL.Path[len("/agent/"):]
	if agentID == "" {
		http.NotFound(w, r)
		return
	}

	// Mock agent data - this will be replaced with actual database queries
	mockAgent := Agent{
		ID:                     agentID,
		Status:                 "active",
		WorktreePath:          "/Users/dev/project1",
		FeatureDescription:    "User authentication system",
		CurrentTask:           "Implementing JWT token validation",
		LastActivityAt:        time.Now().Add(-10 * time.Minute),
		LastActivityFormatted: "10m ago",
		StatusIcon:            "circle-fill",
	}

	// Mock action data
	type Action struct {
		Description        string
		ActionType         string
		TimestampFormatted string
		TypeIcon           string
		TypeColor          string
		Details            string
	}

	mockActions := []Action{
		{
			Description:        "Started working on JWT token validation",
			ActionType:         "task_start",
			TimestampFormatted: "10m ago",
			TypeIcon:          "play-fill",
			TypeColor:         "primary",
			Details:           "Beginning implementation of token validation middleware",
		},
		{
			Description:        "Created auth middleware file",
			ActionType:         "file_operation",
			TimestampFormatted: "8m ago",
			TypeIcon:          "file-earmark-plus",
			TypeColor:         "success",
			Details:           "Created middleware/auth.go",
		},
		{
			Description:        "Updated status to active",
			ActionType:         "status_update",
			TimestampFormatted: "15m ago",
			TypeIcon:          "arrow-clockwise",
			TypeColor:         "info",
		},
	}

	// Mock commit data
	type Commit struct {
		Hash               string
		Message            string
		TimestampFormatted string
	}

	mockCommits := []Commit{
		{
			Hash:               "a3f7d2e1",
			Message:            "Add JWT token validation middleware",
			TimestampFormatted: "12m ago",
		},
		{
			Hash:               "b8c9e4f2",
			Message:            "Update authentication routes",
			TimestampFormatted: "25m ago",
		},
	}

	data := struct {
		Title   string
		Agent   Agent
		Actions []Action
		Commits []Commit
	}{
		Title:   "Agent Detail",
		Agent:   mockAgent,
		Actions: mockActions,
		Commits: mockCommits,
	}

	if err := s.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleAPIAgents handles API calls for agent management
func (s *Server) handleAPIAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract agent ID and action from URL
	_ = r.URL.Path[len("/api/agents/"):]

	switch r.Method {
	case "POST":
		// For now, just return success
		// This will be implemented with actual database operations
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message": "Action completed"}`))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetWindowID handles requests to get current window ID
func (s *Server) handleGetWindowID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Application string `json:"application"`
		WindowTitle string `json:"window_title,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Default to current application if not specified
	if req.Application == "" {
		req.Application = "current"
	}

	// Get window information
	windowInfo, err := s.windowManager.GetCurrentWindowID(req.Application, req.WindowTitle)
	if err != nil {
		resp := map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Return success response
	resp := map[string]interface{}{
		"success":     true,
		"message":     "Window ID retrieved successfully",
		"window_id":   windowInfo.ID,
		"window_info": windowInfo.Title,
		"application": windowInfo.Application,
		"position":    windowInfo.Position,
	}

	json.NewEncoder(w).Encode(resp)
}

// handleBringToFront handles requests to bring a window to front
func (s *Server) handleBringToFront(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Application string `json:"application"`
		WindowTitle string `json:"window_title,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Bring window to front
	if err := s.windowManager.BringToFront(req.Application, req.WindowTitle); err != nil {
		resp := map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Return success response
	resp := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Brought %s to front", req.Application),
	}

	json.NewEncoder(w).Encode(resp)
}

// handleListWindows handles requests to list all windows of an application
func (s *Server) handleListWindows(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get application from query parameter
	application := r.URL.Query().Get("application")
	if application == "" {
		application = "current"
	}

	// Get window list
	windows, err := s.windowManager.GetAllWindows(application)
	if err != nil {
		resp := map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Return success response
	resp := map[string]interface{}{
		"success": true,
		"message": "Windows retrieved successfully",
		"windows": windows,
	}

	json.NewEncoder(w).Encode(resp)
}