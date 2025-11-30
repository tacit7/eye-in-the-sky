# Agent Web Dashboard - Implementation Plan

## Purpose

Replace the current multi-page web UI with a **single cockpit view** showing everything about an agent in one screen. No more clicking between tabs on separate pages - all critical info visible at once.

This answers:
- What is this agent doing right now?
- What sessions has it run?
- What tasks/commits/logs are in the active session?
- What's the session context and notes?
- How many tokens has it burned?

## Design Philosophy

**"One cockpit where you can actually see what the agent is doing"** - not 7 separate walls of text.

---

## Phase 1: Lock the Layout

### Desktop Layout (3-Pane)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  HEADER: Agent f91bbdf5 | Status: active | Project: eye-in-the-sky          │
│  Active Session: session-123 | Duration: 2h 13m | Tokens: 12.5k             │
│  [Start New Session] [Open in TUI] [Copy Session ID]                        │
├──────────────┬────────────────────────────────────────┬─────────────────────┤
│              │                                        │                     │
│  SESSIONS    │  MAIN WORK AREA                        │  CONTEXT PANEL      │
│  (Left)      │  (Center)                              │  (Right)            │
│              │                                        │                     │
│  • session-1 │  [Tasks] [Commits] [Logs]              │  Session Context:   │
│  • session-2 │  ────────────────────────────          │  Phase: implement   │
│  • session-3 │                                        │  Progress: 60%      │
│    (active)  │  ┌─────────────┬──────────────┐        │  Tasks: 3/8         │
│  • session-4 │  │ Task List   │ Task Details │        │                     │
│              │  │             │              │        │  Notes:             │
│  [All]       │  │ • Task 1    │ Title: ...   │        │  - Note 1           │
│  [Active]    │  │ • Task 2    │ Desc: ...    │        │  - Note 2           │
│  [Bookmarks] │  │   (selected)│ Priority: 80 │        │  [Add Note]         │
│              │  │ • Task 3    │              │        │                     │
│              │  │             │ Annotations: │        │  Stats:             │
│              │  │             │ - Annot 1    │        │  Tokens: 12.5k      │
│              │  │             │ - Annot 2    │        │  Cost: $0.15        │
│              │  └─────────────┴──────────────┘        │  Duration: 2h 13m   │
│              │                                        │                     │
│  220-260px   │  Flexible width                        │  260-320px          │
└──────────────┴────────────────────────────────────────┴─────────────────────┘
```

**Pane Widths:**
- **Left (Sessions)**: 220-260px fixed/min width
- **Center (Main)**: Flexible, takes remaining space
- **Right (Context)**: 260-320px fixed/min width

**Resizable:** Yes, with shadcn-svelte Resizable component (Phase 5)

### Mobile / Small Width Behavior

**Stack vertically** when width < 1024px:

```
┌─────────────────────────────┐
│  HEADER (collapsed)         │
│  Agent f91bbdf5 | active    │
│  [≡ Menu]                   │
├─────────────────────────────┤
│                             │
│  MAIN WORK AREA             │
│  [Tasks] [Commits] [Logs]   │
│                             │
│  Task List + Details        │
│  (no split on mobile)       │
│                             │
├─────────────────────────────┤
│                             │
│  ACCORDION SECTIONS         │
│  ▼ Sessions (3)             │
│    • session-1              │
│    • session-2              │
│    • session-3 (active)     │
│                             │
│  ▼ Context                  │
│    Phase: implement         │
│    Progress: 60%            │
│                             │
│  ▼ Notes (2)                │
│    - Note 1                 │
│    - Note 2                 │
│    [Add Note]               │
│                             │
│  ▼ Stats                    │
│    Tokens: 12.5k            │
│    Cost: $0.15              │
└─────────────────────────────┘
```

**Do NOT try to make 3 resizable panes on mobile.** Stack them instead.

---

## Phase 2: LiveView Scaffolding

### Extend Existing AgentLive.Show

**Route:** `/agents/:id` (already exists)

**Add query param for session selection:** `/agents/:id?s=:session_id`

### Required Assigns

```elixir
%{
  # Core data
  agent: agent,                          # Current agent
  sessions: sessions_for_agent,          # All sessions for this agent
  active_session: active_session,        # Currently selected session

  # Session-specific data
  tasks: tasks_for_active_session,
  commits: commits_for_active_session,
  logs: logs_for_active_session,
  notes: notes_for_active_session,
  session_context: session_context,
  metrics: metrics_for_active_session,

  # UI state
  active_tab: :tasks,                    # :tasks | :commits | :logs
  selected_task_index: 0,                # For split-pane selection
  selected_commit_index: 0
}
```

### Data Loading Functions

Create in `lib/eye_in_the_sky_web/agents.ex`:

```elixir
# Core loaders
def get_agent_dashboard_data(agent_id) do
  agent = get_agent!(agent_id)
  sessions = Sessions.list_sessions_for_agent(agent_id)
  active_session = List.first(sessions) # Most recent

  %{
    agent: agent,
    sessions: sessions,
    active_session: active_session
  }
