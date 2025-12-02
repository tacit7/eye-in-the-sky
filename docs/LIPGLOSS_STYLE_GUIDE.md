# Lipgloss Style Guide for Eye in the Sky

This guide establishes consistent styling patterns across all UI components in the Eye in the Sky TUI application.

## 1. Centralize Your Styling

Create a package like `internal/ui/theme` or `internal/ui/styles` that contains **all the base styles** (colors, padding norms, borders) used across components.

### Example:

```go
package theme

import "github.com/charmbracelet/lipgloss"

var (
    PrimaryColor   = lipgloss.Color("#00ADD8")
    SecondaryColor = lipgloss.Color("#FFFFFF")
    BorderColor    = lipgloss.Color("#444444")

    Header = lipgloss.NewStyle().
        Foreground(PrimaryColor).
        Bold(true)

    Panel = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(BorderColor).
        Padding(1,2)
)
```

**Reason**: Avoid "magic style" redefined in each component; ensures consistency and ease of change.

## 2. Component-level Styles Build on Base Styles

In each UI component (tabs, button, modal), **inherit from the base styles** rather than re-defining brand new ones.

### Example in components/button.go:

```go
styleNormal := theme.Panel.Copy().Background(theme.SecondaryColor).Foreground(theme.PrimaryColor)
styleFocused := styleNormal.Copy().Background(theme.PrimaryColor).Foreground(theme.SecondaryColor).Bold(true)
```

## 3. Naming Conventions

Style variables should clearly reflect their role, not their look. E.g.: `BtnPrimary`, `TabActive`, `TableHeader`.

### Examples:
- **Good**: `BtnPrimary`, `BtnDisabled`, `TabActive`, `TabInactive`, `TableHeader`
- **Bad**: `BlueBox1` (describes appearance, not purpose)

Use suffixes like `Active`, `Disabled`, `Focused` for variant states.

### Avoid:
- Names like `BlueBox1` — tie to role not color

## 4. Layout Consistency

Define padding & margin rules and reuse them; e.g., `Padding(1,2)`, `Margin(1)`.

### Guidelines:
- Limit component width/height via `Width()` / `Height()` when appropriate to avoid weird wrapping
- Use `lipgloss.JoinHorizontal` and `lipgloss.JoinVertical` for layout alignment rather than manual spacing

### Example:

```go
style := lipgloss.NewStyle().MaxWidth(m.width - 4).Padding(1)
```

## 5. Colors and Themes

- Prefer **hex colors** (`"#00ADD8"`) or `lipgloss.AdaptiveColor(Light: "...", Dark: "...")` when building light/dark theme support
- **Avoid using raw terminal color codes** when you have a theme color
- Use colors from your theme palette; avoid random colors
- If terminal doesn't support true color, Lipgloss will degrade gracefully

### Example:

```go
PrimaryColor = lipgloss.AdaptiveColor{
    Light: "#00ADD8",
    Dark:  "#5DADE2",
}
```

## 6. Style Reuse and Copying

Always use `Copy()` when you want to branch from a base style to avoid shared mutations.

### Example:

```go
header := theme.Header.Copy().Underline(true)
```

**Why**: Prevents modifying the base style when you only want a variant.

## 7. Avoid Over-styling

- Don't overuse bold, blink, italics etc unless they add meaningful emphasis
- Use borders thoughtfully: too many border-boxes create clutter in terminal UIs

### Guideline:
- Borders for panels/containers: ✅
- Borders for every text element: ❌

## 8. Responsive to Terminal Size

When possible respect `Width()` or `MaxWidth()` to prevent awkward line wraps.

### Example:

```go
style := lipgloss.NewStyle().MaxWidth(m.width - 4).Padding(1)
```

**Why**: Ensures components adapt to terminal dimensions.

## 9. Archiving and Grouping

In your theme package, group style variants by component type:

```
Theme/
  Buttons.go    // BtnPrimary, BtnSecondary, BtnDisabled, etc.
  Tabs.go       // TabActive, TabInactive, TabFocused
  Panels.go     // PanelNormal, PanelFocused
  Textures.go   // TextNormal, TextBold, TextDimmed, CodeBlock
```

