<script>
  import { onMount, afterUpdate } from 'svelte'
  import ChannelsSidebar from '../ChannelsSidebar.svelte'

  export let channels = []
  export let activeChannelId = null
  export let messages = []
  export let unreadCounts = {}
  export let agentStatusCounts = {}
  export let live

  let messagesContainer
  let shouldAutoScroll = true
  let inputValue = ''

  // Modal state
  let showAgentModal = false
  let agentType = 'claude'
  let agentModel = 'sonnet'
  let agentInstructions = ''

  function openAgentModal() {
    showAgentModal = true
    agentType = 'claude'
    agentModel = 'sonnet'
    agentInstructions = ''
  }

  function closeAgentModal() {
    showAgentModal = false
  }

  function createAgent() {
    live.pushEvent('create_agent', {
      agent_type: agentType,
      model: agentModel,
      instructions: agentInstructions,
      channel_id: activeChannelId
    })
    closeAgentModal()
  }

  function handleSubmit(e) {
    const body = inputValue.trim()

    if (body) {
      live.pushEvent('send_channel_message', {
        channel_id: activeChannelId,
        body: body
      })
      inputValue = ''
      shouldAutoScroll = true
    }
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
  /* Theme colors */
  :global(:root) {
    --bg-shell: #f8f8f8;
    --bg-surface: #ffffff;
    --bg-sidebar: #111827;
    --text-primary: #1d1c1d;
    --text-secondary: #616061;
    --text-tertiary: #9ca3af;
    --border-subtle: #dddddd;
    --accent-primary: #0f766e;
    --accent-soft: #e0f2f1;
  }

  :global([data-theme="dark"]) {
    --bg-shell: #1a1d21;
    --bg-surface: #222529;
    --bg-sidebar: #1a1d21;
    --text-primary: #f8f8f8;
    --text-secondary: #dcddde;
    --text-tertiary: #9ca3af;
    --border-subtle: #2f3437;
    --accent-primary: #14b8a6;
    --accent-soft: #064e3b;
  }

  .agent-messages-container {
    display: flex;
    height: 100vh;
    background-color: var(--bg-shell);
  }

  .main-content {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .channel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem 1.5rem;
    border-bottom: 1px solid var(--border-subtle);
    background-color: var(--bg-surface);
  }

  .channel-title {
    font-size: 1.125rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .agent-status {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    font-size: 0.875rem;
  }

  .new-agent-btn {
    padding: 0.5rem 1rem;
    background-color: var(--accent-primary);
    color: white;
    border: none;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.15s;
    white-space: nowrap;
  }

  .new-agent-btn:hover {
    opacity: 0.9;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.75rem;
    border-radius: 0.375rem;
    font-weight: 600;
  }

  .status-badge.status-active {
    background-color: #d1fae5;
    color: #065f46;
  }

  :global(.dark) .status-badge.status-active {
    background-color: #064e3b;
    color: #6ee7b7;
  }

  .status-badge.status-idle {
    background-color: #dbeafe;
    color: #1e40af;
  }

  :global(.dark) .status-badge.status-idle {
    background-color: #1e3a8a;
    color: #93c5fd;
  }

  .status-badge.status-working {
    background-color: #fed7aa;
    color: #92400e;
  }

  :global(.dark) .status-badge.status-working {
    background-color: #78350f;
    color: #fcd34d;
  }

  .status-badge.status-nats {
    background-color: #e0e7ff;
    color: #3730a3;
  }

  :global(.dark) .status-badge.status-nats {
    background-color: #312e81;
    color: #a5b4fc;
  }

  .status-separator {
    color: var(--text-tertiary);
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
    margin: 1.5rem 0;
  }

  .date-separator::before,
  .date-separator::after {
    content: '';
    flex: 1;
    border-bottom: 1px solid var(--border-subtle);
  }

  .date-separator span {
    padding: 0 1rem;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-secondary);
  }

  /* Message */
  .message {
    margin-bottom: 0.5rem;
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    transition: background-color 0.15s;
  }

  .message:hover {
    background-color: var(--bg-shell);
  }

  .message-header {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
  }

  .sender-name {
    font-weight: 700;
    font-size: 0.9375rem;
    color: var(--text-primary);
  }

  .message-time {
    font-size: 0.75rem;
    color: var(--text-secondary);
  }

  .message-body {
    font-size: 0.9375rem;
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.5;
  }

  .input-area {
    border-top: 1px solid var(--border-subtle);
    background-color: var(--bg-surface);
    padding: 1rem 1.5rem;
  }

  .input-form {
    display: flex;
    align-items: flex-end;
    gap: 0.5rem;
  }

  .message-input {
    flex: 1;
    padding: 0.75rem 1rem;
    border-radius: 0.375rem;
    background-color: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    font-size: 0.9375rem;
    color: var(--text-primary);
  }

  .message-input:focus {
    outline: 2px solid var(--accent-primary);
    outline-offset: 0;
    border-color: transparent;
  }

  .send-button {
    padding: 0.75rem 1.5rem;
    border-radius: 0.375rem;
    background-color: var(--accent-primary);
    border: none;
    color: white;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.15s, opacity 0.15s;
  }

  .send-button:hover:not(:disabled) {
    opacity: 0.9;
  }

  .send-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

<div class="agent-messages-container">
  <!-- Channels Sidebar -->
  <ChannelsSidebar
    {channels}
    {activeChannelId}
    {unreadCounts}
    {live}
  />

  <!-- Main Content -->
  <div class="main-content">
    <!-- Channel Header -->
    <div class="channel-header">
      <div class="channel-title">
        {#if activeChannelId}
          {channels.find(c => c.id === activeChannelId)?.name || 'Project'}
        {:else}
          Select a project
        {/if}
      </div>
      <div class="header-right">
        <div class="agent-status">
          <span class="status-badge status-active">
            {agentStatusCounts.active || 0} active
          </span>
          <span class="status-badge status-idle">
            {agentStatusCounts.idle || 0} idle
          </span>
          <span class="status-badge status-working">
            {agentStatusCounts.working || 0} running
          </span>
          <span class="status-separator">·</span>
          <span class="status-badge status-nats">
            NATS: Live
          </span>
        </div>
        <button class="new-agent-btn" on:click={openAgentModal}>
          + New Agent
        </button>
      </div>
    </div>

    <!-- Messages Container -->
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
              <span>{formatDate(message.inserted_at)}</span>
            </div>
          {/if}

          <!-- Message -->
          <div class="message">
            <div class="message-header">
              <span class="sender-name">
                {message.sender_role === 'user' ? 'You' : `Agent (${message.provider || 'unknown'})`}
              </span>
              <span class="message-time">{formatTime(message.inserted_at)}</span>
            </div>

            <div class="message-body">{message.body}</div>
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
              Messages from agents will appear here
            </p>
          </div>
        </div>
      {/if}
    </div>

    <!-- Input Area -->
    <div class="input-area">
      <form on:submit|preventDefault={handleSubmit} class="input-form">
        <input
          type="text"
          bind:value={inputValue}
          placeholder="Send instruction to agents..."
          class="message-input"
          autocomplete="off"
        />

        <button type="submit" class="send-button" disabled={!inputValue || inputValue.trim() === ''}>
          Send
        </button>
      </form>
    </div>
  </div>

  <!-- Agent Creation Modal -->
  {#if showAgentModal}
    <div class="modal modal-open">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4">Create New Agent</h3>

        <!-- Agent Type -->
        <div class="form-control w-full mb-4">
          <label class="label" for="agent-type">
            <span class="label-text">Agent Type</span>
          </label>
          <select id="agent-type" class="select select-bordered w-full" bind:value={agentType}>
            <option value="claude">Claude</option>
            <option value="codex">Codex</option>
          </select>
        </div>

        <!-- Model -->
        <div class="form-control w-full mb-4">
          <label class="label" for="model">
            <span class="label-text">Model</span>
          </label>
          <select id="model" class="select select-bordered w-full" bind:value={agentModel}>
            {#if agentType === 'claude'}
              <option value="sonnet">Sonnet</option>
              <option value="opus">Opus</option>
              <option value="haiku">Haiku</option>
            {:else}
              <option value="gpt-4">GPT-4</option>
              <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
            {/if}
          </select>
        </div>

        <!-- Instructions -->
        <div class="form-control w-full mb-4">
          <label class="label" for="instructions">
            <span class="label-text">Instructions</span>
          </label>
          <textarea
            id="instructions"
            class="textarea textarea-bordered h-24"
            placeholder="Enter agent instructions..."
            bind:value={agentInstructions}
          ></textarea>
        </div>

        <div class="modal-action">
          <button class="btn" on:click={closeAgentModal}>Cancel</button>
          <button class="btn btn-primary" on:click={createAgent}>Create</button>
        </div>
      </div>
    </div>
  {/if}
</div>
