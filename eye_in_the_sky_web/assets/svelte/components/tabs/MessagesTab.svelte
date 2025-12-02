<script>
  import { onMount, afterUpdate } from 'svelte'

  export let messages
  export let live

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
    const date = new Date(timestamp)
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    })
  }

  function formatDate(date) {
    const today = new Date()
    const msgDate = new Date(date)
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
  .whatsapp-container {
    display: flex;
    flex-direction: column;
    height: 600px;
    background-color: #efeae2;
  }

  :global(.dark) .whatsapp-container {
    background-color: #0b141a;
  }

  .messages-scroll {
    flex: 1;
    overflow-y: auto;
    padding: 1rem 2rem;
  }

  /* Date separator */
  .date-separator {
    display: flex;
    justify-content: center;
    margin: 1.5rem 0;
  }

  .date-separator span {
    background-color: #e5e7eb;
    color: #6b7280;
    font-size: 11px;
    padding: 0.25rem 0.75rem;
    border-radius: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  :global(.dark) .date-separator span {
    background-color: #374151;
    color: #9ca3af;
  }

  /* Message group (replaces .message-wrapper) */
  .message-group {
    margin-bottom: 1.5rem;
    display: flex;
  }

  .message-group.outbound {
    justify-content: flex-end;
    flex-direction: column;
    align-items: flex-end;
  }

  .message-group.inbound {
    flex-direction: column;
    align-items: flex-start;
  }

  /* Sender label (once per group) */
  .group-sender-label {
    font-size: 11px;
    color: #6b7280;
    margin-bottom: 0.25rem;
    padding-left: 0.5rem;
  }

  :global(.dark) .group-sender-label {
    color: #9ca3af;
  }

  /* Container for multiple bubbles */
  .message-bubbles-container {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    max-width: 65%;
  }

  .message-bubble {
    width: fit-content;
    padding: 0.5rem 0.75rem;
    border-radius: 0.5rem;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
    position: relative;
  }

  .message-bubble.outbound {
    background-color: #d9fdd3;
  }

  .message-bubble.inbound {
    background-color: white;
  }

  :global(.dark) .message-bubble.outbound {
    background-color: #005c4b;
  }

  :global(.dark) .message-bubble.inbound {
    background-color: #202c33;
  }

  .message-text {
    font-size: 14px;
    color: #111827;
    white-space: pre-wrap;
    word-break: break-word;
    padding-right: 3rem;
  }

  :global(.dark) .message-text {
    color: #f3f4f6;
  }

  .message-meta {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.25rem;
    margin-top: 0.25rem;
  }

  .message-time {
    font-size: 11px;
    color: #6b7280;
  }

  :global(.dark) .message-time {
    color: #9ca3af;
  }

  .input-area {
    border-top: 1px solid #d1d5db;
    background-color: #f0f2f5;
    padding: 0.75rem;
  }

  :global(.dark) .input-area {
    border-color: #374151;
    background-color: #202c33;
  }

  .input-form {
    display: flex;
    align-items: flex-end;
    gap: 0.5rem;
  }

  .message-input {
    flex: 1;
    padding: 0.75rem 1rem;
    border-radius: 0.5rem;
    background-color: white;
    border: none;
    font-size: 15px;
    color: #111827;
  }

  .message-input:focus {
    outline: 2px solid #00a884;
    outline-offset: 0;
  }

  :global(.dark) .message-input {
    background-color: #2a3942;
    color: #f3f4f6;
  }

  .send-button {
    width: 3rem;
    height: 3rem;
    border-radius: 50%;
    background-color: #00a884;
    border: none;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .send-button:hover {
    background-color: #008f72;
  }

  .provider-selector {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .dropdown-content li button {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
  }

  .empty-state-content {
    text-align: center;
  }

  .empty-icon {
    width: 5rem;
    height: 5rem;
    margin: 0 auto 1rem;
    border-radius: 50%;
    background-color: #e5e7eb;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  :global(.dark) .empty-icon {
    background-color: #374151;
  }
</style>

<div class="whatsapp-container">
  <!-- Messages container -->
  <div
    bind:this={messagesContainer}
    on:scroll={handleScroll}
    class="messages-scroll"
  >
    {#if messages && messages.length > 0}
      {#each messages as messageGroup}
        <!-- Date separator -->
        {#if messageGroup.show_date_separator}
          <div class="date-separator">
            <span>{formatDate(messageGroup.date)}</span>
          </div>
        {/if}

        <!-- Message group -->
        <div class="message-group {messageGroup.direction}">
          <!-- Sender label (inbound only, once per group) -->
          {#if messageGroup.direction === 'inbound'}
            <div class="group-sender-label">
              {messageGroup.sender_role === 'user' ? 'You' : 'Agent'}
              {#if messageGroup.provider}
                · {messageGroup.provider}
              {/if}
            </div>
          {/if}

          <!-- Message bubbles -->
          <div class="message-bubbles-container">
            {#each messageGroup.messages as message, idx}
              <div class="message-bubble {messageGroup.direction}">
                <div class="message-text">{message.body}</div>

                <!-- Show time/status only on last message -->
                {#if idx === messageGroup.messages.length - 1}
                  <div class="message-meta">
                    <span class="message-time">{formatTime(message.inserted_at)}</span>

                    {#if messageGroup.direction === 'outbound'}
                      {#if messageGroup.status === 'pending'}
                        <svg style="width: 12px; height: 12px; color: #6b7280;" fill="currentColor" viewBox="0 0 12 12">
                          <path d="M6 0a6 6 0 100 12A6 6 0 006 0zm0 10.8A4.8 4.8 0 1110.8 6 4.8 4.8 0 016 10.8z"/>
                        </svg>
                      {:else if messageGroup.status === 'failed'}
                        <svg style="width: 12px; height: 12px; color: #ef4444;" fill="currentColor" viewBox="0 0 12 12">
                          <path d="M6 0a6 6 0 100 12A6 6 0 006 0zm3.707 8.293L8.293 9.707 6 7.414 3.707 9.707l-1.414-1.414L4.586 6 2.293 3.707l1.414-1.414L6 4.586l2.293-2.293 1.414 1.414L7.414 6l2.293 2.293z"/>
                        </svg>
                      {:else}
                        <svg style="width: 16px; height: 16px; color: #6b7280;" fill="currentColor" viewBox="0 0 16 16">
                          <path d="M15.01 3.316l-.478-.372a.365.365 0 0 0-.51.063L8.666 9.879a.32.32 0 0 1-.484.033l-.358-.325a.319.319 0 0 0-.484.032l-.378.483a.418.418 0 0 0 .036.541l1.32 1.266c.143.14.361.125.484-.033l6.272-8.048a.366.366 0 0 0-.064-.512zm-4.1 0l-.478-.372a.365.365 0 0 0-.51.063L4.566 9.879a.32.32 0 0 1-.484.033L1.891 7.769a.366.366 0 0 0-.515.006l-.423.433a.364.364 0 0 0 .006.514l3.258 3.185c.143.14.361.125.484-.033l6.272-8.048a.365.365 0 0 0-.063-.51z"/>
                        </svg>
                      {/if}
                    {/if}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      {/each}
    {:else}
      <div class="empty-state">
        <div class="empty-state-content">
          <div class="empty-icon">
            <svg style="width: 2.5rem; height: 2.5rem; color: #9ca3af;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
            </svg>
          </div>
          <h3 style="font-size: 1.125rem; font-weight: 600; color: #374151;">No messages yet</h3>
          <p style="font-size: 0.875rem; color: #6b7280; margin-top: 0.25rem;">
            Start a conversation with the agent below
          </p>
        </div>
      </div>
    {/if}
  </div>

  <!-- Input area -->
  <div class="input-area">
    <form on:submit|preventDefault={handleSubmit} class="input-form">
      <!-- Provider selector with badge -->
      <div class="provider-selector">
        <div class="dropdown dropdown-top">
          <label tabindex="0" class="btn btn-ghost btn-sm gap-2" title="Select AI Provider">
            <svg style="width: 1rem; height: 1rem;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
                  <svg style="width: 1rem; height: 1rem;" fill="currentColor" viewBox="0 0 20 20">
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
                  <svg style="width: 1rem; height: 1rem;" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
                  </svg>
                {/if}
              </button>
            </li>
          </ul>
        </div>
      </div>

      <!-- Message input -->
      <input
        type="text"
        name="body"
        placeholder="Send instruction to agent..."
        class="message-input"
        autocomplete="off"
      />

      <!-- Send button -->
      <button type="submit" class="send-button" aria-label="Send message">
        <svg style="width: 1.25rem; height: 1.25rem;" fill="currentColor" viewBox="0 0 24 24">
          <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
        </svg>
      </button>
    </form>
  </div>
</div>
