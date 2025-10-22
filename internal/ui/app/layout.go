package app

import (
	"github.com/charmbracelet/lipgloss"
)

// Layout represents the calculated dimensions of the UI components
type Layout struct {
	HeaderH   int // Height of header
	TabsH     int // Height of tabs
	FooterH   int // Height of footer
	ContentH  int // Available height for content
	TotalH    int // Total terminal height
	TotalW    int // Total terminal width
}

// LayoutManager manages dynamic layout calculations
type LayoutManager struct {
	terminalWidth  int
	terminalHeight int
	cachedLayout   *Layout
	headerContent  string
	tabsContent    string
	footerContent  string
}

// NewLayoutManager creates a new layout manager
func NewLayoutManager(width, height int) *LayoutManager {
	return &LayoutManager{
		terminalWidth:  width,
		terminalHeight: height,
	}
}

// CalculateLayout calculates the layout based on actual component heights
func (lm *LayoutManager) CalculateLayout(header, tabs, footer string) Layout {
	// Cache the content for reuse
	lm.headerContent = header
	lm.tabsContent = tabs
	lm.footerContent = footer

	// Measure actual heights
	headerHeight := lm.measureHeight(header)
	tabsHeight := lm.measureHeight(tabs)
	footerHeight := lm.measureHeight(footer)

	// Calculate available content height
	// Account for spacing between components
	spacing := 2 // newlines between components
	usedHeight := headerHeight + tabsHeight + footerHeight + spacing
	contentHeight := lm.terminalHeight - usedHeight

	// Ensure minimum content height
	contentHeight = lm.ensureMinimumHeight(contentHeight)

	layout := Layout{
		HeaderH:  headerHeight,
		TabsH:    tabsHeight,
		FooterH:  footerHeight,
		ContentH: contentHeight,
		TotalH:   lm.terminalHeight,
		TotalW:   lm.terminalWidth,
	}

	// Cache the layout
	lm.cachedLayout = &layout

	return layout
}

// measureHeight measures the actual height of rendered content
func (lm *LayoutManager) measureHeight(content string) int {
	if content == "" {
		return 0
	}
	return lipgloss.Height(content)
}

// ensureMinimumHeight ensures content area has a minimum height
func (lm *LayoutManager) ensureMinimumHeight(height int) int {
	const minContentHeight = 5
	if height < minContentHeight {
		return minContentHeight
	}
	return height
}

// SafeContentHeight returns a safe content height that prevents overflow
func (lm *LayoutManager) SafeContentHeight() int {
	if lm.cachedLayout == nil {
		// Fallback if layout hasn't been calculated
		return lm.ensureMinimumHeight(lm.terminalHeight - 10)
	}
	return lm.cachedLayout.ContentH
}

// UpdateTerminalSize updates the terminal dimensions
func (lm *LayoutManager) UpdateTerminalSize(width, height int) {
	if lm.terminalWidth != width || lm.terminalHeight != height {
		lm.terminalWidth = width
		lm.terminalHeight = height
		// Invalidate cache on resize
		lm.cachedLayout = nil
	}
}

// GetCachedLayout returns the cached layout if available
func (lm *LayoutManager) GetCachedLayout() *Layout {
	return lm.cachedLayout
}

// InvalidateCache forces recalculation on next layout request
func (lm *LayoutManager) InvalidateCache() {
	lm.cachedLayout = nil
}

// GetAvailableContentArea returns the available area for content
func (lm *LayoutManager) GetAvailableContentArea() (width, height int) {
	if lm.cachedLayout != nil {
		return lm.terminalWidth - 4, lm.cachedLayout.ContentH // -4 for borders
	}
	// Fallback calculation
	return lm.terminalWidth - 4, lm.SafeContentHeight()
}

// ShouldRecalculate determines if layout should be recalculated
func (lm *LayoutManager) ShouldRecalculate(header, tabs, footer string) bool {
	// Recalculate if cache is invalid
	if lm.cachedLayout == nil {
		return true
	}

	// Recalculate if component content has changed significantly
	if header != lm.headerContent || tabs != lm.tabsContent || footer != lm.footerContent {
		return true
	}

	return false
}

// LayoutForModel calculates layout from a Model instance
func LayoutForModel(m *Model) Layout {
	if m.layoutManager == nil {
		m.layoutManager = NewLayoutManager(m.width, m.height)
	}

	// Update terminal size if changed
	m.layoutManager.UpdateTerminalSize(m.width, m.height)

	// Check if we need to recalculate
	header := m.renderHeader()
	tabs := ""
	if m.currentView == ViewList {
		tabs = m.listTabs.Render()
	} else if m.currentView == ViewDetail {
		tabs = m.tabs.Render()
	}
	footer := m.renderFooter()

	if m.layoutManager.ShouldRecalculate(header, tabs, footer) {
		return m.layoutManager.CalculateLayout(header, tabs, footer)
	}

	// Use cached layout
	if cached := m.layoutManager.GetCachedLayout(); cached != nil {
		return *cached
	}

	// Fallback calculation
	return m.layoutManager.CalculateLayout(header, tabs, footer)
}