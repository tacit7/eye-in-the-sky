
# 🧭 Eye in the Sky — Bubble Tea Migration Plan

### Author: Engineering
### Goal
Re-implement the Eye in the Sky TUI using **Bubble Tea** for better stability, color, and configurability.
This version keeps SQLite storage, config files under `~/.config/.eyeinthesky/`, and a polling loop to refresh data.

## 1. Repository Layout
```
eye-in-the-sky/
├── cmd/eye-ui/main.go
├── internal/ui/
│   ├── app/
│   │   ├── model.go
│   │   ├── update.go
│   │   ├── view.go
│   │   ├── keymap.go
│   │   ├── config.go
│   │   ├── theme.go
│   │   └── polling.go
│   ├── components/
│   │   ├── agent_list.go
│   │   ├── agent_detail.go
│   │   ├── context_list.go
│   │   ├── compression_list.go
│   │   └── log_list.go
│   └── util/
│       ├── editor.go
│       ├── cmdrunner.go
│       └── db.go
└── internal/database/
```

## 2. Configuration System
Describes how config, keybindings, and theme files should be structured in `~/.config/.eyeinthesky/`.

## 3. Bubble Tea Program Skeleton
Includes sample code for `main.go`, `model.go`, `update.go`, and `view.go`.

## 4. Development Plan
Day-by-day breakdown for implementing each module, ensuring feature parity with the old TUI.

## 5. Components
Outlines behavior for `agent_list`, `agent_detail`, `context_list`, `compression_list`, and `log_list` components.

## 6. Polling & Commands
Explains polling mechanism, message flow, and asynchronous updates with Bubble Tea.

## 7. Visual Design
Details theme management, status bar, and accent color rules with Lip Gloss.

## 8. Implementation Rules
Defines conventions, DB access patterns, and component isolation guidelines.
