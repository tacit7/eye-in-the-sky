# Phoenix LiveView + Svelte + shadcn Migration Plan

## Overview

This document outlines the complete plan to create a Phoenix LiveView web application that coexists with the current Go/Bubble Tea TUI, providing a modern web interface for the Eye in the Sky agent monitoring system.

## Architecture Decision

- **Keep Go MCP server** - Handles MCP protocol and existing tools (22 MCP tools remain unchanged)
- **New Phoenix app** - Web UI only, pure presentation layer
- **Communication**: NATS JetStream for real-time event streaming between Go and Phoenix
- **Database**: Migrate existing SQLite database (`~/.config/eye-in-the-sky/eits.db`)
- **Deployment**: TUI + Web coexist - both read from same database, allowing users to choose their interface

## Technology Stack

### Core Frameworks
- **Phoenix**: 1.8.1 (latest stable as of Jan 2025)
- **Phoenix LiveView**: 1.1.14 (latest stable)
- **Svelte**: 5.43.6 (latest with runes system)
- **LiveSvelte**: 0.16.0 (Phoenix + Svelte integration)
- **shadcn-svelte**: 1.0.11 (UI component library)
- **Elixir**: OTP 25+ required
- **Node.js**: v19+ required for SSR

### Database & Messaging
- **Ecto SQLite3** adapter (v0.22.0)
- **NATS JetStream** for event streaming
- **Phoenix PubSub** for LiveView subscriptions

### Key Technology Features

**Phoenix 1.8 Highlights:**
- Built-in Tailwind CSS support
- Dark mode support
- Bandit as default web server
- Improved accessibility primitives

**LiveView 1.1 Highlights:**
- Focus management (`<.focus_wrap>`, `JS.focus`)
- Colocated hooks
- Enhanced accessibility
- Stable 1.x API

**Svelte 5 Highlights:**
- Runes system (`$state`, `$derived`, `$effect`)
- Faster and smaller than Svelte 4
- Improved reactivity

**shadcn-svelte Components:**
- Resizable (for split-pane layouts)
- Tabs, Table, Scroll Area
- Full keyboard navigation
- Accessible by default

## System Architecture

```
┌─────────────────────┐
│   Claude Code       │
│   Instances         │
└──────────┬──────────┘
           │ MCP Protocol (stdio)
           ▼
┌─────────────────────┐     ┌─────────────────────┐
│   Go MCP Server     │────▶│   SQLite Database   │
│   (Existing)        │     │   eits.db           │
└──────────┬──────────┘     └──────────┬──────────┘
           │                           │
           │ NATS JetStream            │ Ecto Queries
           │ Event Streaming           │
           ▼                           ▼
┌─────────────────────┐     ┌─────────────────────┐
│  Phoenix LiveView   │────▶│  Phoenix PubSub     │
│  Web Application    │     │  (Internal)         │
└──────────┬──────────┘     └─────────────────────┘
           │
           │ WebSocket
           ▼
┌─────────────────────┐     ┌─────────────────────┐
│   Browser           │     │   TUI (Existing)    │
│   (New Interface)   │     │   Bubble Tea        │
└─────────────────────┘     └─────────────────────┘
```

## Implementation Phases

### Phase 1: Phoenix Project Setup (Day 1-2)

**Objective**: Create Phoenix project with Ecto connected to existing database

**Tasks:**
1. Generate Phoenix project with SQLite
   ```bash
   mix phx.new eye_in_the_sky_web --database sqlite3
   ```

2. Configure Ecto to use existing database
   ```elixir
   # config/dev.exs
   config :eye_in_the_sky_web, Eye.Repo,
     database: Path.expand("~/.config/eye-in-the-sky/eits.db"),
     journal_mode: :wal,        # WAL for concurrent access
     temp_store: :memory,
     foreign_keys: :on,
     busy_timeout: 2000,
     pool_size: 5
   ```

3. Generate Ecto schemas from existing tables:
   - `agents` - Agent instances
   - `sessions` - Session tracking
   - `projects` - Git repositories
   - `commits` - Git commits
   - `tasks` - Task management
   - `workflow_states` - Task states
   - `actions` - Agent actions
   - `logs` - Session logs
   - `session_context` - Session state
   - `agent_context` - Agent context per project
   - `notes` - Polymorphic notes
   - `session_metrics` - Token usage

4. Create Phoenix contexts (business logic layer):
   - `Agents` context - Agent CRUD, status updates
   - `Sessions` context - Session management
   - `Commits` context - Git commit tracking
   - `Tasks` context - Task lifecycle, workflow
   - `Logs` context - Action and session logs
   - `Metrics` context - Token usage tracking

5. Set up initial routes
   ```elixir
   # lib/eye_web/router.ex
   scope "/", EyeWeb do
     pipe_through :browser

     live "/", AgentLive.Index, :index
     live "/agents/:id", AgentLive.Show, :show
   end
   ```

**Deliverables:**
- Working Phoenix app with Ecto schemas
- Basic LiveView routes defined
- Database connection verified

---

### Phase 2: NATS Integration (Day 3-4)

**Objective**: Enable real-time communication between Go MCP server and Phoenix

**Tasks:**

1. Add NATS client to Phoenix
   ```elixir
   # mix.exs
   {:gnat, "~> 1.8"}
   ```

