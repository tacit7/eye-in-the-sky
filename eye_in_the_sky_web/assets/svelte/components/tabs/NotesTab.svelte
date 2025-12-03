<script>
  import { marked } from 'marked'
  import hljs from 'highlight.js'

  export let notes = []

  // Configure marked with syntax highlighting
  marked.setOptions({
    gfm: true,           // GitHub Flavored Markdown
    breaks: true,        // Convert \n to <br>
    headerIds: true,     // Add IDs to headers
    mangle: false,       // Don't mangle email addresses
    pedantic: false,     // Use original markdown.pl behavior
    highlight: function(code, lang) {
      if (lang && hljs.getLanguage(lang)) {
        try {
          return hljs.highlight(code, { language: lang }).value
        } catch (err) {
          console.error('Highlight error:', err)
        }
      }
      return hljs.highlightAuto(code).value
    }
  })

  function formatDate(dateStr) {
    if (!dateStr) return ''
    try {
      const date = new Date(dateStr)
      return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    } catch (e) {
      return dateStr
    }
  }

  function renderMarkdown(content) {
    try {
      return marked.parse(content || '')
    } catch (e) {
      console.error('Markdown parse error:', e)
      return content || ''
    }
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
    return id ? id.substring(0, 8) : ''
  }
</script>

{#if notes && notes.length > 0}
  <div class="space-y-4">
    {#each notes as note (note.id)}
      <div class="card bg-base-100 border border-base-300 shadow-sm hover:shadow-md transition-shadow">
        <div class="card-body p-4">
          <!-- Markdown content -->
          <div class="prose prose-sm max-w-none dark:prose-invert
                      prose-headings:font-semibold
                      prose-h1:text-2xl prose-h1:mb-3
                      prose-h2:text-xl prose-h2:mb-2
                      prose-h3:text-lg prose-h3:mb-2
                      prose-p:mb-2
                      prose-ul:mb-2 prose-ol:mb-2
                      prose-li:mb-1
                      prose-code:bg-base-200 prose-code:px-1 prose-code:py-0.5 prose-code:rounded
                      prose-pre:bg-base-300 prose-pre:p-3 prose-pre:rounded-lg
                      prose-blockquote:border-l-4 prose-blockquote:border-primary prose-blockquote:pl-4">
            {@html renderMarkdown(note.body)}
          </div>

          <!-- Footer with ID and Timestamp -->
          <div class="card-actions justify-between mt-3 pt-3 border-t border-base-300">
            <!-- Note ID Badge -->
            <button
              class="badge badge-ghost badge-sm hover:badge-primary cursor-pointer font-mono transition-colors"
              on:click={(e) => copyToClipboard(note.id, e)}
              title="Copy ID: {note.id}"
            >
              #{getShortId(note.id)}
            </button>

            <!-- Timestamp -->
            {#if note.created_at}
              <div class="badge badge-ghost badge-sm">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {formatDate(note.created_at)}
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/each}
  </div>
{:else}
  <!-- Empty state -->
  <div class="hero min-h-[400px] bg-base-200 rounded-lg">
    <div class="hero-content text-center">
      <div class="max-w-md">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mx-auto mb-4 text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <h2 class="text-2xl font-bold text-base-content">No notes yet</h2>
        <p class="py-3 text-base-content/70">
          Add a note to capture decisions, blockers, or next steps for this session.
        </p>
        <button class="btn btn-primary btn-sm">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Add Your First Note
        </button>
      </div>
    </div>
  </div>
{/if}
