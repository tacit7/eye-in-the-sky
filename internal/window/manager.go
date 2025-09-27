package window

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// WindowInfo represents information about a window
type WindowInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Application string `json:"application"`
	Position    string `json:"position"`
}

// Manager handles window operations
type Manager struct{}

// NewManager creates a new window manager
func NewManager() *Manager {
	return &Manager{}
}

// GetCurrentWindowID gets the current window ID using AppleScript
func (m *Manager) GetCurrentWindowID(application, windowTitle string) (*WindowInfo, error) {
	var script string

	switch strings.ToLower(application) {
	case "ghostty":
		if windowTitle != "" {
			// Get Ghostty window by title/content
			script = fmt.Sprintf(`
				tell application "System Events"
					tell application process "Ghostty"
						set frontWin to first window whose name contains "%s"
						set winPos to position of frontWin
						set winTitle to name of frontWin
						set winID to (item 1 of winPos as string) & "," & (item 2 of winPos as string)
						return winID & "|" & winTitle
					end tell
				end tell
			`, windowTitle)
		} else {
			// Get current Ghostty window
			script = `
				tell application "System Events"
					tell application process "Ghostty"
						set frontWin to first window
						set winPos to position of frontWin
						set winTitle to name of frontWin
						set winID to (item 1 of winPos as string) & "," & (item 2 of winPos as string)
						return winID & "|" & winTitle
					end tell
				end tell
			`
		}

	case "terminal":
		if windowTitle != "" {
			// Get Terminal window by content
			script = fmt.Sprintf(`
				tell application "Terminal"
					set targetWindow to first window whose contents contains "%s"
					set winID to id of targetWindow
					set winTitle to name of targetWindow
					return (winID as string) & "|" & winTitle
				end tell
			`, windowTitle)
		} else {
			// Get current Terminal window using TTY
			script = `
				set currentTTY to do shell script "tty"
				tell application "Terminal"
					set targetWindow to first window whose tty of tab 1 is currentTTY
					set winID to id of targetWindow
					set winTitle to name of targetWindow
					return (winID as string) & "|" & winTitle
				end tell
			`
		}

	case "safari", "browser":
		script = `
			tell application "Safari"
				set frontWin to front window
				set winID to id of frontWin
				set winTitle to name of frontWin
				return (winID as string) & "|" & winTitle
			end tell
		`

	case "current", "":
		// Get frontmost window of any application
		script = `
			tell application "System Events"
				set frontApp to first application process whose frontmost is true
				set appName to name of frontApp
				set frontWin to first window of frontApp
				set winPos to position of frontWin
				set winTitle to name of frontWin
				set winID to (item 1 of winPos as string) & "," & (item 2 of winPos as string)
				return appName & ":" & winID & "|" & winTitle
			end tell
		`

	default:
		// Custom application
		script = fmt.Sprintf(`
			tell application "System Events"
				tell application process "%s"
					set frontWin to first window
					set winPos to position of frontWin
					set winTitle to name of frontWin
					set winID to (item 1 of winPos as string) & "," & (item 2 of winPos as string)
					return winID & "|" & winTitle
				end tell
			end tell
		`, application)
	}

	// Execute AppleScript
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("AppleScript execution failed: %w", err)
	}

	// Parse output (format: "windowID|windowTitle")
	result := strings.TrimSpace(string(output))
	parts := strings.SplitN(result, "|", 2)

	if len(parts) < 2 {
		return nil, fmt.Errorf("unexpected AppleScript output format: %s", result)
	}

	windowID := parts[0]
	windowTitleResult := parts[1]

	// Extract app name if it's included in the ID
	appName := application
	if strings.Contains(windowID, ":") {
		appParts := strings.SplitN(windowID, ":", 2)
		appName = appParts[0]
		windowID = appParts[1]
	}

	// Extract numeric window ID if needed for position-based IDs
	position := windowID
	re := regexp.MustCompile(`\d+`)
	if matches := re.FindString(windowID); matches != "" && !strings.Contains(windowID, ",") {
		windowID = matches
	}

	return &WindowInfo{
		ID:          windowID,
		Title:       windowTitleResult,
		Application: appName,
		Position:    position,
	}, nil
}

// BringToFront brings a window to the front by application and optional title
func (m *Manager) BringToFront(application, windowTitle string) error {
	var script string

	switch strings.ToLower(application) {
	case "ghostty":
		if windowTitle != "" {
			script = fmt.Sprintf(`
				tell application "System Events"
					tell application process "Ghostty"
						set targetWin to first window whose name contains "%s"
						perform action "AXRaise" of targetWin
					end tell
				end tell
			`, windowTitle)
		} else {
			script = `
				tell application "System Events"
					tell application process "Ghostty"
						set frontmost to true
					end tell
				end tell
			`
		}

	case "terminal":
		script = `
			tell application "Terminal"
				activate
			end tell
		`

	case "safari", "browser":
		script = `
			tell application "Safari"
				activate
			end tell
		`

	case "dashboard":
		// Special case: bring browser to dashboard URL
		script = `
			tell application "Safari"
				activate
				tell front window
					set current tab to (first tab whose URL contains "localhost:8080")
				end tell
			end tell
		`

	default:
		script = fmt.Sprintf(`
			tell application "System Events"
				tell application process "%s"
					set frontmost to true
				end tell
			end tell
		`, application)
	}

	// Execute AppleScript
	cmd := exec.Command("osascript", "-e", script)
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to bring window to front: %w", err)
	}

	return nil
}

// GetAllWindows gets information about all windows of a specific application
func (m *Manager) GetAllWindows(application string) ([]*WindowInfo, error) {
	var script string

	switch strings.ToLower(application) {
	case "ghostty":
		script = `
			tell application "System Events"
				tell application process "Ghostty"
					set windowList to {}
					repeat with w in windows
						set winPos to position of w
						set winTitle to name of w
						set winID to (item 1 of winPos as string) & "," & (item 2 of winPos as string)
						set end of windowList to winID & "|" & winTitle
					end repeat
					return windowList as string
				end tell
			end tell
		`

	case "terminal":
		script = `
			tell application "Terminal"
				set windowList to {}
				repeat with w in windows
					set winID to id of w
					set winTitle to name of w
					set end of windowList to (winID as string) & "|" & winTitle
				end repeat
				return windowList as string
			end tell
		`

	default:
		script = fmt.Sprintf(`
			tell application "System Events"
				tell application process "%s"
					set windowList to {}
					repeat with w in windows
						set winPos to position of w
						set winTitle to name of w
						set winID to (item 1 of winPos as string) & "," & (item 2 of winPos as string)
						set end of windowList to winID & "|" & winTitle
					end repeat
					return windowList as string
				end tell
			end tell
		`, application)
	}

	// Execute AppleScript
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("AppleScript execution failed: %w", err)
	}

	// Parse output
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []*WindowInfo{}, nil
	}

	// Split by comma (AppleScript list separator)
	windowStrings := strings.Split(result, ", ")
	windows := make([]*WindowInfo, 0, len(windowStrings))

	for _, windowStr := range windowStrings {
		parts := strings.SplitN(windowStr, "|", 2)
		if len(parts) >= 2 {
			windows = append(windows, &WindowInfo{
				ID:          parts[0],
				Title:       parts[1],
				Application: application,
				Position:    parts[0],
			})
		}
	}

	return windows, nil
}