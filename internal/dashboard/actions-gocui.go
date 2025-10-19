package dashboard

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jroimartin/gocui"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/window"
)

// continueSession opens a new shell and runs claude -c with the session ID
func (a *App) continueSession(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}

	agent := a.agents[a.selectedIdx]

	// Get session ID
	sessionID := ""
	if agent.CurrentSessionID != nil {
		sessionID = *agent.CurrentSessionID
	} else {
		return fmt.Errorf("no session ID for agent %s", agent.ID)
	}

	// Suspend GUI to run external command
	a.gui.Close()

	// Run claude --resume command
	cmd := exec.Command("claude", "--resume", sessionID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\n\033[1;36m→ Resuming session: %s\033[0m\n", sessionID)
	if err := cmd.Run(); err != nil {
		fmt.Printf("\033[1;31m✗ Failed to continue session: %v\033[0m\n", err)
	}

	fmt.Println("\n\033[1;32m✓ Session ended. Press Enter to return to dashboard...\033[0m")
	fmt.Scanln()

	// Reinitialize GUI
	newGui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	a.gui = newGui
	a.gui.SetManagerFunc(a.layout)
	a.gui.Cursor = true
	a.gui.Mouse = false

	// Clear any detail view state
	a.currentAgentID = ""

	// Reinitialize keybinding system
	a.keys = NewKeyRegistry(newGui)
	a.modes = NewModeManager()
	a.views = NewViewManager(a)

	// Re-register all keybinding tags
	a.RegisterGlobalKeys()
	a.RegisterMainKeys()
	a.RegisterDetailsKeys()
	a.RegisterNavigationKeys()
	a.RegisterEditKeys()

	// Activate initial tags
	a.keys.BindTag("global")
	a.keys.BindTag("main")
	a.keys.BindTag("navigation")

	// Force layout to run to create views
	if err := a.layout(a.gui); err != nil {
		return err
	}

	// Refresh data
	return a.refreshAgents()
}

// startSession starts a new claude session with -s flag and exits dashboard
func (a *App) startSession(g *gocui.Gui, v *gocui.View) error {
	// Determine which agent to use based on current view
	var sessionID string

	// If we're in detail view, use currentAgentID
	if a.currentAgentID != "" {
		agent, err := a.db.GetAgent(a.currentAgentID)
		if err != nil {
			return fmt.Errorf("failed to get agent: %v", err)
		}
		if agent.CurrentSessionID != nil {
			sessionID = *agent.CurrentSessionID
		} else {
			return fmt.Errorf("no session ID for agent %s", agent.ID)
		}
	} else {
		// In main list view, use selected agent
		if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
			return nil
		}
		agent := a.agents[a.selectedIdx]
		if agent.CurrentSessionID != nil {
			sessionID = *agent.CurrentSessionID
		} else {
			return fmt.Errorf("no session ID for agent %s", agent.ID)
		}
	}

	// Close GUI
	a.gui.Close()

	// Run claude -s command
	cmd := exec.Command("claude", "-s", sessionID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\n\033[1;36m→ Starting session: %s\033[0m\n", sessionID)
	if err := cmd.Run(); err != nil {
		fmt.Printf("\033[1;31m✗ Failed to start session: %v\033[0m\n", err)
		return gocui.ErrQuit
	}

	// Exit dashboard
	return gocui.ErrQuit
}

// createNewSession creates a completely new session with marker file and launches Claude
func (a *App) createNewSession(g *gocui.Gui, v *gocui.View) error {
	// Generate agent ID using sha1sum of current timestamp (first 8 chars)
	timestamp := time.Now().Unix()
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%d", timestamp)))
	agentID := hex.EncodeToString(hash.Sum(nil))[:8]

	// Generate session ID using UUID
	sessionID := uuid.New().String()

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	// Create marker directory in current project
	markerDir := filepath.Join(cwd, ".claude", "eye-in-the-sky")
	if err := os.MkdirAll(markerDir, 0755); err != nil {
		return fmt.Errorf("failed to create marker directory: %v", err)
	}

	// Create marker file: .claude/eye-in-the-sky/session-{agent_id}-{session_id}
	markerFile := filepath.Join(markerDir, fmt.Sprintf("session-%s-%s", agentID, sessionID))
	file, err := os.Create(markerFile)
	if err != nil {
		return fmt.Errorf("failed to create marker file: %v", err)
	}
	file.Close()

	// Close GUI
	a.gui.Close()

	// Run claude --session-id command
	cmd := exec.Command("claude", "--session-id", sessionID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\n\033[1;36m→ Starting new session\033[0m\n")
	fmt.Printf("  Agent ID:   %s\n", agentID)
	fmt.Printf("  Session ID: %s\n", sessionID)
	fmt.Printf("  Marker:     %s\n\n", markerFile)

	if err := cmd.Run(); err != nil {
		fmt.Printf("\033[1;31m✗ Failed to start new session: %v\033[0m\n", err)
		return gocui.ErrQuit
	}

	// Exit dashboard
	return gocui.ErrQuit
}

// goToWindow brings the agent's window to front using AppleScript
func (a *App) goToWindow(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}

	agent := a.agents[a.selectedIdx]

	// Check if agent has window ID
	if agent.WindowID == nil || *agent.WindowID == "" {
		return a.showMessage("No window ID for this agent")
	}

	// Use window manager to bring window to front
	wm := window.NewManager()

	// Determine application based on source and window ID format
	windowID := *agent.WindowID
	application := "Claude Desktop"

	if agent.Source == "worktree" {
		application = "Ghostty" // Default terminal for worktree agents
	} else if agent.Source == "desktop" {
		// For desktop agents, check window ID format to determine terminal
		if strings.Contains(windowID, ":/dev/tty") {
			// Format "windowID:/dev/ttyXXX" indicates iTerm2 with tab
			application = "iTerm2"
		} else if len(windowID) > 0 && windowID[0] >= '0' && windowID[0] <= '9' && !strings.Contains(windowID, ",") && !strings.Contains(windowID, ":") {
			// Pure numeric IDs indicate iTerm2 or Terminal (old format)
			application = "iTerm2"
		} else if !strings.HasPrefix(windowID, "iTerm2:") && !strings.HasPrefix(windowID, "Terminal:") {
			// Non-prefixed, non-numeric IDs are window titles (Ghostty)
			application = "Ghostty"
		}
	}

	if err := wm.BringToFront(application, windowID); err != nil {
		return a.showMessage(fmt.Sprintf("Failed to bring window to front: %v", err))
	}

	return a.showMessage("Window brought to front")
}

// showMessage displays a temporary message to the user
func (a *App) showMessage(msg string) error {
	maxX, _ := a.gui.Size()

	// Create notification view at top, just below header
	v, err := a.gui.SetView("notification", 0, 3, maxX-1, 5)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	v.Frame = false
	v.BgColor = gocui.ColorYellow
	v.FgColor = gocui.ColorBlack
	v.Clear()
	fmt.Fprintf(v, " %s", msg)

	// Set a timer to clear the message after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		a.gui.Update(func(g *gocui.Gui) error {
			g.DeleteView("notification")
			return nil
		})
	}()

	return nil
}

// showLogs shows the logs view (placeholder for now)
func (a *App) showLogs(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}

	agent := a.agents[a.selectedIdx]

	// Get actions/logs for this agent
	actions, err := a.db.ListActions(agent.ID)
	if err != nil {
		return a.showMessage(fmt.Sprintf("Failed to load logs: %v", err))
	}

	// Create or update logs view
	maxX, maxY := g.Size()
	if v, err := g.SetView(viewLogs, maxX/4, maxY/4, 3*maxX/4, 3*maxY/4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf(" Logs - %s ", agent.ID[:8])
		v.Wrap = true
		v.Autoscroll = false

		// Write logs
		for _, action := range actions {
			fmt.Fprintf(v, "[%s] %s: %s\n",
				action.Timestamp.Format("15:04:05"),
				action.ActionType,
				action.Description)
		}

		// Set as current view
		if _, err := g.SetCurrentView(viewLogs); err != nil {
			return err
		}

		// Delete any existing keybinding and add close keybinding
		g.DeleteKeybinding(viewLogs, 'q', gocui.ModNone)
		if err := g.SetKeybinding(viewLogs, 'q', gocui.ModNone, a.closeLogs); err != nil {
			return err
		}
	}

	return nil
}

// closeLogs closes the logs view
func (a *App) closeLogs(g *gocui.Gui, v *gocui.View) error {
	// Delete keybinding
	g.DeleteKeybinding(viewLogs, 'q', gocui.ModNone)

	if err := g.DeleteView(viewLogs); err != nil {
		return err
	}
	if _, err := g.SetCurrentView(viewMain); err != nil {
		return err
	}
	return nil
}

// viewAgentDetails displays detailed information about an agent

// scrollUp scrolls the detail view up with cursor navigation
func (a *App) scrollUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if cy > 0 {
			if err := v.SetCursor(cx, cy-1); err != nil {
				if oy > 0 {
					if err := v.SetOrigin(ox, oy-1); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// scrollDown scrolls the detail view down with cursor navigation
func (a *App) scrollDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			if err := v.SetOrigin(ox, oy+1); err != nil {
				// Ignore error if we're at the bottom
			}
		}
	}
	return nil
}

// pageDown scrolls down one page (half of view height)
func (a *App) pageDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, height := v.Size()
		pageSize := height / 2
		if pageSize < 1 {
			pageSize = 1
		}

		ox, oy := v.Origin()
		// Move origin down by page size
		if err := v.SetOrigin(ox, oy+pageSize); err != nil {
			// Try to move to the end
			maxY := len(v.BufferLines())
			if maxY > height {
				v.SetOrigin(ox, maxY-height)
			}
		}
	}
	return nil
}

