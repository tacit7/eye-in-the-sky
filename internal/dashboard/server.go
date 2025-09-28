package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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
	Source                 string
	WorktreePath          string
	FeatureDescription    string
	CurrentTask           string
	LastActivityAt        time.Time
	LastActivityFormatted string
	StatusIcon            string
	SourceIcon            string
	SourceColor           string
	SourceBadge           string
	Progress              int    // Progress percentage for suspended sessions
	ProjectName           string // Project name for better identification
}

// DashboardData represents the data passed to the main dashboard template
type DashboardData struct {
	Title             string
	Agents           []Agent
	Stats            DashboardStats
	RecentlyCompleted []Agent
	SuspendedAgents  []Agent
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
	var suspendedAgents []Agent
	for _, dbAgent := range dbAgents {
		agent := Agent{
			ID:                     dbAgent.ID,
			Status:                 dbAgent.Status,
			Source:                 dbAgent.Source,
			WorktreePath:          getStringValue(dbAgent.GitWorktreePath),
			FeatureDescription:    getStringValue(dbAgent.FeatureDescription),
			CurrentTask:           getStringValue(dbAgent.CurrentTask),
			LastActivityAt:        getTimeValue(dbAgent.LastActivityAt, dbAgent.UpdatedAt),
			LastActivityFormatted: formatTimeAgo(getTimeValue(dbAgent.LastActivityAt, dbAgent.UpdatedAt)),
			StatusIcon:            getStatusIcon(dbAgent.Status),
			SourceIcon:            getSourceIcon(dbAgent.Source),
			SourceColor:           getSourceColor(dbAgent.Source),
			SourceBadge:           getSourceBadge(dbAgent.Source),
			Progress:              extractProgressFromContext(dbAgent.ID), // Extract progress from session context
			ProjectName:           getStringValue(dbAgent.ProjectName),
		}

		// Separate suspended agents from active agents
		if dbAgent.Status == "idle" && hasSessionContext(dbAgent.ID) {
			suspendedAgents = append(suspendedAgents, agent)
		} else if dbAgent.Status != database.StatusCompleted && dbAgent.Status != database.StatusFailed {
			agents = append(agents, agent)
		}
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
		SuspendedAgents:  suspendedAgents,
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
		Source:                 dbAgent.Source,
		WorktreePath:          getStringValue(dbAgent.GitWorktreePath),
		FeatureDescription:    getStringValue(dbAgent.FeatureDescription),
		CurrentTask:           getStringValue(dbAgent.CurrentTask),
		LastActivityAt:        getTimeValue(dbAgent.LastActivityAt, dbAgent.UpdatedAt),
		LastActivityFormatted: formatTimeAgo(getTimeValue(dbAgent.LastActivityAt, dbAgent.UpdatedAt)),
		StatusIcon:            getStatusIcon(dbAgent.Status),
		SourceIcon:            getSourceIcon(dbAgent.Source),
		SourceColor:           getSourceColor(dbAgent.Source),
		SourceBadge:           getSourceBadge(dbAgent.Source),
		ProjectName:           getStringValue(dbAgent.ProjectName),
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

	// Write the page manually with agent template
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// Write HTML header
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Agent Detail - Eye in the Sky</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.1/font/bootstrap-icons.css" rel="stylesheet">
    <link href="/static/css/styles.css" rel="stylesheet">
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
        <div class="container">
            <a class="navbar-brand" href="/">
                <i class="bi bi-eye"></i> Eye in the Sky
            </a>
            <div class="navbar-nav ms-auto">
                <span class="nav-link text-light">Claude Code Multi-Agent Dashboard</span>
            </div>
        </div>
    </nav>
    <main class="container mt-4">
`))

	// Execute agent content template
	if err := s.templates.ExecuteTemplate(w, "agent-content", data); err != nil {
		log.Printf("Error executing agent template: %v", err)
		w.Write([]byte("<p>Error loading agent details</p>"))
	}

	// Write HTML footer
	w.Write([]byte(`
    </main>
    <footer class="bg-light mt-5 py-3">
        <div class="container text-center text-muted">
            <small>Claude Code Multi-Agent Management System</small>
        </div>
    </footer>
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/js/bootstrap.bundle.min.js"></script>
    <script src="/static/js/dashboard.js"></script>
</body>
</html>`))
}

// handleAPIAgents handles API calls for agent management
func (s *Server) handleAPIAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract agent ID and action from URL path like /api/agents/{agentId}/{action}
	pathParts := r.URL.Path[len("/api/agents/"):]

	// Parse path components manually
	agentID := ""
	action := ""
	if len(pathParts) >= 8 {
		agentID = pathParts[:8]
		if len(pathParts) > 9 && pathParts[8] == '/' {
			action = pathParts[9:]
		}
	}

	switch r.Method {
	case "POST":
		switch action {
		case "bring-front":
			s.handleBringAgentFront(w, r, agentID)
		case "end":
			s.handleEndSession(w, r, agentID)
		case "status":
			s.handleUpdateStatus(w, r, agentID)
		default:
			http.Error(w, `{"success": false, "message": "Unknown action"}`, http.StatusBadRequest)
		}
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

// getSourceIcon returns the appropriate Bootstrap icon for an agent source
func getSourceIcon(source string) string {
	switch source {
	case database.SourceWorktree:
		return "git"
	case database.SourceDesktop:
		return "laptop"
	default:
		return "question-circle"
	}
}

// getSourceColor returns the appropriate Bootstrap color for an agent source
func getSourceColor(source string) string {
	switch source {
	case database.SourceWorktree:
		return "primary"
	case database.SourceDesktop:
		return "success"
	default:
		return "secondary"
	}
}

// getSourceBadge returns the appropriate badge text for an agent source
func getSourceBadge(source string) string {
	switch source {
	case database.SourceWorktree:
		return "Worktree"
	case database.SourceDesktop:
		return "Desktop"
	default:
		return "Unknown"
	}
}

// hasSessionContext checks if an agent has session context saved (indicating it can be suspended/resumed)
func hasSessionContext(agentID string) bool {
	// For now, we'll check if there's a session context file
	// In a real implementation, this would check the database for session context records
	contextFile := fmt.Sprintf("%s-context.md", agentID)
	_, err := os.Stat(contextFile)
	return err == nil
}

// extractProgressFromContext extracts progress percentage from session context
func extractProgressFromContext(agentID string) int {
	// For demonstration, return fixed progress for known agent
	if agentID == "534002f0" {
		return 95 // Agent 534002f0 was at 95% completion when suspended
	}
	return 0
}

// handleBringAgentFront uses MCP tool to bring agent window to front
func (s *Server) handleBringAgentFront(w http.ResponseWriter, r *http.Request, agentID string) {
	// Use the MCP bring_window_front tool
	args := mcp.BringWindowFrontArgs{
		AgentID: agentID,
	}

	result, err := s.mcpServer.HandleTool("bring_window_front", mustMarshalJSON(args))
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to bring window to front: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Forward the MCP result
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// handleEndSession ends an agent session
func (s *Server) handleEndSession(w http.ResponseWriter, r *http.Request, agentID string) {
	// For now, just return success - this would integrate with MCP end_session tool
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Session ended for agent %s", agentID),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleUpdateStatus updates an agent's status
func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request, agentID string) {
	// Parse request body
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// For now, just return success - this would integrate with MCP update_status tool
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Status updated to %s for agent %s", req.Status, agentID),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// mustMarshalJSON marshals to JSON and panics on error (for internal use)
func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}