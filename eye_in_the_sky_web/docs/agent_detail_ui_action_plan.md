# Agent Detail UI — Junior Developer Action Plan

Goal
- Improve the Agent Detail page to be clearer, more consistent, and more actionable, with minimal risk and incremental commits.

Scope
- Files you will edit:
  - `eye_in_the_sky_web/assets/svelte/components/AgentDetail.svelte`
  - `eye_in_the_sky_web/assets/css/app.css`
  - `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/live/agent_live/show.ex`
  - (Optional, if needed) `eye_in_the_sky_web/assets/js/app.js`

Prerequisites
- Able to run the web app locally:
  - In `eye_in_the_sky_web/` run: `mix setup` then `mix phx.server`
  - Visit http://localhost:4000
- Keep changes incremental and test visually after each step.

Acceptance Criteria
- Status pill is color‑semantic: active = green, completed = gray, failed = red, idle = amber.
- Tab counts appear as uniform, small rounded badges; hide when zero; visually consistent on all tabs.
- Time: header shows a friendly relative “Started … ago” with a tooltip or title showing the exact timestamp; Duration live‑updates while active.
- Header stays visible while scrolling (sticky) and has a subtle shadow once content scrolls.
- Content panel uses lighter borders, more padding, and consistent spacing.
- Header has clear primary actions: End Session, New Task, Add Note (wire up events; non‑destructive actions can be placeholders with a flash/toast for now).
- A11y: Tabs are keyboard focusable and indicate selection.

Milestone 1 — Status Pill (color + clarity)
1) Open `assets/svelte/components/AgentDetail.svelte`.
2) Add a small helper for status classes:

```svelte
  const statusStyles = {
    active: "bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-900/30 dark:text-emerald-300 dark:border-emerald-800",
    completed: "bg-gray-100 text-gray-700 border border-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-600",
    failed: "bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-900/30 dark:text-rose-300 dark:border-rose-800",
    idle: "bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-900/30 dark:text-amber-300 dark:border-amber-800"
  }
```

3) Replace the existing status `<span>` class expression with:

```svelte
  <span class={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${statusStyles[header.status] || statusStyles.completed}`} aria-label={`Status: ${header.status}`}>
    {header.status}
  </span>
```

Milestone 2 — Normalize Tab Count Badges
1) In `AgentDetail.svelte`, keep `tabs` but ensure counts render as small pills and hide when 0:

```svelte
  function countFor(key) {
    return counts?.[key] || 0
  }
```

2) Update the tab button’s badge snippet:

```svelte
  {#if t.countKey && countFor(t.countKey) > 0}
    <span class={`rounded-full px-1.5 py-0.5 text-[11px] font-bold ${activeTab === t.key ? 'bg-white/20 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}`}>
      {countFor(t.countKey)}
    </span>
  {/if}
```

3) Add basic ARIA to tabs:

```svelte
  <div role="tablist" aria-label="Session sections" class="inline-flex gap-1">
    <!-- tab buttons -->
    <button role="tab" aria-selected={activeTab === t.key} ...>
```

Milestone 3 — Relative Time + Live Duration
1) Parse the `header.started` which may be an ISO string or a Go‑style string. Add helpers:

```svelte
  function parseDateLike(v) {
    if (!v) return null
    // Try ISO first
    const d1 = new Date(v)
    if (!isNaN(d1)) return d1
    // Try Go format: "YYYY-MM-DD HH:MM:SS ..."
    const parts = String(v).split(" ")
    if (parts.length >= 2) {
      const isoish = parts[0] + "T" + parts[1]
      const d2 = new Date(isoish)
      if (!isNaN(d2)) return d2
    }
    return null
  }

  function relativeFrom(date) {
    if (!date) return "—"
    const secs = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000))
    const m = Math.floor(secs / 60), s = secs % 60
    const h = Math.floor(m / 60), mm = m % 60
    if (h > 0) return `${h}h ${mm}m ago`
    if (m > 0) return `${m}m ago`
    return `${s}s ago`
  }

  let tick = 0
  import { onMount } from "svelte"
  onMount(() => {
    const id = setInterval(() => tick++, 1000)
    return () => clearInterval(id)
  })
