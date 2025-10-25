package app

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/components"
)

// newTestModel creates a Model instance for testing with default values
func newTestModel(t *testing.T) *Model {
	return &Model{
		selectedIndex: 0,
		listTabs:      newTestTabs(),
		tabs:          newTestTabs(),
		listLayout: ListLayout{
			HeaderH: 1,
			TabsH:   2,
			FooterH: 1,
			Content: Bounds{X: 2, Y: 5, W: 76, H: 12},
			RowH:    1,
		},
		rightPaneOffset:      0,
		styles:               defaultTestStyles(),
		overviewStyles:       defaultTestOverviewStyles(),
		statusMsg:            "",
		currentView:          ViewDetail,
		tasksIndex:           0,
		tasksOffset:          0,
		projectTicketsIndex:  0,
		projectTicketsOffset: 0,
		listOffset:           0,
		detailOffset:         0,
	}
}

// newTestTabs creates a TabsModel for testing
func newTestTabs() components.TabsModel {
	return components.NewTabsModel(
		[]string{"[O]verview", "[C]ommits", "[L]ogs", "[N]otes", "[A]ctions", "[T]asks", "[P]rojects"},
		"#ff0000",
		"#ffffff",
	)
}

// defaultTestStyles creates a full Styles struct for testing
func defaultTestStyles() Styles {
	return Styles{
		SectionTitle: lipgloss.NewStyle(),
		Label:        lipgloss.NewStyle(),
		Value:        lipgloss.NewStyle(),
		Primary:      lipgloss.NewStyle(),
		Warning:      lipgloss.NewStyle(),
		Success:      lipgloss.NewStyle(),
		Error:        lipgloss.NewStyle(),
		Subtle:       lipgloss.NewStyle(),
		Git:          lipgloss.NewStyle(),
		Text:         lipgloss.NewStyle(),
		Border:       lipgloss.NewStyle(),
		Selected:     lipgloss.NewStyle(),
		Active:       lipgloss.NewStyle(),
		Working:      lipgloss.NewStyle(),
		Idle:         lipgloss.NewStyle(),
		Stale:        lipgloss.NewStyle(),
		Unknown:      lipgloss.NewStyle(),
		Completed:    lipgloss.NewStyle(),
		Failed:       lipgloss.NewStyle(),
		Secondary:    lipgloss.NewStyle(),
		Title:        lipgloss.NewStyle(),
		ContentBox:   lipgloss.NewStyle(),
		InfoBox:      lipgloss.NewStyle(),
		ErrorBox:     lipgloss.NewStyle(),
		Header:       lipgloss.NewStyle(),
		Footer:       lipgloss.NewStyle(),
		StatusBar:    lipgloss.NewStyle(),
		KeyHelp:      lipgloss.NewStyle(),
		Code:         lipgloss.NewStyle(),
		Highlight:    lipgloss.NewStyle(),
		Bold:         lipgloss.NewStyle(),
	}
}

// defaultTestOverviewStyles creates a test OverviewStyles with minimal styling
func defaultTestOverviewStyles() OverviewStyles {
	return OverviewStyles{
		SectionTitle: lipgloss.NewStyle(),
		Label:        lipgloss.NewStyle(),
		Value:        lipgloss.NewStyle(),
		Primary:      lipgloss.NewStyle(),
		Warning:      lipgloss.NewStyle(),
		Success:      lipgloss.NewStyle(),
		Error:        lipgloss.NewStyle(),
		Subtle:       lipgloss.NewStyle(),
		Git:          lipgloss.NewStyle(),
		Secondary:    lipgloss.NewStyle(),
		Danger:       lipgloss.NewStyle(),
		Info:         lipgloss.NewStyle(),
		Success2:     lipgloss.NewStyle(),
	}
}

// assertOffset is a helper to verify viewport/scroll positions
func assertOffset(t *testing.T, got, want int, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", msg, got, want)
	}
}

// assertIndex is a helper to verify list indices
func assertIndex(t *testing.T, got, want int, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", msg, got, want)
	}
}

// assertBool is a helper to verify boolean values
func assertBool(t *testing.T, got, want bool, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

// assertString is a helper to verify string values
func assertString(t *testing.T, got, want, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", msg, got, want)
	}
}
