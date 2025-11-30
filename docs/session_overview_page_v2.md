# Session Overview Page (All Agents, All Projects)

## Purpose

Single place to see **all active and recent sessions** across all agents and projects.

This view answers:

- What is currently running or recently active
- Which project and agent a session belongs to
- How long it has been running
- How many tokens it has burned
- How to jump into or restart work from here

## Route and Module

- Route: `/sessions`
- LiveView: `EyeWeb.SessionLive.Index`
- Navigates into agent detail: `/agents/:agent_id?s=:session_id`

## Columns and Behaviors

Table shows one row per session, newest first (by last action).

Columns:

1. **Session ID**
   - Full `sessions.id`
   - Click → navigate to agent detail for this session:
     - `GET /agents/:agent_id?s=:session_id`
   - Secondary action: “Copy ID” button in the row (client-side clipboard).

2. **Status**
   - `sessions.status`
   - Rendered as a small badge:
     - `running` → green
     - `idle` → blue
     - `failed` → red
     - everything else → gray

3. **Project**
   - `projects.name` (via `sessions.project_id`)
   - If no project: show `—`

4. **Session name**
   - `sessions.name` (optional human label)
   - If null: show `—`

5. **Last action**
   - `sessions.last_action_at`
   - Human readable timestamp, UTC or app default timezone.
   - If null: show `—`

6. **Duration**
   - Derived as:
     - from `sessions.started_at` to `sessions.ended_at` if ended
     - else from `sessions.started_at` to now
   - Render as:
     - `2h 13m`, `7m`, or `42s`
   - If `started_at` is missing: show `—`

7. **Token usage**
   - From `session_metrics` (if present):
     - `total_tokens`
     - `prompt_tokens`
     - `completion_tokens`
   - Display format:
     - `12345 (8000 / 4345)`
       - first = total
       - inside parentheses = `prompt / completion`
   - If metrics row missing: show `0` or `—` depending on your preference.

8. **Actions**
   - **Copy ID** button
     - Uses a LiveView hook for `navigator.clipboard.writeText(session_id)`
   - **New session** button
     - Starts a new session for that agent and navigates to it.

## Data Source

Read-only query that joins sessions, agents, projects, and optional metrics.

### Function: `Eye.Sessions.list_session_overview_rows/1`

Signature:

```elixir
@spec list_session_overview_rows(keyword()) :: [session_overview_row()]
```

Return shape:

```elixir
@type session_overview_row :: %{
        session_id: String.t(),
        session_status: String.t(),
        session_name: String.t() | nil,
        agent_id: String.t(),
        project_name: String.t() | nil,
        last_action_at: NaiveDateTime.t() | DateTime.t() | nil,
        started_at: NaiveDateTime.t() | DateTime.t(),
        ended_at: NaiveDateTime.t() | DateTime.t() | nil,
        duration_seconds: integer() | nil,
        total_tokens: integer() | nil,
        prompt_tokens: integer() | nil,
        completion_tokens: integer() | nil
      }
```

Implementation sketch:

```elixir
def list_session_overview_rows(opts \ []) do
  limit = Keyword.get(opts, :limit, 200)

  from(s in Session,
    join: a in Agent, on: a.id == s.agent_id,
    left_join: p in Project, on: p.id == s.project_id,
    left_join: m in SessionMetric, on: m.session_id == s.id,
    order_by: [desc: s.last_action_at],
    limit: ^limit,
    select: %{
      session_id: s.id,
      session_status: s.status,
      session_name: s.name,
      agent_id: a.id,
      project_name: p.name,
      last_action_at: s.last_action_at,
      started_at: s.started_at,
      ended_at: s.ended_at,
      duration_seconds:
        fragment(
          "strftime('%s', coalesce(?, CURRENT_TIMESTAMP)) - strftime('%s', ?)",
          s.ended_at,
          s.started_at
        ),
      total_tokens: m.total_tokens,
      prompt_tokens: m.prompt_tokens,
      completion_tokens: m.completion_tokens
    }
  )
  |> Repo.all()
end
```

If you do not have `session_metrics` yet, remove the join and the token fields. Duration logic can also be done in Elixir instead of SQL; this is just a shortcut for SQLite.

## LiveView: `EyeWeb.SessionLive.Index`

Minimal LiveView that loads rows and wires up the actions.

```elixir
defmodule EyeWeb.SessionLive.Index do
  use EyeWeb, :live_view

  alias Eye.Sessions

  @impl true
  def mount(_params, _session, socket) do
    sessions = Sessions.list_session_overview_rows()

    {:ok,
     socket
     |> assign(:sessions, sessions)
     |> assign(:page_title, "Session overview")}
  end

  @impl true
  def handle_event("start_session", %{"agent_id" => agent_id}, socket) do
    {:ok, session} = Sessions.start_session_for_agent(agent_id)

    {:noreply,
     push_navigate(socket,
       to: ~p"/agents/#{session.agent_id}?s=#{session.id}"
     )}
  end

  @impl true
  def handle_event("start_session_global", _params, socket) do
    # For now just send the user to the agents list.
    {:noreply, push_navigate(socket, to: ~p"/")}
  end

  # You can add a handle_event("copy_id", ...) later if you want logging
  # for copy actions; the copy itself is client-side via a hook.

  # Formatting helpers

  defp status_badge_class("running"),
    do:
      "inline-flex items-center rounded px-2 py-0.5 text-xs bg-emerald-100 text-emerald-700"

  defp status_badge_class("idle"),
    do: "inline-flex items-center rounded px-2 py-0.5 text-xs bg-sky-100 text-sky-700"

  defp status_badge_class("failed"),
    do: "inline-flex items-center rounded px-2 py-0.5 text-xs bg-red-100 text-red-700"

  defp status_badge_class(_),
    do: "inline-flex items-center rounded px-2 py-0.5 text-xs bg-slate-100 text-slate-700"

  defp format_timestamp(nil), do: "—"

  defp format_timestamp(%NaiveDateTime{} = dt),
    do: Calendar.strftime(dt, "%Y-%m-%d %H:%M:%S")

  defp format_timestamp(%DateTime{} = dt),
    do: Calendar.strftime(dt, "%Y-%m-%d %H:%M:%S")

  defp format_duration(nil), do: "—"

  defp format_duration(seconds) when is_integer(seconds) do
    minutes = div(seconds, 60)
    hrs = div(minutes, 60)
    mins = rem(minutes, 60)

    cond do
      hrs > 0 -> "#{hrs}h #{mins}m"
      mins > 0 -> "#{mins}m"
      true -> "#{seconds}s"
    end
  end
end
```

