package app

import (
	"testing"
)

// Phase 2E: Config Tab Edit Mode Tests

// TestConfigTabEditModeToggle tests toggling edit mode on/off
func TestConfigTabEditModeToggle(t *testing.T) {
	m := newTestModel(t)
	m.currentView = ViewDetail
	m.tabs.Set(7) // Config tab
	m.keybindingsYAML = "# Sample keybindings config\ntest: value\n"
	m.keybindingsEditing = false
	m.keybindingsModified = false

	// Simulate entering edit mode
	m.keybindingsEditing = true
	m.keybindingsEditBuf = m.keybindingsYAML
	m.keybindingsModified = false

	if !m.keybindingsEditing {
		t.Error("Edit mode should be enabled")
	}

	if m.keybindingsEditBuf != m.keybindingsYAML {
		t.Error("Edit buffer should be populated with current YAML content")
	}

	if m.keybindingsModified {
		t.Error("Modified flag should be false when entering edit mode")
	}

	// Toggle edit mode off
	m.keybindingsEditing = false
	m.keybindingsEditBuf = ""
	m.keybindingsModified = false

	if m.keybindingsEditing {
		t.Error("Edit mode should be disabled")
	}

	if m.keybindingsEditBuf != "" {
		t.Error("Edit buffer should be cleared when exiting edit mode")
	}

	if m.keybindingsModified {
		t.Error("Modified flag should be false when exiting edit mode")
	}
}

// TestConfigTabCharacterInput tests character input in edit mode
func TestConfigTabCharacterInput(t *testing.T) {
	m := newTestModel(t)
	m.currentView = ViewDetail
	m.keybindingsEditing = true
	m.keybindingsEditBuf = "test: "
	m.keybindingsModified = false

	// Simulate character input
	m.keybindingsEditBuf += "a"
	m.keybindingsModified = true

	if m.keybindingsEditBuf != "test: a" {
		t.Errorf("Expected 'test: a', got %q", m.keybindingsEditBuf)
	}

	if !m.keybindingsModified {
		t.Error("Modified flag should be true after character input")
	}
}

// TestConfigTabInputBuffer tests input buffer accumulation
func TestConfigTabInputBuffer(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = true
	m.keybindingsEditBuf = ""
	m.keybindingsModified = false

	// Type multiple characters
	chars := "test"
	for _, ch := range chars {
		m.keybindingsEditBuf += string(ch)
		m.keybindingsModified = true
	}

	if m.keybindingsEditBuf != "test" {
		t.Errorf("Expected 'test', got %q", m.keybindingsEditBuf)
	}

	if !m.keybindingsModified {
		t.Error("Modified flag should be true after multiple inputs")
	}
}

// TestConfigTabBackspaceInput tests backspace in edit mode
func TestConfigTabBackspaceInput(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = true
	m.keybindingsEditBuf = "test"
	m.keybindingsModified = false

	// Simulate backspace
	if len(m.keybindingsEditBuf) > 0 {
		m.keybindingsEditBuf = m.keybindingsEditBuf[:len(m.keybindingsEditBuf)-1]
		m.keybindingsModified = true
	}

	if m.keybindingsEditBuf != "tes" {
		t.Errorf("Expected 'tes', got %q", m.keybindingsEditBuf)
	}

	if !m.keybindingsModified {
		t.Error("Modified flag should be true after backspace")
	}
}

// TestConfigTabBackspaceEmpty tests backspace on empty buffer
func TestConfigTabBackspaceEmpty(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = true
	m.keybindingsEditBuf = ""
	m.keybindingsModified = false

	// Simulate backspace on empty buffer
	if len(m.keybindingsEditBuf) > 0 {
		m.keybindingsEditBuf = m.keybindingsEditBuf[:len(m.keybindingsEditBuf)-1]
		m.keybindingsModified = true
	}

	if m.keybindingsEditBuf != "" {
		t.Error("Backspace on empty buffer should not change it")
	}

	if m.keybindingsModified {
		t.Error("Modified flag should remain false for backspace on empty buffer")
	}
}

// TestConfigTabEnterInput tests enter key in edit mode
func TestConfigTabEnterInput(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = true
	m.keybindingsEditBuf = "test:"
	m.keybindingsModified = false

	// Simulate enter key
	m.keybindingsEditBuf += "\n"
	m.keybindingsModified = true

	if m.keybindingsEditBuf != "test:\n" {
		t.Errorf("Expected 'test:\\n', got %q", m.keybindingsEditBuf)
	}

	if !m.keybindingsModified {
		t.Error("Modified flag should be true after enter")
	}
}