Components then import theme rather than define their own.

### Benefits:
- Easy to find and modify related styles
- Enforces separation of concerns
- Makes theming changes simpler

## 10. Documentation & Examples

- For every style you add, include a **comment explaining when to use it**
- Add a small example in the README showing how to render with that style

### Example:

```go
// BtnPrimary is the default button style for primary actions (submit, confirm)
var BtnPrimary = lipgloss.NewStyle().
    Background(PrimaryColor).
    Foreground(SecondaryColor).
    Padding(0, 2).
    Bold(true)

// Usage:
// button := theme.BtnPrimary.Render("Submit")
```

### Encourage devs:
> "If you think you need a *new* style, ask: can I reuse or extend an existing one?"

---

## Quick Reference

### Base Theme Structure

```go
package theme

import "github.com/charmbracelet/lipgloss"

var (
    // Colors
    PrimaryColor   = lipgloss.Color("#00ADD8")
    SecondaryColor = lipgloss.Color("#FFFFFF")
    BorderColor    = lipgloss.Color("#444444")
    ErrorColor     = lipgloss.Color("#FF5555")
    SuccessColor   = lipgloss.Color("#50FA7B")

    // Base Styles
    Header = lipgloss.NewStyle().
        Foreground(PrimaryColor).
        Bold(true)

    Panel = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(BorderColor).
        Padding(1, 2)
)
```

### Component Style Example

```go
package components

import "github.com/tacit7/eye-in-the-sky/internal/ui/theme"

// Build on theme
var (
    ButtonNormal = theme.Panel.Copy().
        Background(theme.SecondaryColor).
        Foreground(theme.PrimaryColor)

    ButtonFocused = ButtonNormal.Copy().
        Background(theme.PrimaryColor).
        Foreground(theme.SecondaryColor).
        Bold(true)
)
```

---

---

## Eye in the Sky Theme Package

### Current Structure

Our theme package is located at `internal/ui/theme/` with this structure:

```
internal/ui/theme/
├── theme.go        # Base colors, styles, and OverviewStyles compatibility
├── buttons.go      # Button variants (BtnPrimary, BtnSecondary, etc.)
├── tabs.go         # Tab styles (TabActive, TabInactive, TabFocused)
├── panels.go       # Panel/container styles (PanelNormal, PanelHighlight, etc.)
└── textures.go     # Text typography styles (TextTitle, TextLabel, TextWarning, etc.)
```

### Available Color Palette

Located in `theme.go`:

```go
var (
    PrimaryColor   = lipgloss.Color("#00ADD8") // Cyan - primary actions
    SecondaryColor = lipgloss.Color("#FFFFFF") // White - text
    BorderColor    = lipgloss.Color("#444444") // Dark gray - borders
    ErrorColor     = lipgloss.Color("#FF5555") // Red - errors
    SuccessColor   = lipgloss.Color("#50FA7B") // Green - success
    MutedColor     = lipgloss.Color("#888888") // Light gray - dimmed text
    BackgroundDark = lipgloss.Color("#1E1E1E") // Very dark gray
)
```

### How to Change Colors

**Option 1: Change globally** (affects all components)

Edit `internal/ui/theme/theme.go`:

```go
// Change primary color from cyan to purple
PrimaryColor = lipgloss.Color("#BD93F9")
```

**Option 2: Add new color**

```go
// In theme.go
InfoColor = lipgloss.Color("#8BE9FD") // Light cyan for info messages
```

**Option 3: Adaptive colors** (light/dark terminal support)

```go
PrimaryColor = lipgloss.AdaptiveColor{
    Light: "#00ADD8", // Shown in light terminals
    Dark:  "#5DADE2", // Shown in dark terminals
}
```

### How to Add New Styles

**Step 1**: Identify which file the style belongs to

- Button styles → `buttons.go`
- Tab styles → `tabs.go`
- Panel/container styles → `panels.go`
- Text/typography styles → `textures.go`
- Base/shared styles → `theme.go`