## Route

```elixir
# lib/eye_web/router.ex
scope "/", EyeWeb do
  pipe_through :browser

  live "/", AgentLive.Index, :index
  live "/agents/:id", AgentLive.Show, :show

  live "/sessions", SessionLive.Index, :index
end
```

## Template: `session_live/index.html.heex`

```heex
<div class="flex items-center justify-between mb-4">
  <h1 class="text-xl font-semibold">Session overview</h1>

  <button
    phx-click="start_session_global"
    class="inline-flex items-center rounded border px-3 py-1 text-sm font-medium bg-primary text-primary-foreground hover:opacity-90"
  >
    Start new session
  </button>
</div>

<div class="overflow-x-auto border rounded-md">
  <table class="min-w-full text-sm">
    <thead class="bg-muted">
      <tr>
        <th class="px-3 py-2 text-left font-semibold">Session ID</th>
        <th class="px-3 py-2 text-left font-semibold">Status</th>
        <th class="px-3 py-2 text-left font-semibold">Project</th>
        <th class="px-3 py-2 text-left font-semibold">Session name</th>
        <th class="px-3 py-2 text-left font-semibold">Last action</th>
        <th class="px-3 py-2 text-left font-semibold">Duration</th>
        <th class="px-3 py-2 text-left font-semibold">Tokens</th>
        <th class="px-3 py-2 text-left font-semibold">Actions</th>
      </tr>
    </thead>
    <tbody>
      <%= for session <- @sessions do %>
        <tr class="border-t hover:bg-accent/40">
          <td class="px-3 py-2 font-mono text-xs">
            <.link
              navigate={~p"/agents/#{session.agent_id}?s=#{session.session_id}"}
              class="underline decoration-dotted"
            >
              <%= session.session_id %>
            </.link>
          </td>

          <td class="px-3 py-2">
            <span class={status_badge_class(session.session_status)}>
              <%= session.session_status %>
            </span>
          </td>

          <td class="px-3 py-2">
            <%= session.project_name || "—" %>
          </td>

          <td class="px-3 py-2">
            <%= session.session_name || "—" %>
          </td>

          <td class="px-3 py-2 text-xs text-muted-foreground">
            <%= format_timestamp(session.last_action_at) %>
          </td>

          <td class="px-3 py-2 text-xs">
            <%= format_duration(session.duration_seconds) %>
          </td>

          <td class="px-3 py-2 text-xs text-muted-foreground">
            <%= session.total_tokens || 0 %>
            <%= if session.prompt_tokens && session.completion_tokens do %>
              (<%= session.prompt_tokens %> / <%= session.completion_tokens %>)
            <% end %>
          </td>

          <td class="px-3 py-2 text-xs">
            <div class="flex gap-2">
              <button
                phx-hook="CopyToClipboard"
                data-session-id={session.session_id}
                class="border px-2 py-1 rounded hover:bg-muted"
              >
                Copy ID
              </button>

              <button
                phx-click="start_session"
                phx-value-agent_id={session.agent_id}
                class="border px-2 py-1 rounded hover:bg-muted"
              >
                New session
              </button>
            </div>
          </td>
        </tr>
      <% end %>
    </tbody>
  </table>
</div>
```

## Copy to Clipboard Hook

```javascript
// assets/js/hooks/copy_to_clipboard.js
export const CopyToClipboard = {
  mounted() {
    this.handleClick = () => {
      const id = this.el.dataset.sessionId
      if (!id) return

      if (navigator.clipboard?.writeText) {
        navigator.clipboard
          .writeText(id)
          .catch((err) => console.error("Failed to copy session id", err))
      }
    }

    this.el.addEventListener("click", this.handleClick)
  },

  destroyed() {
    if (this.handleClick) {
      this.el.removeEventListener("click", this.handleClick)
    }
  }
}
```

Hook registration:

```javascript
// assets/js/app.js
import { Socket } from "phoenix"
import { LiveSocket } from "phoenix_live_view"
import topbar from "../vendor/topbar"
import { CopyToClipboard } from "./hooks/copy_to_clipboard"

let Hooks = {}
Hooks.CopyToClipboard = CopyToClipboard

let csrfToken = document
  .querySelector("meta[name='csrf-token']")
  .getAttribute("content")

let liveSocket = new LiveSocket("/live", Socket, {
  params: { _csrf_token: csrfToken },
  hooks: Hooks
})

liveSocket.connect()
window.liveSocket = liveSocket
```

## UX Summary

- Shows **all sessions** across all agents and projects in one place.
- User can:
  - Click Session ID to jump into the agent detail for that session.
  - Copy Session ID with one click.
  - Start a new session from a given agent directly from the table.
- Columns include:
  - Session ID, Status, Project, Session name, Last action, Duration, Token usage, and row actions.

This is enough for a dev to implement the view without guessing intent or behavior.