```

2) Use these in the header:

```svelte
  {#if header.started}
    <div class="rounded-lg bg-gray-100 dark:bg-gray-800 px-3 py-2" title={String(header.started) + " UTC"}>
      <span class="text-gray-600 dark:text-gray-400 font-medium">Started</span>
      <span class="ml-2 text-gray-900 dark:text-gray-100 font-semibold">{relativeFrom(parseDateLike(header.started))}</span>
    </div>
  {/if}

  <div class="rounded-lg bg-gray-100 dark:bg-gray-800 px-3 py-2">
    <span class="text-gray-600 dark:text-gray-400 font-medium">Duration</span>
    <span class="ml-2 text-gray-900 dark:text-gray-100 font-semibold">
      {header.status === 'active' ? relativeFrom(parseDateLike(header.started)) : (header.duration ?? '—')}
    </span>
  </div>
```

Milestone 4 — Sticky Header + Spacing
1) Make the header sticky by adjusting the header wrapper in `AgentDetail.svelte`:

```svelte
  <div class="bg-white/80 dark:bg-gray-800/80 backdrop-blur sticky top-0 z-30 border-b border-gray-200 dark:border-gray-700 px-6 py-5">
```

2) Add a little gap between tabs and content and lighten content borders:
   - Ensure the container around tab content has `mt-3`–`mt-4` and `p-4 md:p-6`.
   - Prefer `border-gray-200 dark:border-gray-700` and add `shadow-sm` instead of heavy outlines.

Milestone 5 — Header Actions (wire-ups)
1) Add a right‑side button group in the header:

```svelte
  <div class="flex shrink-0 items-center gap-2">
    <button class="rounded-md bg-gray-900 dark:bg-gray-700 px-3 py-1.5 text-sm font-semibold text-white hover:bg-gray-800 dark:hover:bg-gray-600"
            on:click={() => live.pushEvent('end_session')} aria-label="End session">
      End Session
    </button>
    <button class="rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-1.5 text-sm font-semibold text-gray-800 dark:text-gray-100 hover:bg-gray-50"
            on:click={() => live.pushEvent('new_task')} aria-label="Create new task">
      New Task
    </button>
    <button class="rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-1.5 text-sm font-semibold text-gray-800 dark:text-gray-100 hover:bg-gray-50"
            on:click={() => live.pushEvent('add_note')} aria-label="Add note">
      Add Note
    </button>
  </div>
```

2) In `lib/eye_in_the_sky_web_web/live/agent_live/show.ex`, add placeholder handlers that keep the UI responsive:

```elixir
  @impl true
  def handle_event("new_task", _params, socket) do
    # TODO: open a modal or navigate to tasks tab in a future PR
    {:noreply, socket}
  end

  @impl true
  def handle_event("add_note", _params, socket) do
    # TODO: open a note input in a future PR
    {:noreply, socket}
  end
```

Milestone 6 — CSS Touch‑ups (borders, spacing)
1) In `assets/css/app.css`, confirm the existing borders are not globally heavier than needed. If the main content still feels heavy, add a small rule to soften panel borders globally used by this page:

```css
/* Optional: soften common panel borders used on Agent Detail */
.agent-detail-wrapper .panel-soft { border-color: rgb(229 231 235 / 1); } /* gray-200 */
```

2) Apply `panel-soft` to the main content panels in `AgentDetail.svelte` if needed.

Milestone 7 — A11y and Keyboard Support
- Ensure buttons and tabs have clear labels and `aria-selected` on the active tab.
- Verify focus rings are visible (Tailwind defaults are OK). Avoid removing outlines.

Milestone 8 — Manual QA
- Navigate to an agent session and verify:
  - Status pill color matches state, text is readable in light/dark modes.
  - Started shows a friendly relative time; hovering reveals exact timestamp (via title attribute).
  - Duration updates every second while active.
  - Tab counts appear only when > 0 and style is consistent.
  - Header stays visible while scrolling.
  - End Session/New Task/Add Note buttons appear and do not break the page. End Session triggers the existing LV event.
  - No console errors; network and LiveView events are clean.

Commit Plan
- Commit by milestone where possible. Example messages:
  - "UI: color‑semantic status pill + sticky header"
  - "UI: tab count badges + spacing"
  - "UX: relative time + live duration"
  - "Actions: add header actions + LV stubs"

Notes
- Follow Tailwind utility classes (no inline `<style>`). Avoid DaisyUI components.
- Keep Svelte logic simple and predictable; prefer small helpers over complex stores.
- Avoid server changes unless necessary; the LV stubs above are safe placeholders.