**Step 2**: Create the style using existing colors

```go
// In textures.go
// StatusPending for pending status indicators
// Usage: theme.StatusPending.Render("⏳ pending")
StatusPending = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#F1FA8C")). // Yellow
    Bold(true)
```

**Step 3**: Document with comment and usage example

Always include:
- What the style is for
- Usage example showing `theme.StyleName.Render()`

### How to Use Styles in Components

**In stateless tabs** (like logs_tab.go):

```go
package tabs

import "github.com/tacit7/eye-in-the-sky/internal/ui/theme"

func RenderLogs(ctx *DataContext, overviewStyles OverviewStyles) string {
    // Use overviewStyles (passed from parent)
    header := overviewStyles.Primary.Render("Logs")

    // Color-code by type
    switch logType {
    case "error":
        typeStyled = overviewStyles.Error.Render(logType)
    case "info":
        typeStyled = overviewStyles.Primary.Render(logType)
    }

    return header + "\n" + typeStyled
}
```

**In components with direct theme access**:

```go
package components

import "github.com/tacit7/eye-in-the-sky/internal/ui/theme"

func (m Model) View() string {
    title := theme.Header.Render("Agent Dashboard")

    // Create variant using Copy()
    activeTab := theme.TabActive.Copy().Underline(true)

    return title + "\n" + activeTab.Render("Overview")
}
```

### OverviewStyles Compatibility Layer

For tabs that receive styles from parent, we use `OverviewStyles`:

**Definition** (in `theme.go`):

```go
type OverviewStyles struct {
    SectionTitle lipgloss.Style
    Label        lipgloss.Style
    Value        lipgloss.Style
    Subtle       lipgloss.Style
    Success      lipgloss.Style
    Warning      lipgloss.Style
    Primary      lipgloss.Style
    Secondary    lipgloss.Style
    Git          lipgloss.Style
    Error        lipgloss.Style
}

func GetOverviewStyles() OverviewStyles {
    return OverviewStyles{
        SectionTitle: Header,
        Label:        TextLabel,
        Value:        TextValue,
        Subtle:       TextDimmed,
        Success:      TextSuccess,
        Warning:      TextWarning,
        Primary:      lipgloss.NewStyle().Foreground(PrimaryColor),
        Secondary:    lipgloss.NewStyle().Foreground(SecondaryColor),
        Git:          lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
        Error:        TextError,
    }
}
```

**Why?** Allows tabs to be stateless pure functions while still accessing theme styles.

**Usage in parent**:

```go
// In parent component
styles := theme.GetOverviewStyles()
tabContent := tabs.RenderLogs(ctx, styles)
```

### Common Patterns

#### Pattern 1: Color-Coded Status

```go
var statusStyle lipgloss.Style
switch status {
case "active":
    statusStyle = overviewStyles.Success
case "failed":
    statusStyle = overviewStyles.Error
case "idle":
    statusStyle = overviewStyles.Subtle
default:
    statusStyle = overviewStyles.Primary
}

rendered := statusStyle.Render("● " + status)
```

#### Pattern 2: Responsive Width

```go
// Use context width for responsive rendering
header := overviewStyles.Primary.Render("Time        Type        Message")
separator := overviewStyles.Subtle.Render(strings.Repeat("─", ctx.Width-4))
```

#### Pattern 3: Creating Variants

```go
// Start with base style
baseStyle := theme.Panel.Copy()

// Create focused variant
focusedStyle := baseStyle.Copy().
    BorderForeground(theme.PrimaryColor).
    Bold(true)

// Create error variant
errorStyle := baseStyle.Copy().
    BorderForeground(theme.ErrorColor)
```

#### Pattern 4: Conditional Styling

```go
func renderItem(text string, isSelected bool, styles OverviewStyles) string {
    if isSelected {
        return styles.Primary.Render("▶ " + text)
    }
    return styles.Subtle.Render("  " + text)
}
```

### Testing Style Changes

After modifying styles:

