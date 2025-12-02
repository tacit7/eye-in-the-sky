<script>
  import { onMount, afterUpdate } from 'svelte'

  export let messages
  export let live

  $: console.log('Messages received:', messages)

  let messagesContainer
  let shouldAutoScroll = true
  let selectedProvider = 'claude'

  function handleSubmit(e) {
    const formData = new FormData(e.target)
    const body = formData.get('body')

    if (body.trim()) {
      live.pushEvent('send_message', { body, provider: selectedProvider })
      e.target.reset()
      shouldAutoScroll = true
    }
  }

  function selectProvider(provider) {
    selectedProvider = provider
  }

  function formatTime(timestamp) {
    if (!timestamp) return 'Unknown'
    const date = new Date(timestamp)
    if (isNaN(date.getTime())) return 'Invalid Date'
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    })
  }

  function formatDate(date) {
    if (!date) return 'Unknown'
    const today = new Date()
    const msgDate = new Date(date)
    if (isNaN(msgDate.getTime())) return 'Invalid Date'

    const diffDays = Math.floor((today - msgDate) / (1000 * 60 * 60 * 24))

    if (diffDays === 0) return 'Today'
    if (diffDays === 1) return 'Yesterday'
    return msgDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  }

  function scrollToBottom() {
    if (messagesContainer && shouldAutoScroll) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight
    }
  }

  function handleScroll() {
    if (messagesContainer) {
      const isAtBottom = messagesContainer.scrollHeight - messagesContainer.scrollTop <= messagesContainer.clientHeight + 100
      shouldAutoScroll = isAtBottom
    }
  }

  afterUpdate(() => {
    scrollToBottom()
  })

  onMount(() => {
    scrollToBottom()
  })
</script>

<style>
  .messages-container {
    display: flex;
    flex-direction: column;
    height: 600px;
  }

  .messages-scroll {
    flex: 1;
    overflow-y: auto;
    padding: 1rem 1.5rem;
  }

  /* Date separator */
  .date-separator {
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 1.5rem 0;
  }

  .date-badge {
    font-size: 0.75rem;
    font-weight: 600;
    padding: 0.25rem 0.75rem;
  }
</style>

<div class="messages-container">
  <!-- Messages scroll area -->
  <div
    bind:this={messagesContainer}
    on:scroll={handleScroll}
    class="messages-scroll"
  >
    {#if messages && messages.length > 0}
      {#each messages as message, idx}
        <!-- Date separator -->
        {#if idx === 0 || formatDate(messages[idx - 1].inserted_at) !== formatDate(message.inserted_at)}
          <div class="date-separator">
            <div class="badge badge-ghost badge-sm date-badge">
              {formatDate(message.inserted_at)}
            </div>
          </div>
        {/if}

        <!-- Message card -->
        <div class="card bg-base-100 shadow-sm mb-3 {message.sender_role === 'user' ? 'border-l-4 border-primary' : 'border-l-4 border-base-300'}">
          <div class="card-body p-4">
            <div class="flex items-center gap-2 mb-2">
              <h3 class="card-title text-sm">
                {message.sender_role === 'user' ? 'You' : message.sender_role === 'agent' ? 'Agent' : message.sender_role}
              </h3>
              {#if message.provider}
                <span class="badge badge-xs badge-ghost">{message.provider}</span>
              {/if}
              <time class="text-xs opacity-50 ml-auto">{formatTime(message.inserted_at)}</time>
            </div>
            <p class="text-sm whitespace-pre-wrap">{message.body}</p>
          </div>
        </div>
      {/each}
    {:else}
      <!-- Empty state -->
      <div class="flex items-center justify-center h-full">
        <div class="text-center max-w-md">
          <svg class="w-12 h-12 mx-auto mb-4 text-base-content/30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
          </svg>
          <h3 class="text-lg font-semibold text-base-content">No messages yet</h3>
          <p class="mt-1 text-sm text-base-content/70">
            Start a conversation with the agent below
          </p>
        </div>
      </div>
    {/if}
  </div>

  <!-- Input area -->
  <div class="border-t border-base-300 p-4">
    <form on:submit|preventDefault={handleSubmit} class="flex items-center gap-2">
      <!-- Provider selector dropdown -->
      <div class="dropdown dropdown-top">
        <label tabindex="0" class="btn btn-ghost btn-sm gap-2" title="Select AI Provider">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
          </svg>
          <span class="badge badge-sm badge-primary">{selectedProvider}</span>
        </label>
        <ul tabindex="0" class="dropdown-content menu p-2 shadow bg-base-100 rounded-box w-40 mb-2">
          <li>
            <button type="button" on:click={() => selectProvider('claude')}
              class="flex items-center justify-between">
              <span>Claude</span>
              {#if selectedProvider === 'claude'}
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
                </svg>
              {/if}
            </button>
          </li>
          <li>
            <button type="button" on:click={() => selectProvider('openai')}
              class="flex items-center justify-between">
              <span>OpenAI</span>
              {#if selectedProvider === 'openai'}
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
                </svg>
              {/if}
            </button>
          </li>
        </ul>
      </div>

      <!-- Message input -->
      <input
        type="text"
        name="body"
        placeholder="Send instruction to agent..."
        class="input input-bordered flex-1"
        autocomplete="off"
      />

      <!-- Send button -->
      <button type="submit" class="btn btn-primary btn-sm" aria-label="Send message">
        <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
          <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
        </svg>
      </button>
    </form>
  </div>
</div>