// pageUp scrolls up one page (half of view height)
func (a *App) pageUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, height := v.Size()
		pageSize := height / 2
		if pageSize < 1 {
			pageSize = 1
		}

		ox, oy := v.Origin()
		newY := oy - pageSize
		if newY < 0 {
			newY = 0
		}
		v.SetOrigin(ox, newY)
	}
	return nil
}

// jumpToNextSection jumps to the next section in detail view (Actions -> Logs -> Compactions -> Actions...)
func (a *App) jumpToNextSection(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}

	// Get current cursor position
	_, cy := v.Cursor()
	_, oy := v.Origin()
	currentLine := cy + oy

	// Define section order
	sections := []string{"actions", "logs", "compactions", "contexts"}

	// Find which section we should jump to next
	var targetLine int
	foundNext := false

	for _, section := range sections {
		if sectionLine, exists := a.sectionLines[section]; exists {
			if sectionLine > currentLine {
				targetLine = sectionLine
				foundNext = true
				break
			}
		}
	}

	// If no next section found, cycle back to first section
	if !foundNext {
		if firstLine, exists := a.sectionLines["actions"]; exists {
			targetLine = firstLine
		} else {
			return nil // No sections available
		}
	}

	// Move cursor to target line
	ox, _ := v.Origin()
	_, viewHeight := v.Size()

	// Set origin so target line is at top of view
	newOriginY := targetLine
	if newOriginY < 0 {
		newOriginY = 0
	}

	// Set cursor to the section header line
	newCursorY := 0
	if targetLine < viewHeight {
		// If target fits in view from origin 0, adjust cursor instead
		newOriginY = 0
		newCursorY = targetLine
	}

	v.SetOrigin(ox, newOriginY)
	v.SetCursor(0, newCursorY)

	return nil
}