1. **Build**: `go build -o bin/eye-in-the-sky ./cmd/server`
2. **Run**: `./bin/eye-in-the-sky`
3. **Navigate**: Test all views (overview, agent details, all tabs)
4. **Check**:
   - Colors render correctly
   - No text overflow/wrapping issues
   - Borders align properly
   - Status indicators are clear

### Performance Tips

✅ **DO**:
- Reuse existing styles with `.Copy()`
- Cache rendered strings when content doesn't change
- Use `lipgloss.JoinHorizontal/Vertical` for layout

❌ **DON'T**:
- Create new `lipgloss.NewStyle()` in tight loops
- Modify base styles directly (always `.Copy()` first)
- Render same content repeatedly without caching

### Troubleshooting

#### Colors look wrong in terminal

**Cause**: Terminal doesn't support true color
**Solution**: Use `AdaptiveColor` or basic terminal colors

#### Borders don't align

**Cause**: Width calculations off
**Solution**: Account for border width in layout calculations

```go
// Reserve space for borders (2 chars per side + 1 for separator)
contentWidth := m.width - 6
```

#### Styles conflict/override each other

**Cause**: Not using `.Copy()` before modifying
**Solution**: Always copy base styles

```go
// WRONG - modifies base style
theme.Panel.Bold(true)

// CORRECT - creates variant
myPanel := theme.Panel.Copy().Bold(true)
```

---

## Summary

Following these guidelines ensures:
- **Consistency** across all UI components
- **Maintainability** through centralized theme management
- **Flexibility** for future theming changes
- **Clarity** through semantic naming
- **Performance** through style reuse

When in doubt, check if you can reuse or extend an existing style before creating a new one.

### Quick Checklist for New Styles

- [ ] Added to correct file (buttons.go, tabs.go, panels.go, textures.go, or theme.go)
- [ ] Uses existing colors from theme palette
- [ ] Has semantic name (not appearance-based)
- [ ] Includes comment explaining usage
- [ ] Includes usage example in comment
- [ ] Built from existing styles using `.Copy()` when possible
- [ ] Tested in actual TUI rendering
- [ ] Documented in this guide if it's a new pattern

---

## Lipgloss Table Styling

### Agent List Table Implementation

The agent list uses `github.com/charmbracelet/lipgloss/table` for native table rendering with full border support and custom styling.

**Location**: `internal/ui/components/agent_renderer.go` → `RenderAgentTable()`

### Table Configuration

```go
import "github.com/charmbracelet/lipgloss/table"

// Choose border style based on Nerd Font mode
border := lipgloss.RoundedBorder()  // ╭╮╰╯ rounded corners
if !config.UseNerdFonts {
    border = lipgloss.ASCIIBorder()  // +|-  ASCII fallback
}

// Create table with borders
t := table.New().
    Border(border).
    BorderStyle(lipgloss.NewStyle().
        Foreground(lipgloss.Color("#4A5057")).      // Border color
        Background(lipgloss.Color("#282A2C"))).     // Border background
    BorderTop(false).       // No top border
    BorderBottom(true).     // Bottom border
    BorderLeft(true).       // Left border
    BorderRight(true).      // Right border
    BorderHeader(false).    // No header separator
    BorderColumn(true).     // Column separators
    BorderRow(false).       // No row separators
    Headers(headers...).
    Rows(rows...)
```

### Color Scheme

**Background Colors**:
- App background: `#363638` (defined in `internal/ui/config/theme_config.go`)
- Table cells: `#282A2C` (darker than app background for subtle inset effect)
- Border color: `#4A5057` (lighter gray for visibility)

**Customization via Environment Variable**:
```bash
# Override app background color
export EITS_BG_COLOR="#292C33"
```

### Cell Styling with StyleFunc

```go
t.StyleFunc(func(row, col int) lipgloss.Style {
    // Header row styling
    if row == table.HeaderRow {
        return r.Styles.GetPrimary().Bold(true).
            Background(lipgloss.Color("#282A2C"))
    }

    // Selected row highlighting
    if row == selectedIndex {
        return r.Styles.GetSelected()
    }

    // Default cell background
    return lipgloss.NewStyle().
        Background(lipgloss.Color("#282A2C"))
})
```

