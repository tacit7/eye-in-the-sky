package modal

import (
	"github.com/charmbracelet/bubbles/textinput"
)

// ModalType represents the type of modal currently displayed
type ModalType int

const (
	ModalNone ModalType = iota
	ModalForm
	ModalHelp
	ModalError
)

// FormField represents a single field in a form
type FormField struct {
	Name        string
	Label       string
	Placeholder string
	Multiline   bool
	Required    bool
}

// FormSpec defines the structure of a form modal
type FormSpec struct {
	Title  string
	Fields []FormField
	Submit string
	Cancel string
	Source string // "new-session", "new-ticket", etc.
}

// Modal represents the modal subsystem state
type Modal struct {
	Active bool
	Type   ModalType

	// Generic modal data
	Title   string
	Content string

	// Form-specific data
	FormSpec FormSpec
	Inputs   []textinput.Model
	FocusIdx int

	// Dimensions for rendering
	Width  int
	Height int
}

// Internal messages for form submission
type formSubmittedMsg struct {
	Source string
	Data   map[string]string
}

// FormSubmitted is emitted when a form is successfully submitted
type FormSubmitted struct {
	Source string
	Data   map[string]string
}