// truncateString truncates a string to the specified length

// ptrToString converts a string pointer to string
func ptrToString(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}

// archiveAgent archives (soft deletes) the selected agent
func (a *App) archiveAgent(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}

	agent := a.agents[a.selectedIdx]

	// Update agent status to archived
	if err := a.db.UpdateAgentStatus(agent.ID, database.StatusArchived, nil); err != nil {
		return a.showMessage(fmt.Sprintf("Failed to archive agent: %v", err))
	}

	// Refresh the agent list
	if err := a.refreshAgents(); err != nil {
		return err
	}

	// Adjust selected index if needed
	if a.selectedIdx >= len(a.agents) && a.selectedIdx > 0 {
		a.selectedIdx = len(a.agents) - 1
	}

	// Update cursor position
	if mainView, err := g.View(viewMain); err == nil {
		mainView.SetCursor(0, a.selectedIdx+2) // +2 for header
	}

	return a.showMessage(fmt.Sprintf("Agent %s archived", agent.ID[:8]))
}

// markAgentDone marks the selected agent as completed
func (a *App) markAgentDone(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}

	agent := a.agents[a.selectedIdx]

	// Update agent status to completed
	if err := a.db.UpdateAgentStatus(agent.ID, database.StatusCompleted, nil); err != nil {
		return a.showMessage(fmt.Sprintf("Failed to mark agent as done: %v", err))
	}

	// Refresh the agent list
	if err := a.refreshAgents(); err != nil {
		return err
	}

	// Adjust selected index if needed
	if a.selectedIdx >= len(a.agents) && a.selectedIdx > 0 {
		a.selectedIdx = len(a.agents) - 1
	}

	// Update cursor position
	if mainView, err := g.View(viewMain); err == nil {
		mainView.SetCursor(0, a.selectedIdx+2) // +2 for header
	}

	return a.showMessage(fmt.Sprintf("Agent %s marked as completed", agent.ID[:8]))
}