end

def load_session_data(session_id) do
  %{
    tasks: Tasks.list_tasks_for_session(session_id),
    commits: Commits.list_commits_for_session(session_id),
    logs: Logs.list_logs_for_session(session_id),
    notes: Notes.list_notes_for_session(session_id),
    session_context: Contexts.get_session_context(session_id),
    metrics: Metrics.get_session_metrics(session_id)
  }
end
```

**Key principle:** Lazy load per session, NOT per tab. Tabs are just different views of the same session data.

### LiveView Events

```elixir
# Session selection
def handle_event("select_session", %{"session_id" => id}, socket) do
  session_data = load_session_data(id)

  {:noreply,
   socket
   |> assign(:active_session, get_session!(id))
   |> assign(session_data)
   |> assign(:active_tab, :tasks) # Reset to tasks tab
  }
end

# Tab switching
def handle_event("change_tab", %{"tab" => tab}, socket) do
  {:noreply, assign(socket, :active_tab, String.to_existing_atom(tab))}
end

# Start new session (mocked for now)
def handle_event("start_session", _params, socket) do
  # TODO: Wire to Go MCP server or create session via API
  {:noreply, socket}
end

# Optional: Add note
def handle_event("add_note", %{"body" => body}, socket) do
  Notes.create_note(%{
    parent_type: "session",
    parent_id: socket.assigns.active_session.id,
    body: body
  })

  notes = Notes.list_notes_for_session(socket.assigns.active_session.id)
  {:noreply, assign(socket, :notes, notes)}
end
```

---

## Phase 3: Skeleton UI (No Fancy Yet)

### Header Section

```heex
<!-- lib/eye_in_the_sky_web_web/live/agent_live/show.html.heex -->
<div class="bg-white border-b border-gray-200 px-6 py-4">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-xl font-semibold text-gray-900">
        Agent <%= String.slice(@agent.id, 0..7) %>
      </h1>
      <div class="mt-1 flex items-center gap-4 text-sm text-gray-500">
        <span class={status_badge_class(@agent.status)}>
          <%= @agent.status %>
        </span>
        <span>Project: <%= @agent.project_name || "—" %></span>
      </div>
    </div>

    <div class="flex gap-2">
      <button
        phx-click="start_session"
        class="px-3 py-2 text-sm bg-indigo-600 text-white rounded hover:bg-indigo-700"
      >
        Start New Session
      </button>
      <button
        phx-hook="CopyToClipboard"
        data-session-id={@active_session.id}
        class="px-3 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50"
      >
        Copy Session ID
      </button>
    </div>
  </div>

  <%= if @active_session do %>
    <div class="mt-3 flex items-center gap-6 text-sm text-gray-600">
      <span>Session: <%= @active_session.name || String.slice(@active_session.id, 0..11) %></span>
      <span>Duration: <%= format_duration(@active_session.started_at, @active_session.ended_at) %></span>
      <span>Tokens: <%= @metrics.total_tokens || 0 %></span>
      <span>Last action: <%= format_timestamp(@active_session.started_at) %></span>
    </div>
  <% end %>
</div>
```

### Basic 3-Column Layout (CSS Grid)

```heex
<div class="grid grid-cols-[minmax(220px,260px)_minmax(0,1fr)_minmax(260px,320px)] gap-4 h-[calc(100vh-180px)] p-4">
  <!-- Left: Sessions Sidebar -->
  <div class="overflow-y-auto border border-gray-200 rounded-lg">
    <.svelte
      name="SessionsSidebar"
      props={%{
        sessions: @sessions,
        activeSessionId: @active_session.id
      }}
      socket={@socket}
    />
  </div>

  <!-- Center: Main Work Area -->
  <div class="overflow-y-auto border border-gray-200 rounded-lg">
    <.svelte
      name="MainWorkArea"
      props={%{
        activeTab: @active_tab,
        tasks: @tasks,
        commits: @commits,
        logs: @logs
      }}
      socket={@socket}
    />
  </div>

  <!-- Right: Context Panel -->
  <div class="overflow-y-auto border border-gray-200 rounded-lg">
    <.svelte
      name="ContextPanel"
      props={%{
        sessionContext: @session_context,
        notes: @notes,
        metrics: @metrics
      }}
      socket={@socket}
    />
  </div>