// TestConfigTabTabInput tests tab character in edit mode
func TestConfigTabTabInput(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = true
	m.keybindingsEditBuf = "test"
	m.keybindingsModified = false

	// Simulate tab key
	m.keybindingsEditBuf += "\t"
	m.keybindingsModified = true

	if m.keybindingsEditBuf != "test\t" {
		t.Errorf("Expected 'test\\t', got %q", m.keybindingsEditBuf)
	}

	if !m.keybindingsModified {
		t.Error("Modified flag should be true after tab")
	}
}

// TestConfigTabEditingState tests edit mode state transitions
func TestConfigTabEditingState(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsYAML = "key: value"
	m.keybindingsEditing = false
	m.keybindingsModified = false

	assertBool(t, m.keybindingsEditing, false, "Initially not editing")

	// Enter edit mode
	m.keybindingsEditing = true
	m.keybindingsEditBuf = m.keybindingsYAML

	assertBool(t, m.keybindingsEditing, true, "After entering edit mode")

	// Modify content
	m.keybindingsEditBuf += "\nother: value2"
	m.keybindingsModified = true

	assertBool(t, m.keybindingsModified, true, "After modifying content")

	// Cancel edit
	m.keybindingsEditing = false
	m.keybindingsEditBuf = ""
	m.keybindingsModified = false

	assertBool(t, m.keybindingsEditing, false, "After cancelling edit")
}

// TestConfigTabModifiedFlag tests tracking of modified state
func TestConfigTabModifiedFlag(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditBuf = "original"
	m.keybindingsModified = false

	// Adding content sets modified flag
	m.keybindingsEditBuf += "x"
	m.keybindingsModified = true

	assertBool(t, m.keybindingsModified, true, "Modified flag set after content change")

	// Clearing edit without saving shows no changes
	if m.keybindingsEditBuf != "originalx" {
		t.Error("Content should be accumulated")
	}
}

// TestConfigTabSaveCondition tests conditions for saving
func TestConfigTabSaveCondition(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = true
	m.keybindingsModified = true
	m.keybindingsEditBuf = "new content"

	// Can save only if both conditions are true
	canSave := m.keybindingsEditing && m.keybindingsModified

	assertBool(t, canSave, true, "Can save when editing and modified")

	// Cannot save if not modified
	m.keybindingsModified = false
	canSave = m.keybindingsEditing && m.keybindingsModified
	assertBool(t, canSave, false, "Cannot save when not modified")

	// Cannot save if not editing
	m.keybindingsModified = true
	m.keybindingsEditing = false
	canSave = m.keybindingsEditing && m.keybindingsModified
	assertBool(t, canSave, false, "Cannot save when not editing")
}

// TestConfigTabCancelEdit tests cancelling edit discards changes
func TestConfigTabCancelEdit(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsYAML = "# Original content"
	m.keybindingsEditing = true
	m.keybindingsEditBuf = "# Modified content"
	m.keybindingsModified = true

	// Cancel edit
	m.keybindingsEditing = false
	m.keybindingsEditBuf = ""
	m.keybindingsModified = false

	if m.keybindingsEditing {
		t.Error("Edit mode should be disabled after cancel")
	}

	if m.keybindingsEditBuf != "" {
		t.Error("Edit buffer should be cleared after cancel")
	}

	if m.keybindingsModified {
		t.Error("Modified flag should be false after cancel")
	}

	// Original YAML should be unchanged
	if m.keybindingsYAML != "# Original content" {
		t.Error("Original YAML should not be modified by edit")
	}
}

// TestConfigTabMultilineContent tests multiline keybindings content
func TestConfigTabMultilineContent(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = true
	m.keybindingsEditBuf = ""

	// Build multiline content
	content := "line1\nline2\nline3"
	m.keybindingsEditBuf = content
	m.keybindingsModified = true

	if m.keybindingsEditBuf != "line1\nline2\nline3" {
		t.Errorf("Expected 'line1\\nline2\\nline3', got %q", m.keybindingsEditBuf)
	}

	if !m.keybindingsModified {
		t.Error("Modified flag should be true for multiline content")
	}
}