// viewCompactionFile displays the compaction JSONL file content
func (a *App) viewCompactionFile(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}

	// Get current cursor position
	_, cy := v.Cursor()
	_, oy := v.Origin()
	currentLine := cy + oy

	// Check if this line maps to a compaction
	compactionIdx, exists := a.compactionLineMap[currentLine]
	if !exists || compactionIdx >= len(a.currentCompactions) {
		return nil // Not on a compaction line
	}

	// Get the compaction
	compaction := a.currentCompactions[compactionIdx]

	// Construct the JSONL file path (absolute path)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return a.showMessage(fmt.Sprintf("Failed to get home directory: %v", err))
	}
	jsonlPath := filepath.Join(homeDir, "projects/eye-in-the-sky/data/compactions", compaction.SessionID+".jsonl")

	fmt.Fprintf(os.Stderr, "[viewCompactionFile] Looking for file at: %s\n", jsonlPath)

	// Format with jq for better readability
	cmd := exec.Command("jq", "-r",
		`select(.type == "user" or .type == "assistant") |
		"[\(.timestamp // "no-timestamp")] \(.type | ascii_upcase):\n" +
		if .message.content then
			(if (.message.content | type) == "string" then .message.content
			 elif (.message.content | type) == "array" then (.message.content | map(select(.type == "text") | .text) | join("\n"))
			 else "" end)
		else "" end + "\n---"`,
		jsonlPath)

	output, err := cmd.Output()
	if err != nil {
		// Fallback to raw content if jq fails
		content, readErr := os.ReadFile(jsonlPath)
		if readErr != nil {
			return a.showMessage(fmt.Sprintf("Failed to read compaction file: %v", readErr))
		}
		return a.showCompactionView(compaction.SessionID, string(content))
	}

	// Create compaction view with formatted content
	return a.showCompactionView(compaction.SessionID, string(output))
}

