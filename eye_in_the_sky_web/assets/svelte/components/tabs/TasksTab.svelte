<script>
  export let tasks = []
  export let live

  let selectedTask = null
  let hoveredTask = null

  function getPriorityFlag(priority) {
    if (priority >= 70) return { color: '#d1453b', icon: '🚩', label: 'P1' }
    if (priority >= 40) return { color: '#eb8909', icon: '🔶', label: 'P2' }
    if (priority >= 20) return { color: '#246fe0', icon: '🔵', label: 'P3' }
    return { color: '#808080', icon: '⚪', label: 'P4' }
  }

  function formatDate(dateStr) {
    if (!dateStr) return null
    try {
      const date = new Date(dateStr)
      const today = new Date()
      const tomorrow = new Date(today)
      tomorrow.setDate(tomorrow.getDate() + 1)

      if (date.toDateString() === today.toDateString()) return 'Today'
      if (date.toDateString() === tomorrow.toDateString()) return 'Tomorrow'

      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
    } catch {
      return null
    }
  }

  function isOverdue(dateStr) {
    if (!dateStr) return false
    return new Date(dateStr) < new Date()
  }

  function handleTaskClick(task) {
    selectedTask = selectedTask?.id === task.id ? null : task
  }

  function toggleComplete(task, event) {
    event.stopPropagation()
    // TODO: Implement task completion via live push
    console.log('Toggle complete:', task.id)
  }

  function formatUUID(id) {
    if (!id) return ''
    // Remove any existing dashes and format as UUID
    const clean = id.replace(/-/g, '')
    if (clean.length !== 32) return id // Return as-is if not valid length
    return `${clean.slice(0,8)}-${clean.slice(8,12)}-${clean.slice(12,16)}-${clean.slice(16,20)}-${clean.slice(20)}`
  }

  function copyToClipboard(text, event) {
    event.stopPropagation()
    const formattedId = formatUUID(text)
    navigator.clipboard.writeText(formattedId).then(() => {
      console.log('Copied:', formattedId)
    }).catch(err => {
      console.error('Failed to copy:', err)
    })
  }

  function getShortId(id) {
    if (!id) return ''
    const idStr = String(id)
    // For integer IDs, show last 8 chars for better uniqueness
    // For UUIDs, show first 8 chars
    if (idStr.includes('-')) {
      return idStr.substring(0, 8)
    } else {
      return idStr.slice(-8)
    }
  }
</script>

<div class="max-w-4xl mx-auto">
  <!-- Task List -->
  <div class="space-y-0">
    {#each tasks as task}
      {@const priority = getPriorityFlag(task.priority)}
      {@const dueDate = formatDate(task.due_at)}
      {@const overdue = isOverdue(task.due_at)}

      <div
        class="group relative border-b border-base-200 hover:bg-base-200/50 transition-colors cursor-pointer"
        on:click={() => handleTaskClick(task)}
        on:mouseenter={() => hoveredTask = task.id}
        on:mouseleave={() => hoveredTask = null}
      >
        <div class="flex items-start gap-3 py-3 px-4">
          <!-- Checkbox -->
          <button
            class="mt-0.5 flex-shrink-0 w-5 h-5 rounded-full border-2 border-base-content/30 hover:border-primary transition-colors flex items-center justify-center"
            on:click={(e) => toggleComplete(task, e)}
            aria-label="Complete task"
          >
            {#if task.state_name === 'done' || task.completed_at}
              <svg class="w-3.5 h-3.5 text-success" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
            {/if}
          </button>

          <!-- Task Content -->
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-2">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <h3 class="text-sm font-normal text-base-content group-hover:text-base-content/90 {task.completed_at ? 'line-through opacity-60' : ''}">
                    {task.title}
                  </h3>
                  {#if task.priority >= 20}
                    <span
                      class="flex-shrink-0 text-xs"
                      style="color: {priority.color}"
                      title="{priority.label} priority"
                    >
                      🚩
                    </span>
                  {/if}
                </div>

                {#if task.description && selectedTask?.id === task.id}
                  <p class="text-xs text-base-content/60 mt-1 whitespace-pre-wrap">
                    {task.description}
                  </p>
                {/if}

                <!-- Task Meta -->
                <div class="flex items-center gap-3 mt-2 text-xs text-base-content/50">
                  <!-- Task ID Badge -->
                  <button
                    class="badge badge-ghost badge-xs hover:badge-primary cursor-pointer font-mono transition-colors"
                    on:click={(e) => copyToClipboard(task.id, e)}
                    title="Copy ID: {task.id}"
                  >
                    #{getShortId(task.id)}
                  </button>

                  {#if dueDate}
                    <span class="flex items-center gap-1 {overdue ? 'text-error' : ''}">
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                      {dueDate}
                    </span>
                  {/if}

                  {#if task.state_name}
                    <span class="badge badge-ghost badge-xs">
                      {task.state_name}
                    </span>
                  {/if}

                  {#if task.tags && task.tags.length > 0}
                    <div class="flex items-center gap-1">
                      {#each task.tags as tag}
                        <span class="badge badge-outline badge-xs">
                          {tag.name}
                        </span>
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>

              <!-- Actions (visible on hover) -->
              {#if hoveredTask === task.id || selectedTask?.id === task.id}
                <div class="flex items-center gap-1 flex-shrink-0">
                  <button
                    class="btn btn-ghost btn-xs btn-square"
                    on:click={(e) => {e.stopPropagation()}}
                    title="Edit task"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    class="btn btn-ghost btn-xs btn-square text-error/70 hover:text-error"
                    on:click={(e) => {e.stopPropagation()}}
                    title="Delete task"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              {/if}
            </div>
          </div>
        </div>
      </div>
    {/each}

    {#if !tasks || tasks.length === 0}
      <div class="text-center py-12">
        <div class="text-base-content/40 mb-2">
          <svg class="w-12 h-12 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
        </div>
        <h3 class="text-sm font-medium text-base-content/60 mb-1">No tasks yet</h3>
        <p class="text-xs text-base-content/40">Click "New Task" to get started</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .badge-xs {
    font-size: 0.65rem;
    padding: 0.125rem 0.375rem;
  }
</style>
