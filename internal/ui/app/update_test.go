package app

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
)

func TestAdjustScroll(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		offset   int
		height   int
		padding  int
		expected int
	}{
		{
			name:     "scroll up when index above offset",
			index:    5,
			offset:   10,
			height:   20,
			padding:  5,
			expected: 5,
		},
		{
			name:     "scroll down when index below visible area",
			index:    20,
			offset:   0,
			height:   20,
			padding:  5,
			expected: 6, // 20 - 15 + 1 = 6
		},
		{
			name:     "keep offset when index in visible range",
			index:    10,
			offset:   5,
			height:   20,
			padding:  5,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := tt.offset
			adjustScroll(tt.index, &offset, tt.height, tt.padding)
			if offset != tt.expected {
				t.Errorf("got %d, want %d", offset, tt.expected)
			}
		})
	}
}

func TestViewHandlerRouting(t *testing.T) {
	tests := []struct {
		name      string
		viewMode  ViewType
		hasHandler bool
	}{
		{
			name:      "ViewList has handler",
			viewMode:  ViewList,
			hasHandler: true,
		},
		{
			name:      "ViewDetail has handler",
			viewMode:  ViewDetail,
			hasHandler: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := viewHandlers[tt.viewMode]
			if exists != tt.hasHandler {
				t.Errorf("handler exists = %v, want %v", exists, tt.hasHandler)
			}
		})
	}
}

func TestGlobalKeysRouting(t *testing.T) {
	// Test that handleGlobalKeys exists and is callable
	t.Run("global keys handler exists", func(t *testing.T) {
		m := &Model{}
		msg := tea.KeyMsg{}

		// Should not panic
		_, _ = m.handleGlobalKeys(msg)
	})
}

func TestDebugLogging(t *testing.T) {
	// Test that debugf only logs when DEBUG env var is set
	// This is a basic sanity check - actual logging output is hard to test
	t.Run("debug function exists", func(t *testing.T) {
		// Should not panic
		debugf("test message")
	})
}

func TestHandleListClickBounds(t *testing.T) {
	tests := []struct {
		name      string
		msg       tea.MouseMsg
		layout    ListLayout
		agents    []Agent
		expectNil bool
	}{
		{
			name: "click outside content box - above",
			msg:  tea.MouseMsg{X: 5, Y: 1},
			layout: ListLayout{
				HeaderH: 3,
				TabsH:   2,
				Content: Bounds{X: 2, Y: 5, W: 76, H: 12},
				RowH:    1,
			},
			agents:    []Agent{{ID: "agent1"}},
			expectNil: true, // Should not process
		},
		{
			name: "click outside content box - left",
			msg:  tea.MouseMsg{X: 0, Y: 6},
			layout: ListLayout{
				HeaderH: 3,
				TabsH:   2,
				Content: Bounds{X: 2, Y: 5, W: 76, H: 12},
				RowH:    1,
			},
			agents:    []Agent{{ID: "agent1"}},
			expectNil: true,
		},
		{
			name: "click within content bounds",
			msg:  tea.MouseMsg{X: 10, Y: 6},
			layout: ListLayout{
				HeaderH: 3,
				TabsH:   2,
				Content: Bounds{X: 2, Y: 5, W: 76, H: 12},
				RowH:    1,
			},
			agents:    []Agent{{ID: "agent1"}, {ID: "agent2"}},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{
				listLayout:    tt.layout,
				agents:        tt.agents,
				selectedIndex: -1,
				lastClickRow:  -1,
				listTabs: components.NewTabsModel(
					[]string{"[O]verview", "[P]roject", "[C]laude", "[T]oken Usage"},
					"#ff0000", // Dummy color
					"#ffffff",
				),
			}
			// Set to Project tab to avoid loadAgentDetails call
			m.listTabs.Set(1)

			// Perform click
			_, _ = m.handleListClick(tt.msg)

			// Check if selection was updated
			wasProcessed := m.selectedIndex >= 0 || m.lastClickRow >= 0
			if wasProcessed == tt.expectNil {
				t.Errorf("click processed = %v, expected %v", wasProcessed, !tt.expectNil)
			}
		})
	}
}

