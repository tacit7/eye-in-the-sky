<script>
  import { onMount, onDestroy } from "svelte"
  import NotesTab from "./tabs/NotesTab.svelte"

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

  const statusStyles = {
    active: "bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-900/30 dark:text-emerald-300 dark:border-emerald-800",
    working: "bg-indigo-50 text-indigo-700 border border-indigo-200 dark:bg-indigo-900/30 dark:text-indigo-300 dark:border-indigo-800",
    completed: "bg-gray-100 text-gray-700 border border-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-600",
    failed: "bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-900/30 dark:text-rose-300 dark:border-rose-800",
    idle: "bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-900/30 dark:text-amber-300 dark:border-amber-800"
  }

  const short = (id) => (id ? id.slice(0, 8) : "")

  function countFor(key) {
    return counts?.[key] || 0
  }

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

  function elapsedTime(date) {
    if (!date) return "—"
    const secs = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000))
    const m = Math.floor(secs / 60), s = secs % 60
    const h = Math.floor(m / 60), mm = m % 60
    if (h > 0) return `${h}h ${mm}m`
    if (m > 0) return `${m}m`
    return `${s}s`
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

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
  <div class="mx-auto max-w-6xl px-6 py-6">
    <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm">
      <!-- Header -->
      <div class="bg-white/80 dark:bg-gray-800/80 backdrop-blur sticky top-0 z-30 border-b border-gray-200 dark:border-gray-700 shadow-sm px-6 py-5">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <a href="/" class="inline-flex items-center gap-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100">
              <span>←</span>
              <span>Back to Agents</span>
            </a>

            <div class="mt-3 flex flex-wrap items-center gap-2">
              <h1 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">{header.agent_type}</h1>

              <span class="rounded-full border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-800 px-2 py-0.5 font-mono text-xs text-gray-700 dark:text-gray-300">
                {short(header.agent_id)}
              </span>

              <span class={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${statusStyles[header.status] || statusStyles.completed}`} aria-label={`Status: ${header.status}`}>
                {header.status}
              </span>
            </div>

            <!-- Meta chips -->
            <div class="mt-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2 text-xs sm:text-sm">
              <div class="rounded-lg bg-gray-50 dark:bg-gray-800/50 px-3 py-1.5">
                <span class="text-gray-600 dark:text-gray-400 font-medium">Session</span>
                <span class="ml-2 font-mono text-gray-900 dark:text-gray-100 font-semibold">#{short(header.session_id)}</span>
              </div>

              <div class="rounded-lg bg-gray-50 dark:bg-gray-800/50 px-3 py-1.5">
                <span class="text-gray-600 dark:text-gray-400 font-medium">Project</span>
                <span class="ml-2 text-gray-900 dark:text-gray-100 font-semibold">{header.project ?? "Unassigned"}</span>
              </div>

              {#if header.started}
                <div class="rounded-lg bg-gray-50 dark:bg-gray-800/50 px-3 py-1.5">
                  <span class="text-gray-600 dark:text-gray-400 font-medium">Started</span>
                  <time datetime={parseDateLike(header.started)?.toISOString()} title={String(header.started) + " UTC"} class="ml-2 text-gray-900 dark:text-gray-100 font-semibold">
                    {relativeFrom(parseDateLike(header.started))}
                  </time>
                </div>
              {/if}

              <div class="rounded-lg bg-gray-50 dark:bg-gray-800/50 px-3 py-1.5">
                <span class="text-gray-600 dark:text-gray-400 font-medium">Duration</span>
                <span class="ml-2 text-gray-900 dark:text-gray-100 font-semibold">
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
            <button
              class="rounded-md bg-gray-900 dark:bg-gray-700 px-3 py-1.5 text-sm font-semibold text-white hover:bg-gray-800 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
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
              class="rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-1.5 text-sm font-semibold text-gray-800 dark:text-gray-100 hover:bg-gray-50 dark:hover:bg-gray-700"
              on:click={() => live.pushEvent('new_task')}
              aria-label="Create new task"
            >
              New Task
            </button>
            <button
              class="rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-1.5 text-sm font-semibold text-gray-800 dark:text-gray-100 hover:bg-gray-50 dark:hover:bg-gray-700"
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
        <div role="tablist" aria-label="Session sections" class="inline-flex gap-1 border-b border-gray-200 dark:border-gray-700">
          {#each tabs as t}
            <button
              role="tab"
              id={`tab-${t.key}`}
              aria-selected={activeTab === t.key}
              aria-controls={`panel-${t.key}`}
              class="flex items-center gap-2 px-3 py-2 text-sm font-semibold transition-all border-b-2 -mb-px
              {activeTab === t.key ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400' : 'border-transparent text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:border-gray-300 dark:hover:border-gray-600'}"
              on:click={() => changeTab(t.key)}
            >
              <span>{t.label}</span>
              {#if t.countKey && countFor(t.countKey) > 0}
                <span class={`rounded-full px-1.5 py-0.5 text-[11px] font-bold ${activeTab === t.key ? 'bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300' : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}`}>
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
          {#if tasks && tasks.length > 0}
            <div class="space-y-2">
              {#each tasks as task}
                <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 px-4 py-2.5 hover:border-gray-400 dark:hover:border-gray-500 hover:bg-gray-50 dark:hover:bg-gray-750 hover:shadow-sm transition-all">
                  <div class="flex items-start justify-between gap-4">
                    <div class="flex-1 min-w-0">
                      <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">{task.title}</h3>
                      {#if task.description}
                        <p class="mt-1 text-sm text-gray-700 dark:text-gray-300 line-clamp-2">{task.description}</p>
                      {/if}
                      {#if task.tags && task.tags.length > 0}
                        <div class="mt-2 flex flex-wrap gap-1">
                          {#each task.tags as tag}
                            <span class="inline-flex items-center rounded-md border border-gray-200 bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200">
                              {tag.name}
                            </span>
                          {/each}
                        </div>
                      {/if}
                    </div>
                    <span class="shrink-0 rounded-full px-2.5 py-1 text-xs font-bold bg-blue-600 text-white">
                      {task.state_name || 'todo'}
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="rounded-lg border border-dashed border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-800 px-6 py-10 text-center">
              <div class="mx-auto max-w-md">
                <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">No tasks yet</h2>
                <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
                  Tasks created during this session will appear here.
                </p>
              </div>
            </div>
          {/if}
        {:else if activeTab === "commits"}
          {#if commits && commits.length > 0}
            <div class="space-y-2">
              {#each commits as commit}
                <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 px-4 py-2.5 hover:border-gray-400 dark:hover:border-gray-500 hover:bg-gray-50 dark:hover:bg-gray-750 hover:shadow-sm transition-all">
                  <div class="flex items-start gap-4">
                    <code class="shrink-0 rounded-md bg-gray-700 dark:bg-gray-600 px-2 py-1 font-mono text-xs font-semibold text-white">
                      {commit.commit_hash?.slice(0, 7)}
                    </code>
                    <p class="flex-1 text-sm font-medium text-gray-900 dark:text-gray-100">{commit.commit_message}</p>
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="rounded-lg border border-dashed border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-800 px-6 py-10 text-center">
              <div class="mx-auto max-w-md">
                <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">No commits yet</h2>
                <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
                  Git commits made during this session will appear here.
                </p>
              </div>
            </div>
          {/if}
        {:else if activeTab === "logs"}
          {#if logs && logs.length > 0}
            <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-900 p-4">
              <div class="space-y-1 font-mono text-xs">
                {#each logs as log}
                  <div class="flex gap-3">
                    <span class="shrink-0 text-gray-500">[{log.type}]</span>
                    <span class="flex-1 text-gray-100">{log.message}</span>
                  </div>
                {/each}
              </div>
            </div>
          {:else}
            <div class="rounded-lg border border-dashed border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-800 px-6 py-10 text-center">
              <div class="mx-auto max-w-md">
                <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">No logs yet</h2>
                <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
                  Session logs will appear here as the agent works.
                </p>
              </div>
            </div>
          {/if}
        {:else if activeTab === "context"}
          {#if context?.context}
            <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 px-6 py-4">
              <div class="prose prose-gray dark:prose-invert max-w-none text-sm">
                <pre class="whitespace-pre-wrap text-gray-900 dark:text-gray-100">{context.context}</pre>
              </div>
            </div>
          {:else}
            <div class="rounded-lg border border-dashed border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-800 px-6 py-10 text-center">
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
          <div class="flex flex-col h-[600px]">
            <!-- Messages list -->
            <div class="flex-1 overflow-y-auto space-y-3 p-4 bg-gray-50 dark:bg-gray-900 rounded-t-lg">
              {#if messages && messages.length > 0}
                {#each messages as message}
                  <div class="flex {message.direction === 'outbound' ? 'justify-end' : 'justify-start'}">
                    <div class="max-w-[70%] rounded-lg px-4 py-3 {message.direction === 'outbound' ? 'bg-indigo-600 text-white' : 'bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700'}">
                      <div class="flex items-center gap-2 mb-1">
                        <span class="text-xs font-semibold opacity-80">
                          {message.sender_role === 'user' ? 'You' : 'Agent'}
                        </span>
                        {#if message.provider}
                          <span class="text-xs opacity-70">· {message.provider}</span>
                        {/if}
                        <span class="text-xs opacity-70">
                          · {new Date(message.inserted_at).toLocaleTimeString()}
                        </span>
                      </div>
                      <p class="text-sm whitespace-pre-wrap">{message.body}</p>
                      {#if message.status === 'pending'}
                        <span class="inline-block mt-2 text-xs px-2 py-0.5 rounded-full bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400">
                          Pending
                        </span>
                      {:else if message.status === 'failed'}
                        <span class="inline-block mt-2 text-xs px-2 py-0.5 rounded-full bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400">
                          Failed
                        </span>
                      {/if}
                    </div>
                  </div>
                {/each}
              {:else}
                <div class="flex items-center justify-center h-full">
                  <div class="text-center">
                    <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100">No messages yet</h3>
                    <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
                      Start a conversation with the agent below.
                    </p>
                  </div>
                </div>
              {/if}
            </div>

            <!-- Message input -->
            <div class="border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 rounded-b-lg p-4">
              <form on:submit|preventDefault={(e) => {
                const formData = new FormData(e.target);
                const body = formData.get('body');
                const provider = formData.get('provider');
                if (body.trim()) {
                  live.pushEvent('send_message', { body, provider });
                  e.target.reset();
                }
              }}>
                <div class="flex gap-2">
                  <select
                    name="provider"
                    class="rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100"
                  >
                    <option value="claude">Claude</option>
                    <option value="openai">OpenAI</option>
                  </select>
                  <input
                    type="text"
                    name="body"
                    placeholder="Type your message..."
                    class="flex-1 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500 dark:focus:border-indigo-400 dark:focus:ring-indigo-400"
                    autocomplete="off"
                  />
                  <button
                    type="submit"
                    class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                  >
                    Send
                  </button>
                </div>
              </form>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>
