package modal

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// New creates a new Modal
func New() *Modal {
	return &Modal{
		Active:       false,
		Type:         ModalNone,
		Inputs:       []textinput.Model{},
		helpViewport: viewport.New(80, 20), // Initialize with default dimensions
	}
}

// Open displays a modal of the specified type
func (m *Modal) Open(modalType ModalType) {
	m.Active = true
	m.Type = modalType
}

// Close hides the modal
func (m *Modal) Close() {
	m.Active = false
	m.Type = ModalNone
}

// IsActive returns true if a modal is currently open
func (m *Modal) IsActive() bool {
	return m.Active
}

// OpenForm displays a form modal with the given spec
func (m *Modal) OpenForm(spec FormSpec) {
	m.Active = true
	m.Type = ModalForm
	m.FormSpec = spec
	m.Title = spec.Title

	// Initialize text inputs for each field
	m.Inputs = make([]textinput.Model, len(spec.Fields))
	for i, field := range spec.Fields {
		ti := textinput.New()
		ti.Placeholder = field.Placeholder
		ti.CharLimit = 256

		if i == 0 {
			ti.Focus()
		}

		m.Inputs[i] = ti
	}
	m.FocusIdx = 0
}

// OpenError displays an error modal
func (m *Modal) OpenError(title, content string) {
	m.Active = true
	m.Type = ModalError
	m.Title = title
	m.Content = content
}

// OpenHelp displays a help modal
func (m *Modal) OpenHelp(title, content string) {
	m.Active = true
	m.Type = ModalHelp
	m.Title = title
	m.Content = content
}

// NextField moves focus to the next field in a form
func (m *Modal) NextField() {
	if len(m.Inputs) == 0 {
		return
	}
	m.Inputs[m.FocusIdx].Blur()
	m.FocusIdx = (m.FocusIdx + 1) % len(m.Inputs)
	m.Inputs[m.FocusIdx].Focus()
}

// PrevField moves focus to the previous field in a form
func (m *Modal) PrevField() {
	if len(m.Inputs) == 0 {
		return
	}
	m.Inputs[m.FocusIdx].Blur()
	m.FocusIdx = (m.FocusIdx - 1 + len(m.Inputs)) % len(m.Inputs)
	m.Inputs[m.FocusIdx].Focus()
}

// GetFormData returns the current form field values
func (m *Modal) GetFormData() map[string]string {
	data := make(map[string]string)
	for i, input := range m.Inputs {
		if i < len(m.FormSpec.Fields) {
			data[m.FormSpec.Fields[i].Name] = input.Value()
		}
	}
	return data
}

// ValidateForm checks that all required fields are filled
func (m *Modal) ValidateForm() bool {
	for i, field := range m.FormSpec.Fields {
		if field.Required && m.Inputs[i].Value() == "" {
			return false
		}
	}
	return true
}

// Update handles messages for the modal
func (m *Modal) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return nil
	}

	// Handle text input updates if form is active
	if m.Type == ModalForm && m.FocusIdx < len(m.Inputs) {
		var cmd tea.Cmd
		m.Inputs[m.FocusIdx], cmd = m.Inputs[m.FocusIdx].Update(msg)
		return cmd
	}

	return nil
}

// handleKey processes key presses within a modal
func (m *Modal) handleKey(key tea.KeyMsg) tea.Cmd {
	switch key.String() {
	case "esc":
		m.Close()
		return nil
	case "tab":
		if m.Type == ModalForm {
			m.NextField()
		}
		return nil
	case "shift+tab":
		if m.Type == ModalForm {
			m.PrevField()
		}
		return nil
	case "enter":
		if m.Type == ModalForm {
			if m.ValidateForm() {
				data := m.GetFormData()
				m.Close()
				return func() tea.Msg {
					return FormSubmitted{
						Source: m.FormSpec.Source,
						Data:   data,
					}
				}
			}
		}
		return nil

	// Help modal scrolling
	case "j", "down":
		if m.Type == ModalHelp {
			m.HandleHelpScroll("down")
		}
		return nil
	case "k", "up":
		if m.Type == ModalHelp {
			m.HandleHelpScroll("up")
		}
		return nil
	case "pgdown":
		if m.Type == ModalHelp {
			m.HandleHelpScroll("page_down")
		}
		return nil
	case "pgup":
		if m.Type == ModalHelp {
			m.HandleHelpScroll("page_up")
		}
		return nil
	}

	return nil
}

// View renders the modal
func (m *Modal) View() string {
	if !m.Active {
		return ""
	}

	switch m.Type {
	case ModalForm:
		return m.viewForm()
	case ModalError:
		return m.viewError()
	case ModalHelp:
		return m.viewHelp()
	default:
		return ""
	}
}

// viewForm renders a form modal
func (m *Modal) viewForm() string {
	var content string
	content += m.Title + "\n\n"

	for i, field := range m.FormSpec.Fields {
		if i < len(m.Inputs) {
			content += field.Label + ":\n"
			content += m.Inputs[i].View() + "\n\n"
		}
	}

	content += "[Tab/Shift+Tab] Navigate | [Enter] " + m.FormSpec.Submit + " | [Esc] " + m.FormSpec.Cancel

	return content
}

// viewError renders an error modal
func (m *Modal) viewError() string {
	return m.Title + "\n\n" + m.Content + "\n\n[Esc] Close"
}

// viewHelp renders a help modal with scrollable viewport
func (m *Modal) viewHelp() string {
	// Render viewport content if available
	if m.helpViewport.Width > 0 {
		return m.Title + "\n\n" + m.helpViewport.View()
	}
	// Fallback if viewport not initialized
	return m.Title + "\n\n" + m.Content + "\n\n[Esc] Close"
}
