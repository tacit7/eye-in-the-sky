package modal

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestModalInitialization tests that a modal can be created and is inactive by default
func TestModalInitialization(t *testing.T) {
	m := New()

	if m == nil {
		t.Fatal("expected new modal to not be nil")
	}

	if m.IsActive() {
		t.Error("expected new modal to be inactive")
	}

	if m.Type != ModalNone {
		t.Errorf("expected Type to be ModalNone, got %v", m.Type)
	}
}

// TestOpenForm tests opening a form modal
func TestOpenForm(t *testing.T) {
	m := New()

	spec := FormSpec{
		Title:  "Test Form",
		Submit: "Submit",
		Cancel: "Cancel",
		Source: "test-form",
		Fields: []FormField{
			{
				Name:        "field1",
				Label:       "Field 1",
				Placeholder: "Enter value",
				Required:    true,
			},
		},
	}

	m.OpenForm(spec)

	if !m.IsActive() {
		t.Error("expected modal to be active after OpenForm")
	}

	if m.Type != ModalForm {
		t.Errorf("expected Type to be ModalForm, got %v", m.Type)
	}

	if m.FormSpec.Title != "Test Form" {
		t.Errorf("expected form title 'Test Form', got '%s'", m.FormSpec.Title)
	}
}

// TestOpenHelp tests opening a help modal
func TestOpenHelp(t *testing.T) {
	m := New()

	m.OpenHelp("Test Help", "Help content goes here")

	if !m.IsActive() {
		t.Error("expected modal to be active after OpenHelp")
	}

	if m.Type != ModalHelp {
		t.Errorf("expected Type to be ModalHelp, got %v", m.Type)
	}

	if m.Title != "Test Help" {
		t.Errorf("expected title 'Test Help', got '%s'", m.Title)
	}
}

// TestOpenError tests opening an error modal
func TestOpenError(t *testing.T) {
	m := New()

	m.OpenError("Test Error", "Something went wrong")

	if !m.IsActive() {
		t.Error("expected modal to be active after OpenError")
	}

	if m.Type != ModalError {
		t.Errorf("expected Type to be ModalError, got %v", m.Type)
	}

	if m.Title != "Test Error" {
		t.Errorf("expected title 'Test Error', got '%s'", m.Title)
	}
}

// TestFormFieldAccess tests getting and setting field values via GetFormData
func TestFormFieldAccess(t *testing.T) {
	m := New()

	spec := FormSpec{
		Title:  "Test Form",
		Submit: "Submit",
		Cancel: "Cancel",
		Source: "test-form",
		Fields: []FormField{
			{
				Name:        "name",
				Label:       "Name",
				Placeholder: "Enter name",
				Required:    true,
			},
			{
				Name:        "description",
				Label:       "Description",
				Placeholder: "Enter description",
				Multiline:   true,
				Required:    false,
			},
		},
	}

	m.OpenForm(spec)

	// Set field values through inputs (simulating user input)
	if len(m.Inputs) >= 2 {
		m.Inputs[0].SetValue("John Doe")
		m.Inputs[1].SetValue("Test description")
	}

	// Get all form data
	formData := m.GetFormData()
	if formData["name"] != "John Doe" {
		t.Errorf("expected name 'John Doe', got '%s'", formData["name"])
	}

	if formData["description"] != "Test description" {
		t.Errorf("expected description 'Test description', got '%s'", formData["description"])
	}
}

// TestFormNavigation tests Tab and Shift+Tab navigation between fields
func TestFormNavigation(t *testing.T) {
	m := New()

	spec := FormSpec{
		Title:  "Test Form",
		Submit: "Submit",
		Cancel: "Cancel",
		Source: "test-form",
		Fields: []FormField{
			{Name: "field1", Label: "Field 1", Required: true},
			{Name: "field2", Label: "Field 2", Required: false},
			{Name: "field3", Label: "Field 3", Required: false},
		},
	}

	m.OpenForm(spec)

	initialFocus := m.FocusIdx
	if initialFocus != 0 {
		t.Errorf("expected initial focus on first field, got %d", initialFocus)
	}

	// Move to next field
	m.NextField()
	if m.FocusIdx != 1 {
		t.Errorf("expected focus on field 2 after NextField, got %d", m.FocusIdx)
	}

	// Move to previous field
	m.PrevField()
	if m.FocusIdx != 0 {
		t.Errorf("expected focus on field 1 after PrevField, got %d", m.FocusIdx)
	}

	// Move to last field
	m.NextField()
	m.NextField()
	if m.FocusIdx != 2 {
		t.Errorf("expected focus on field 3, got %d", m.FocusIdx)
	}

	// Should wrap around
	m.NextField()
	if m.FocusIdx != 0 {
		t.Errorf("expected focus to wrap to field 1, got %d", m.FocusIdx)
	}
}

// TestFormValidation tests form field validation
func TestFormValidation(t *testing.T) {
	m := New()

	spec := FormSpec{
		Title:  "Test Form",
		Submit: "Submit",
		Cancel: "Cancel",
		Source: "test-form",
		Fields: []FormField{
			{
				Name:     "required_field",
				Label:    "Required",
				Required: true,
			},
			{
				Name:     "optional_field",
				Label:    "Optional",
				Required: false,
			},
		},
	}

	m.OpenForm(spec)

	// Missing required field should fail validation
	if len(m.Inputs) >= 1 {
		m.Inputs[1].SetValue("value")
	}
	isValid := m.ValidateForm()
	if isValid {
		t.Error("expected validation to fail for missing required field")
	}

	// With required field filled, should pass
	if len(m.Inputs) >= 1 {
		m.Inputs[0].SetValue("value")
	}
	isValid = m.ValidateForm()
	if !isValid {
		t.Error("expected validation to pass with required field filled")
	}
}

// TestCloseModal tests closing a modal
func TestCloseModal(t *testing.T) {
	m := New()

	m.OpenForm(FormSpec{
		Title:  "Test Form",
		Submit: "Submit",
		Cancel: "Cancel",
		Source: "test",
		Fields: []FormField{
			{Name: "field", Label: "Field", Required: true},
		},
	})

	if !m.IsActive() {
		t.Fatal("expected modal to be active")
	}

	m.Close()

	if m.IsActive() {
		t.Error("expected modal to be closed after Close()")
	}

	if m.Type != ModalNone {
		t.Errorf("expected Type to be ModalNone, got %v", m.Type)
	}
}

// TestModalUpdateMessaging tests that modal properly handles key messages
func TestModalUpdateMessaging(t *testing.T) {
	m := New()

	spec := FormSpec{
		Title:  "Test Form",
		Submit: "Submit",
		Cancel: "Cancel",
		Source: "test",
		Fields: []FormField{
			{Name: "field", Label: "Field", Required: true},
		},
	}

	m.OpenForm(spec)

	// Test Tab key moves focus
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	m.Update(keyMsg)
	// Tab should move focus (though it's hard to test without inspecting internal state in detail)

	// Test Esc key closes modal
	keyMsg = tea.KeyMsg{Type: tea.KeyEscape}
	m.Update(keyMsg)
	if m.IsActive() {
		t.Error("expected Esc to close modal")
	}
}

// TestModalWindowSize tests that modal handles window size messages
func TestModalWindowSize(t *testing.T) {
	m := New()

	m.OpenHelp("Help", "Content")

	// Update with window size message
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	m.Update(sizeMsg)

	if m.Width != 80 {
		t.Errorf("expected width 80, got %d", m.Width)
	}

	if m.Height != 24 {
		t.Errorf("expected height 24, got %d", m.Height)
	}
}
