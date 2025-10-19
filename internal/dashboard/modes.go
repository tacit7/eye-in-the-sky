package dashboard

// Mode represents the current UI mode/state
type Mode int

const (
	ModeMain Mode = iota
	ModeDetails
	ModeLogs
	ModeEdit
	ModeCmdline
	ModeTasks
)

// String returns the mode name
func (m Mode) String() string {
	switch m {
	case ModeMain:
		return "main"
	case ModeDetails:
		return "details"
	case ModeLogs:
		return "logs"
	case ModeEdit:
		return "edit"
	case ModeCmdline:
		return "cmdline"
	case ModeTasks:
		return "tasks"
	default:
		return "unknown"
	}
}

// ModeManager manages the mode stack for view lifecycle tracking
type ModeManager struct {
	stack []Mode
}

// NewModeManager creates a new mode manager starting in main mode
func NewModeManager() *ModeManager {
	return &ModeManager{
		stack: []Mode{ModeMain},
	}
}

// Push adds a new mode to the stack
func (m *ModeManager) Push(mode Mode) {
	m.stack = append(m.stack, mode)
}

// Pop removes and returns the current mode, restores previous
// Returns the mode that is now current after popping
func (m *ModeManager) Pop() Mode {
	if len(m.stack) <= 1 {
		// Never pop the last mode (main)
		return m.stack[0]
	}

	// Remove current mode
	m.stack = m.stack[:len(m.stack)-1]

	// Return new current mode
	return m.Current()
}

// Current returns the current (top of stack) mode
func (m *ModeManager) Current() Mode {
	if len(m.stack) == 0 {
		return ModeMain
	}
	return m.stack[len(m.stack)-1]
}

// Previous returns the previous mode (one below current)
// Returns current mode if stack has only one element
func (m *ModeManager) Previous() Mode {
	if len(m.stack) <= 1 {
		return m.Current()
	}
	return m.stack[len(m.stack)-2]
}

// Reset clears the stack back to main mode
func (m *ModeManager) Reset() {
	m.stack = []Mode{ModeMain}
}

// Depth returns the current stack depth
func (m *ModeManager) Depth() int {
	return len(m.stack)
}