### Icon Integration

Status icons use Nerd Font glyphs with orange color:

```go
// Status icon with orange color
func (r *AgentLineRenderer) getStatusIcon(status string) string {
    icon := getIcon(status)  // Returns "\uf069" (nf-fa-asterisk)
    orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00"))
    return orangeStyle.Render(icon)
}
```

**Icon Mappings**:
- Nerd Font mode: `\uf069` (orange asterisk) for active/working/idle
- Plain mode: `*/@/o` ASCII characters

### Border Variants

**Rounded (Nerd Font mode)**:
```
╭──────┬─────────┬──────────╮
│ Col1 │ Col2    │ Col3     │
├──────┼─────────┼──────────┤
│ data │ data    │ data     │
╰──────┴─────────┴──────────╯
```

**ASCII (Plain mode)**:
```
+------+---------+----------+
| Col1 | Col2    | Col3     |
+------+---------+----------+
| data | data    | data     |
+------+---------+----------+
```

### Usage in Tabs

**In `agents_tab.go`**:

```go
func RenderAgentsTab(agents []domain.Agent, selectedIndex int,
                     listOffset int, styles components.Styles,
                     layout LayoutInfo) string {

    renderer := components.NewAgentLineRenderer(styles)

    // Configure visible columns
    renderer.ConfigureColumns(components.ColumnConfig{
        Status:  true,
        Session: true,
        ID:      true,
        Task:    true,
        Source:  true,
    })

    // Calculate viewport slice
    visibleAgents := agents[listOffset:end]
    adjustedSelectedIndex := selectedIndex - listOffset

    // Render with lipgloss table
    return renderer.RenderAgentTable(visibleAgents, adjustedSelectedIndex)
}
```

### Description Truncation

Long descriptions are automatically truncated to 50 characters:

```go
func (r *AgentLineRenderer) formatTask(agent domain.Agent) string {
    desc := agent.FeatureDesc
    if desc == "" {
        desc = agent.CurrentTask
    }

    // Truncate to max 50 characters
    maxWidth := 50
    if len(desc) > maxWidth {
        return desc[:maxWidth-3] + "..."
    }
    return desc
}
```

### Best Practices for Tables

1. **Use RoundedBorder() for modern look** - softer appearance than sharp corners
2. **Disable BorderHeader if no separator needed** - cleaner header integration
3. **Always set BorderColumn(true)** - helps distinguish columns
4. **Match border background to cell background** - creates cohesive appearance
5. **Use StyleFunc for row-based styling** - selection, alternating colors, etc.
6. **Truncate long text fields** - prevents table layout breaking
7. **Support both Nerd Font and ASCII modes** - maximum compatibility

### Common Pitfalls

❌ **DON'T**: Set border background without matching cell background
```go
// Border stands out awkwardly
BorderStyle(lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")))
```

✅ **DO**: Match border and cell backgrounds
```go
borderBg := "#282A2C"
BorderStyle(lipgloss.NewStyle().Background(lipgloss.Color(borderBg)))
// ... and in StyleFunc:
return lipgloss.NewStyle().Background(lipgloss.Color(borderBg))
```

❌ **DON'T**: Forget to handle viewport slicing
```go
// Will render entire agent list
return renderer.RenderAgentTable(agents, selectedIndex)
```

✅ **DO**: Slice for viewport and adjust selection index
```go
visibleAgents := agents[listOffset:end]
adjustedSelectedIndex := selectedIndex - listOffset
return renderer.RenderAgentTable(visibleAgents, adjustedSelectedIndex)
```

### Testing Table Styles

```bash
# Test with Nerd Fonts
go build -o bin/eye-ui ./cmd/eye-ui && ./bin/eye-ui

# Test with ASCII mode
NERD_FONTS=0 ./bin/eye-ui

# Test with custom background
EITS_BG_COLOR="#1E1E1E" ./bin/eye-ui
```

---