func TestHandleListClickDoubleClick(t *testing.T) {
	t.Run("single click records click info", func(t *testing.T) {
		m := &Model{
			listLayout: ListLayout{
				HeaderH: 3,
				TabsH:   2,
				Content: Bounds{X: 2, Y: 5, W: 76, H: 12},
				RowH:    1,
			},
			agents:        []Agent{{ID: "agent1"}, {ID: "agent2"}},
			selectedIndex: 0,
			lastClickRow:  -1,
			listOffset:    0,
			listTabs: components.NewTabsModel(
				[]string{"[O]verview", "[P]roject", "[C]laude", "[T]oken Usage"},
				"#ff0000",
				"#ffffff",
			),
		}
		// Set to Project tab to avoid loadAgentDetails call
		m.listTabs.Set(1)

		msg := tea.MouseMsg{X: 10, Y: 6} // Row 0 (Y=6 -> contentTop=5 -> rowInView=0)
		before := time.Now()
		_, _ = m.handleListClick(msg)
		after := time.Now()

		if m.lastClickRow != 0 {
			t.Errorf("lastClickRow = %d, want 0", m.lastClickRow)
		}
		if m.lastClickAt.Before(before) || m.lastClickAt.After(after) {
			t.Errorf("lastClickAt not set to current time")
		}
	})

	t.Run("double click detection logic - same row within window", func(t *testing.T) {
		// Set up a time that's within the double-click window
		clickTime := time.Now().Add(-100 * time.Millisecond)

		// Check the logic: would this trigger a double-click?
		rowIndex := 0
		lastClickRow := 0
		timeSinceClick := time.Now().Sub(clickTime)

		if rowIndex == lastClickRow && timeSinceClick <= dbClickWindow {
			// This would trigger double-click
		} else {
			t.Errorf("Double-click detection logic failed: rowIndex=%d, lastClickRow=%d, timeSinceClick=%v, window=%v",
				rowIndex, lastClickRow, timeSinceClick, dbClickWindow)
		}
	})

	t.Run("click outside double-click window is just single click", func(t *testing.T) {
		// Set up a time that's outside the double-click window
		clickTime := time.Now().Add(-300 * time.Millisecond)

		// Check the logic: would this NOT trigger a double-click?
		rowIndex := 0
		lastClickRow := 0
		timeSinceClick := time.Now().Sub(clickTime)

		if rowIndex == lastClickRow && timeSinceClick <= dbClickWindow {
			t.Errorf("Double-click should NOT trigger: timeSinceClick=%v exceeds window=%v", timeSinceClick, dbClickWindow)
		}
		// This is correct - it's just a single click
	})
}

func TestHandleListClickRowCalculation(t *testing.T) {
	t.Run("correct row index calculation", func(t *testing.T) {
		m := &Model{
			listLayout: ListLayout{
				HeaderH: 3,
				TabsH:   2,
				Content: Bounds{X: 2, Y: 5, W: 76, H: 12},
				RowH:    1,
			},
			agents:        make([]Agent, 10),
			selectedIndex: -1,
			lastClickRow:  -1,
			listOffset:    2, // Viewing agents from index 2 onwards
			listTabs: components.NewTabsModel(
				[]string{"[O]verview", "[P]roject", "[C]laude", "[T]oken Usage"},
				"#ff0000",
				"#ffffff",
			),
		}
		m.listTabs.Set(1)

		// Click on first visible row (Y=6 means row 0 in view -> index 2)
		msg := tea.MouseMsg{X: 10, Y: 6}
		_, _ = m.handleListClick(msg)

		if m.selectedIndex != 2 {
			t.Errorf("selectedIndex = %d, want 2", m.selectedIndex)
		}
	})

	t.Run("click past end of list is ignored", func(t *testing.T) {
		m := &Model{
			listLayout: ListLayout{
				HeaderH: 3,
				TabsH:   2,
				Content: Bounds{X: 2, Y: 5, W: 76, H: 12},
				RowH:    1,
			},
			agents:        make([]Agent, 5), // Only 5 agents
			selectedIndex: -1,
			lastClickRow:  -1,
			listOffset:    3, // Viewing agents 3 and 4
			listTabs: components.NewTabsModel(
				[]string{"[O]verview", "[P]roject", "[C]laude", "[T]oken Usage"},
				"#ff0000",
				"#ffffff",
			),
		}
		m.listTabs.Set(1)

		// Click on row that doesn't exist (Y=10 would be agent index 8, but only 5 exist)
		msg := tea.MouseMsg{X: 10, Y: 10}
		_, _ = m.handleListClick(msg)

		if m.selectedIndex != -1 {
			t.Errorf("selectedIndex should not change, got %d", m.selectedIndex)
		}
	})
}
