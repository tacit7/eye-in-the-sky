# Eye in the Sky - UI Architecture

## Overview

This directory contains the TUI (Terminal User Interface) application for Eye in the Sky, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

The architecture has been refactored to separate concerns, improve maintainability, and prepare for future scalability.

## Component Architecture

### Core Components

#### 1. Table Rendering System (`table.go`)

The foundation for rendering tabular data throughout the application.

**Key Types:**
- `TableColumn` - Defines column properties (title, width, alignment, truncation)
- `TableBuilder` - Fluent API for building tables with various styles
- `BorderStyle` - Enum for border types (None, Simple, Rounded, Double, Heavy)
- `Alignment` - Text alignment options (Left, Center, Right)

**Features:**
- Configurable column widths and alignment
- Multiple border styles for different use cases
- Row styling callbacks for selection and alternating colors
- Returns both rendered string AND computed height for layout calculations
- Reusable across all views

**Usage Example:**
```go
tb := NewTableBuilder()
tb.AddColumn("Name", 30, AlignLeft, true)
tb.AddColumn("Status", 12, AlignCenter, false)
tb.SetBorderStyle(BorderSimple)
tb.SetRowStyleFunc(func(rowIndex int) lipgloss.Style {
    if rowIndex%2 == 0 {
        return alternateStyle
    }
    return normalStyle
})

result, height := tb.RenderTable(rows)
```

#### 2. Layout Manager (`layout.go`)

Dynamically calculates UI component dimensions to eliminate magic numbers.

**Key Type:**
- `LayoutManager` - Manages layout calculations and caching
- `Layout` - Holds calculated dimensions (header, tabs, footer, content heights)

**Features:**
- Measures actual component heights using `lipgloss.Height()`
- Caches layout calculations to avoid jitter
- Invalidates cache on terminal resize
- Ensures minimum content height for small terminals

**Usage Example:**
```go
layout := LayoutForModel(m)
// layout.ContentH contains the available content area height
// layout.HeaderH, TabsH, FooterH contain measured component heights
```

#### 3. Agent Renderer (`agent_renderer.go`)

Specialized renderer for agent data with configurable columns.

**Key Types:**
- `AgentLineRenderer` - Renders agent data into table rows
- `ColumnConfig` - Controls which columns are visible
- Map-based status styling (replaces switch statements)

**Features:**
- Data-driven column configuration
- Status-specific styling with color mapping
- Icon rendering for parent/child relationships
- Integrates with table system for consistent formatting

**Usage Example:**
```go
renderer := NewAgentLineRenderer(&m.styles)
renderer.ConfigureColumns(ColumnConfig{
    Icon: true,
    Status: true,
    ID: true,
    Task: true,
})

table := renderer.RenderAgentTable(agents, selectedIndex)
```

#### 4. Data Preparation (`data_prep.go`)

Separates data transformation from rendering logic.

**Key Functions:**
- `PrepareAgentData()` - Filters and sorts agents
- `BuildAgentHierarchy()` - Creates parent-child structure
- `RenderSubagentGroup()` - Prepares hierarchical display
- `GetVisibleAgents()` - Returns agents based on view settings

**Features:**
- `AgentProvider` interface for testability
- `FilterConfig` for filtering by status/source
- `SortConfig` for multiple sort strategies
- Pure functions enable easy testing

**Usage Example:**
```go
agents := GetVisibleAgents(m)
// Returns filtered and sorted agents ready for display
```

#### 5. Viewport (`viewport.go`)

Manages the visible window for future line-based virtualization.

**Key Type:**
- `Viewport` - Tracks offset, height, and total items

**Features:**
- Calculates visible range for rendering
- Scroll management (up, down, top, bottom)
- Percentage scrolled calculation
- Prepared for future optimizations (not fully implemented yet)

**Future Usage:**
```go
viewport := NewViewport(visibleHeight, len(agents))
start, end := viewport.VisibleRange()
visibleAgents := agents[start:end]
// Only render visible items for better performance
```

## View Organization

### File Structure

