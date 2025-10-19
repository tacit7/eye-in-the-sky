package util

import (
	"fmt"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/window"
)

// WindowFocuser provides window focusing functionality
type WindowFocuser interface {
	Focus(windowID string) error
}

// macOSWindowFocuser implements WindowFocuser for macOS using AppleScript
type macOSWindowFocuser struct {
	manager *window.Manager
}

// NewWindowFocuser creates a new WindowFocuser for the current platform
func NewWindowFocuser() WindowFocuser {
	if runtime.GOOS == "darwin" {
		return &macOSWindowFocuser{
			manager: window.NewManager(),
		}
	}
	return &stubWindowFocuser{}
}

// Focus brings a window to the front on macOS
func (f *macOSWindowFocuser) Focus(windowID string) error {
	if windowID == "" {
		return fmt.Errorf("empty window ID")
	}

	// Use existing window manager to bring window to front
	// Assume windowID format is "application:identifier"
	// For now, we'll use Claude as the application
	return f.manager.BringToFront("Claude", windowID)
}

// stubWindowFocuser is a stub implementation for unsupported platforms
type stubWindowFocuser struct{}

// Focus returns a not implemented error on unsupported platforms
func (f *stubWindowFocuser) Focus(windowID string) error {
	return fmt.Errorf("window focusing not implemented on %s", runtime.GOOS)
}

// focusResult represents the result of a window focus operation
type FocusResult struct {
	Success bool
	Message string
	Err     error
}

// FocusWindowCmd returns a tea.Cmd that focuses a window
func FocusWindowCmd(focuser WindowFocuser, windowID string) tea.Cmd {
	return func() tea.Msg {
		if windowID == "" {
			return FocusResult{
				Success: false,
				Message: "No window ID",
				Err:     fmt.Errorf("empty window ID"),
			}
		}

		if err := focuser.Focus(windowID); err != nil {
			return FocusResult{
				Success: false,
				Message: "Failed to focus window",
				Err:     err,
			}
		}

		return FocusResult{
			Success: true,
			Message: "Window focused",
		}
	}
}
