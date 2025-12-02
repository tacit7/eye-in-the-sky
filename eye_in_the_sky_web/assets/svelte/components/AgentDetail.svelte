<script>
  import { onMount, onDestroy } from "svelte"
  import TasksTab from "./tabs/TasksTab.svelte"
  import CommitsTab from "./tabs/CommitsTab.svelte"
  import LogsTab from "./tabs/LogsTab.svelte"
  import NotesTab from "./tabs/NotesTab.svelte"
  import MessagesTab from "./tabs/MessagesTab.svelte"
  import { parseDateLike, relativeFrom, elapsedTime, shortId } from "../utils/datetime.js"
  import { emptyStateStyle } from "../utils/styles.js"

  function statusToBadgeVariant(status) {
    const map = {
      active: 'badge-success',
      working: 'badge-warning',
      idle: 'badge-info',
      stale: 'badge-warning badge-outline',
      completed: 'badge-ghost',
      failed: 'badge-error'
    }
    return map[status] || 'badge-ghost'
  }

  export let header
  export let activeTab
  export let counts
  export let tasks
  export let commits
  export let logs
  export let context
  export let notes
  export let messages
  export let live

  const tabs = [
    { key: "tasks", label: "Tasks", countKey: "tasks" },
    { key: "commits", label: "Commits", countKey: "commits" },
    { key: "logs", label: "Logs", countKey: "logs" },
    { key: "context", label: "Context", countKey: null },
    { key: "notes", label: "Notes", countKey: "notes" },
    { key: "messages", label: "Messages", countKey: "messages" },
  ]

  function countFor(key) {
    return counts?.[key] || 0
  }

  let tick = 0
  let intervalId

  onMount(() => {
    // Only start the interval if status is active
    if (header.status === 'active') {
      intervalId = setInterval(() => tick++, 1000)
    }
  })

  onDestroy(() => {
    if (intervalId) {
      clearInterval(intervalId)
    }
  })

  function changeTab(tab) {
    live.pushEvent("change_tab", { tab })
  }

  function copySessionId() {
    navigator.clipboard.writeText(header.session_id)
  }
</script>

<div class="min-h-screen bg-base-200">
  <div class="mx-auto max-w-6xl px-6 py-6">
    <div class="card bg-base-100 shadow-sm">
      <div class="card-body">
        <!-- Header -->
        <div class="bg-base-100/80 backdrop-blur sticky top-0 z-30 border-b border-base-300 -m-6 mb-0 p-6">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <a href="/" class="inline-flex items-center gap-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100">
              <span>←</span>
              <span>Back to Agents</span>
            </a>

            <div class="mt-3 flex flex-wrap items-center gap-2">
              <h1 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">{header.agent_type}</h1>

              <span class="badge badge-outline font-mono">
                {shortId(header.agent_id)}
              </span>

              <span class="badge {statusToBadgeVariant(header.status)}" aria-label={`Status: ${header.status}`}>
                {header.status}
              </span>
            </div>

            <!-- Meta chips -->
            <div class="mt-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2 text-xs sm:text-sm">
              <div class="bg-base-200 rounded-lg px-3 py-1.5">
                <span class="text-base-content/70 font-medium">Session</span>
                <span class="ml-2 font-mono font-semibold">#{shortId(header.session_id)}</span>
              </div>

              <div class="bg-base-200 rounded-lg px-3 py-1.5">
                <span class="text-base-content/70 font-medium">Project</span>
                <span class="ml-2 font-semibold">{header.project ?? "Unassigned"}</span>
              </div>

              {#if header.started}
                <div class="bg-base-200 rounded-lg px-3 py-1.5">
                  <span class="text-base-content/70 font-medium">Started</span>
                  <time datetime={parseDateLike(header.started)?.toISOString()} title={String(header.started) + " UTC"} class="ml-2 font-semibold">
                    {relativeFrom(parseDateLike(header.started))}
                  </time>
                </div>
              {/if}

              <div class="bg-base-200 rounded-lg px-3 py-1.5">
                <span class="text-base-content/70 font-medium">Duration</span>
                <span class="ml-2 font-semibold">
                  {#if header.status === 'active'}
                    {tick >= 0 ? elapsedTime(parseDateLike(header.started)) : '—'}
                  {:else}
                    {header.duration ?? '—'}
                  {/if}
                </span>
              </div>
            </div>
          </div>

          <div class="flex shrink-0 items-center gap-2">
            <label class="swap swap-rotate btn btn-ghost btn-sm btn-circle">
              <input type="checkbox" class="theme-controller" value="dark" />
              <!-- sun icon -->
              <svg class="swap-on h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              <!-- moon icon -->
              <svg class="swap-off h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
              </svg>
            </label>

            <button
              class="btn btn-primary btn-sm"
              on:click={() => {
                if (confirm('Are you sure you want to end this session?')) {
                  live.pushEvent('end_session')
                }
              }}
              disabled={header.status === 'completed' || header.status === 'failed'}
              aria-label="End session"
            >
              End Session
            </button>
            <button
              class="btn btn-outline btn-sm"
              on:click={() => live.pushEvent('new_task')}
              aria-label="Create new task"
            >
              New Task
            </button>
            <button
              class="btn btn-outline btn-sm"
              on:click={() => live.pushEvent('add_note')}
              aria-label="Add note"
            >
              Add Note
            </button>
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="px-6 pt-4 pb-2">
        <div role="tablist" aria-label="Session sections" class="tabs tabs-bordered">
          {#each tabs as t}
            <button
              role="tab"
              id={`tab-${t.key}`}
              aria-selected={activeTab === t.key}
              aria-controls={`panel-${t.key}`}
              class="tab {activeTab === t.key ? 'tab-active' : ''} flex items-center gap-2"
              on:click={() => changeTab(t.key)}
            >
              <span>{t.label}</span>
              {#if t.countKey && countFor(t.countKey) > 0}
                <span class="badge badge-sm {activeTab === t.key ? 'badge-primary' : 'badge-ghost'}">
                  {countFor(t.countKey)}
                </span>
              {/if}
            </button>
          {/each}
        </div>
      </div>

      <!-- Content -->
      <div role="tabpanel" id={`panel-${activeTab}`} aria-labelledby={`tab-${activeTab}`} class="px-6 py-6">
        {#if activeTab === "tasks"}
          <TasksTab {tasks} />
        {:else if activeTab === "commits"}
          <CommitsTab {commits} />
        {:else if activeTab === "logs"}
          <LogsTab {logs} />
        {:else if activeTab === "context"}
          {#if context?.context}
            <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 px-6 py-4">
              <div class="prose prose-gray dark:prose-invert max-w-none text-sm">
                <pre class="whitespace-pre-wrap text-gray-900 dark:text-gray-100">{context.context}</pre>
              </div>
            </div>
          {:else}
            <div class={emptyStateStyle}>
              <div class="mx-auto max-w-md">
                <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">No context saved</h2>
                <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
                  Session context will be saved here for resumption.
                </p>
              </div>
            </div>
          {/if}
        {:else if activeTab === "notes"}
          <NotesTab {notes} />
        {:else if activeTab === "messages"}
          <MessagesTab {messages} {live} />
        {/if}
      </div>
      </div>
    </div>
  </div>
</div>