// showCompactionView displays compaction file content in a modal view
func (a *App) showCompactionView(sessionID, content string) error {
	maxX, maxY := a.gui.Size()

	// Create compaction view (modal, 90% of screen)
	margin := 5
	v, err := a.gui.SetView("compaction", margin, margin, maxX-margin, maxY-margin)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	v.Title = fmt.Sprintf(" Compaction: %s ", sessionID)
	v.Wrap = false
	v.Autoscroll = false
	v.Highlight = false
	v.Clear()

	// Write content
	fmt.Fprint(v, content)

	// Set cursor to beginning
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	// Set as current view
	if _, err := a.gui.SetCurrentView("compaction"); err != nil {
		return err
	}

	// Add keybindings to close
	for _, key := range a.keymap.Quit {
		if err := a.gui.SetKeybinding("compaction", rune(key[0]), gocui.ModNone, a.closeCompactionView); err != nil {
			return err
		}
	}

	// Add scroll keybindings
	a.gui.SetKeybinding("compaction", gocui.KeyArrowUp, gocui.ModNone, a.scrollUp)
	a.gui.SetKeybinding("compaction", gocui.KeyArrowDown, gocui.ModNone, a.scrollDown)
	for _, key := range a.keymap.Up {
		a.gui.SetKeybinding("compaction", rune(key[0]), gocui.ModNone, a.scrollUp)
	}
	for _, key := range a.keymap.Down {
		a.gui.SetKeybinding("compaction", rune(key[0]), gocui.ModNone, a.scrollDown)
	}

	// Add page scroll keybindings (space for down, b for up)
	a.gui.SetKeybinding("compaction", ' ', gocui.ModNone, a.pageDown)
	a.gui.SetKeybinding("compaction", 'b', gocui.ModNone, a.pageUp)
	a.gui.SetKeybinding("compaction", gocui.KeyPgdn, gocui.ModNone, a.pageDown)
	a.gui.SetKeybinding("compaction", gocui.KeyPgup, gocui.ModNone, a.pageUp)

	// Update status bar
	statusView, err := a.gui.View(viewStatus)
	if err != nil {
		return err
	}
	statusView.Clear()
	fmt.Fprint(statusView, " [j/k] Scroll • [Space/b] Page • [PgUp/PgDn] Page • [q] Close")

	return nil
}

// closeCompactionView closes the compaction view and returns to details
func (a *App) closeCompactionView(g *gocui.Gui, v *gocui.View) error {
	// Delete keybindings
	for _, key := range a.keymap.Quit {
		g.DeleteKeybinding("compaction", rune(key[0]), gocui.ModNone)
	}
	g.DeleteKeybinding("compaction", gocui.KeyArrowUp, gocui.ModNone)
	g.DeleteKeybinding("compaction", gocui.KeyArrowDown, gocui.ModNone)
	g.DeleteKeybinding("compaction", ' ', gocui.ModNone)
	g.DeleteKeybinding("compaction", 'b', gocui.ModNone)
	g.DeleteKeybinding("compaction", gocui.KeyPgdn, gocui.ModNone)
	g.DeleteKeybinding("compaction", gocui.KeyPgup, gocui.ModNone)
	for _, key := range a.keymap.Up {
		g.DeleteKeybinding("compaction", rune(key[0]), gocui.ModNone)
	}
	for _, key := range a.keymap.Down {
		g.DeleteKeybinding("compaction", rune(key[0]), gocui.ModNone)
	}

	// Delete the view
	if err := g.DeleteView("compaction"); err != nil {
		return err
	}

	// Return to details view
	if _, err := g.SetCurrentView("details"); err != nil {
		return err
	}

	// Restore status bar
	statusView, err := g.View(viewStatus)
	if err != nil {
		return err
	}
	statusView.Clear()
	fmt.Fprint(statusView, " [j/k] Scroll • [Enter] View item • [R] Refresh • [s] Start • [w] Window • [L] Logs • [q] Back")

	return nil
}