</div>
```

**Mobile Responsive:**
```heex
<!-- Use Tailwind responsive classes -->
<div class="grid grid-cols-1 lg:grid-cols-[minmax(220px,260px)_minmax(0,1fr)_minmax(260px,320px)] gap-4 ...">
```

When `lg:` breakpoint not met (< 1024px), collapses to single column stack.

---

## Phase 4: Svelte Components (v1)

### Component 1: SessionsSidebar.svelte

```svelte
<!-- assets/svelte/components/SessionsSidebar.svelte -->
<script>
  import { ScrollArea } from '$lib/components/ui/scroll-area'
  import { Button } from '$lib/components/ui/button'

  export let sessions = []
  export let activeSessionId
  export let live

  function selectSession(sessionId) {
    live.pushEvent('select_session', { session_id: sessionId })
  }

  function getStatusBadge(session) {
    // Derive status from ended_at
    return session.ended_at ? 'ended' : 'active'
  }
</script>

<div class="p-4">
  <h3 class="text-sm font-semibold mb-3">Sessions</h3>

  <div class="flex gap-2 mb-4">
    <Button variant="outline" size="sm">All</Button>
    <Button variant="outline" size="sm">Active</Button>
    <Button variant="outline" size="sm">Bookmarked</Button>
  </div>

  <ScrollArea class="h-[calc(100vh-300px)]">
    <div class="space-y-2">
      {#each sessions as session}
        <button
          class="w-full text-left p-3 rounded border hover:bg-accent"
          class:bg-accent={session.id === activeSessionId}
          class:border-indigo-500={session.id === activeSessionId}
          on:click={() => selectSession(session.id)}
        >
          <div class="flex items-center justify-between mb-1">
            <span class="text-sm font-medium truncate">
              {session.name || session.id.slice(0, 11)}
            </span>
            <span class="text-xs px-2 py-0.5 rounded"
                  class:bg-emerald-100={getStatusBadge(session) === 'active'}
                  class:text-emerald-700={getStatusBadge(session) === 'active'}
                  class:bg-gray-100={getStatusBadge(session) === 'ended'}
                  class:text-gray-600={getStatusBadge(session) === 'ended'}>
              {getStatusBadge(session)}
            </span>
          </div>
          <div class="text-xs text-muted-foreground">
            {session.started_at?.slice(0, 16)}
          </div>
        </button>
      {/each}
    </div>
  </ScrollArea>
</div>
```

**Props:**
- `sessions` - Array of session objects
- `activeSessionId` - Currently selected session
- `live` - LiveView socket for pushEvent

**Behavior:**
- Shows list of sessions for this agent
- Highlights active session
- Filters: All, Active, Bookmarked
- Click → `live.pushEvent('select_session', { session_id })`

### Component 2: MainWorkArea.svelte

```svelte
<!-- assets/svelte/components/MainWorkArea.svelte -->
<script>
  import { Tabs } from '$lib/components/ui/tabs'
  import { Resizable } from '$lib/components/ui/resizable'
  import TasksTab from './tabs/TasksTab.svelte'
  import CommitsTab from './tabs/CommitsTab.svelte'
  import LogsTab from './tabs/LogsTab.svelte'

  export let activeTab = 'tasks'
  export let tasks = []
  export let commits = []
  export let logs = []
  export let live

  function handleTabChange(tab) {
    live.pushEvent('change_tab', { tab })
  }
</script>

<Tabs.Root value={activeTab} onValueChange={handleTabChange}>
  <Tabs.List class="px-4 pt-4">
    <Tabs.Trigger value="tasks">Tasks</Tabs.Trigger>
    <Tabs.Trigger value="commits">Commits</Tabs.Trigger>
    <Tabs.Trigger value="logs">Logs</Tabs.Trigger>
  </Tabs.List>

  <Tabs.Content value="tasks" class="p-4">
    <TasksTab {tasks} {live} />
  </Tabs.Content>

  <Tabs.Content value="commits" class="p-4">
    <CommitsTab {commits} {live} />
  </Tabs.Content>

  <Tabs.Content value="logs" class="p-4">
    <LogsTab {logs} {live} />
  </Tabs.Content>
</Tabs.Root>
```

**Props:**
- `activeTab` - Current tab (:tasks | :commits | :logs)
- `tasks`, `commits`, `logs` - Data arrays
- `live` - LiveView socket

### Component 3: TasksTab.svelte (Split-Pane)

```svelte
<!-- assets/svelte/components/tabs/TasksTab.svelte -->
<script>
  import { Resizable } from '$lib/components/ui/resizable'
  import { ScrollArea } from '$lib/components/ui/scroll-area'

  export let tasks = []
  export let live

  let selectedTask = tasks[0] || null

  function priorityColor(priority) {
    if (priority >= 70) return 'text-red-500'
    if (priority >= 40) return 'text-yellow-500'
    return 'text-blue-500'
  }
</script>

<Resizable.PaneGroup direction="horizontal">
  <!-- Left: Task List -->
  <Resizable.Pane defaultSize={50} minSize={30}>
    <ScrollArea class="h-full">
      <div class="p-2 space-y-1">
        <div class="text-xs text-muted-foreground mb-2">
          {tasks.length} tasks
        </div>
        {#each tasks as task}
          <button
            class="w-full text-left p-2 rounded hover:bg-accent"
            class:bg-accent={selectedTask?.id === task.id}
            on:click={() => selectedTask = task}
          >
            <div class="flex items-center gap-2">
              <span class={priorityColor(task.priority)}>●</span>
              <span class="text-sm font-medium truncate">{task.title}</span>
            </div>
            <div class="text-xs text-muted-foreground ml-6">
              {task.state_name || 'No state'}
            </div>
          </button>
        {/each}
      </div>
    </ScrollArea>
  </Resizable.Pane>

  <Resizable.Handle />

  <!-- Right: Task Details -->
  <Resizable.Pane defaultSize={50} minSize={30}>
    <ScrollArea class="h-full p-4">
      {#if selectedTask}
        <h4 class="font-semibold mb-2">{selectedTask.title}</h4>
        <p class="text-sm text-muted-foreground mb-4">{selectedTask.description || 'No description'}</p>

        <div class="space-y-2 text-sm mb-4">
          <div><strong>State:</strong> {selectedTask.state_name}</div>
          <div><strong>Priority:</strong> {selectedTask.priority}</div>
          <div><strong>Tags:</strong> {selectedTask.tags?.join(', ') || 'None'}</div>
        </div>

        {#if selectedTask.annotations?.length > 0}
          <div class="mt-4">
            <h5 class="text-sm font-semibold mb-2">Annotations</h5>
            {#each selectedTask.annotations as annotation}
              <div class="border-l-2 border-accent pl-3 mb-3">
                <p class="text-sm">{annotation.body}</p>
                <span class="text-xs text-muted-foreground">{annotation.created_at}</span>
              </div>
            {/each}
          </div>
        {/if}
      {:else}
        <p class="text-muted-foreground">Select a task to view details</p>
      {/if}
    </ScrollArea>
  </Resizable.Pane>
</Resizable.PaneGroup>
```

### Component 4: CommitsTab.svelte (Split-Pane)

```svelte
<!-- assets/svelte/components/tabs/CommitsTab.svelte -->
<script>
  import { Resizable } from '$lib/components/ui/resizable'
  import { ScrollArea } from '$lib/components/ui/scroll-area'

  export let commits = []
  export let live

  let selectedCommit = commits[0] || null
  let diffContent = ''

  async function selectCommit(commit) {
    selectedCommit = commit
    // Fetch diff from API
    const response = await fetch(`/api/commits/${commit.commit_hash}/diff`)
    diffContent = await response.text()
  }
</script>

<Resizable.PaneGroup direction="horizontal">
  <!-- Left: Commit List -->
  <Resizable.Pane defaultSize={40} minSize={25}>
    <ScrollArea class="h-full">
      <div class="p-2 space-y-1">
        {#each commits as commit}
          <button
            class="w-full text-left p-3 rounded hover:bg-accent"
            class:bg-accent={selectedCommit?.commit_hash === commit.commit_hash}
            on:click={() => selectCommit(commit)}
          >
            <div class="font-mono text-xs text-muted-foreground">
              {commit.commit_hash.slice(0, 8)}
            </div>
            <div class="text-sm font-medium mt-1">{commit.commit_message}</div>
            <div class="text-xs text-muted-foreground mt-1">
              {commit.created_at}
            </div>
          </button>
        {/each}
      </div>
    </ScrollArea>
  </Resizable.Pane>

  <Resizable.Handle />

  <!-- Right: Git Diff -->
  <Resizable.Pane defaultSize={60} minSize={40}>
    <ScrollArea class="h-full">
      <pre class="p-4 font-mono text-xs whitespace-pre-wrap">{diffContent || 'Select a commit to view diff'}</pre>
    </ScrollArea>
  </Resizable.Pane>
</Resizable.PaneGroup>
```

### Component 5: LogsTab.svelte

```svelte
<!-- assets/svelte/components/tabs/LogsTab.svelte -->
<script>
  import { ScrollArea } from '$lib/components/ui/scroll-area'
  import { Button } from '$lib/components/ui/button'

  export let logs = []
  export let live

  let filterLevel = 'all'

  $: filteredLogs = filterLevel === 'all'
    ? logs
    : logs.filter(log => log.log_level === filterLevel)
</script>

<div class="h-full flex flex-col">
  <div class="flex gap-2 p-2 border-b">
    <Button variant={filterLevel === 'all' ? 'default' : 'outline'} size="sm" on:click={() => filterLevel = 'all'}>
      All
    </Button>
    <Button variant={filterLevel === 'info' ? 'default' : 'outline'} size="sm" on:click={() => filterLevel = 'info'}>
      Info
    </Button>
    <Button variant={filterLevel === 'error' ? 'default' : 'outline'} size="sm" on:click={() => filterLevel = 'error'}>
      Error
    </Button>
  </div>

  <ScrollArea class="flex-1 p-4">
    {#each filteredLogs as log}
      <div class="mb-3 text-sm">
        <div class="flex items-center gap-2 mb-1">
          <span class="text-xs font-mono text-muted-foreground">{log.timestamp}</span>
          <span class="text-xs px-2 py-0.5 rounded bg-gray-100">{log.log_level}</span>
        </div>
        <p>{log.message}</p>
      </div>
    {/each}
  </ScrollArea>
</div>
```

### Component 6: ContextPanel.svelte

```svelte
<!-- assets/svelte/components/ContextPanel.svelte -->
<script>
  import { ScrollArea } from '$lib/components/ui/scroll-area'
  import { Button } from '$lib/components/ui/button'

  export let sessionContext
  export let notes = []
  export let metrics
  export let live

  let noteBody = ''

  function addNote() {
    if (!noteBody.trim()) return
    live.pushEvent('add_note', { body: noteBody })
    noteBody = ''
  }
</script>

<ScrollArea class="h-full">
  <div class="p-4 space-y-6">
    <!-- Session Context Section -->
    <div>
      <h4 class="text-sm font-semibold mb-2">Session Context</h4>
      {#if sessionContext?.context}
        <div class="text-sm text-muted-foreground whitespace-pre-wrap">
          {sessionContext.context}
        </div>
      {:else}
        <p class="text-sm text-muted-foreground">No context available</p>
      {/if}
    </div>

    <!-- Notes Section -->
    <div>
      <h4 class="text-sm font-semibold mb-2">Notes ({notes.length})</h4>
      <div class="space-y-2 mb-3">
        {#each notes as note}
          <div class="border-l-2 border-accent pl-3 py-1">
            <p class="text-sm">{note.body}</p>
            <span class="text-xs text-muted-foreground">{note.created_at}</span>
          </div>
        {/each}
      </div>

      <div class="space-y-2">
        <textarea
          bind:value={noteBody}
          placeholder="Add a note..."
          class="w-full p-2 text-sm border rounded resize-none"
          rows="3"
        />
        <Button size="sm" on:click={addNote}>
          Add Note
        </Button>
      </div>
    </div>

    <!-- Stats Section -->
    <div>
      <h4 class="text-sm font-semibold mb-2">Stats</h4>
      <dl class="space-y-1 text-sm">
        <div class="flex justify-between">
          <dt class="text-muted-foreground">Tokens:</dt>
          <dd class="font-mono">{metrics?.total_tokens || 0}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-muted-foreground">Cost:</dt>
          <dd class="font-mono">${(metrics?.cost || 0).toFixed(2)}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-muted-foreground">Duration:</dt>
          <dd>{metrics?.duration || '—'}</dd>
        </div>
      </dl>
    </div>
  </div>
</ScrollArea>
```

---

## Phase 5: Make It Resizable and Nicer

### Replace Static Grid with Resizable

Create a top-level `AgentDashboard.svelte` component:

```svelte
<!-- assets/svelte/components/AgentDashboard.svelte -->
<script>
  import { Resizable } from '$lib/components/ui/resizable'
  import SessionsSidebar from './SessionsSidebar.svelte'
  import MainWorkArea from './MainWorkArea.svelte'
  import ContextPanel from './ContextPanel.svelte'

  export let agent
  export let sessions
  export let activeSession
  export let tasks
  export let commits
  export let logs
  export let notes
  export let sessionContext
  export let metrics
  export let activeTab
  export let live
</script>

<Resizable.PaneGroup direction="horizontal" class="h-[calc(100vh-180px)]">
  <!-- Left: Sessions -->
  <Resizable.Pane defaultSize={20} minSize={15} maxSize={30}>
    <SessionsSidebar
      {sessions}
      activeSessionId={activeSession?.id}
      {live}
    />
  </Resizable.Pane>

  <Resizable.Handle />

  <!-- Center: Main Work Area -->
  <Resizable.Pane defaultSize={50} minSize={30}>
    <MainWorkArea
      {activeTab}
      {tasks}
      {commits}
      {logs}
      {live}
    />
  </Resizable.Pane>

  <Resizable.Handle />

  <!-- Right: Context Panel -->
  <Resizable.Pane defaultSize={30} minSize={20} maxSize={40}>
    <ContextPanel
      {sessionContext}
      {notes}
      {metrics}
      {live}
    />
  </Resizable.Pane>
</Resizable.PaneGroup>
```

**Use in LiveView:**
```heex
<.svelte
  name="AgentDashboard"
  props={%{
    agent: @agent,
    sessions: @sessions,
    activeSession: @active_session,
    tasks: @tasks,
    commits: @commits,
    logs: @logs,
    notes: @notes,
    sessionContext: @session_context,
    metrics: @metrics,
    activeTab: @active_tab
  }}
  socket={@socket}
/>
```

### Keyboard Shortcuts

Add to `AgentDashboard.svelte`:

```svelte
<script>
  import { onMount } from 'svelte'

  onMount(() => {
    function handleKeyDown(e) {
      // Tab shortcuts
      if (e.key === 't') {
        live.pushEvent('change_tab', { tab: 'tasks' })
      } else if (e.key === 'c') {
        live.pushEvent('change_tab', { tab: 'commits' })
      } else if (e.key === 'l') {
        live.pushEvent('change_tab', { tab: 'logs' })
      }
      // Session navigation
      else if (e.key === 'j') {
        // Move down in sessions list
        live.pushEvent('navigate_session', { direction: 'down' })
      } else if (e.key === 'k') {
        // Move up in sessions list
        live.pushEvent('navigate_session', { direction: 'up' })
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  })
</script>
```

---

## Phase 6: Data Parity with TUI

### Checklist

**Tasks Tab:**
- ✅ Task list with priority sorting
- ✅ Task details in right pane
- ✅ Annotations visible
- ✅ State badges
- ⬜ Task filtering by state
- ⬜ Task creation/editing

**Commits Tab:**
- ✅ Commit list
- ⬜ Git diff viewer with syntax highlighting
- ⬜ Color coding (green additions, red deletions)
- ⬜ Scrollable diff

**Logs Tab:**
- ✅ Log list with timestamps
- ✅ Level filtering (info/error/all)
- ⬜ Category filtering
- ⬜ Search/filter by text

**Notes:**
- ✅ Visible in context panel
- ✅ Add note functionality
- ⬜ Edit/delete notes
- ⬜ Note attachments

**Session Context:**
- ✅ Markdown rendering
- ⬜ Edit context inline
- ⬜ Context history/versions

**Metrics:**
- ✅ Basic stats (tokens, cost, duration)
- ⬜ Charts/visualizations
- ⬜ Cost breakdown

### Missing from Current Web Implementation

**Must add:**
1. Task annotations query and display
2. Git diff API endpoint (`/api/commits/:hash/diff`)
3. Diff syntax highlighting (Prism.js or similar)
4. Session context loading
5. Notes CRUD operations
6. Metrics calculation (duration from timestamps)

**Nice to have:**
7. Task filtering controls
8. Log search
9. Context editing
10. Charts for token usage over time

---

## Phase 7: Polish

### Performance Optimizations

**Limits:**
- Sessions list: Last 50 sessions (add "Load more" button)
- Logs: Last 100 logs per session
- Commits: All (usually not that many)
- Tasks: All (filtered by session)

**Caching:**
```elixir
# Cache git diffs in ETS
def get_commit_diff_cached(hash) do
  case :ets.lookup(:diff_cache, hash) do
    [{^hash, diff, timestamp}] when timestamp > now() - 300_000 ->
      diff
    _ ->
      diff = Git.show_diff(hash)
      :ets.insert(:diff_cache, {hash, diff, now()})
      diff
  end
end
```

**Throttle updates:**
```elixir
# Don't repaint on every NATS message
def handle_info({:nats_event, _}, socket) do
  Process.send_after(self(), :refresh_data, 500)
  {:noreply, socket}
end

def handle_info(:refresh_data, socket) do
  {:noreply, reload_session_data(socket)}
end
```

### Real-Time Updates (NATS Integration)

**Subscribe to relevant topics in mount:**
```elixir
if connected?(socket) do
  Phoenix.PubSub.subscribe(Eye.PubSub, "agents.#{agent_id}")
  Phoenix.PubSub.subscribe(Eye.PubSub, "sessions.#{active_session.id}")
end
```

**Handle events:**
```elixir
def handle_info({:new_commit, commit}, socket) do
  {:noreply, update(socket, :commits, fn commits -> [commit | commits] end)}
end

def handle_info({:new_log, log}, socket) do
  {:noreply, update(socket, :logs, fn logs -> [log | logs] end)}
end

def handle_info({:task_updated, task}, socket) do
  {:noreply, update_task_in_list(socket, task)}
end
```

---

## Design Constraints

### DO

✅ Keep all session data visible in one screen
✅ Use split-pane for complex data (tasks, commits)
✅ Lazy load per session (not per tab)
✅ Make sessions list prominent and easy to switch
✅ Show metrics prominently in header and context panel
✅ Use keyboard shortcuts for power users
✅ Make it responsive (stack on mobile)

### DO NOT

❌ Add more tabs to center pane (stick to Tasks/Commits/Logs)
❌ Make 3 resizable panes on mobile (stack instead)
❌ Load all sessions' data at once (load active session only)
❌ Repaint entire UI on every event (throttle updates)
❌ Hide critical info in nested menus
❌ Ignore TUI patterns (they work well)

---

## File Structure

```
eye_in_the_sky_web/
├── lib/
│   ├── eye_in_the_sky_web/
│   │   ├── agents.ex                    # Add: get_agent_dashboard_data/1, load_session_data/1
│   │   └── commits.ex                   # Add: get_commit_diff/1
│   └── eye_in_the_sky_web_web/
│       ├── live/
│       │   └── agent_live/
│       │       └── show.ex               # Extend: new assigns, events, data loaders
│       └── controllers/
│           └── api/
│               └── commit_controller.ex  # New: /api/commits/:hash/diff
├── assets/
│   └── svelte/
│       └── components/
│           ├── AgentDashboard.svelte     # Top-level resizable layout
│           ├── SessionsSidebar.svelte    # Left pane
│           ├── MainWorkArea.svelte       # Center pane with tabs
│           ├── ContextPanel.svelte       # Right pane
│           └── tabs/
│               ├── TasksTab.svelte       # Split: list ↔ details
│               ├── CommitsTab.svelte     # Split: list ↔ diff
│               └── LogsTab.svelte        # Simple list + filters
```

---

## Implementation Order

**Do it in this exact order to avoid chaos:**

1. ✅ Lock the layout (this document)
2. Extend `AgentLive.Show` with new assigns and events
3. Build skeleton UI with static CSS grid (3 columns)
4. Create Svelte components one by one (Sessions → Main → Context)
5. Wire components to LiveView with `<.svelte>`
6. Add Resizable to make panes adjustable
7. Add keyboard shortcuts
8. Achieve data parity with TUI (annotations, diffs, etc.)
9. Add real-time updates via NATS
10. Polish visual hierarchy and performance

**Anything else is volunteering for chaos.**

---

**Document Version**: 1.0
**Last Updated**: 2025-11-13
**Status**: Locked - Do not "just add one more tab"
