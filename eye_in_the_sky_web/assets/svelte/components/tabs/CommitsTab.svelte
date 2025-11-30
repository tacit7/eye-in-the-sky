<script>
  export let commits = []
  export let live

  let selectedCommit = commits[0] || null

  $: if (commits.length > 0 && !selectedCommit) {
    selectedCommit = commits[0]
  }
</script>

<div class="grid grid-cols-[40%_60%] gap-4 h-full">
  <!-- Left: Commit List -->
  <div class="overflow-y-auto">
    <div class="space-y-1">
      {#each commits as commit}
        <button
          class="w-full text-left p-3 rounded border border-transparent hover:bg-gray-50 transition-colors"
          class:bg-indigo-50={selectedCommit?.commit_hash === commit.commit_hash}
          class:border-indigo-200={selectedCommit?.commit_hash === commit.commit_hash}
          on:click={() => (selectedCommit = commit)}
        >
          <div class="font-mono text-xs text-gray-500">
            {commit.commit_hash.slice(0, 8)}
          </div>
          <div class="text-sm font-medium mt-1">{commit.commit_message}</div>
          {#if commit.created_at}
            <div class="text-xs text-gray-500 mt-1">
              {commit.created_at.slice(0, 16)}
            </div>
          {/if}
        </button>
      {/each}
    </div>
  </div>

  <!-- Right: Diff Viewer (placeholder) -->
  <div class="overflow-y-auto border-l border-gray-200 pl-4">
    {#if selectedCommit}
      <div class="font-mono text-xs">
        <div class="mb-2 text-gray-600">
          Commit: {selectedCommit.commit_hash.slice(0, 8)}
        </div>
        <pre class="whitespace-pre-wrap text-gray-500">Git diff viewer coming soon...</pre>
      </div>
    {:else}
      <p class="text-gray-500">Select a commit to view diff</p>
    {/if}
  </div>
</div>