// viewFileOrContext determines whether to view a compaction file or context based on cursor position
func (a *App) viewFileOrContext(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		fmt.Fprintf(os.Stderr, "[viewFileOrContext] View is nil\n")
		return nil
	}

	// Get current cursor position
	_, cy := v.Cursor()
	_, oy := v.Origin()
	currentLine := cy + oy

	fmt.Fprintf(os.Stderr, "[viewFileOrContext] Current line: %d\n", currentLine)
	fmt.Fprintf(os.Stderr, "[viewFileOrContext] Compaction line map size: %d\n", len(a.compactionLineMap))
	fmt.Fprintf(os.Stderr, "[viewFileOrContext] Compaction line map: %v\n", a.compactionLineMap)
	fmt.Fprintf(os.Stderr, "[viewFileOrContext] Context line map size: %d\n", len(a.contextLineMap))
	fmt.Fprintf(os.Stderr, "[viewFileOrContext] Context line map: %v\n", a.contextLineMap)
	fmt.Fprintf(os.Stderr, "[viewFileOrContext] Current contexts count: %d\n", len(a.currentContexts))

	// Check if this line maps to a compaction
	if compactionIdx, exists := a.compactionLineMap[currentLine]; exists && compactionIdx < len(a.currentCompactions) {
		fmt.Fprintf(os.Stderr, "[viewFileOrContext] Found compaction at index %d\n", compactionIdx)
		return a.viewCompactionFile(g, v)
	}

	// Check if this line maps to a context
	if contextIdx, exists := a.contextLineMap[currentLine]; exists && contextIdx < len(a.currentContexts) {
		fmt.Fprintf(os.Stderr, "[viewFileOrContext] Found context at index %d\n", contextIdx)
		return a.viewContextFile(g, v)
	}

	fmt.Fprintf(os.Stderr, "[viewFileOrContext] No mapping found for line %d\n", currentLine)
	return nil // Not on a viewable line
}