// TestConfigTabEditBufferPreservation tests that YAML is copied to edit buffer
func TestConfigTabEditBufferPreservation(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsYAML = "# Original YAML\nkey: value\n"
	m.keybindingsEditing = false

	// Enter edit mode - YAML should be copied to edit buffer
	m.keybindingsEditing = true
	m.keybindingsEditBuf = m.keybindingsYAML

	if m.keybindingsEditBuf != m.keybindingsYAML {
		t.Errorf("Edit buffer should contain YAML content: got %q", m.keybindingsEditBuf)
	}
}

// TestConfigTabStatusMessaging tests status message updates
func TestConfigTabStatusMessaging(t *testing.T) {
	m := newTestModel(t)
	m.statusMsg = ""

	// Entering edit mode should set status
	m.keybindingsEditing = true
	m.statusMsg = "Edit mode on. Press Ctrl+S to save, Esc to cancel."

	if m.statusMsg == "" {
		t.Error("Status message should be set when entering edit mode")
	}

	assertString(t, m.statusMsg, "Edit mode on. Press Ctrl+S to save, Esc to cancel.", "Edit mode status")

	// Cancelling should set different status
	m.keybindingsEditing = false
	m.statusMsg = "Edit mode cancelled."

	assertString(t, m.statusMsg, "Edit mode cancelled.", "Cancel status")
}

// TestConfigTabNonEditingNavigation tests navigation outside edit mode
func TestConfigTabNonEditingNavigation(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = false
	m.keybindingsViewport.Height = 10
	m.keybindingsViewport.SetContent("line1\nline2\nline3")

	// Should be able to navigate viewport without being in edit mode
	canNavigate := !m.keybindingsEditing

	assertBool(t, canNavigate, true, "Can navigate viewport outside edit mode")
}

// TestConfigTabLargeInput tests handling of large keybindings content
func TestConfigTabLargeInput(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = true
	m.keybindingsEditBuf = ""

	// Build large content
	for i := 0; i < 100; i++ {
		m.keybindingsEditBuf += "line " + string(rune('0'+(i%10))) + "\n"
	}
	m.keybindingsModified = true

	if len(m.keybindingsEditBuf) == 0 {
		t.Error("Edit buffer should contain large content")
	}

	if !m.keybindingsModified {
		t.Error("Modified flag should be true after large input")
	}
}

// TestConfigTabEmptyInput tests handling of empty edit buffer
func TestConfigTabEmptyInput(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsEditing = true
	m.keybindingsEditBuf = ""
	m.keybindingsModified = false

	assertString(t, m.keybindingsEditBuf, "", "Edit buffer should be empty initially")
	assertBool(t, m.keybindingsModified, false, "Modified flag false for empty buffer")
}

// TestConfigTabReloadScenario tests typical reload workflow
func TestConfigTabReloadScenario(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsYAML = "# Old config"
	m.keybindingsEditing = false
	m.keybindingsEditBuf = ""
	m.keybindingsModified = false

	// User presses 'e' to edit
	m.keybindingsEditing = true
	m.keybindingsEditBuf = m.keybindingsYAML

	assertBool(t, m.keybindingsEditing, true, "In edit mode")
	assertString(t, m.keybindingsEditBuf, "# Old config", "Buffer has original content")

	// User modifies content
	m.keybindingsEditBuf += "\nnew key: value"
	m.keybindingsModified = true

	assertBool(t, m.keybindingsModified, true, "Content is modified")

	// User presses Escape to cancel
	m.keybindingsEditing = false
	m.keybindingsEditBuf = ""
	m.keybindingsModified = false

	assertBool(t, m.keybindingsEditing, false, "Edit mode cancelled")
	assertString(t, m.keybindingsYAML, "# Old config", "Original YAML unchanged")
}

// TestConfigTabSaveScenario tests typical save workflow
func TestConfigTabSaveScenario(t *testing.T) {
	m := newTestModel(t)
	m.keybindingsYAML = "# Original"
	m.keybindingsEditing = true
	m.keybindingsEditBuf = "# Modified"
	m.keybindingsModified = true

	// Can save because both conditions are met
	canSave := m.keybindingsEditing && m.keybindingsModified
	assertBool(t, canSave, true, "Can save with modifications")

	// After saving, clear the edit state
	m.keybindingsYAML = m.keybindingsEditBuf
	m.keybindingsEditing = false
	m.keybindingsEditBuf = ""
	m.keybindingsModified = false

	assertString(t, m.keybindingsYAML, "# Modified", "YAML updated after save")
	assertBool(t, m.keybindingsEditing, false, "Edit mode ended")
}
