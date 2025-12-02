package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// NoteScope represents the scope of a note (Global, Project, Agent)
type NoteScope string

const (
	NoteScopeGlobal  NoteScope = "global"
	NoteScopeProject NoteScope = "project"
	NoteScopeAgent   NoteScope = "agent"
)

// NoteSubmitMsg is emitted when a note is submitted
type NoteSubmitMsg struct {
	Scope      NoteScope
	ParentID   string
	ParentType string
	Body       string
}

// NoteModal represents a modal for creating notes
type NoteModal struct {
	Visible   bool
	textarea  textarea.Model
	scopeDd   Dropdown
	width     int
	height    int
	agentID   string
	sessionID string
	projectID string
}

// NewNoteModal creates a new note modal
func NewNoteModal(width, height int) NoteModal {
	ta := textarea.New()
	ta.Placeholder = "Enter your note here... (markdown supported)"
	ta.Focus()
	ta.CharLimit = 5000
	ta.SetWidth(width - 10)
	ta.SetHeight(height - 15)

	scopeDd := NewDropdown([]string{"Global", "Project", "Agent"})
	scopeDd.SetSelected(2) // Default to Agent

	return NoteModal{
		Visible:  false,
		textarea: ta,
		scopeDd:  scopeDd,
		width:    width,
		height:   height,
	}
}

// SetContext sets the context for the note (agent, session, project)
func (m *NoteModal) SetContext(agentID, sessionID, projectID string) {
	m.agentID = agentID
	m.sessionID = sessionID
	m.projectID = projectID
}

// Show displays the modal
func (m *NoteModal) Show() {
	m.Visible = true
	m.textarea.Focus()
	m.textarea.SetValue("")
}

// Hide hides the modal
func (m *NoteModal) Hide() {
	m.Visible = false
	m.textarea.Blur()
}

// Update handles modal updates
func (m NoteModal) Update(msg tea.Msg) (NoteModal, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle dropdown shortcuts when dropdown is not focused
		if !m.scopeDd.Open {
			switch msg.String() {
			case "ctrl+g":
				m.scopeDd.SetSelected(0) // Global
				return m, nil
			case "ctrl+p":
				m.scopeDd.SetSelected(1) // Project
				return m, nil
			case "ctrl+a":
				m.scopeDd.SetSelected(2) // Agent
				return m, nil
			case "ctrl+s", "ctrl+enter":
				// Submit note
				return m, m.submitNote()
			case "esc":
				m.Hide()
				return m, nil
			case "tab":
				// Toggle dropdown
				m.scopeDd, cmd = m.scopeDd.Update(tea.KeyMsg{Type: tea.KeyEnter})
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		// Update dropdown if open
		if m.scopeDd.Open {
			m.scopeDd, cmd = m.scopeDd.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			// Update textarea
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the modal
func (m NoteModal) View() string {
	if !m.Visible {
		return ""
	}

	// Title
	title := theme.TextTitle.Render("Create Note")

	// Scope selector
	scopeLabel := theme.TextLabel.Render("Scope:")
	scopeDD := m.scopeDd.View()
	scopeLine := lipgloss.JoinHorizontal(lipgloss.Left, scopeLabel, " ", scopeDD)

	// Shortcuts hint
	hints := theme.TextMuted.Render("Ctrl+G:Global | Ctrl+P:Project | Ctrl+A:Agent | Tab:Toggle | Ctrl+S:Submit | Esc:Cancel")

	// Textarea
	textareaView := theme.PanelNoBorder.Render(m.textarea.View())

	// Combine content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		scopeLine,
		"",
		textareaView,
		"",
		hints,
	)

	// Wrap in modal panel
	modal := theme.PanelModal.Copy().
		Width(m.width - 4).
		Height(m.height - 4).
		Render(content)

	// Center modal
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}

// submitNote creates a NoteSubmitMsg with the appropriate parent info
func (m *NoteModal) submitNote() tea.Cmd {
	body := m.textarea.Value()
	if body == "" {
		return nil
	}

	scope := m.scopeDd.SelectedValue()
	var parentID, parentType string

	switch scope {
	case "Global":
		parentType = "global"
		parentID = "" // NULL
	case "Project":
		parentType = "projects"
		parentID = m.projectID
	case "Agent":
		// Will create two notes: one for agent, one for session
		// Return agent info first, handler will create both
		parentType = "agents"
		parentID = m.agentID
	}

	m.Hide()

	return func() tea.Msg {
		return NoteSubmitMsg{
			Scope:      NoteScope(scope),
			ParentID:   parentID,
			ParentType: parentType,
			Body:       body,
		}
	}
}

// Init initializes the modal
func (m NoteModal) Init() tea.Cmd {
	return textarea.Blink
}
