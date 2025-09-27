package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/mcp"
	"github.com/tacit7/eye-in-the-sky/internal/window"
)

// Server represents the dashboard HTTP server
type Server struct {
	templates     *template.Template
	port          string
	windowManager *window.Manager
	db            *database.DB
	mcpServer     *mcp.Server
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
		mcpServer:     mcp.NewServer(db),
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

	// MCP tool API routes
	http.HandleFunc("/api/mcp/tools/", s.handleMCPTools)

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

	// Get agent from database
	dbAgent, err := s.db.GetAgent(agentID)
	if err != nil {
		log.Printf("Error fetching agent %s: %v", agentID, err)
		http.NotFound(w, r)
		return
	}

	// Convert database agent to dashboard agent
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

	// Get actions from database
	type Action struct {
		Description        string
		ActionType         string
		TimestampFormatted string
		TypeIcon           string
		TypeColor          string
		Details            string
	}

	dbActions, err := s.db.ListActions(agentID)
	if err != nil {
		log.Printf("Error fetching actions for agent %s: %v", agentID, err)
		// Continue with empty actions rather than failing
	}

	var actions []Action
	for _, dbAction := range dbActions {
		action := Action{
			Description:        dbAction.Description,
			ActionType:         dbAction.ActionType,
			TimestampFormatted: formatTimeAgo(dbAction.Timestamp),
			TypeIcon:          getActionTypeIcon(dbAction.ActionType),
			TypeColor:         getActionTypeColor(dbAction.ActionType),
			Details:           getStringValue(dbAction.Details),
		}
		actions = append(actions, action)
	}

	// Get commits from database
	type Commit struct {
		Hash               string
		Message            string
		TimestampFormatted string
	}

	dbCommits, err := s.db.ListCommits(agentID)
	if err != nil {
		log.Printf("Error fetching commits for agent %s: %v", agentID, err)
		// Continue with empty commits rather than failing
	}

	var commits []Commit
	for _, dbCommit := range dbCommits {
		commit := Commit{
			Hash:               dbCommit.CommitHash,
			Message:            getStringValue(dbCommit.CommitMessage),
			TimestampFormatted: formatTimeAgo(dbCommit.Timestamp),
		}
		commits = append(commits, commit)
	}

	data := struct {
		Title   string
		Agent   Agent
		Actions []Action
		Commits []Commit
	}{
		Title:   "Agent Detail",
		Agent:   agent,
		Actions: actions,
		Commits: commits,
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

// handleMCPTools handles MCP tool calls via the dashboard
func (s *Server) handleMCPTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tool name from URL path
	toolName := r.URL.Path[len("/api/mcp/tools/"):]
	if toolName == "" {
		http.Error(w, "Tool name required", http.StatusBadRequest)
		return
	}

	// Read request body
	var requestBody json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Call MCP tool
	result, err := s.mcpServer.HandleTool(toolName, requestBody)
	if err != nil {
		resp := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Return result
	resp := map[string]interface{}{
		"success": true,
		"result":  result,
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

// getActionTypeIcon returns the appropriate Bootstrap icon for an action type
func getActionTypeIcon(actionType string) string {
	switch actionType {
	case database.ActionTaskStart:
		return "play-fill"
	case database.ActionFileOperation:
		return "file-earmark-plus"
	case database.ActionGitCommit:
		return "git"
	case database.ActionStatusUpdate:
		return "arrow-clockwise"
	default:
		return "activity"
	}
}

// getActionTypeColor returns the appropriate Bootstrap color for an action type
func getActionTypeColor(actionType string) string {
	switch actionType {
	case database.ActionTaskStart:
		return "primary"
	case database.ActionFileOperation:
		return "success"
	case database.ActionGitCommit:
		return "warning"
	case database.ActionStatusUpdate:
		return "info"
	default:
		return "secondary"
	}
}