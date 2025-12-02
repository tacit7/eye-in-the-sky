package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// TaskAnnotationSubmitMsg is emitted when a task annotation is submitted
type TaskAnnotationSubmitMsg struct {
	TaskID domain.TaskID
	Body   string
}

// TaskAnnotationModal represents a modal for creating task annotations
type TaskAnnotationModal struct {
	Visible  bool
	textarea textarea.Model
	width    int
	height   int
	taskID   domain.TaskID
	taskName string
}

// NewTaskAnnotationModal creates a new task annotation modal
func NewTaskAnnotationModal(width, height int) TaskAnnotationModal {
	ta := textarea.New()
	ta.Placeholder = "Enter annotation... (markdown supported)"
	ta.Focus()
	ta.CharLimit = 5000
	ta.SetWidth(width - 10)
	ta.SetHeight(height - 10)

	return TaskAnnotationModal{
		Visible:  false,
		textarea: ta,
		width:    width,
		height:   height,
	}
}

// SetTask sets the task to annotate
func (m *TaskAnnotationModal) SetTask(taskID domain.TaskID, taskName string) {
	m.taskID = taskID
	m.taskName = taskName
}

// Show displays the modal
func (m *TaskAnnotationModal) Show() {
	m.Visible = true
	m.textarea.Focus()
	m.textarea.SetValue("")
}

// Hide hides the modal
func (m *TaskAnnotationModal) Hide() {
	m.Visible = false
	m.textarea.Blur()
}

// Update handles modal updates
func (m TaskAnnotationModal) Update(msg tea.Msg) (TaskAnnotationModal, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s", "ctrl+enter":
			// Submit annotation
			return m, m.submitAnnotation()
		case "esc":
			m.Hide()
			return m, nil
		default:
			// Update textarea
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// View renders the modal
func (m TaskAnnotationModal) View() string {
	if !m.Visible {
		return ""
	}

	// Title
	title := theme.TextTitle.Render("Annotate Task")

	// Task name
	taskLabel := theme.TextLabel.Render("Task: ")
	taskValue := theme.TextValue.Render(m.taskName)
	taskLine := lipgloss.JoinHorizontal(lipgloss.Left, taskLabel, taskValue)

	// Shortcuts hint
	hints := theme.TextMuted.Render("Ctrl+S or Ctrl+Enter: Submit | Esc: Cancel")

	// Textarea
	textareaView := theme.PanelNoBorder.Render(m.textarea.View())

	// Combine content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		taskLine,
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

// submitAnnotation creates a TaskAnnotationSubmitMsg
func (m *TaskAnnotationModal) submitAnnotation() tea.Cmd {
	body := m.textarea.Value()
	if body == "" {
		return nil
	}

	m.Hide()

	return func() tea.Msg {
		return TaskAnnotationSubmitMsg{
			TaskID: m.taskID,
			Body:   body,
		}
	}
}

// Init initializes the modal
func (m TaskAnnotationModal) Init() tea.Cmd {
	return textarea.Blink
}
