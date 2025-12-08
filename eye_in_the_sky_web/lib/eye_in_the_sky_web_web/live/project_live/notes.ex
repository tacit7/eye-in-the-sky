defmodule EyeInTheSkyWebWeb.ProjectLive.Notes do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Projects
  alias EyeInTheSkyWeb.Notes
  alias EyeInTheSkyWeb.Repo
  import Ecto.Query

  @impl true
  def mount(%{"id" => id}, _session, socket) do
    # Parse project ID safely
    project_id = case Integer.parse(id) do
      {int, ""} -> int
      _ -> nil
    end

    socket = if project_id do
      project = Projects.get_project!(project_id)
      |> Repo.preload([:agents, :commits])

      # Load tasks manually due to type mismatch
      tasks = Projects.get_project_tasks(project_id)

      socket
      |> assign(:page_title, "Notes - #{project.name}")
      |> assign(:project, project)
      |> assign(:tasks, tasks)
      |> assign(:search_query, "")
      |> assign(:notes, [])
      |> load_notes()
    else
      socket
      |> assign(:page_title, "Project Not Found")
      |> assign(:project, nil)
      |> assign(:tasks, [])
      |> assign(:search_query, "")
      |> assign(:notes, [])
      |> put_flash(:error, "Invalid project ID")
    end

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
    <.live_component module={EyeInTheSkyWebWeb.Components.Navbar} id="navbar" current_project={@project} />

    <EyeInTheSkyWebWeb.Components.ProjectNav.render
      project={@project}
      tasks={@tasks}
      current_tab={:notes}
    />

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
                  <pre id={"note-highlight-#{note.id}"} phx-hook="Highlight" class="whitespace-pre-wrap text-sm text-base-content p-0 font-mono leading-relaxed"><code class="language-plaintext"><%= note.body %></code></pre>
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