2. Create NATS consumer GenServer
   ```elixir
   # lib/eye/nats/consumer.ex
   defmodule Eye.NATS.Consumer do
     use GenServer

     def start_link(_opts) do
       GenServer.start_link(__MODULE__, %{}, name: __MODULE__)
     end

     def init(_state) do
       {:ok, conn} = Gnat.start_link(%{host: "localhost", port: 4222})

       # Subscribe to agent event topics
       Gnat.sub(conn, self(), "agents.status.>")
       Gnat.sub(conn, self(), "agents.commits.>")
       Gnat.sub(conn, self(), "agents.tasks.>")
       Gnat.sub(conn, self(), "agents.logs.>")

       {:ok, %{conn: conn}}
     end

     def handle_info({:msg, %{topic: topic, body: body}}, state) do
       # Parse event and broadcast to Phoenix.PubSub
       event = Jason.decode!(body)
       Phoenix.PubSub.broadcast(Eye.PubSub, topic, event)
       {:noreply, state}
     end
   end
   ```

3. Update Go MCP server to publish NATS events
   ```go
   // internal/mcp/nats_publisher.go
   package mcp

   import (
       "encoding/json"
       "github.com/nats-io/nats.go"
   )

   type NATSPublisher struct {
       conn *nats.Conn
   }

   func (p *NATSPublisher) PublishAgentStatus(agentID, status string) error {
       data, _ := json.Marshal(map[string]string{
           "agent_id": agentID,
           "status": status,
       })
       return p.conn.Publish("agents.status." + agentID, data)
   }

   func (p *NATSPublisher) PublishCommit(agentID, hash, message string) error {
       data, _ := json.Marshal(map[string]string{
           "agent_id": agentID,
           "hash": hash,
           "message": message,
       })
       return p.conn.Publish("agents.commits." + agentID, data)
   }
   ```

4. Configure Phoenix PubSub subscriptions in LiveViews
   ```elixir
   # lib/eye_web/live/agent_live/index.ex
   def mount(_params, _session, socket) do
     if connected?(socket) do
       Phoenix.PubSub.subscribe(Eye.PubSub, "agents.status.>")
     end

     {:ok, fetch_agents(socket)}
   end

   def handle_info(%{"agent_id" => id, "status" => status}, socket) do
     {:noreply, update_agent_status(socket, id, status)}
   end
   ```

**Deliverables:**
- NATS consumer running in Phoenix
- Go MCP server publishing events to NATS
- LiveViews receiving real-time updates

---

### Phase 3: LiveSvelte + shadcn Setup (Day 5)

**Objective**: Integrate Svelte components with shadcn UI library

**Tasks:**

1. Install LiveSvelte
   ```elixir
   # mix.exs
   {:live_svelte, "~> 0.16.0"}
   ```

   ```bash
   mix deps.get
   mix live_svelte.setup
   ```

2. Configure custom build script
   ```javascript
   // assets/build.js
   const esbuild = require('esbuild')
   const sveltePlugin = require('esbuild-svelte')

   esbuild.build({
     entryPoints: ['js/app.js'],
     bundle: true,
     outdir: '../priv/static/assets',
     plugins: [sveltePlugin()],
     loader: { '.svg': 'file', '.woff': 'file', '.woff2': 'file' },
   }).catch(() => process.exit(1))
   ```

3. Install shadcn-svelte
   ```bash
   cd assets
   npm install -D shadcn-svelte
   npx shadcn-svelte@latest init
   ```

4. Add shadcn components
   ```bash
   npx shadcn-svelte@latest add resizable
   npx shadcn-svelte@latest add tabs
   npx shadcn-svelte@latest add table
   npx shadcn-svelte@latest add scroll-area
   npx shadcn-svelte@latest add button
   ```

5. Create Svelte component structure
   ```
   assets/
   ├── svelte/
   │   └── components/
   │       ├── AgentTree.svelte
   │       ├── CommitDiff.svelte
   │       ├── TaskList.svelte
   │       ├── LogViewer.svelte
   │       └── ui/                  # shadcn components
   │           ├── resizable/
   │           ├── tabs/
   │           ├── table/
   │           └── scroll-area/
   ```

6. Configure Tailwind to process Svelte files
   ```javascript
   // assets/tailwind.config.js
   module.exports = {
     content: [
       './js/**/*.js',
       './svelte/**/*.svelte',
       '../lib/*_web.ex',
       '../lib/*_web/**/*.*ex',
     ],
     // ... rest of config
   }
   ```

**Deliverables:**
- LiveSvelte integration working
- shadcn-svelte components available
- Build pipeline configured

---

### Phase 4: Agent List View (Day 6-8)

**Objective**: Implement overview page showing all agents with real-time updates

**Tasks:**

1. Create LiveView index page
   ```elixir
   # lib/eye_web/live/agent_live/index.ex
   defmodule EyeWeb.AgentLive.Index do
     use EyeWeb, :live_view

     def mount(_params, _session, socket) do
       if connected?(socket) do
         Phoenix.PubSub.subscribe(Eye.PubSub, "agents.status.>")
       end

       agents = Eye.Agents.list_agents_with_sessions()
       {:ok, assign(socket, agents: agents, selected_index: 0)}
     end

     def handle_event("keydown", %{"key" => "j"}, socket) do
       {:noreply, update(socket, :selected_index, &min(&1 + 1, length(socket.assigns.agents) - 1))}
     end

     def handle_event("keydown", %{"key" => "k"}, socket) do
       {:noreply, update(socket, :selected_index, &max(&1 - 1, 0))}
     end

     def handle_event("keydown", %{"key" => "Enter"}, socket) do
       agent = Enum.at(socket.assigns.agents, socket.assigns.selected_index)
       {:noreply, push_navigate(socket, to: ~p"/agents/#{agent.id}")}
     end
   end
   ```

