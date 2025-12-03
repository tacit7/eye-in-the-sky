<script>
  import { onMount } from 'svelte'

  export let live

  let isExpanded = false
  let bookmarkedAgents = []
  let selectedAgent = null
  let showChatModal = false

  // Chat state
  let messages = []
  let inputValue = ''

  const MAX_BOOKMARKS = 4

  onMount(() => {
    loadBookmarks()

    // Listen for bookmark changes from other components
    window.addEventListener('bookmarks-updated', handleBookmarksUpdated)

    return () => {
      window.removeEventListener('bookmarks-updated', handleBookmarksUpdated)
    }
  })

  function loadBookmarks() {
    try {
      const stored = localStorage.getItem('eye-in-the-sky-bookmarks')
      if (stored) {
        bookmarkedAgents = JSON.parse(stored)
      }
    } catch (e) {
      console.error('Failed to load bookmarks:', e)
      bookmarkedAgents = []
    }
  }

  function handleBookmarksUpdated(event) {
    loadBookmarks()
  }

  function toggleFAB() {
    isExpanded = !isExpanded
  }

  function handleAgentClick(agent) {
    selectedAgent = agent
    showChatModal = true
    isExpanded = false

    // Load messages for this agent/session
    // TODO: Implement message loading via LiveView
    messages = []
  }

  function closeChatModal() {
    showChatModal = false
    selectedAgent = null
    messages = []
    inputValue = ''
  }

  function handleSubmit() {
    if (!inputValue.trim() || !selectedAgent) return

    // Send message via LiveView
    live.pushEvent('send_direct_message', {
      session_id: selectedAgent.session_id,
      body: inputValue.trim()
    })

    // Add message to local display immediately (optimistic update)
    messages = [...messages, {
      body: inputValue.trim(),
      sender_role: 'user',
      inserted_at: new Date().toISOString()
    }]

    inputValue = ''
  }

  // Calculate position for each petal in the flower
  function getPetalPosition(index, total) {
    const angle = (index * (360 / total)) - 90 // Start from top
    const radius = 80 // Distance from center
    const x = Math.cos(angle * Math.PI / 180) * radius
    const y = Math.sin(angle * Math.PI / 180) * radius
    return { x, y }
  }

  function getAgentInitials(name) {
    if (!name) return '?'
    return name.split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase()
  }

  function getStatusColor(status) {
    switch(status) {
      case 'active': return '#10b981' // green
      case 'idle': return '#3b82f6'   // blue
      case 'working': return '#f59e0b' // orange
      default: return '#6b7280'        // gray
    }
  }
</script>

