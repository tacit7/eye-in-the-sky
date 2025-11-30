# Agent 37265fd3 — Critique & Action Plan

## Goals
- Improve information density, visual hierarchy, and consistency on the Agent Detail view.
- Reduce triage time with clearer tabs, states, and actionable lists.
- Maintain accessibility (contrast, focus) and LiveView performance (streams).

## Quick Wins (Day 1–2)
- Page container: wrap content in `max-w-7xl mx-auto px-6`.
- Meta chips: adopt one compact pill style `inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-sm` with optional `<.icon>`.
- Sticky tabs: add bottom divider, bold active tab, count badges, and `aria-current="page"`.
- Copy action: use existing `CopyToClipboard` hook for session ID.

## Iteration 2 (Day 3–5)
- Convert Tasks/Commits/Logs to LiveView streams; track counts in assigns.
- Cards: lighter borders, subtle hover elevation (`transition-all duration-150`), consistent tag layout with `flex flex-wrap gap-2`.
- Empty/loading: add empty states and skeletons using existing `phx-*` variants.
- URL state: persist filters/search in params; use `<.link patch={...}>` and `handle_params/3`.

## Iteration 3 (Day 5+)
- Logs UX: monospace, wrap/nowrap toggle, level filters, sticky time column.
- Keyboard/a11y: visible focus rings (`focus-visible:outline focus-visible:outline-2`), label controls, `aria-live="polite"` for counts.
- Micro-interactions: hover/press states on buttons, card lift on hover, smooth tab transitions.

## Acceptance Criteria
- Layout uses `max-w-7xl mx-auto px-6`; meta row is a responsive grid.
- Unified chip style across the view with consistent spacing and contrast ≥ 4.5:1.
- Tabs are sticky with clear active state and count badges; navigation uses patch.
- Lists are streamed; counts are accurate and updated without full reloads.
- Empty and loading states are present and visually consistent.

## Risks & Assumptions
- NATS/log feeds may be intermittent; UI should handle empty/late-arriving data gracefully.
- Ensure no usage of deprecated LiveView APIs (`live_patch`, `live_redirect`).

## References
- `docs/agent_detail_ui_review.md` for design notes and HEEx snippet.