2. Create Svelte component for hierarchical agent tree
   ```svelte
   <!-- assets/svelte/components/AgentTree.svelte -->
   <script>
     import { Table } from '$lib/components/ui/table'

     export let agents = []
     export let selectedIndex = 0

     function getStatusColor(status) {
       const colors = {
         active: 'text-green-500',
         working: 'text-yellow-500',
         idle: 'text-blue-500',
         completed: 'text-gray-500',
         failed: 'text-red-500'
       }
       return colors[status] || 'text-gray-500'
     }

     function isSubagent(agent) {
       return agent.parent_agent_id !== null
     }
   </script>

   <Table.Root>
     <Table.Header>
       <Table.Row>
         <Table.Head>Status</Table.Head>
         <Table.Head>Agent ID</Table.Head>
         <Table.Head>Description</Table.Head>
         <Table.Head>Project</Table.Head>
         <Table.Head>Last Update</Table.Head>
       </Table.Row>
     </Table.Header>
     <Table.Body>
       {#each agents as agent, i}
         <Table.Row class={i === selectedIndex ? 'bg-accent' : ''}>
           <Table.Cell>
             <span class={getStatusColor(agent.status)}>●</span>
             {#if isSubagent(agent)}
               <span class="text-green-500 ml-2">│</span>
             {/if}
           </Table.Cell>
           <Table.Cell class="font-mono text-sm">{agent.id.slice(0, 8)}</Table.Cell>
           <Table.Cell>{agent.description}</Table.Cell>
           <Table.Cell>{agent.project_name || '-'}</Table.Cell>
           <Table.Cell class="text-sm text-muted-foreground">
             {new Date(agent.updated_at).toLocaleString()}
           </Table.Cell>
         </Table.Row>
       {/each}
     </Table.Body>
   </Table.Root>
   ```

3. Implement keyboard navigation
   ```heex
   <!-- lib/eye_web/live/agent_live/index.html.heex -->
   <div phx-window-keydown="keydown" phx-throttle="100">
     <.svelte
       name="AgentTree"
       props={%{agents: @agents, selectedIndex: @selected_index}}
       socket={@socket}
     />
   </div>
   ```

4. Add real-time status updates
   ```elixir
   def handle_info(%{"agent_id" => id, "status" => status}, socket) do
     agents = Enum.map(socket.assigns.agents, fn agent ->
       if agent.id == id do
         %{agent | status: status, updated_at: DateTime.utc_now()}
       else
         agent
       end
     end)

     {:noreply, assign(socket, agents: agents)}
   end
   ```

5. Style with status color coding
   - Green: Active
   - Yellow: Working
   - Blue: Idle
   - Gray: Completed
   - Red: Failed

6. Add hierarchical display with green │ for subagents

**Deliverables:**
- Working agent list view
- Real-time status updates
- Keyboard navigation (j/k/Enter)
- Visual indicators for agent status
- Parent-child relationship display

---

### Phase 5: Agent Detail View (Day 9-14)

**Objective**: Implement tabbed detail view with 7 tabs matching TUI functionality

**Tasks:**

1. Create LiveView show page
   ```elixir
   # lib/eye_web/live/agent_live/show.ex
   defmodule EyeWeb.AgentLive.Show do
     use EyeWeb, :live_view

     def mount(%{"id" => id}, _session, socket) do
       if connected?(socket) do
         Phoenix.PubSub.subscribe(Eye.PubSub, "agents.#{id}")
       end

       agent = Eye.Agents.get_agent!(id)

       socket =
         socket
         |> assign(agent: agent, active_tab: :overview)
         |> load_tab_data(:overview)

       {:ok, socket}
     end

     def handle_event("switch_tab", %{"tab" => tab}, socket) do
       tab_atom = String.to_existing_atom(tab)
       {:noreply, socket |> assign(active_tab: tab_atom) |> load_tab_data(tab_atom)}
     end

     defp load_tab_data(socket, :overview), do: assign(socket, :overview_data, fetch_overview(socket.assigns.agent))
     defp load_tab_data(socket, :tasks), do: assign(socket, :tasks, fetch_tasks(socket.assigns.agent))
     defp load_tab_data(socket, :commits), do: assign(socket, :commits, fetch_commits(socket.assigns.agent))
     # ... other tabs
   end
   ```

2. Implement tab navigation with shadcn Tabs
   ```svelte
   <!-- assets/svelte/components/AgentDetailTabs.svelte -->
   <script>
     import { Tabs } from '$lib/components/ui/tabs'
     export let activeTab = 'overview'
     export let live

     const tabs = [
       { id: 'overview', label: 'Overview', key: 'o' },
       { id: 'tasks', label: 'Tasks', key: 't' },
       { id: 'logs', label: 'Logs', key: 'l' },
       { id: 'commits', label: 'Commits', key: 'c' },
       { id: 'notes', label: 'Notes', key: 'n' },
       { id: 'sessions', label: 'Sessions', key: 's' },
       { id: 'actions', label: 'Actions', key: 'a' }
     ]

     function handleTabChange(tabId) {
       live.pushEvent('switch_tab', { tab: tabId })
     }
   </script>

   <Tabs.Root value={activeTab} onValueChange={handleTabChange}>
     <Tabs.List>
       {#each tabs as tab}
         <Tabs.Trigger value={tab.id}>
           [{tab.key.toUpperCase()}] {tab.label}
         </Tabs.Trigger>
       {/each}
     </Tabs.List>
   </Tabs.Root>
   ```

