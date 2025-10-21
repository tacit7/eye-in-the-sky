package app

// adjustScroll is a shared helper that adjusts scroll offset to keep selected item visible
func adjustScroll(index int, offset *int, height int, padding int) {
	visibleHeight := height - padding
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	// If selected item is above visible area, scroll up
	if index < *offset {
		*offset = index
	}

	// If selected item is below visible area, scroll down
	if index >= *offset+visibleHeight {
		*offset = index - visibleHeight + 1
	}
}

// adjustListScroll adjusts list scroll offset to keep selected item visible
func (m *Model) adjustListScroll() {
	// Reserve space for header (3 lines) and footer (2 lines)
	adjustScroll(m.selectedIndex, &m.listOffset, m.height, 5)
}

// adjustTasksScroll adjusts tasks scroll offset to keep selected item visible
func (m *Model) adjustTasksScroll() {
	adjustScroll(m.tasksIndex, &m.tasksOffset, m.height, 8)
}

// adjustCommitsScroll adjusts commits scroll offset to keep selected item visible
func (m *Model) adjustCommitsScroll() {
	adjustScroll(m.commitsIndex, &m.commitsOffset, m.height, 8)
}

// adjustNotesScroll adjusts notes scroll offset to keep selected item visible
func (m *Model) adjustNotesScroll() {
	adjustScroll(m.notesIndex, &m.notesOffset, m.height, 8)
}
