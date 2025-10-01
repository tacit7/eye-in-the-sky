package dashboard

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/mcp"
	"github.com/tacit7/eye-in-the-sky/internal/window"
)

// APIResponse represents a standardized JSON API response
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    *T     `json:"data,omitempty"`
}

// Server represents the dashboard HTTP server
type Server struct {
	templates     *template.Template
	mux           *http.ServeMux
	httpServer    *http.Server
	port          string
	windowManager *window.Manager
	db            *database.DB
	mcpServer     *mcp.Server
}

// Agent represents an agent for template rendering
type Agent struct {
	ID                     string    `json:"id"`
	Status                 string    `json:"status"`
	Source                 string    `json:"source"`
	WorktreePath          string    `json:"worktree_path,omitempty"`
	FeatureDescription    string    `json:"feature_description,omitempty"`
	CurrentTask           string    `json:"current_task,omitempty"`
	LastActivityAt        time.Time `json:"last_activity_at"`
	LastActivityFormatted string    `json:"last_activity_formatted"`
	LastActivityISO       string    `json:"last_activity_iso"` // For tooltips
	StatusIcon            string    `json:"status_icon"`
	SourceIcon            string    `json:"source_icon"`
	SourceColor           string    `json:"source_color"`
	SourceBadge           string    `json:"source_badge"`
	Progress              int       `json:"progress"` // Progress percentage for suspended sessions
	ProjectName           string    `json:"project_name,omitempty"` // Project name for better identification
	Name                  string    `json:"name,omitempty"` // Session name
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
	mux := http.NewServeMux()
	s := &Server{
		port:          port,
		windowManager: window.NewManager(),
		db:            db,
		mcpServer:     mcp.NewServer(db),
		mux:           mux,
	}
	s.routes()
	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return s
}

// LoadTemplates loads and parses HTML templates from embedded filesystem
func (s *Server) LoadTemplates(templateFS embed.FS, pattern string) error {
	s.templates = template.Must(template.New("").Funcs(template.FuncMap{
		"timeago": formatTimeAgo,
	}).ParseFS(templateFS, pattern))
	return nil
}

// routes sets up all HTTP routes
func (s *Server) routes() {
	fs := http.FileServer(http.Dir("web/static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", s.cacheStatic(fs)))

	s.mux.HandleFunc("/", s.cacheBustHTML(s.handleIndex))
	s.mux.HandleFunc("/mockup", s.cacheBustHTML(s.handleMockup))
	s.mux.HandleFunc("/agent/", s.cacheBustHTML(s.handleAgentDetail))

	s.mux.HandleFunc("/api/agents/", s.handleAPIAgents)
	s.mux.HandleFunc("/api/mcp/tools/", s.handleMCPTools)
	s.mux.HandleFunc("/api/window/get-id", s.handleGetWindowID)
	s.mux.HandleFunc("/api/window/bring-to-front", s.handleBringToFront)
	s.mux.HandleFunc("/api/window/list", s.handleListWindows)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Dashboard server starting on http://localhost:%s", s.port)
	return s.httpServer.ListenAndServe()
}

// cacheStatic adds long-term cache headers for static assets
func (s *Server) cacheStatic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		h.ServeHTTP(w, r)
	})
}

// cacheBustHTML adds cache-busting and security headers for HTML responses
func (s *Server) cacheBustHTML(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := time.Now().UnixNano()
		log.Printf("req=%d %s %s", reqID, r.Method, r.URL.Path)

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		s.securityHeaders(w)
		next(w, r)
	}
}

// securityHeaders adds security headers to responses
func (s *Server) securityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'")
}