3. **Overview Tab** - Agent metadata and recent activity
   ```elixir
   defp fetch_overview(agent) do
     %{
       agent: agent,
       recent_commits: Eye.Commits.list_recent_commits(agent.id, limit: 5),
       recent_logs: Eye.Logs.list_recent_logs(agent.id, limit: 10),
       notes_summary: Eye.Notes.count_notes(agent.id)
     }
   end
   ```

4. **Tasks Tab** - Split-pane with task list + details
   ```svelte
   <!-- assets/svelte/components/TaskList.svelte -->
   <script>
     import { Resizable } from '$lib/components/ui/resizable'
     import { ScrollArea } from '$lib/components/ui/scroll-area'

     export let tasks = []
     let selectedTask = null

     function priorityColor(priority) {
       if (priority >= 70) return 'text-red-500'
       if (priority >= 40) return 'text-yellow-500'
       return 'text-blue-500'
     }
   </script>

   <Resizable.PaneGroup direction="horizontal">
     <Resizable.Pane defaultSize={50}>
       <ScrollArea class="h-full">
         <div class="space-y-2 p-4">
           {#each tasks as task}
             <button
               class="w-full text-left p-2 hover:bg-accent rounded"
               class:bg-accent={selectedTask?.id === task.id}
               on:click={() => selectedTask = task}
             >
               <div class="flex items-center gap-2">
                 <span class={priorityColor(task.priority)}>●</span>
                 <span class="font-medium">{task.title}</span>
               </div>
               <div class="text-sm text-muted-foreground ml-6">
                 {task.state_name}
               </div>
             </button>
           {/each}
         </div>
       </ScrollArea>
     </Resizable.Pane>

     <Resizable.Handle />

     <Resizable.Pane defaultSize={50}>
       <ScrollArea class="h-full p-4">
         {#if selectedTask}
           <h3 class="font-bold mb-2">{selectedTask.title}</h3>
           <p class="text-sm text-muted-foreground mb-4">{selectedTask.description}</p>
           <div class="space-y-2">
             <div><strong>State:</strong> {selectedTask.state_name}</div>
             <div><strong>Priority:</strong> {selectedTask.priority}</div>
             <div><strong>Tags:</strong> {selectedTask.tags?.join(', ') || 'None'}</div>
             <div><strong>Created:</strong> {new Date(selectedTask.created_at).toLocaleString()}</div>
           </div>

           {#if selectedTask.notes?.length > 0}
             <div class="mt-4">
               <h4 class="font-semibold mb-2">Notes:</h4>
               {#each selectedTask.notes as note}
                 <div class="border-l-2 border-accent pl-4 mb-2">
                   <p class="text-sm">{note.body}</p>
                   <span class="text-xs text-muted-foreground">{new Date(note.created_at).toLocaleString()}</span>
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

5. **Commits Tab** - Split-pane with commit list + git diff
   ```svelte
   <!-- assets/svelte/components/CommitDiff.svelte -->
   <script>
     import { Resizable } from '$lib/components/ui/resizable'
     import { ScrollArea } from '$lib/components/ui/scroll-area'

     export let commits = []
     let selectedCommit = null
     let diffContent = ''

     async function selectCommit(commit) {
       selectedCommit = commit
       // Fetch diff from Phoenix
       const response = await fetch(`/api/commits/${commit.hash}/diff`)
       diffContent = await response.text()
     }
   </script>

   <Resizable.PaneGroup direction="horizontal">
     <Resizable.Pane defaultSize={40}>
       <ScrollArea class="h-full">
         {#each commits as commit}
           <button
             class="w-full text-left p-3 border-b hover:bg-accent"
             class:bg-accent={selectedCommit?.hash === commit.hash}
             on:click={() => selectCommit(commit)}
           >
             <div class="font-mono text-xs text-muted-foreground">{commit.hash.slice(0, 8)}</div>
             <div class="font-medium">{commit.message}</div>
             <div class="text-xs text-muted-foreground">{new Date(commit.created_at).toLocaleString()}</div>
           </button>
         {/each}
       </ScrollArea>
     </Resizable.Pane>

     <Resizable.Handle />

     <Resizable.Pane defaultSize={60}>
       <ScrollArea class="h-full">
         <pre class="p-4 font-mono text-xs">{diffContent || 'Select a commit to view diff'}</pre>
       </ScrollArea>
     </Resizable.Pane>
   </Resizable.PaneGroup>
   ```

6. **Logs Tab** - Session logs with filtering
   ```elixir
   defp fetch_logs(agent) do
     Eye.Logs.list_session_logs(agent.id)
   end
   ```

7. **Notes Tab** - Session notes display
   ```elixir
   defp fetch_notes(agent) do
     Eye.Notes.list_notes(parent_type: "agents", parent_id: agent.id)
   end
   ```

8. **Sessions Tab** - Session history
   ```elixir
   defp fetch_sessions(agent) do
     Eye.Sessions.list_sessions_for_agent(agent.id)
   end
   ```

9. **Actions Tab** - Action log with timestamps
   ```elixir
   defp fetch_actions(agent) do
     Eye.Actions.list_actions(agent.id)
   end
   ```

10. Implement keyboard shortcuts for tabs (o/t/l/c/n/s/a)
    ```heex
    <div phx-window-keydown="tab_shortcut" phx-throttle="100">
      <!-- Tab content -->
    </div>
    ```

    ```elixir
    def handle_event("tab_shortcut", %{"key" => key}, socket) do
      tab = case key do
        "o" -> :overview
        "t" -> :tasks
        "l" -> :logs
        "c" -> :commits
        "n" -> :notes
        "s" -> :sessions
        "a" -> :actions
        _ -> socket.assigns.active_tab
      end

      {:noreply, socket |> assign(active_tab: tab) |> load_tab_data(tab)}
    end
    ```

**Deliverables:**
- Working agent detail page
- All 7 tabs functional
- Split-pane layouts for Tasks and Commits
- Keyboard shortcuts working
- Real-time updates per tab

---

### Phase 6: Feature Parity (Day 15-18)

**Objective**: Match all TUI features and polish the UI

**Tasks:**

1. Add horizontal scroll for wide content (h/l keys)
   ```elixir
   def handle_event("keydown", %{"key" => "h"}, socket) do
     {:noreply, push_event(socket, "scroll", %{direction: "left", amount: 100})}
   end

   def handle_event("keydown", %{"key" => "l"}, socket) do
     {:noreply, push_event(socket, "scroll", %{direction: "right", amount: 100})}
   end
   ```

2. Implement git diff syntax highlighting
   - Use Prism.js or similar library
   - Highlight additions (green), deletions (red), context (gray)
   - Line numbers on left side

3. Add task priority sorting and filtering
   ```elixir
   # Sort: High > Medium > Low, deleted tasks at bottom
   defp sort_tasks(tasks) do
     tasks
     |> Enum.sort_by(fn task ->
       {task.deleted_at != nil, -task.priority, task.created_at}
     end)
   end
   ```

4. Implement log type filtering
   ```svelte
   <script>
     let logTypes = ['all', 'info', 'warn', 'error']
     let selectedType = 'all'

     $: filteredLogs = selectedType === 'all'
       ? logs
       : logs.filter(log => log.type === selectedType)
   </script>

   <div class="flex gap-2 mb-4">
     {#each logTypes as type}
       <Button
         variant={selectedType === type ? 'default' : 'outline'}
         on:click={() => selectedType = type}
       >
         {type}
       </Button>
     {/each}
   </div>
   ```

5. Add real-time updates to all tabs
   ```elixir
   def handle_info({:new_commit, commit}, socket) do
     if socket.assigns.active_tab == :commits do
       {:noreply, update(socket, :commits, fn commits -> [commit | commits] end)}
     else
       {:noreply, socket}
     end
   end
   ```

6. Implement loading states
   ```svelte
   <script>
     export let loading = false
   </script>

   {#if loading}
     <div class="flex items-center justify-center h-full">
       <Spinner />
       <span class="ml-2">Loading...</span>
     </div>
   {:else}
     <!-- Content -->
   {/if}
   ```

7. Add error handling
   ```elixir
   def handle_info({:error, reason}, socket) do
     {:noreply, put_flash(socket, :error, "Failed to load data: #{reason}")}
   end
   ```

8. Accessibility improvements
   - ARIA labels for all interactive elements
   - Focus management with `<.focus_wrap>`
   - Keyboard navigation announcement via screen readers
   - Semantic HTML structure

**Deliverables:**
- Full feature parity with TUI
- Syntax highlighting for diffs
- Filtering and sorting working
- Loading states and error handling
- Accessibility compliant

---

### Phase 7: Testing & Polish (Day 19-21)

**Objective**: Comprehensive testing and quality assurance

**Tasks:**

1. Write LiveView integration tests
   ```elixir
   # test/eye_web/live/agent_live_test.exs
   defmodule EyeWeb.AgentLiveTest do
     use EyeWeb.ConnCase
     import Phoenix.LiveViewTest

     test "displays agent list", %{conn: conn} do
       agent = agent_fixture()

       {:ok, view, html} = live(conn, ~p"/")

       assert html =~ agent.id
       assert html =~ agent.description
     end

     test "navigates to agent detail", %{conn: conn} do
       agent = agent_fixture()

       {:ok, view, _html} = live(conn, ~p"/")

       view
       |> element("button[data-agent-id='#{agent.id}']")
       |> render_click()

       assert_redirected(view, ~p"/agents/#{agent.id}")
     end
   end
   ```

2. Test NATS event handling
   ```elixir
   test "receives real-time status updates" do
     agent = agent_fixture()

     {:ok, view, _html} = live(conn, ~p"/")

     # Simulate NATS event
     Phoenix.PubSub.broadcast(Eye.PubSub, "agents.status.#{agent.id}", %{
       "agent_id" => agent.id,
       "status" => "working"
     })

     assert render(view) =~ "working"
   end
   ```

3. Test PubSub subscriptions
   ```elixir
   test "subscribes to agent updates on mount" do
     agent = agent_fixture()

     {:ok, view, _html} = live(conn, ~p"/agents/#{agent.id}")

     # Verify subscription
     assert Phoenix.PubSub.subscribers(Eye.PubSub, "agents.#{agent.id}") != []
   end
   ```

4. Performance testing with multiple concurrent agents
   - Simulate 50+ agents with frequent updates
   - Measure render times (should be < 100ms)
   - Monitor memory usage
   - Check WebSocket message throughput

5. Browser testing
   - Chrome: Keyboard shortcuts, rendering
   - Firefox: WebSocket stability
   - Safari: Accessibility features
   - Mobile browsers: Responsive layout

6. Keyboard navigation testing
   - Test all shortcuts (j/k/o/t/l/c/n/s/a/h/l)
   - Verify focus management
   - Test Tab/Shift+Tab navigation
   - Ensure no keyboard traps

7. Accessibility audit
   - Run axe DevTools
   - Test with screen reader (VoiceOver on macOS)
   - Verify keyboard-only navigation
   - Check color contrast ratios
   - Validate ARIA attributes

**Deliverables:**
- Comprehensive test suite passing
- Performance benchmarks documented
- Browser compatibility verified
- Accessibility audit complete

---

### Phase 8: Production Setup (Day 22-25)

**Objective**: Deploy Phoenix app to production alongside existing TUI

**Tasks:**

1. Configure production release
   ```elixir
   # config/prod.exs
   config :eye_in_the_sky_web, EyeWeb.Endpoint,
     url: [host: "localhost", port: 4000],
     cache_static_manifest: "priv/static/cache_manifest.json",
     server: true

   config :eye_in_the_sky_web, Eye.Repo,
     database: System.get_env("DATABASE_PATH") || Path.expand("~/.config/eye-in-the-sky/eits.db"),
     journal_mode: :wal,
     pool_size: 10
   ```

2. Create release configuration
   ```elixir
   # mix.exs
   def project do
     [
       # ...
       releases: [
         eye_web: [
           include_executables_for: [:unix],
           applications: [runtime_tools: :permanent]
         ]
       ]
     ]
   end
   ```

3. Build production release
   ```bash
   mix deps.get --only prod
   MIX_ENV=prod mix compile
   MIX_ENV=prod mix assets.deploy
   MIX_ENV=prod mix release
   ```

4. Create systemd service
   ```ini
   # /etc/systemd/system/eye-web.service
   [Unit]
   Description=Eye in the Sky Web UI
   After=network.target nats.service

   [Service]
   Type=exec
   User=eye
   Group=eye
   WorkingDirectory=/opt/eye-in-the-sky-web
   Environment="DATABASE_PATH=/home/eye/.config/eye-in-the-sky/eits.db"
   Environment="SECRET_KEY_BASE=<generated-secret>"
   Environment="PHX_HOST=localhost"
   Environment="PORT=4000"
   ExecStart=/opt/eye-in-the-sky-web/bin/eye_web start
   Restart=on-failure
   RestartSec=5
   StandardOutput=journal
   StandardError=journal

   [Install]
   WantedBy=multi-user.target
   ```

5. Configure reverse proxy (nginx)
   ```nginx
   # /etc/nginx/sites-available/eye-web
   upstream phoenix {
     server 127.0.0.1:4000;
   }

   server {
     listen 80;
     server_name eye.yourdomain.com;

     location / {
       proxy_pass http://phoenix;
       proxy_http_version 1.1;
       proxy_set_header Upgrade $http_upgrade;
       proxy_set_header Connection "upgrade";
       proxy_set_header Host $host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_set_header X-Forwarded-Proto $scheme;
     }

     location ~* ^/assets/ {
       root /opt/eye-in-the-sky-web/lib/eye_web-0.1.0/priv/static;
       expires 1y;
       add_header Cache-Control "public, immutable";
     }
   }
   ```

6. Set up SSL with Let's Encrypt
   ```bash
   sudo certbot --nginx -d eye.yourdomain.com
   ```

7. Production NATS configuration
   ```bash
   # Ensure NATS is running
   systemctl status nats

   # Configure NATS JetStream persistence
   # /etc/nats/nats-server.conf
   jetstream {
     store_dir: "/var/lib/nats"
     max_memory_store: 1GB
     max_file_store: 10GB
   }
   ```

8. Deployment documentation
   ```markdown
   # Deployment Guide

   ## Prerequisites
   - Elixir 1.14+ and Erlang/OTP 25+
   - Node.js 19+ for asset compilation
   - NATS server with JetStream enabled
   - Nginx or Caddy for reverse proxy

   ## Installation Steps
   1. Clone repository
   2. Build production release
   3. Copy to /opt/eye-in-the-sky-web
   4. Configure environment variables
   5. Set up systemd service
   6. Configure nginx
   7. Enable and start services

   ## Monitoring
   - Logs: journalctl -u eye-web -f
   - Metrics: Phoenix LiveDashboard at /dashboard
   - Health check: curl http://localhost:4000/api/health
   ```

9. Update main README
   ```markdown
   # Eye in the Sky

   ## Interfaces

   ### TUI (Terminal UI)
   ```bash
   go run cmd/eye-ui/main.go
   ```

   ### Web UI (Phoenix LiveView)
   ```bash
   cd web
   mix phx.server
   # Visit http://localhost:4000
   ```

   ## Architecture
   - Go MCP Server: Handles Claude Code MCP protocol
   - TUI: Terminal interface (Bubble Tea)
   - Web UI: Browser interface (Phoenix LiveView + Svelte)
   - Database: Shared SQLite (~/.config/eye-in-the-sky/eits.db)
   - Messaging: NATS JetStream for real-time events
   ```

**Deliverables:**
- Production-ready Phoenix release
- Systemd service configured
- Reverse proxy set up with SSL
- Documentation complete
- Both TUI and Web UI working simultaneously

---

## Project File Structure

```
eye-in-the-sky/
├── cmd/
│   ├── server/main.go              # Go MCP server (existing)
│   └── eye-ui/main.go              # TUI (existing)
├── internal/                        # Go packages (existing)
│   ├── mcp/
│   │   ├── server.go
│   │   ├── nats_tools.go
│   │   └── nats_publisher.go       # NEW: Publish events to NATS
│   ├── ui/
│   ├── database/
│   └── utils/
├── web/                             # NEW: Phoenix application
│   ├── mix.exs
│   ├── mix.lock
│   ├── README.md
│   ├── config/
│   │   ├── config.exs
│   │   ├── dev.exs
│   │   ├── test.exs
│   │   └── prod.exs
│   ├── lib/
│   │   ├── eye_web.ex
│   │   ├── eye_web/
│   │   │   ├── application.ex
│   │   │   ├── endpoint.ex
│   │   │   ├── router.ex
│   │   │   ├── telemetry.ex
│   │   │   ├── components/
│   │   │   │   └── core_components.ex
│   │   │   ├── live/
│   │   │   │   ├── agent_live/
│   │   │   │   │   ├── index.ex           # Agent list view
│   │   │   │   │   ├── show.ex            # Agent detail view
│   │   │   │   │   └── components.ex      # Shared LiveView components
│   │   │   │   └── dashboard_live.ex
│   │   │   ├── channels/
│   │   │   │   └── agent_channel.ex
│   │   │   └── controllers/
│   │   │       └── api/
│   │   │           └── commit_controller.ex  # API for git diffs
│   │   └── eye/
│   │       ├── application.ex
│   │       ├── repo.ex
│   │       ├── nats/
│   │       │   └── consumer.ex            # NATS event consumer
│   │       ├── agents/
│   │       │   ├── agent.ex               # Schema
│   │       │   ├── session.ex             # Schema
│   │       │   └── context.ex             # Business logic
│   │       ├── commits/
│   │       │   ├── commit.ex              # Schema
│   │       │   └── context.ex
│   │       ├── tasks/
│   │       │   ├── task.ex                # Schema
│   │       │   ├── workflow_state.ex      # Schema
│   │       │   └── context.ex
│   │       ├── logs/
│   │       │   ├── log.ex                 # Schema
│   │       │   ├── session_log.ex         # Schema
│   │       │   └── context.ex
│   │       └── metrics/
│   │           ├── session_metric.ex      # Schema
│   │           └── context.ex
│   ├── assets/
│   │   ├── js/
│   │   │   ├── app.js
│   │   │   └── hooks/                     # LiveView hooks
│   │   │       └── scroll.js
│   │   ├── svelte/
│   │   │   ├── components/
│   │   │   │   ├── AgentTree.svelte
│   │   │   │   ├── AgentDetailTabs.svelte
│   │   │   │   ├── CommitDiff.svelte
│   │   │   │   ├── TaskList.svelte
│   │   │   │   ├── LogViewer.svelte
│   │   │   │   └── ui/                    # shadcn-svelte components
│   │   │   │       ├── resizable/
│   │   │   │       │   ├── index.ts
│   │   │   │       │   └── resizable.svelte
│   │   │   │       ├── tabs/
│   │   │   │       ├── table/
│   │   │   │       ├── scroll-area/
│   │   │   │       └── button/
│   │   │   └── _build/                    # Gitignored
│   │   ├── css/
│   │   │   └── app.css
│   │   ├── build.js                       # Custom esbuild script
│   │   ├── package.json
│   │   ├── package-lock.json
│   │   └── tailwind.config.js
│   ├── priv/
│   │   ├── static/                        # Compiled assets
│   │   └── repo/
│   │       └── migrations/                # Empty (using existing DB)
│   └── test/
│       ├── eye_web/
│       │   └── live/
│       │       └── agent_live_test.exs
│       └── test_helper.exs
├── docs/
│   └── PHOENIX_MIGRATION_PLAN.md          # This document
└── README.md                               # Updated with web UI info
```

## Technical Configuration Details

### Ecto SQLite Configuration

```elixir
# config/dev.exs and config/prod.exs
config :eye_in_the_sky_web, Eye.Repo,
  database: Path.expand("~/.config/eye-in-the-sky/eits.db"),

  # WAL mode enables concurrent reads while writing
  journal_mode: :wal,

  # Store temp tables in memory for better performance
  temp_store: :memory,

  # Enforce foreign key constraints
  foreign_keys: :on,

  # Timeout for acquiring locks (2 seconds)
  busy_timeout: 2000,

  # Connection pool size
  pool_size: 5,  # dev
  pool_size: 10  # prod
```

### NATS Topic Structure

```
agents.status.<agent_id>       # Agent status changes
agents.commits.<agent_id>      # New git commits
agents.tasks.<agent_id>        # Task updates
agents.logs.<agent_id>         # New log entries
agents.created                 # New agent registered
agents.deleted.<agent_id>      # Agent removed
```

### Phoenix PubSub Topics

```
"agents.status.>"              # All status updates (subscribe in index)
"agents.<agent_id>"            # Specific agent updates (subscribe in show)
"commits.<agent_id>"           # Commit updates
"tasks.<agent_id>"             # Task updates
"logs.<agent_id>"              # Log updates
```

### LiveView Performance Optimizations

```elixir
# Use temporary assigns for large lists
def mount(_params, _session, socket) do
  {:ok, assign(socket, agents: fetch_agents()) |> temporary_assigns([agents: []])}
end

# Debounce rapid updates
def handle_info({:agent_updated, _agent}, socket) do
  Process.send_after(self(), :refresh_agents, 500)
  {:noreply, socket}
end

def handle_info(:refresh_agents, socket) do
  {:noreply, assign(socket, agents: fetch_agents())}
end

# Use ETS for caching frequently accessed data
def fetch_agent_cached(id) do
  case :ets.lookup(:agent_cache, id) do
    [{^id, agent, timestamp}] when timestamp > now() - 5_000 ->
      agent
    _ ->
      agent = Eye.Agents.get_agent!(id)
      :ets.insert(:agent_cache, {id, agent, now()})
      agent
  end
end
```

## Key Technical Decisions

### 1. Why Keep Go MCP Server?
- **Mature implementation**: Go MCP server is stable and working
- **Official SDK**: Go SDK is official and well-maintained
- **No Elixir MCP SDK**: Would require custom implementation
- **Separation of concerns**: Phoenix focuses on presentation
- **Lower risk**: No need to rewrite critical MCP protocol handling

### 2. Why NATS JetStream?
- **Already in go.mod**: Infrastructure already present
- **Persistent streams**: Messages aren't lost if Phoenix is down
- **Scalability**: Can add multiple Phoenix nodes easily
- **Decoupling**: Go and Phoenix are loosely coupled
- **Message replay**: Can replay events for debugging

### 3. Why LiveView + Svelte (not pure LiveView)?
- **Complex client interactions**: Svelte handles rich UI better
- **Keyboard shortcuts**: Svelte can manage complex key handling
- **Split-pane layouts**: shadcn-svelte Resizable component
- **Component ecosystem**: Access to shadcn-svelte components
- **Server-side state**: LiveView still manages application state

### 4. Why SQLite (not PostgreSQL)?
- **Existing database**: Migration is simpler
- **Single-node deployment**: Not building distributed system yet
- **WAL mode**: Supports concurrent reads during writes
- **Simplicity**: No additional database server to manage
- **Future migration path**: Can move to Postgres if needed

### 5. Why Coexist with TUI?
- **User choice**: Some users prefer terminal, others prefer web
- **Gradual adoption**: Users can try web UI without losing TUI
- **Different use cases**: TUI for quick checks, web for detailed analysis
- **Lower risk**: TUI remains as fallback if web has issues
- **Shared database**: Both interfaces see same data

## Success Criteria

### Functional Requirements
✅ Feature parity with TUI (all 7 tabs functional)
✅ Real-time updates via NATS (< 500ms latency)
✅ Keyboard navigation working in browser
✅ Split-pane layouts with resizable handles
✅ Agent list with hierarchical display
✅ Git diff viewer with syntax highlighting
✅ Task management with priority sorting
✅ Log filtering and search

### Non-Functional Requirements
✅ Page load time < 2 seconds
✅ LiveView render time < 100ms
✅ Support 50+ concurrent agents
✅ WebSocket reconnection handling
✅ Mobile-responsive layout
✅ WCAG 2.1 Level AA accessibility
✅ Browser support: Chrome, Firefox, Safari (latest 2 versions)

### Operational Requirements
✅ TUI continues to work alongside web UI
✅ No changes required to Go MCP server core
✅ Production deployment with systemd
✅ SSL/TLS enabled
✅ Health check endpoint
✅ Comprehensive test coverage (>80%)
✅ Documentation complete

## Risk Mitigation

### Risk: LiveSvelte + shadcn-svelte Integration Issues
**Mitigation**: Start with simple Svelte component integration before adding shadcn. Test each component individually. Have fallback plan to use SaladUI (pure LiveView components) if integration fails.

### Risk: NATS Message Volume Overwhelming Phoenix
**Mitigation**: Implement rate limiting and debouncing in NATS consumer. Use ETS caching for frequently accessed data. Monitor WebSocket message rate and implement backpressure if needed.

### Risk: SQLite Database Locking with Concurrent Access
**Mitigation**: Enable WAL mode (already configured). Use IMMEDIATE transaction mode for balanced read/write. Monitor for "database is locked" errors. Have migration path to PostgreSQL documented.

### Risk: Keyboard Shortcuts Conflicting with Browser
**Mitigation**: Use key combinations (Ctrl+key) for critical shortcuts. Provide click-based alternatives for all keyboard actions. Test across browsers for key capture issues. Document any browser-specific limitations.

### Risk: Performance Degradation with Many Agents
**Mitigation**: Implement pagination for agent list (default 50 per page). Use temporary assigns for lists. Add ETS caching layer. Monitor memory usage and implement limits if needed.

## Timeline Summary

| Phase | Days | Description |
|-------|------|-------------|
| Phase 1 | 2 | Phoenix project setup, Ecto schemas, contexts |
| Phase 2 | 2 | NATS integration, PubSub configuration |
| Phase 3 | 1 | LiveSvelte + shadcn-svelte setup |
| Phase 4 | 3 | Agent list view with real-time updates |
| Phase 5 | 6 | Agent detail view with 7 tabs |
| Phase 6 | 4 | Feature parity, syntax highlighting, polish |
| Phase 7 | 3 | Testing, accessibility, performance |
| Phase 8 | 4 | Production deployment, documentation |
| **Total** | **25 days** | **Approximately 5 weeks** |

## Next Steps

1. **Confirm architecture decision** with team
2. **Set up development environment** (Elixir, Node.js, NATS)
3. **Create Phoenix project** and configure Ecto
4. **Generate schemas** from existing database
5. **Begin Phase 1 implementation**

## Additional Resources

- [Phoenix LiveView Documentation](https://hexdocs.pm/phoenix_live_view/)
- [LiveSvelte GitHub](https://github.com/woutdp/live_svelte)
- [shadcn-svelte Documentation](https://www.shadcn-svelte.com/)
- [Ecto SQLite3 Adapter](https://hexdocs.pm/ecto_sqlite3/)
- [NATS JetStream Elixir Client (gnat)](https://github.com/nats-io/nats.ex)

---

**Document Version**: 1.0
**Last Updated**: 2025-01-13
**Author**: Claude Code
**Status**: Ready for Implementation
