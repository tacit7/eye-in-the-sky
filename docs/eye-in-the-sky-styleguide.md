# Eye in the Sky Dashboard - Dark Theme Style Guide

## 🎨 Color Palette

| Purpose              | Hex / HSL             | Notes |
|----------------------|-----------------------|-------|
| **Background base**  | `#1B1F28` (hsl(222, 22%, 13%)) | Deep navy-gray, not pure black (keeps it readable). |
| **Card surface**     | `#242B38` (hsl(220, 21%, 18%)) | Slightly lighter for panels, feeds, agent cards. |
| **Borders/Dividers** | `#3A3F4C` (hsl(220, 13%, 27%)) | Subtle separators. |
| **Text primary**     | `#E5E9F0` (hsl(210, 20%, 94%)) | Almost white, not blinding. |
| **Text secondary**   | `#A0A9B8` (hsl(220, 10%, 70%)) | Muted gray for meta info. |
| **Active (green)**   | `#3DDC97` (hsl(160, 70%, 55%)) | Teal green for healthy agents. |
| **Idle (gold)**      | `#E0B95B` (hsl(42, 70%, 62%))  | Warm gold for idle status. |
| **Error (red)**      | `#F96363` (hsl(0, 83%, 68%))   | Soft red for failures. |
| **Completed (gray)** | `#9CA3AF` (hsl(220, 9%, 65%))  | Neutral gray. |
| **Highlight/link**   | `#83AAFF` (hsl(220, 100%, 75%)) | Cool blue accent. |

---

## 🔤 Typography

- **Font family**: `Inter`, `Roboto`, or system sans.
- **Sizes**:
  - Headline: `text-2xl font-semibold`
  - Section titles: `text-xl font-medium`
  - Body: `text-base`
  - Meta: `text-sm text-gray-400`
- **Line height**: 1.5 for readability.

---

## 📐 Layout

- **Header bar**: `bg-[#1A1F25]`, pinned top, logo + title on left, settings button on right.
- **Dashboard**:
  - Left/main column: agent cards grid.
  - Right column or bottom panel: activity feed.
- **Agent detail**:
  - Header: agent name + status badge.
  - Tab bar: Logs | Notes | Context.
  - Content panel: scrollable logs by default.

---

## 🧩 Components

### Status Badge
- **Shape**: pill (`rounded-full px-2 py-1`).
- **Color**: uses status palette above.
- **Icon + text**: `🟢 Active`, `🟡 Idle`, `🔴 Failed`, `⚪ Completed`.

### Agent Card
- **Container**: `bg-surface p-4 rounded-lg shadow-sm hover:shadow-md`.
- **Sections**:
  - Title (agent name).
  - Status badge + task.
  - Source + last updated.

### Activity Feed Item
- **Row layout**: timestamp left, icon+message right.
- **Row background**: alternate subtle shades (`#242B38` vs `#2C323C`).
- **Icons**:
  - Commit → git icon (blue accent).
  - Action → ⚡ (yellow).
  - Error → ⚠️ (red).

### Tabs
- **Style**: segmented control.
- **Active tab**: underline in primary accent (`#3DDC97`).
- **Inactive tabs**: muted gray text.

---

## 🌙 Dark Mode Behaviors

- Always dark, no toggle (for now).
- Maintain contrast ratios: WCAG AA for text on backgrounds.
- Accents pop against deep navy-gray, not neon on black.

---

## 🛠 Tailwind Tokens

```js
theme: {
  extend: {
    colors: {
      background: '#1B1F28',
      surface: '#242B38',
      border: '#3A3F4C',
      text: {
        primary: '#E5E9F0',
        secondary: '#A0A9B8'
      },
      status: {
        active: '#3DDC97',
        idle: '#E0B95B',
        error: '#F96363',
        completed: '#9CA3AF'
      },
      highlight: '#83AAFF'
    }
  }
}
```

---

This style guide defines the dark theme design system for **Eye in the Sky Dashboard**. It ensures a consistent and modern look, distinct from RouteWise.