// viewContextFile displays the saved session context details
func (a *App) viewContextFile(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}

	// Get current cursor position
	_, cy := v.Cursor()
	_, oy := v.Origin()
	currentLine := cy + oy

	// Check if this line maps to a context
	contextIdx, exists := a.contextLineMap[currentLine]
	if !exists || contextIdx >= len(a.currentContexts) {
		return nil // Not on a context line
	}

	// Get the context
	ctx := a.currentContexts[contextIdx]

	// Format the context for display
	var content strings.Builder

	// Header
	content.WriteString(fmt.Sprintf("Session ID: %s\n", ctx.SessionID))
	content.WriteString(fmt.Sprintf("Agent ID: %s\n", ctx.AgentID))
	content.WriteString(fmt.Sprintf("Created: %s\n", ctx.CreatedAt.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Updated: %s\n", ctx.UpdatedAt.Format("2006-01-02 15:04:05")))
	content.WriteString(strings.Repeat("─", 80) + "\n\n")

	// Current Phase
	if ctx.CurrentPhase != nil {
		content.WriteString(fmt.Sprintf("📍 Current Phase: %s\n\n", *ctx.CurrentPhase))
	}

	// Progress
	if ctx.OverallProgress != nil {
		content.WriteString(fmt.Sprintf("📊 Overall Progress: %.1f%%\n\n", *ctx.OverallProgress*100))
	}

	// Learned Context
	if ctx.LearnedContext != nil && *ctx.LearnedContext != "" {
		content.WriteString("🧠 Learned Context:\n")
		content.WriteString(*ctx.LearnedContext)
		content.WriteString("\n\n")
	}

	// Current Goals
	if ctx.CurrentGoals != nil && *ctx.CurrentGoals != "" {
		content.WriteString("🎯 Current Goals:\n")
		var goals []string
		if err := json.Unmarshal([]byte(*ctx.CurrentGoals), &goals); err == nil {
			for _, goal := range goals {
				content.WriteString(fmt.Sprintf("  • %s\n", goal))
			}
		} else {
			content.WriteString(*ctx.CurrentGoals)
		}
		content.WriteString("\n")
	}

	// Pending Tasks
	if ctx.PendingTasks != nil && *ctx.PendingTasks != "" {
		content.WriteString("📋 Pending Tasks:\n")
		var tasks []string
		if err := json.Unmarshal([]byte(*ctx.PendingTasks), &tasks); err == nil {
			for _, task := range tasks {
				content.WriteString(fmt.Sprintf("  ☐ %s\n", task))
			}
		} else {
			content.WriteString(*ctx.PendingTasks)
		}
		content.WriteString("\n")
	}

	// Completed Tasks
	if ctx.CompletedTasks != nil && *ctx.CompletedTasks != "" {
		content.WriteString("✅ Completed Tasks:\n")
		var tasks []string
		if err := json.Unmarshal([]byte(*ctx.CompletedTasks), &tasks); err == nil {
			for _, task := range tasks {
				content.WriteString(fmt.Sprintf("  ✓ %s\n", task))
			}
		} else {
			content.WriteString(*ctx.CompletedTasks)
		}
		content.WriteString("\n")
	}

	// Next Actions
	if ctx.NextActions != nil && *ctx.NextActions != "" {
		content.WriteString("⏭️  Next Actions:\n")
		var actions []string
		if err := json.Unmarshal([]byte(*ctx.NextActions), &actions); err == nil {
			for i, action := range actions {
				content.WriteString(fmt.Sprintf("  %d. %s\n", i+1, action))
			}
		} else {
			content.WriteString(*ctx.NextActions)
		}
		content.WriteString("\n")
	}

	// Blockers
	if ctx.Blockers != nil && *ctx.Blockers != "" {
		content.WriteString("🚧 Blockers:\n")
		content.WriteString(*ctx.Blockers)
		content.WriteString("\n\n")
	}

	// Important Files
	if ctx.ImportantFiles != nil && *ctx.ImportantFiles != "" {
		content.WriteString("📁 Important Files:\n")
		var files []string
		if err := json.Unmarshal([]byte(*ctx.ImportantFiles), &files); err == nil {
			for _, file := range files {
				content.WriteString(fmt.Sprintf("  • %s\n", file))
			}
		} else {
			content.WriteString(*ctx.ImportantFiles)
		}
		content.WriteString("\n")
	}

	// Dependencies
	if ctx.Dependencies != nil && *ctx.Dependencies != "" {
		content.WriteString("🔗 Dependencies:\n")
		var deps []string
		if err := json.Unmarshal([]byte(*ctx.Dependencies), &deps); err == nil {
			for _, dep := range deps {
				content.WriteString(fmt.Sprintf("  • %s\n", dep))
			}
		} else {
			content.WriteString(*ctx.Dependencies)
		}
		content.WriteString("\n")
	}

	// Milestones
	if ctx.Milestones != nil && *ctx.Milestones != "" {
		content.WriteString("🏆 Milestones:\n")
		content.WriteString(*ctx.Milestones)
		content.WriteString("\n\n")
	}

	// Key Decisions
	if ctx.KeyDecisions != nil && *ctx.KeyDecisions != "" {
		content.WriteString("💡 Key Decisions:\n")
		content.WriteString(*ctx.KeyDecisions)
		content.WriteString("\n\n")
	}

	// Environment
	if ctx.Environment != nil && *ctx.Environment != "" {
		content.WriteString("🌍 Environment:\n")
		content.WriteString(*ctx.Environment)
		content.WriteString("\n\n")
	}

	// Metrics
	if ctx.Metrics != nil && *ctx.Metrics != "" {
		content.WriteString("📈 Metrics:\n")
		content.WriteString(*ctx.Metrics)
		content.WriteString("\n\n")
	}

	// Auto-save status
	if ctx.AutoSave {
		content.WriteString("💾 Auto-saved checkpoint\n")
	}

	// Create context view with formatted content
	return a.showContextView(ctx.SessionID, content.String())
}

// showContextView displays context details in a modal view
func (a *App) showContextView(sessionID, content string) error {
	maxX, maxY := a.gui.Size()

	// Create context view (modal, 90% of screen)
	margin := 5
	v, err := a.gui.SetView("context", margin, margin, maxX-margin, maxY-margin)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	v.Title = fmt.Sprintf(" Session Context: %s ", sessionID)
	v.Wrap = true
	v.Autoscroll = false
	v.Highlight = false
	v.Clear()

	// Write content
	fmt.Fprint(v, content)

	// Set cursor to beginning
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	// Set as current view
	if _, err := a.gui.SetCurrentView("context"); err != nil {
		return err
	}

	// Add keybindings to close
	for _, key := range a.keymap.Quit {
		if err := a.gui.SetKeybinding("context", rune(key[0]), gocui.ModNone, a.closeContextView); err != nil {
			return err
		}
	}

	// Add scroll keybindings
	a.gui.SetKeybinding("context", gocui.KeyArrowUp, gocui.ModNone, a.scrollUp)
	a.gui.SetKeybinding("context", gocui.KeyArrowDown, gocui.ModNone, a.scrollDown)
	for _, key := range a.keymap.Up {
		a.gui.SetKeybinding("context", rune(key[0]), gocui.ModNone, a.scrollUp)
	}
	for _, key := range a.keymap.Down {
		a.gui.SetKeybinding("context", rune(key[0]), gocui.ModNone, a.scrollDown)
	}

	// Add page scroll keybindings (space for down, b for up)
	a.gui.SetKeybinding("context", ' ', gocui.ModNone, a.pageDown)
	a.gui.SetKeybinding("context", 'b', gocui.ModNone, a.pageUp)
	a.gui.SetKeybinding("context", gocui.KeyPgdn, gocui.ModNone, a.pageDown)
	a.gui.SetKeybinding("context", gocui.KeyPgup, gocui.ModNone, a.pageUp)

	// Update status bar
	statusView, err := a.gui.View(viewStatus)
	if err != nil {
		return err
	}
	statusView.Clear()
	fmt.Fprint(statusView, " [j/k] Scroll • [Space/b] Page • [PgUp/PgDn] Page • [q] Close")

	return nil
}

// closeContextView closes the context view and returns to details
func (a *App) closeContextView(g *gocui.Gui, v *gocui.View) error {
	// Delete keybindings
	for _, key := range a.keymap.Quit {
		g.DeleteKeybinding("context", rune(key[0]), gocui.ModNone)
	}
	g.DeleteKeybinding("context", gocui.KeyArrowUp, gocui.ModNone)
	g.DeleteKeybinding("context", gocui.KeyArrowDown, gocui.ModNone)
	g.DeleteKeybinding("context", ' ', gocui.ModNone)
	g.DeleteKeybinding("context", 'b', gocui.ModNone)
	g.DeleteKeybinding("context", gocui.KeyPgdn, gocui.ModNone)
	g.DeleteKeybinding("context", gocui.KeyPgup, gocui.ModNone)
	for _, key := range a.keymap.Up {
		g.DeleteKeybinding("context", rune(key[0]), gocui.ModNone)
	}
	for _, key := range a.keymap.Down {
		g.DeleteKeybinding("context", rune(key[0]), gocui.ModNone)
	}

	// Delete the view
	if err := g.DeleteView("context"); err != nil {
		return err
	}

	// Return to details view
	if _, err := g.SetCurrentView("details"); err != nil {
		return err
	}

	// Restore status bar
	statusView, err := g.View(viewStatus)
	if err != nil {
		return err
	}
	statusView.Clear()
	fmt.Fprint(statusView, " [j/k] Scroll • [Enter] View item • [R] Refresh • [s] Start • [w] Window • [L] Logs • [q] Back")

	return nil
}
