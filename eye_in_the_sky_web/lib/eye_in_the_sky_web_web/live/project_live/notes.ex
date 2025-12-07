defmodule EyeInTheSkyWebWeb.ProjectLive.Notes do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Projects
  alias EyeInTheSkyWeb.Notes
  alias EyeInTheSkyWeb.Repo
  import Ecto.Query

  @impl true
  def mount(%{"id" => id}, _session, socket) do
    project_id = String.to_integer(id)
    project = Projects.get_project!(project_id)
    |> Repo.preload([:agents, :commits])

    # Load tasks manually due to type mismatch
    tasks = Projects.get_project_tasks(project_id)

    socket =
      socket
      |> assign(:page_title, "Notes - #{project.name}")
      |> assign(:project, project)
      |> assign(:tasks, tasks)
      |> assign(:search_query, "")
      |> assign(:notes, [])
      |> load_notes()

    {:ok, socket}
  end

  @impl true
  def handle_event("search", %{"query" => query}, socket) do
    socket =
      socket
      |> assign(:search_query, query)
      |> load_notes()

    {:noreply, socket}
  end

  defp load_notes(socket) do
    project = socket.assigns.project
    agent_ids = Enum.map(project.agents, & &1.id)

    # Get all session IDs for agents in this project
    session_ids =
      from(s in EyeInTheSkyWeb.Sessions.Session,
        where: s.agent_id in ^agent_ids,
        select: s.id
      )
      |> Repo.all()

    query = socket.assigns.search_query

    notes = if query != "" and String.trim(query) != "" do
      Notes.search_notes(query, agent_ids)
    else
      from(n in EyeInTheSkyWeb.Notes.Note,
        where: (n.parent_type == "agent" and n.parent_id in ^agent_ids) or
               (n.parent_type == "session" and n.parent_id in ^session_ids),
        order_by: [desc: n.created_at]
      )
      |> Repo.all()
    end

    assign(socket, :notes, notes)
  end

  @impl true
  def render(assigns) do
    ~H"""
    <style>
      .collapse input[type="checkbox"] {
        appearance: none;
        cursor: pointer;
        width: 100%;
        height: 100%;
        position: absolute;
        opacity: 0;
        margin: 0;
      }
    </style>
    <.live_component module={EyeInTheSkyWebWeb.Components.Navbar} id="navbar" />

    <!-- GitHub-style Project Navigation -->
    <div class="border-b border-base-300 bg-base-100">
      <div class="px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between py-3">
          <!-- Project Name -->
          <div class="flex items-center gap-2">
            <svg class="w-5 h-5 text-base-content/60" fill="currentColor" viewBox="0 0 16 16">
              <path d="M2 2.5A2.5 2.5 0 0 1 4.5 0h8.75a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-.75.75h-2.5a.75.75 0 0 1 0-1.5h1.75v-2h-8a1 1 0 0 0-.714 1.7.75.75 0 1 1-1.072 1.05A2.495 2.495 0 0 1 2 11.5Zm10.5-1h-8a1 1 0 0 0-1 1v6.708A2.486 2.486 0 0 1 4.5 9h8ZM5 12.25a.25.25 0 0 1 .25-.25h3.5a.25.25 0 0 1 .25.25v3.25a.25.25 0 0 1-.4.2l-1.45-1.087a.249.249 0 0 0-.3 0L5.4 15.7a.25.25 0 0 1-.4-.2Z" />
            </svg>
            <h1 class="text-xl font-semibold text-base-content"><%= @project.name %></h1>
          </div>

          <!-- Stats -->
          <div class="flex items-center gap-4 text-sm">
            <span class="text-base-content/60"><%= length(@project.agents) %> agents</span>
            <span class="text-base-content/60"><%= length(@tasks) %> tasks</span>
            <span class="text-base-content/60"><%= length(@project.commits) %> commits</span>
          </div>
        </div>

        <!-- Navigation Tabs -->
        <div class="flex items-center gap-1 -mb-px">
          <a
            href={~p"/projects/#{@project.id}"}
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-transparent hover:border-base-content/20 text-sm text-base-content/60 hover:text-base-content transition-colors"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M1.75 1A1.75 1.75 0 0 0 0 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0 0 16 13.25v-8.5A1.75 1.75 0 0 0 14.25 3H7.5a.25.25 0 0 1-.2-.1l-.9-1.2C6.07 1.26 5.55 1 5 1H1.75Z" />
            </svg>
            Files
          </a>
          <a
            href={~p"/projects/#{@project.id}/agents"}
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-transparent hover:border-base-content/20 text-sm text-base-content/60 hover:text-base-content transition-colors"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0ZM1.5 8a6.5 6.5 0 1 0 13 0 6.5 6.5 0 0 0-13 0Zm7-3.25v2.992l2.028.812a.75.75 0 0 1-.557 1.392l-2.5-1A.751.751 0 0 1 7 8.25v-3.5a.75.75 0 0 1 1.5 0Z" />
            </svg>
            Agents
          </a>
          <a
            href={~p"/projects/#{@project.id}/prompts"}
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-transparent hover:border-base-content/20 text-sm text-base-content/60 hover:text-base-content transition-colors"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v9.5A1.75 1.75 0 0 1 14.25 13H8.06l-2.573 2.573A1.458 1.458 0 0 1 3 14.543V13H1.75A1.75 1.75 0 0 1 0 11.25Zm1.75-.25a.25.25 0 0 0-.25.25v9.5c0 .138.112.25.25.25h2a.75.75 0 0 1 .75.75v2.19l2.72-2.72a.749.749 0 0 1 .53-.22h6.5a.25.25 0 0 0 .25-.25v-9.5a.25.25 0 0 0-.25-.25Z" />
            </svg>
            Prompts
          </a>
          <a
            href={~p"/projects/#{@project.id}/tasks"}
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-transparent hover:border-base-content/20 text-sm text-base-content/60 hover:text-base-content transition-colors"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M2.5 1.75v11.5c0 .138.112.25.25.25h3.17a.75.75 0 0 1 .75.75V16L9.4 13.571c.13-.096.289-.196.601-.196h3.249a.25.25 0 0 0 .25-.25V1.75a.25.25 0 0 0-.25-.25H2.75a.25.25 0 0 0-.25.25Zm-1.5 0C1 .784 1.784 0 2.75 0h10.5C14.216 0 15 .784 15 1.75v11.5A1.75 1.75 0 0 1 13.25 15H10l-3.573 2.573A1.458 1.458 0 0 1 4 16.543V15H2.75A1.75 1.75 0 0 1 1 13.25Z" />
            </svg>
            Tasks
          </a>
          <a
            href={~p"/projects/#{@project.id}/notes"}
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-primary text-sm font-medium text-base-content"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v12.5A1.75 1.75 0 0 1 14.25 16H1.75A1.75 1.75 0 0 1 0 14.25Zm1.75-.25a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h12.5a.25.25 0 0 0 .25-.25V1.75a.25.25 0 0 0-.25-.25ZM3.5 4.75A.75.75 0 0 1 4.25 4h7.5a.75.75 0 0 1 0 1.5h-7.5A.75.75 0 0 1 3.5 4.75ZM4.25 7a.75.75 0 0 0 0 1.5h7.5a.75.75 0 0 0 0-1.5ZM3.5 10.75a.75.75 0 0 1 .75-.75h7.5a.75.75 0 0 1 0 1.5h-7.5a.75.75 0 0 1-.75-.75Z" />
            </svg>
            Notes
          </a>
        </div>
      </div>
    </div>

    <div class="px-4 sm:px-6 lg:px-8 py-8">
      <div class="max-w-6xl mx-auto">
        <!-- Search Input -->
        <div class="mb-6">
          <form phx-change="search" class="w-full">
            <input
              type="text"
              name="query"
              value={@search_query}
              placeholder="Search notes by content..."
              class="input input-bordered w-full"
              autocomplete="off"
            />
          </form>
        </div>

        <%= if length(@notes) > 0 do %>
          <!-- Notes Accordion -->
          <div class="join join-vertical w-full">
            <%= for note <- @notes do %>
              <div class="collapse collapse-arrow join-item border border-base-300">
                <!-- Collapse Title -->
                <input type="checkbox" />
                <div class="collapse-title flex items-center justify-between bg-base-100 hover:bg-base-100/80 transition-colors cursor-pointer">
                  <div class="flex items-center gap-3 flex-1">
                    <svg class="w-4 h-4 text-base-content/60 flex-shrink-0" fill="currentColor" viewBox="0 0 16 16">
                      <path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v12.5A1.75 1.75 0 0 1 14.25 16H1.75A1.75 1.75 0 0 1 0 14.25Zm1.75-.25a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h12.5a.25.25 0 0 0 .25-.25V1.75a.25.25 0 0 0-.25-.25ZM3.5 4.75A.75.75 0 0 1 4.25 4h7.5a.75.75 0 0 1 0 1.5h-7.5A.75.75 0 0 1 3.5 4.75ZM4.25 7a.75.75 0 0 0 0 1.5h7.5a.75.75 0 0 0 0-1.5ZM3.5 10.75a.75.75 0 0 1 .75-.75h7.5a.75.75 0 0 1 0 1.5h-7.5a.75.75 0 0 1-.75-.75Z" />
                    </svg>
                    <div class="flex flex-col gap-1">
                      <h3 class="font-semibold text-sm text-base-content"><%= note.title || extract_title(note.body) %></h3>
                      <div class="flex items-center gap-2 text-xs text-base-content/60">
                        <span class="font-mono"><%= String.slice(note.parent_id, 0..7) %></span>
                        <span>•</span>
                        <span><%= format_timestamp(note.created_at) %></span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Collapse Content -->
                <div class="collapse-content bg-base-50">
                  <pre class="whitespace-pre-wrap text-sm text-base-content p-0 font-mono leading-relaxed"><%= note.body %></pre>
                </div>
              </div>
            <% end %>
          </div>
        <% else %>
          <!-- Empty State -->
          <div class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-base-content/40" fill="currentColor" viewBox="0 0 16 16">
              <path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v12.5A1.75 1.75 0 0 1 14.25 16H1.75A1.75 1.75 0 0 1 0 14.25Zm1.75-.25a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h12.5a.25.25 0 0 0 .25-.25V1.75a.25.25 0 0 0-.25-.25ZM3.5 4.75A.75.75 0 0 1 4.25 4h7.5a.75.75 0 0 1 0 1.5h-7.5A.75.75 0 0 1 3.5 4.75ZM4.25 7a.75.75 0 0 0 0 1.5h7.5a.75.75 0 0 0 0-1.5ZM3.5 10.75a.75.75 0 0 1 .75-.75h7.5a.75.75 0 0 1 0 1.5h-7.5a.75.75 0 0 1-.75-.75Z" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-base-content">No notes yet</h3>
            <p class="mt-1 text-sm text-base-content/60">
              Notes from agents will appear here
            </p>
          </div>
        <% end %>

      </div>
    </div>
    """
  end

  defp format_timestamp(nil), do: ""
  defp format_timestamp(timestamp) when is_binary(timestamp) do
    # Timestamp is stored as string, just display it as-is or format as needed
    timestamp
  end

  defp extract_title(body) when is_nil(body), do: "Untitled"
  defp extract_title(body) when is_binary(body) do
    body
    |> String.trim()
    |> String.split("\n")
    |> List.first()
    |> String.slice(0..50)
    |> then(fn text ->
      if String.length(text) >= 50, do: text <> "...", else: text
    end)
  end
end
