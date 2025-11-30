<script>
  export let tasks = []
  export let live

  let selectedTask = tasks[0] || null

  $: if (tasks.length > 0 && !selectedTask) {
    selectedTask = tasks[0]
  }

  function priorityColor(priority) {
    if (priority >= 70) return 'text-red-500'
    if (priority >= 40) return 'text-yellow-500'
    return 'text-blue-500'
  }
</script>

<div class="grid grid-cols-2 gap-4 h-full">
  <!-- Left: Task List -->
  <div class="overflow-y-auto">
    <div class="text-xs text-gray-500 mb-2">
      {tasks.length} tasks
    </div>
    <div class="space-y-1">
      {#each tasks as task}
        <button
          class="w-full text-left p-2 rounded hover:bg-gray-50 transition-colors"
          class:bg-indigo-50={selectedTask?.id === task.id}
          on:click={() => (selectedTask = task)}
        >
          <div class="flex items-center gap-2">
            <span class={priorityColor(task.priority)}>●</span>
            <span class="text-sm font-medium truncate">{task.title}</span>
          </div>
          {#if task.state}
            <div class="text-xs text-gray-500 ml-6">
              {task.state.name || 'No state'}
            </div>
          {/if}
        </button>
      {/each}
    </div>
  </div>

  <!-- Right: Task Details -->
  <div class="overflow-y-auto border-l border-gray-200 pl-4">
    {#if selectedTask}
      <h4 class="font-semibold mb-2">{selectedTask.title}</h4>
      {#if selectedTask.description}
        <p class="text-sm text-gray-600 mb-4">{selectedTask.description}</p>
      {/if}

      <div class="space-y-2 text-sm mb-4">
        {#if selectedTask.state}
          <div><strong>State:</strong> {selectedTask.state.name}</div>
        {/if}
        <div><strong>Priority:</strong> {selectedTask.priority}</div>
        {#if selectedTask.tags && selectedTask.tags.length > 0}
          <div>
            <strong>Tags:</strong>
            {selectedTask.tags.map((t) => t.name).join(', ')}
          </div>
        {/if}
        {#if selectedTask.created_at}
          <div class="text-xs text-gray-500">
            Created: {selectedTask.created_at.slice(0, 16)}
          </div>
        {/if}
      </div>
    {:else}
      <p class="text-gray-500">Select a task to view details</p>
    {/if}
  </div>
</div>
