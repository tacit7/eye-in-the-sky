# Agent Detail — UI Review & Recommendations

This review summarizes improvements for the Agent Detail view based on the current layout and visual style.

## Key Issues

- Content runs edge‑to‑edge, creating low information density and weak hierarchy.
- Meta “chips” (session, project, started, duration, status) use mixed styles and spacing.
- Tabs aren’t visually anchored; counts lack badges, and the active tab is subtle.
- Cards use heavy borders and inconsistent tag chips; large empty canvas when lists are short.

## Recommendations

- Container and grid: wrap page in `max-w-7xl mx-auto px-6` and render the meta row as a responsive grid (`grid grid-cols-1 md:grid-cols-3 gap-3`).
- Unified chips: one compact pill style (`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-sm`) with small `<.icon>` where useful.
- Breadcrumb: keep “Back to Agents” lightweight; reduce size and contrast.
- Tabs: make sticky with a bottom divider; show counts as badges; emphasize active tab with bold + bottom indicator; add `aria-current="page"`.
- Cards: replace thick borders with subtle border + hover elevation; promote titles, use `flex flex-wrap gap-2` for tags.
- Empty/loading: provide an empty‑state card and skeletons using your existing `@custom-variant phx-*` Tailwind variants.
- A11y: ensure 4.5:1 contrast for text inside chips; keep visible focus rings.
- LiveView: use streams for Tasks/Commits/Logs and track counts in assigns; tabs should use `<.link patch={...}>`.

## Implementation Hints (HEEx)

```heex
<div class="max-w-7xl mx-auto px-6">
  <h1 class="text-2xl font-semibold">Claude Code</h1>
  <div class="mt-2 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
    <span class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5">
      <.icon name="hero-hashtag" class="h-4 w-4" /> <code>{@session_short}</code>
    </span>
    <span class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5">
      <.icon name="hero-folder" class="h-4 w-4" /> {@project_name}
    </span>
    <span class="inline-flex items-center gap-1 rounded-full border px-2 py-0.5">
      <.icon name="hero-clock" class="h-4 w-4" /> Started {@started_at} • {@status}
    </span>
  </div>
</div>
```

Use the existing `CopyToClipboard` hook (assets/js/hooks/copy_to_clipboard.js) for the session ID copy action. Keep all styles Tailwind‑based (no DaisyUI components).

