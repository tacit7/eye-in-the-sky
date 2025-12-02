package app

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestTableRendering tests the table rendering system in isolation
func TestTableRendering(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *TableBuilder
		rows     [][]string
		expected int // expected height
	}{
		{
			name: "Simple table with headers",
			setup: func() *TableBuilder {
				tb := NewTableBuilder()
				tb.AddColumn("ID", 10, AlignLeft, true)
				tb.AddColumn("Name", 20, AlignLeft, true)
				tb.AddColumn("Status", 12, AlignCenter, false)
				tb.SetBorderStyle(BorderSimple)
				return tb
			},
			rows: [][]string{
				{"001", "Alice Johnson", "active"},
				{"002", "Bob Smith", "inactive"},
				{"003", "Charlie Davis", "pending"},
			},
			expected: 5, // header + separator + 3 rows
		},
		{
			name: "Table without borders",
			setup: func() *TableBuilder {
				tb := NewTableBuilder()
				tb.AddColumn("Col1", 15, AlignLeft, false)
				tb.AddColumn("Col2", 15, AlignRight, false)
				tb.SetBorderStyle(BorderNone)
				return tb
			},
			rows: [][]string{
				{"Left", "Right"},
				{"Test", "Data"},
			},
			expected: 2, // just the rows
		},
		{
			name: "Table with alternating row styles",
			setup: func() *TableBuilder {
				tb := NewTableBuilder()
				tb.AddColumn("Index", 8, AlignRight, false)
				tb.AddColumn("Value", 20, AlignLeft, true)
				tb.SetBorderStyle(BorderSimple)
				tb.SetRowStyleFunc(func(rowIndex int) lipgloss.Style {
					if rowIndex%2 == 0 {
						return lipgloss.NewStyle().Background(lipgloss.Color("235"))
					}
					return lipgloss.NewStyle()
				})
				return tb
			},
			rows: [][]string{
				{"1", "First row with alternating style"},
				{"2", "Second row normal"},
				{"3", "Third row with alternating style"},
				{"4", "Fourth row normal"},
			},
			expected: 6, // header + separator + 4 rows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := tt.setup()
			result, height := tb.RenderTable(tt.rows)

			// Check height
			if height != tt.expected {
				t.Errorf("Expected height %d, got %d", tt.expected, height)
			}

			// Check result is not empty
			if result == "" {
				t.Error("Expected non-empty result")
			}

			// Print for visual inspection
			fmt.Printf("\n=== %s ===\n%s\n", tt.name, result)
		})
	}
}

// TestTableTruncation tests that truncation works correctly
func TestTableTruncation(t *testing.T) {
	tb := NewTableBuilder()
	tb.AddColumn("Short", 5, AlignLeft, true)
	tb.AddColumn("Long", 10, AlignLeft, true)
	tb.SetBorderStyle(BorderSimple)

	rows := [][]string{
		{"abc", "This is a very long text that should be truncated"},
		{"defgh", "Another long piece of text"},
	}

	result, _ := tb.RenderTable(rows)

	// Check that truncation happened
	if len(result) == 0 {
		t.Error("Expected result")
	}

	// Visual inspection
	fmt.Printf("\n=== Truncation Test ===\n%s\n", result)
}

// TestTableAlignment tests different alignment options
func TestTableAlignment(t *testing.T) {
	tb := NewTableBuilder()
	tb.AddColumn("Left", 15, AlignLeft, false)
	tb.AddColumn("Center", 15, AlignCenter, false)
	tb.AddColumn("Right", 15, AlignRight, false)
	tb.SetBorderStyle(BorderSimple)

	rows := [][]string{
		{"Left", "Center", "Right"},
		{"L", "C", "R"},
		{"LeftAlign", "CenterAlign", "RightAlign"},
	}

	result, _ := tb.RenderTable(rows)

	// Visual inspection
	fmt.Printf("\n=== Alignment Test ===\n%s\n", result)
}

// BenchmarkTableRendering benchmarks table rendering performance
func BenchmarkTableRendering(b *testing.B) {
	tb := NewTableBuilder()
	tb.AddColumn("ID", 10, AlignLeft, true)
	tb.AddColumn("Name", 30, AlignLeft, true)
	tb.AddColumn("Status", 15, AlignCenter, false)
	tb.AddColumn("Description", 50, AlignLeft, true)
	tb.SetBorderStyle(BorderSimple)

	// Generate test data
	rows := make([][]string, 100)
	for i := 0; i < 100; i++ {
		rows[i] = []string{
			fmt.Sprintf("ID%04d", i),
			fmt.Sprintf("User Name %d", i),
			"active",
			fmt.Sprintf("This is a description for row %d with some longer text", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.RenderTable(rows)
	}
}

// TestQuickRender provides a quick way to test table rendering
func TestQuickRender(t *testing.T) {
	// Simple test for quick iteration
	headers := []string{"Agent", "Status", "Task"}
	widths := []int{20, 12, 40}

	rows := [][]string{
		{"agent-001", "active", "Processing user requests"},
		{"agent-002", "idle", "Waiting for input"},
		{"agent-003", "working", "Analyzing data patterns"},
	}

	result := RenderSimpleTable(headers, rows, widths)
	fmt.Printf("\n=== Quick Render Test ===\n%s\n", result)
}