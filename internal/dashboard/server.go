package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/urielmaldonado/eye-in-the-sky/internal/database"
	"github.com/urielmaldonado/eye-in-the-sky/internal/window"
)

// Server represents the dashboard HTTP server
type Server struct {
	templates     *template.Template
	port          string
	windowManager *window.Manager
	db            *database.DB
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
func NewServer(port string, db *database.DB) *Server {
	return &Server{
		port:          port,
		windowManager: window.NewManager(),
		db:            db,
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

	// Get all agents from database
	dbAgents, err := s.db.ListAgents("")
	if err != nil {
		log.Printf("Error fetching agents: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Convert database agents to dashboard agents
	var agents []Agent
	for _, dbAgent := range dbAgents {
		agent := Agent{
			ID:                     dbAgent.ID,
			Status:                 dbAgent.Status,
			WorktreePath:          getStringValue(dbAgent.GitWorktreePath),
			FeatureDescription:    getStringValue(dbAgent.FeatureDescription),
			CurrentTask:           getStringValue(dbAgent.CurrentTask),
			LastActivityAt:        getTimeValue(dbAgent.LastActivityAt, dbAgent.UpdatedAt),
			LastActivityFormatted: formatTimeAgo(getTimeValue(dbAgent.LastActivityAt, dbAgent.UpdatedAt)),
			StatusIcon:            getStatusIcon(dbAgent.Status),
		}
		agents = append(agents, agent)
	}

	data := DashboardData{
		Title:  "Agent Dashboard",
		Agents: agents,
		Stats: DashboardStats{
			TotalActive:   len(agents), // TODO: Calculate actual stats
			TotalToday:    len(agents),
			AvgDuration:   "N/A",
			InactiveCount: 0,
		},
		RecentlyCompleted: []Agent{}, // TODO: Get completed agents
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

// Helper functions for converting database values

// getStringValue safely gets string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getTimeValue safely gets time value from pointer, with fallback
func getTimeValue(ptr *time.Time, fallback time.Time) time.Time {
	if ptr == nil {
		return fallback
	}
	return *ptr
}

// formatTimeAgo formats a time as "X ago" format
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration.Hours() >= 24 {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	} else if duration.Hours() >= 1 {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else if duration.Minutes() >= 1 {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	} else {
		return "just now"
	}
}

// getStatusIcon returns the appropriate Bootstrap icon for an agent status
func getStatusIcon(status string) string {
	switch status {
	case database.StatusActive:
		return "circle-fill"
	case database.StatusWorking:
		return "play-circle-fill"
	case database.StatusIdle:
		return "pause-circle-fill"
	case database.StatusCompleted:
		return "check-circle-fill"
	case database.StatusFailed:
		return "x-circle-fill"
	default:
		return "question-circle-fill"
	}
}