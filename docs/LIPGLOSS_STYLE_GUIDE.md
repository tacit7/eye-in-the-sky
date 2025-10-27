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

## Summary

Following these guidelines ensures:
- **Consistency** across all UI components
- **Maintainability** through centralized theme management
- **Flexibility** for future theming changes
- **Clarity** through semantic naming
- **Performance** through style reuse

When in doubt, check if you can reuse or extend an existing style before creating a new one.
