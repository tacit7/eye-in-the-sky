package app

// Viewport manages the visible window into a larger dataset
type Viewport struct {
	Offset int // Starting index of visible items
	Height int // Number of visible items
	Total  int // Total number of items in dataset
}

// NewViewport creates a new viewport
func NewViewport(height, total int) *Viewport {
	return &Viewport{
		Offset: 0,
		Height: height,
		Total:  total,
	}
}

// VisibleRange returns the start and end indices of visible items
func (v *Viewport) VisibleRange() (start, end int) {
	start = v.Offset
	end = v.Offset + v.Height

	// Ensure we don't exceed the total
	if end > v.Total {
		end = v.Total
	}

	// Ensure start is valid
	if start < 0 {
		start = 0
	}

	return start, end
}

// ScrollUp moves the viewport up by the specified lines
func (v *Viewport) ScrollUp(lines int) bool {
	oldOffset := v.Offset
	v.Offset -= lines

	if v.Offset < 0 {
		v.Offset = 0
	}

	return v.Offset != oldOffset
}

// ScrollDown moves the viewport down by the specified lines
func (v *Viewport) ScrollDown(lines int) bool {
	oldOffset := v.Offset
	v.Offset += lines

	// Calculate the maximum valid offset
	maxOffset := v.Total - v.Height
	if maxOffset < 0 {
		maxOffset = 0
	}

	if v.Offset > maxOffset {
		v.Offset = maxOffset
	}

	return v.Offset != oldOffset
}

// ScrollToTop moves the viewport to the beginning
func (v *Viewport) ScrollToTop() {
	v.Offset = 0
}

// ScrollToBottom moves the viewport to the end
func (v *Viewport) ScrollToBottom() {
	maxOffset := v.Total - v.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	v.Offset = maxOffset
}

// IsAtTop checks if the viewport is at the beginning
func (v *Viewport) IsAtTop() bool {
	return v.Offset == 0
}

// IsAtBottom checks if the viewport is at the end
func (v *Viewport) IsAtBottom() bool {
	maxOffset := v.Total - v.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	return v.Offset >= maxOffset
}

// Update updates the viewport with new dimensions
func (v *Viewport) Update(height, total int) {
	v.Height = height
	v.Total = total

	// Adjust offset if necessary
	maxOffset := v.Total - v.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if v.Offset > maxOffset {
		v.Offset = maxOffset
	}
}

// PercentScrolled returns the percentage of content scrolled (0-100)
func (v *Viewport) PercentScrolled() int {
	if v.Total <= v.Height {
		return 100 // All content is visible
	}

	maxOffset := v.Total - v.Height
	if maxOffset == 0 {
		return 100
	}

	percent := (v.Offset * 100) / maxOffset
	if percent > 100 {
		percent = 100
	}
	return percent
}

// TODO: Future virtualization improvements
//
// When we implement full line-based virtualization:
// 1. Add caching of rendered lines to avoid re-rendering
// 2. Implement lazy loading of data (only fetch visible + buffer)
// 3. Add smooth scrolling with partial line visibility
// 4. Implement jump-to-item functionality
// 5. Add scrollbar rendering support
//
// Example future usage:
//   viewport := NewViewport(visibleHeight, len(agents))
//   start, end := viewport.VisibleRange()
//   visibleAgents := agents[start:end]
//   // Only render visibleAgents
//
// This structure is designed to be backward compatible,
// so we can gradually migrate to virtualized rendering
// without breaking existing code.