<script>
  export let logs = []
  export let live

  let filterLevel = 'all'

  $: filteredLogs =
    filterLevel === 'all' ? logs : logs.filter((log) => log.type === filterLevel)
</script>

<div class="h-full flex flex-col">
  <div class="flex gap-2 mb-4">
    <button
      class="px-3 py-1 text-xs rounded border"
      class:bg-indigo-600={filterLevel === 'all'}
      class:text-white={filterLevel === 'all'}
      class:border-indigo-600={filterLevel === 'all'}
      class:border-gray-300={filterLevel !== 'all'}
      on:click={() => (filterLevel = 'all')}
    >
      All
    </button>
    <button
      class="px-3 py-1 text-xs rounded border"
      class:bg-indigo-600={filterLevel === 'info'}
      class:text-white={filterLevel === 'info'}
      class:border-indigo-600={filterLevel === 'info'}
      class:border-gray-300={filterLevel !== 'info'}
      on:click={() => (filterLevel = 'info')}
    >
      Info
    </button>
    <button
      class="px-3 py-1 text-xs rounded border"
      class:bg-indigo-600={filterLevel === 'error'}
      class:text-white={filterLevel === 'error'}
      class:border-indigo-600={filterLevel === 'error'}
      class:border-gray-300={filterLevel !== 'error'}
      on:click={() => (filterLevel = 'error')}
    >
      Error
    </button>
  </div>

  <div class="flex-1 overflow-y-auto space-y-3">
    {#each filteredLogs as log}
      <div class="text-sm">
        <div class="flex items-center gap-2 mb-1">
          <span class="text-xs font-mono text-gray-500">
            {log.timestamp ? log.timestamp.slice(0, 19) : '—'}
          </span>
          <span class="text-xs px-2 py-0.5 rounded bg-gray-100">{log.type}</span>
        </div>
        <p>{log.message}</p>
      </div>
    {/each}
  </div>
</div>
