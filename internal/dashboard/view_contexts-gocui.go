package dashboard

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

// viewAllContexts displays all saved contexts in a modal list
func (a *App) viewAllContexts(g *gocui.Gui, v *gocui.View) error {
	if a.currentAgentID == "" {
		return nil
	}

	// Get all contexts for this agent
	contexts, err := a.db.GetSessionContextsForAgent(a.currentAgentID)
	if err != nil || len(contexts) == 0 {
		return a.showMessage("No saved contexts available")
	}

	// Build content
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Saved Contexts for Agent %s (%d total)\n\n", a.currentAgentID[:8], len(contexts)))

	for i, ctx := range contexts {
		content.WriteString(fmt.Sprintf("─────────────────────────────────────────────────────\n"))
		content.WriteString(fmt.Sprintf("#%d - %s\n", i+1, ctx.CreatedAt.Format("2006-01-02 15:04:05")))
		content.WriteString(fmt.Sprintf("─────────────────────────────────────────────────────\n"))

		// Session ID
		content.WriteString(fmt.Sprintf("Session: %s\n", ctx.SessionID))

		// Current Phase
		if ctx.CurrentPhase != nil {
			content.WriteString(fmt.Sprintf("Phase: %s\n", *ctx.CurrentPhase))
		}

		// Progress
		if ctx.OverallProgress != nil {
			content.WriteString(fmt.Sprintf("Progress: %.1f%%\n", *ctx.OverallProgress*100))
		}

		// Current Goals
		if ctx.CurrentGoals != nil && *ctx.CurrentGoals != "" {
			content.WriteString("\nCurrent Goals:\n")
			var goals []string
			if err := json.Unmarshal([]byte(*ctx.CurrentGoals), &goals); err == nil {
				for _, goal := range goals {
					content.WriteString(fmt.Sprintf("  • %s\n", goal))
				}
			}
		}

		// Pending Tasks
		if ctx.PendingTasks != nil && *ctx.PendingTasks != "" {
			content.WriteString("\nPending Tasks:\n")
			var tasks []string
			if err := json.Unmarshal([]byte(*ctx.PendingTasks), &tasks); err == nil {
				for _, task := range tasks {
					content.WriteString(fmt.Sprintf("  ☐ %s\n", task))
				}
			}
		}

		// Completed Tasks
		if ctx.CompletedTasks != nil && *ctx.CompletedTasks != "" {
			content.WriteString("\nCompleted Tasks:\n")
			var tasks []string
			if err := json.Unmarshal([]byte(*ctx.CompletedTasks), &tasks); err == nil {
				for _, task := range tasks {
					content.WriteString(fmt.Sprintf("  ✓ %s\n", task))
				}
			}
		}

		// Next Actions
		if ctx.NextActions != nil && *ctx.NextActions != "" {
			content.WriteString("\nNext Actions:\n")
			var actions []string
			if err := json.Unmarshal([]byte(*ctx.NextActions), &actions); err == nil {
				for j, action := range actions {
					content.WriteString(fmt.Sprintf("  %d. %s\n", j+1, action))
				}
			}
		}

		content.WriteString("\n")
	}

	// Show in modal
	return a.showModalView("contexts-list", "All Saved Contexts", content.String())
}

// viewAllActions displays all actions/logs in a modal list
func (a *App) viewAllActions(g *gocui.Gui, v *gocui.View) error {
	if a.currentAgentID == "" {
		return nil
	}

	// Get all actions for this agent
	actions, err := a.db.GetActionsForAgent(a.currentAgentID, 100) // Get up to 100 actions
	if err != nil || len(actions) == 0 {
		return a.showMessage("No actions logged")
	}

	// Build content
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Actions for Agent %s (%d shown)\n\n", a.currentAgentID[:8], len(actions)))

	for _, action := range actions {
		content.WriteString(fmt.Sprintf("[%s] %s - %s\n",
			action.Timestamp.Format("2006-01-02 15:04:05"),
			action.ActionType,
			action.Description))

		// Show details if available
		if action.Details != nil && *action.Details != "" {
			// Indent details
			details := strings.ReplaceAll(*action.Details, "\n", "\n  ")
			content.WriteString(fmt.Sprintf("  %s\n", details))
		}
		content.WriteString("\n")
	}

	// Show in modal
	return a.showModalView("actions-list", "All Actions & Logs", content.String())
}

// showModalView creates a generic modal view for displaying content
func (a *App) showModalView(viewName, title, content string) error {
	maxX, maxY := a.gui.Size()

	// Create modal view (90% of screen)
	margin := 5
	v, err := a.gui.SetView(viewName, margin, margin, maxX-margin, maxY-margin)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	v.Title = fmt.Sprintf(" %s ", title)
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
	if _, err := a.gui.SetCurrentView(viewName); err != nil {
		return err
	}

	// Add keybindings to close
	for _, key := range a.keymap.Quit {
		if err := a.gui.SetKeybinding(viewName, rune(key[0]), gocui.ModNone, a.closeModalView(viewName)); err != nil {
			return err
		}
	}

	// Add scroll keybindings
	a.gui.SetKeybinding(viewName, gocui.KeyArrowUp, gocui.ModNone, a.scrollUp)
	a.gui.SetKeybinding(viewName, gocui.KeyArrowDown, gocui.ModNone, a.scrollDown)
	for _, key := range a.keymap.Up {
		a.gui.SetKeybinding(viewName, rune(key[0]), gocui.ModNone, a.scrollUp)
	}
	for _, key := range a.keymap.Down {
		a.gui.SetKeybinding(viewName, rune(key[0]), gocui.ModNone, a.scrollDown)
	}

	// Add page scroll keybindings
	a.gui.SetKeybinding(viewName, ' ', gocui.ModNone, a.pageDown)
	a.gui.SetKeybinding(viewName, 'b', gocui.ModNone, a.pageUp)
	a.gui.SetKeybinding(viewName, gocui.KeyPgdn, gocui.ModNone, a.pageDown)
	a.gui.SetKeybinding(viewName, gocui.KeyPgup, gocui.ModNone, a.pageUp)

	// Update status bar
	statusView, err := a.gui.View(viewStatus)
	if err != nil {
		return err
	}
	statusView.Clear()
	fmt.Fprint(statusView, " [j/k] Scroll • [Space/b] Page • [PgUp/PgDn] Page • [q] Close")

	return nil
}

// closeModalView returns a function that closes a named modal view
func (a *App) closeModalView(viewName string) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		return a.views.CloseModalView(
			viewName,                         // Modal view name
			[]string{"navigation"},           // Modal tags (just navigation for scrolling)
			"details",                        // Return to details
			[]string{"details", "navigation"}, // Restore details tags
		)
	}
}
