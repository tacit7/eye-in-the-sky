package shared

import "github.com/charmbracelet/lipgloss"

// Styles interface defines the styling methods needed by views and tabs
// This is implemented by app.Styles to avoid circular dependencies
type Styles interface {
	// Render methods
	RenderSuccess(string) string
	RenderWarning(string) string
	RenderSubtle(string) string
	RenderError(string) string
	RenderPrimary(string) string
	RenderSelected(string) string

	// Get* methods for components.Styles compatibility
	GetSuccess() lipgloss.Style
	GetWarning() lipgloss.Style
	GetSubtle() lipgloss.Style
	GetError() lipgloss.Style
	GetPrimary() lipgloss.Style
	GetSelected() lipgloss.Style
}

// ViewportProvider provides access to viewports from root Model
// This interface allows overview tabs to use existing viewport implementations
// without creating circular dependencies
type ViewportProvider interface {
	// GetUsageViewport returns the usage viewport for token usage tab
	GetUsageViewport() ViewportAccess

	// GetClaudeViewport returns the Claude config viewport
	GetClaudeViewport() ViewportAccess

	// GetKeybindingsViewport returns the keybindings YAML viewport
	GetKeybindingsViewport() ViewportAccess

	// GetProjectViewport returns the project viewport for project tab
	GetProjectViewport() ViewportAccess
}

// ViewportAccess provides read access to a Bubble Tea viewport
type ViewportAccess interface {
	View() string
	AtTop() bool
	AtBottom() bool
}