<style>
  .fab-container {
    position: fixed;
    bottom: 2rem;
    right: 2rem;
    z-index: 1000;
  }

  .fab-button {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: linear-gradient(135deg, #0f766e 0%, #14b8a6 100%);
    border: none;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15), 0 2px 6px rgba(0, 0, 0, 0.1);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-size: 24px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
  }

  .fab-button:hover {
    transform: scale(1.1);
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2), 0 3px 8px rgba(0, 0, 0, 0.15);
  }

  .fab-button.expanded {
    transform: rotate(45deg);
  }

  .fab-icon {
    transition: transform 0.3s ease;
  }

  .bookmark-count {
    position: absolute;
    top: -4px;
    right: -4px;
    background: #ef4444;
    color: white;
    border-radius: 50%;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 700;
    border: 2px solid white;
  }

  .petals-container {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    pointer-events: none;
  }

  .petal {
    position: absolute;
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: white;
    border: 2px solid #e5e7eb;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    opacity: 0;
    transform: translate(-50%, -50%) scale(0);
    pointer-events: none;
    font-size: 14px;
    font-weight: 600;
  }

  .petal.visible {
    opacity: 1;
    pointer-events: auto;
    transform: translate(-50%, -50%) scale(1);
  }

  .petal:hover {
    transform: translate(-50%, -50%) scale(1.15);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    z-index: 10;
  }

  .petal-initials {
    color: #374151;
    font-size: 16px;
    font-weight: 700;
    line-height: 1;
  }

  .petal-status {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-top: 4px;
    border: 1px solid white;
  }

  /* Chat Modal */
  .chat-modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
    animation: fadeIn 0.2s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .chat-modal {
    background: white;
    border-radius: 12px;
    width: 90%;
    max-width: 500px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    animation: slideUp 0.3s ease;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal-header {
    padding: 1.25rem;
    border-bottom: 1px solid #e5e7eb;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .modal-title {
    font-size: 1.125rem;
    font-weight: 700;
    color: #1f2937;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .modal-session-id {
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 0.75rem;
    color: #6b7280;
    background: #f3f4f6;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    color: #9ca3af;
    cursor: pointer;
    padding: 0.25rem;
    line-height: 1;
    transition: color 0.2s;
  }

  .close-btn:hover {
    color: #374151;
  }

  .modal-messages {
    flex: 1;
    overflow-y: auto;
    padding: 1rem;
    background: #f9fafb;
  }

  .modal-input-area {
    padding: 1rem;
    border-top: 1px solid #e5e7eb;
    background: white;
  }

  .modal-input-form {
    display: flex;
    gap: 0.5rem;
  }

  .modal-input {
    flex: 1;
    padding: 0.625rem 1rem;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 0.875rem;
    outline: none;
  }

  .modal-input:focus {
    border-color: #0f766e;
    box-shadow: 0 0 0 3px rgba(15, 118, 110, 0.1);
  }

  .modal-send-btn {
    padding: 0.625rem 1.25rem;
    background: #0f766e;
    color: white;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    font-size: 0.875rem;
    cursor: pointer;
    transition: background 0.2s;
  }

  .modal-send-btn:hover:not(:disabled) {
    background: #0d665f;
  }

  .modal-send-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .empty-messages {
    text-align: center;
    color: #9ca3af;
    padding: 2rem;
    font-size: 0.875rem;
  }

  .message {
    padding: 0.75rem;
    margin-bottom: 0.5rem;
    border-radius: 8px;
    font-size: 0.875rem;
    line-height: 1.5;
  }

  .message.user {
    background: #0f766e;
    color: white;
    margin-left: 2rem;
    text-align: right;
  }

  .message.agent {
    background: white;
    color: #1f2937;
    margin-right: 2rem;
    border: 1px solid #e5e7eb;
  }
</style>

<div class="fab-container">
  <!-- Petals (bookmarked agents) -->
  {#if bookmarkedAgents.length > 0}
    <div class="petals-container">
      {#each bookmarkedAgents as agent, index}
        {@const pos = getPetalPosition(index, bookmarkedAgents.length)}
        <div
          class="petal"
          class:visible={isExpanded}
          style="left: {pos.x}px; top: {pos.y}px; transition-delay: {index * 0.05}s;"
          on:click={() => handleAgentClick(agent)}
          on:keydown={(e) => e.key === 'Enter' && handleAgentClick(agent)}
          role="button"
          tabindex="0"
          title={agent.name || agent.session_id}
        >
          <span class="petal-initials">{getAgentInitials(agent.name)}</span>
          <div class="petal-status" style="background-color: {getStatusColor(agent.status)}"></div>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Main FAB Button -->
  <button
    class="fab-button"
    class:expanded={isExpanded}
    on:click={toggleFAB}
    title={bookmarkedAgents.length > 0 ? 'Open bookmarked agents' : 'No bookmarked agents'}
    disabled={bookmarkedAgents.length === 0}
  >
    <span class="fab-icon">💬</span>
    {#if bookmarkedAgents.length > 0}
      <span class="bookmark-count">{bookmarkedAgents.length}</span>
    {/if}
  </button>
</div>

<!-- Chat Modal -->
{#if showChatModal && selectedAgent}
  <div class="chat-modal-backdrop" on:click={closeChatModal}>
    <div class="chat-modal" on:click|stopPropagation>
      <div class="modal-header">
        <div class="modal-title">
          <span>{selectedAgent.name || 'Agent'}</span>
          <span class="modal-session-id">{selectedAgent.session_id.substring(0, 8)}</span>
        </div>
        <button class="close-btn" on:click={closeChatModal}>&times;</button>
      </div>

      <div class="modal-messages">
        {#if messages.length === 0}
          <div class="empty-messages">
            No messages yet. Start a conversation!
          </div>
        {:else}
          {#each messages as message}
            <div class="message {message.sender_role}">
              {message.body}
            </div>
          {/each}
        {/if}
      </div>

      <div class="modal-input-area">
        <form on:submit|preventDefault={handleSubmit} class="modal-input-form">
          <input
            type="text"
            class="modal-input"
            placeholder="Send message to agent..."
            bind:value={inputValue}
            autocomplete="off"
          />
          <button type="submit" class="modal-send-btn" disabled={!inputValue.trim()}>
            Send
          </button>
        </form>
      </div>
    </div>
  </div>
{/if}