```
internal/ui/app/
├── model.go              - Main application model
├── view_root.go          - Root view dispatcher and common rendering
├── view_list.go          - Agent list view (refactored)
├── view_detail.go        - Agent detail view
├── view_project.go       - Project information tab
├── view_usage.go         - Usage metrics tab
├── view_help.go          - Help overlay
├── view_helpers.go       - Common view utilities
├── table.go              - Table rendering system
├── layout.go             - Layout manager
├── agent_renderer.go     - Agent-specific rendering
├── data_prep.go          - Data preparation utilities
├── viewport.go           - Viewport for virtualization
├── error_handling.go     - Unified error display
├── constants.go          - UI constants
└── README.md             - This file
```

### View Rendering Pipeline

```
User Input → Update() → View() → Renderer Map → View Function
                                                      ↓
                                        Data Prep → Layout → Table → Render
```

## Design Principles

### 1. Separation of Concerns

- **Data Preparation**: Pure functions that transform data
- **Layout Calculation**: Measures and allocates space
- **Rendering**: Converts data to styled strings
- **Composition**: Assembles components into final view

### 2. Reusability

- Table system used across all views
- Agent renderer reusable for different contexts
- Layout manager applicable to all views
- Viewport pattern extensible to other lists

### 3. Testability

- Small, focused functions
- Pure functions for data transformation
- Interfaces for dependency injection
- Test harness for isolated component testing

### 4. Performance

- Layout caching reduces recalculation
- Viewport prepared for line-based virtualization
- Lipgloss handles width calculations efficiently
- Future: String builder pooling if needed

### 5. Maintainability

- Constants extracted for easy configuration
- Map-based dispatch patterns
- Clear component boundaries
- Comprehensive documentation

## Common Patterns

### Adding a New Column to Agent Table

1. Update `ColumnConfig` in `agent_renderer.go`
2. Add rendering logic in `RenderAgent()`
3. Update column widths in `getColumnWidths()`
4. No changes needed to rendering code

### Creating a New View

1. Add view type to `ViewType` enum
2. Create view render function (e.g., `renderMyView()`)
3. Add to renderer map in `model.go`
4. Reuse table system for tabular data

### Changing Layout

1. Modify constants in `constants.go`
2. Layout manager auto-adjusts
3. No manual spacing calculations needed

## Performance Considerations

### Current State

- Renders all agents every frame
- Suitable for <200 agents
- Layout caching prevents jitter
- Lipgloss handles styling efficiently

### Future Optimizations (TODO)

When agent count exceeds 200:

1. **Line Virtualization**: Render only visible rows
2. **Render Caching**: Cache unchanged content
3. **Lazy Loading**: Fetch data on-demand
4. **String Builder Pooling**: Reduce allocations

The current architecture supports these optimizations without breaking changes.

## Testing

### Unit Tests

```bash
# Test table rendering
go test ./internal/ui/app -run TestTableRendering -v

# Test quick render for visual inspection
go test ./internal/ui/app -run TestQuickRender -v

# Benchmark table performance
go test ./internal/ui/app -bench=BenchmarkTableRendering
```

### Snapshot Testing (Future)

Capture rendered output at fixed terminal widths:
- 80 columns (minimum)
- 120 columns (typical)
- 160 columns (wide)

## Migration Notes

### Changes from Previous Architecture

**Before:**
- 314 lines of mixed concerns
- Magic numbers scattered throughout
- Manual string padding and calculations
- Tight coupling between data and rendering
- Duplicate code across tabs

**After:**
- ~100 lines of orchestration in view_list.go
- Dynamic layout calculations
- Lipgloss handles all spacing
- Clear separation: data → layout → render
- Unified table system

### Backward Compatibility

The old `view_list.go` is backed up as `view_list.go.backup`. The refactored version maintains the same external interface but uses the new component system internally.

## Contributing

### Adding Features

1. Use table system for new tabular data
2. Extract constants instead of hardcoding
3. Separate data transformation from rendering
4. Write tests for new components
5. Update this README

### Code Style

- Use Lipgloss helpers for layout
- Avoid manual string math
- Prefer pure functions
- Document exported types and functions
- Keep files under 300 lines

## Future Work

- [ ] Implement full line virtualization
- [ ] Add render caching layer
- [ ] Create scrollbar component
- [ ] Add column resizing
- [ ] Implement column reordering
- [ ] Add snapshot tests
- [ ] Performance profiling and optimization

## References

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Project CLAUDE.md](../../../CLAUDE.md) - Project-wide documentation