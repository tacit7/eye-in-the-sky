package modal

import (
	"github.com/charmbracelet/bubbles/textinput"
)

// NewFormModal creates a new form modal with the given specification
func NewFormModal(spec FormSpec) *Modal {
	m := New()
	m.OpenForm(spec)
	return m
}

// SetFieldValue sets the value of a form field by name
func (m *Modal) SetFieldValue(fieldName string, value string) error {
	for i, field := range m.FormSpec.Fields {
		if field.Name == fieldName && i < len(m.Inputs) {
			m.Inputs[i].SetValue(value)
			return nil
		}
	}
	return nil // Field not found but not an error
}

// GetFieldValue gets the value of a form field by name
func (m *Modal) GetFieldValue(fieldName string) string {
	for i, field := range m.FormSpec.Fields {
		if field.Name == fieldName && i < len(m.Inputs) {
			return m.Inputs[i].Value()
		}
	}
	return ""
}

// BlurAllInputs removes focus from all form inputs
func (m *Modal) BlurAllInputs() {
	for i := range m.Inputs {
		m.Inputs[i].Blur()
	}
}

// FocusField sets focus to a specific field by index
func (m *Modal) FocusField(idx int) {
	if idx < 0 || idx >= len(m.Inputs) {
		return
	}
	m.BlurAllInputs()
	m.FocusIdx = idx
	m.Inputs[idx].Focus()
}

// Common form specifications that can be reused

// NewSessionFormSpec creates a form for starting a new session
func NewSessionFormSpec() FormSpec {
	return FormSpec{
		Title: "New Session",
		Fields: []FormField{
			{
				Name:        "description",
				Label:       "Description",
				Placeholder: "What will you be working on?",
				Multiline:   false,
				Required:    true,
			},
		},
		Submit: "Create",
		Cancel: "Cancel",
		Source: "new-session",
	}
}

// NewTicketFormSpec creates a form for creating a new ticket
func NewTicketFormSpec(project string) FormSpec {
	return FormSpec{
		Title: "New Ticket",
		Fields: []FormField{
			{
				Name:        "description",
				Label:       "Description",
				Placeholder: "What needs to be done?",
				Multiline:   false,
				Required:    true,
			},
			{
				Name:        "project",
				Label:       "Project",
				Placeholder: "Project name",
				Multiline:   false,
				Required:    false,
			},
			{
				Name:        "tags",
				Label:       "Tags (comma-separated)",
				Placeholder: "+tag1, +tag2",
				Multiline:   false,
				Required:    false,
			},
		},
		Submit: "Create",
		Cancel: "Cancel",
		Source: "new-ticket",
	}
}

// CreateTextInput creates a configured text input for form fields
func CreateTextInput(field FormField) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = field.Placeholder
	ti.CharLimit = 256

	if field.Multiline {
		ti.CharLimit = 1024
	}

	return ti
}
