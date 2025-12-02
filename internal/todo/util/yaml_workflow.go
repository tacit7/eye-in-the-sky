package util

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// WorkflowYAML represents the structure of a workflow YAML file.
type WorkflowYAML struct {
	Workflow []WorkflowState `yaml:"workflow"`
}

// WorkflowState represents a single state in a workflow.
type WorkflowState struct {
	Code  string `yaml:"code"`
	Label string `yaml:"label"`
}

// ParseWorkflowYAML parses workflow definition from YAML bytes.
func ParseWorkflowYAML(data []byte) (*WorkflowYAML, error) {
	var wf WorkflowYAML
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	// Validate the parsed workflow
	if err := ValidateWorkflow(&wf); err != nil {
		return nil, err
	}

	return &wf, nil
}

// ValidateWorkflow validates a WorkflowYAML structure.
func ValidateWorkflow(wf *WorkflowYAML) error {
	if len(wf.Workflow) == 0 {
		return fmt.Errorf("workflow must contain at least one state")
	}

	seenCodes := make(map[string]bool)

	for i, state := range wf.Workflow {
		// Validate code
		if err := ValidateWorkflowStateCode(state.Code); err != nil {
			return fmt.Errorf("workflow[%d]: %w", i, err)
		}

		// Validate display name (Label)
		if err := ValidateWorkflowStateDisplayName(state.Label); err != nil {
			return fmt.Errorf("workflow[%d]: %w", i, err)
		}

		// Check for duplicates
		if seenCodes[state.Code] {
			return fmt.Errorf("workflow[%d]: duplicate state code '%s'", i, state.Code)
		}
		seenCodes[state.Code] = true
	}

	return nil
}

// WorkflowStateInput represents a state for syncing to the database.
type WorkflowStateInput struct {
	Code        string
	DisplayName string
}

// ToWorkflowStateInputs converts WorkflowYAML to database input format.
func (wf *WorkflowYAML) ToWorkflowStateInputs() []WorkflowStateInput {
	inputs := make([]WorkflowStateInput, len(wf.Workflow))
	for i, state := range wf.Workflow {
		inputs[i] = WorkflowStateInput{
			Code:        state.Code,
			DisplayName: state.Label,
		}
	}
	return inputs
}

// DefaultWorkflow returns a default workflow structure.
func DefaultWorkflow() *WorkflowYAML {
	return &WorkflowYAML{
		Workflow: []WorkflowState{
			{Code: "todo", Label: "To Do"},
			{Code: "doing", Label: "In Progress"},
			{Code: "review", Label: "Under Review"},
			{Code: "done", Label: "Done"},
		},
	}
}

// ExampleWorkflowYAML returns an example YAML string for documentation.
func ExampleWorkflowYAML() string {
	return `workflow:
  - code: todo
    label: To Do
  - code: doing
    label: In Progress
  - code: review
    label: Under Review
  - code: done
    label: Done
`
}