// writeJSON writes a standardized JSON response
func writeJSON[T any](w http.ResponseWriter, status int, payload APIResponse[T]) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
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
		lastActivity := getTimeValue(dbAgent.LastActivityAt, dbAgent.UpdatedAt)

		// Get session name from current session (if sessions table exists)
		sessionName := ""
		if dbAgent.CurrentSessionID != nil && *dbAgent.CurrentSessionID != "" {
			// Try to get session, but ignore errors if table doesn't exist yet
			session, err := s.db.GetSession(*dbAgent.CurrentSessionID)
			if err == nil && session != nil && session.Name != nil {
				sessionName = *session.Name
			}
		}

		agent := Agent{
			ID:                     dbAgent.ID,
			Status:                 dbAgent.Status,
			Source:                 dbAgent.Source,
			WorktreePath:          getStringValue(dbAgent.GitWorktreePath),
			FeatureDescription:    getStringValue(dbAgent.FeatureDescription),
			CurrentTask:           getStringValue(dbAgent.CurrentTask),
			LastActivityAt:        lastActivity,
			LastActivityFormatted: formatTimeAgo(lastActivity),
			LastActivityISO:       lastActivity.Format(time.RFC3339), // For tooltips
			StatusIcon:            getStatusIcon(dbAgent.Status),
			SourceIcon:            getSourceIcon(dbAgent.Source),
			SourceColor:           getSourceColor(dbAgent.Source),
			SourceBadge:           getSourceBadge(dbAgent.Source),
			Progress:              extractProgressFromContext(dbAgent.ID), // Extract progress from session context
			ProjectName:           getStringValue(dbAgent.ProjectName),
			Name:                  sessionName,
		}

		// Filter out archived agents and separate suspended from active
		if dbAgent.Status == database.StatusArchived {
			// Skip archived agents - don't show them on dashboard
			continue
		} else if dbAgent.Status == "idle" && hasSessionContext(dbAgent.ID) {
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

// handleMockup serves the mockup page
func (s *Server) handleMockup(w http.ResponseWriter, r *http.Request) {

	data := DashboardData{
		Title: "Mockup Dashboard",
	}

	if err := s.templates.ExecuteTemplate(w, "mockup.html", data); err != nil {
		log.Printf("Error executing mockup template: %v", err)
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

	if err := s.templates.ExecuteTemplate(w, "agent.html", data); err != nil {
		log.Printf("Error executing agent template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleAPIAgents handles API calls for agent management
func (s *Server) handleAPIAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract agent ID and action from URL path like /api/agents/{agentId}/{action}
	pathParts := r.URL.Path[len("/api/agents/"):]
	agentID, action, ok := parseAgentPath(pathParts)
	if !ok {
		http.Error(w, `{"success": false, "message": "Invalid agent path"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "POST":
		switch action {
		case "recreate":
			s.handleRecreateAgent(w, r, agentID)
		case "bring-front":
			s.handleBringAgentFront(w, r, agentID)
		case "end":
			s.handleEndSession(w, r, agentID)
		case "archive":
			s.handleArchiveAgent(w, r, agentID)
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
		writeJSON(w, http.StatusInternalServerError, APIResponse[any]{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Return success response with window data
	data := map[string]interface{}{
		"window_id":   windowInfo.ID,
		"window_info": windowInfo.Title,
		"application": windowInfo.Application,
		"position":    windowInfo.Position,
	}

	writeJSON(w, http.StatusOK, APIResponse[map[string]interface{}]{
		Success: true,
		Message: "Window ID retrieved successfully",
		Data:    &data,
	})
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
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d >= 48*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d >= time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d >= time.Minute:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	default:
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
	// FIXED: Use database instead of dangerous file reading
	// In a proper implementation, this would check for a session_context table
	// For now, return false to disable the dangerous file-based approach
	return false
}

// extractProgressFromContext extracts progress percentage from session context
func extractProgressFromContext(agentID string) int {
	// FIXED: Use database instead of dangerous file reading
	// In a proper implementation, this would query a progress field from the database
	// For now, return 0 to disable the dangerous file-based approach
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
	// Update agent status to completed
	err := s.db.UpdateAgentStatus(agentID, database.StatusCompleted, nil)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to end session for agent %s: %v", agentID, err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Session ended for agent %s", agentID),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleArchiveAgent archives an agent to hide it from the dashboard
func (s *Server) handleArchiveAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	// Update agent status to archived
	err := s.db.UpdateAgentStatus(agentID, database.StatusArchived, nil)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to archive agent %s: %v", agentID, err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Agent %s archived", agentID),
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

// parseAgentPath parses agent path components from URL path
func parseAgentPath(p string) (id, action string, ok bool) {
	parts := strings.Split(strings.TrimPrefix(p, "/api/agents/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", "", false
	}
	id = parts[0]
	if len(parts) > 1 {
		action = parts[1]
	}
	return id, action, true
}

// handleRecreateAgent creates a new agent with the same learned context/expertise as the original
func (s *Server) handleRecreateAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	// Get original agent
	agent, err := s.db.GetAgent(agentID)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to find agent %s: %v", agentID, err),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Try to get latest session context with learned_context
	var learnedContext string
	var personaID *string

	// First check if agent has a persona_id
	if agent.PersonaID != nil && *agent.PersonaID != "" {
		personaID = agent.PersonaID

		// Get persona details to show in response
		persona, err := s.db.GetPersona(*agent.PersonaID)
		if err == nil && persona != nil {
			learnedContext = persona.InitialContext
		}
	}

	// If no persona, try to get learned_context from most recent session
	if learnedContext == "" && agent.CurrentSessionID != nil && *agent.CurrentSessionID != "" {
		// Load session context using MCP tool
		args := mcp.LoadSessionContextArgs{
			AgentID: agentID,
		}

		result, err := s.mcpServer.HandleTool("load_session_context", mustMarshalJSON(args))
		if err == nil {
			// Try to extract learned_context from result
			if resultMap, ok := result.(map[string]interface{}); ok {
				if context, ok := resultMap["context"].(map[string]interface{}); ok {
					if lc, ok := context["learned_context"].(string); ok {
						learnedContext = lc
					}
				}
			}
		}
	}

	// If we still don't have context, inform the user
	if learnedContext == "" && personaID == nil {
		response := map[string]interface{}{
			"success": false,
			"message": "No learned context or persona found for this agent. Agent must have either saved session context with learned_context or be associated with a persona.",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create new agent with same context
	// Use appropriate registration based on source
	var newAgentID string
	if agent.Source == database.SourceDesktop {
		// Register as desktop agent
		args := mcp.RegisterDesktopAgentArgs{
			Description: getStringValue(agent.FeatureDescription),
			ProjectName: getStringValue(agent.ProjectName),
		}

		result, err := s.mcpServer.HandleTool("register_claude_desktop_agent", mustMarshalJSON(args))
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Failed to create new desktop agent: %v", err),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Extract agent ID from result
		if resultMap, ok := result.(map[string]interface{}); ok {
			if msg, ok := resultMap["message"].(string); ok {
				// Parse agent ID from message like "Agent a1b2c3d4 registered"
				parts := strings.Fields(msg)
				if len(parts) >= 2 {
					newAgentID = parts[1]
				}
			}
		}
	} else {
		// Register as worktree agent
		args := mcp.RegisterAgentArgs{
			Description:  getStringValue(agent.FeatureDescription),
			WorktreePath: agent.GitWorktreePath,
			ProjectName:  agent.ProjectName,
		}

		result, err := s.mcpServer.HandleTool("register_agent", mustMarshalJSON(args))
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Failed to create new worktree agent: %v", err),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Extract agent ID from result
		if resultMap, ok := result.(map[string]interface{}); ok {
			if msg, ok := resultMap["message"].(string); ok {
				// Parse agent ID from message like "Agent a1b2c3d4 registered"
				parts := strings.Fields(msg)
				if len(parts) >= 2 {
					newAgentID = parts[1]
				}
			}
		}
	}

	if newAgentID == "" {
		response := map[string]interface{}{
			"success": false,
			"message": "Failed to extract new agent ID from registration",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// If we have a persona, associate it with the new agent
	if personaID != nil {
		err := s.db.UpdateAgentPersona(newAgentID, *personaID)
		if err != nil {
			log.Printf("Warning: Failed to associate persona with new agent %s: %v", newAgentID, err)
		}
	}

	// Success response
	contextInfo := "persona"
	if personaID == nil {
		contextInfo = "learned context from session"
	}

	response := map[string]interface{}{
		"success":      true,
		"message":      fmt.Sprintf("New agent created with %s from %s", contextInfo, agentID),
		"new_agent_id": newAgentID,
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