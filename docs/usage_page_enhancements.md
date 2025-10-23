
# Eye in the Sky – Usage Page Enhancements

This document outlines how to enhance the **Usage Page** in the Eye in the Sky TUI. The goal is to make the page look like a modern, dynamic dashboard instead of a static dump of tables.

---

## 1. Add Context

### Quick Summary Bar

Add a top-level summary bar for key metrics:

```
──────────────────────────────────────────────────────────────
💰 Total Cost: $12.349  |  🧠 Tokens: 1,234,987  |  🕓 Last Sync: 2m ago
──────────────────────────────────────────────────────────────
```

**Example (Go):**
```go
summary := lipgloss.JoinHorizontal(lipgloss.Top,
    m.styles.Primary.Render(fmt.Sprintf("💰 Total: $%.4f", totalCost)),
    spacer,
    m.styles.Secondary.Render(fmt.Sprintf("🧠 Tokens: %s", formatNumber(totalTokens))),
    spacer,
    m.styles.Subtle.Render(fmt.Sprintf("🕓 Updated: %s", humanize.Time(m.lastSync))),
)
```

---

## 2. Add Insight

### Trend Indicators

Compare current vs previous month to visualize deltas:

```
↑ +12.4% cost vs last month
↓ -5.7% token usage
```

```go
delta := (current.TotalCost - prev.TotalCost) / prev.TotalCost * 100
arrow := "↑"
color := m.styles.Success
if delta < 0 {
  arrow = "↓"
  color = m.styles.Subtle
}
trend := color.Render(fmt.Sprintf("%s %.1f%% vs last month", arrow, delta))
```

### Cost Distribution

Show a proportional visualization of model cost contribution:

```
Claude-3 Opus   ████████████████████  64%
Claude-3 Haiku  ██████                20%
GPT-4 Turbo     ███                   10%
Other           ██                    6%
```

Use **bubbles/progress** for gradient bars.

---

## 3. Add Visual Structure

### Borders and Grouping

Wrap each section in a rounded Lipgloss box:

```go
sectionBox := lipgloss.NewStyle().
  Border(lipgloss.RoundedBorder()).
  BorderForeground(lipgloss.Color("240")).
  Padding(1, 2).
  MarginTop(1)
```

### Compact Table Headers

Convert repeated section titles into **tabs**:

```
┌────────────────────────────────────────────┐
│ Claude Code: [ Daily ] [ Monthly ] [ Billing ] │
└────────────────────────────────────────────┘
```

### Footer Info Bar

Add navigation hints:

```
←/→ Switch Section  |  ↑/↓ Scroll  |  R Refresh  |  Q Quit
```

Use `bubbles/help` for consistent layout.

---

## 4. Add Interaction

### Scrollable Viewport

Wrap content in a `viewport.Model` for scrolling:

```go
viewport := viewport.New(width, height)
viewport.SetContent(content)
```

### Sorting and Filtering

Enable sorting by column and simple filters (e.g., last 30 days).

### Auto-refresh Timer

Trigger data refresh automatically every 60s:

```
⟳ Refreshing data...
```

---

## 5. Add Aesthetic Layers

### Gradient Headers

```go
header := lipgloss.NewStyle().
  Foreground(lipgloss.Color("15")).
  Background(lipgloss.Color("33")).
  Bold(true).
  Padding(0, 2).
  Render("Usage Dashboard")
```

### Inline Color Tokens

Use color for cost levels:

```go
func colorCost(cost float64) string {
  switch {
  case cost > 2:
    return m.styles.Error.Render(fmt.Sprintf("$%.2f", cost))
  case cost > 1:
    return m.styles.Warning.Render(fmt.Sprintf("$%.2f", cost))
  default:
    return m.styles.Success.Render(fmt.Sprintf("$%.2f", cost))
  }
}
```

### Animated Progress Bar for Billing Block

```
█████████████░░░░░░░░░░░  62% remaining
```

Use `bubbles/progress` to render.

---

## 6. Add Optional Widgets

| Widget | Library | Purpose |
|--------|----------|----------|
| Graph (sparklines) | termui/v3, bubbles-graph | Mini trend charts |
| Mini bar chart | Lipgloss + runes | Cost by day/model |
| Spinner | bubbles/spinner | Show refresh activity |
| Progress bar | bubbles/progress | Billing depletion |
| Tooltip / hover | bubbles/list | Row detail |

---

## 7. Add Personality

Add a banner or identity element:

```
🛰️  Eye in the Sky — Usage Metrics
"Monitoring the minds and the money."
```

---

## 8. Example Layout

```
🛰️ Eye in the Sky — Usage Metrics
──────────────────────────────────────────────────────────────
💰 $12.34 total | 🧠 1,234,987 tokens | ⟳ Updated 2m ago
↑ +12.4% vs last month

[Eye-in-the-Sky Metrics]
┌──────────────────────────────────────────────┐
│ ...table...                                  │
└──────────────────────────────────────────────┘

[Claude Code Daily Usage]
┌──────────────────────────────────────────────┐
│ ...table...                                  │
└──────────────────────────────────────────────┘

[Billing Block]
█████████████░░░░░░░░░░░ 62% remaining

──────────────────────────────────────────────────────────────
←/→ Switch Section | ↑/↓ Scroll | R Refresh | Q Quit
```

---

## Summary of Additions

1. Quick summary bar  
2. Scrollable viewport  
3. Trend indicators  
4. Progress bar for billing  
5. Colored cost tokens  
6. Tabs or grouped sections  
7. Footer help hints  
8. Gradient header + spacing  

---

This will make the `view_usage` page feel like a **real interactive dashboard**, not just a text dump.
