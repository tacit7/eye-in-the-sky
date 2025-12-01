<script>
  export let messages
  export let live

  function handleSubmit(e) {
    const formData = new FormData(e.target)
    const body = formData.get('body')
    const provider = formData.get('provider')

    if (body.trim()) {
      live.pushEvent('send_message', { body, provider })
      e.target.reset()
    }
  }
</script>

<div class="card bg-base-100 shadow-sm h-[600px] overflow-hidden">
  <div class="card-body p-0 flex flex-col">
    <!-- Messages list -->
    <div class="flex-1 overflow-y-auto space-y-4 p-4 bg-base-200">
      {#if messages && messages.length > 0}
        {#each messages as message}
          <div class="chat {message.direction === 'outbound' ? 'chat-end' : 'chat-start'}">
            <div class="chat-header text-xs opacity-70">
              {message.sender_role === 'user' ? 'You' : 'Agent'}
              {#if message.provider}
                · {message.provider}
              {/if}
              · {new Date(message.inserted_at).toLocaleTimeString()}
            </div>

            <div class="chat-bubble {message.direction === 'outbound' ? 'chat-bubble-primary' : ''} whitespace-pre-wrap">
              {message.body}
            </div>

            {#if message.status === 'pending'}
              <div class="chat-footer opacity-70">
                <span class="badge badge-warning badge-sm">Pending</span>
              </div>
            {:else if message.status === 'failed'}
              <div class="chat-footer opacity-70">
                <span class="badge badge-error badge-sm">Failed</span>
              </div>
            {/if}
          </div>
        {/each}
      {:else}
        <div class="hero min-h-full">
          <div class="hero-content text-center">
            <div>
              <h3 class="text-lg font-bold">No messages yet</h3>
              <p class="py-2 text-base-content/70">
                Start a conversation with the agent below.
              </p>
            </div>
          </div>
        </div>
      {/if}
    </div>

    <!-- Message input -->
    <form on:submit|preventDefault={handleSubmit}>
      <div class="card-actions border-t border-base-300 p-4 flex-nowrap">
        <select name="provider" class="select select-bordered select-sm">
          <option value="claude">Claude</option>
          <option value="openai">OpenAI</option>
        </select>

        <input
          type="text"
          name="body"
          placeholder="Type your message..."
          class="input input-bordered flex-1"
          autocomplete="off"
        />

        <button type="submit" class="btn btn-primary">
          Send
        </button>
      </div>
    </form>
  </div>
</div>
