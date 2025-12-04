defmodule EyeInTheSkyWebWeb.ProjectLive.Tasks do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Projects
  alias EyeInTheSkyWeb.Tasks
  alias EyeInTheSkyWeb.Repo

  @impl true
  def mount(%{"id" => id}, _session, socket) do
    project_id = String.to_integer(id)
    project = Projects.get_project!(project_id)
    |> Repo.preload([:agents, :commits])

    socket =
      socket
      |> assign(:page_title, "Tasks - #{project.name}")
      |> assign(:project, project)
      |> assign(:project_id, project_id)
      |> assign(:search_query, "")
      |> assign(:tasks, [])
      |> load_tasks()

    {:ok, socket}
  end

  @impl true
  def handle_event("search", %{"query" => query}, socket) do
    socket =
      socket
      |> assign(:search_query, query)
      |> load_tasks()

    {:noreply, socket}
  end

  defp load_tasks(socket) do
    project_id = socket.assigns.project_id
    query = socket.assigns.search_query

    tasks = if query != "" and String.trim(query) != "" do
      Tasks.search_tasks(query, project_id)
    else
      Projects.get_project_tasks(project_id)
    end

    assign(socket, :tasks, tasks)
  end

  @impl true
  def render(assigns) do
    ~H"""
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
              <path d="M2 5.5a3.5 3.5 0 1 1 5.898 2.549 5.508 5.508 0 0 1 3.034 4.084.75.75 0 1 1-1.482.235 4 4 0 0 0-7.9 0 .75.75 0 0 1-1.482-.236A5.507 5.507 0 0 1 3.102 8.05 3.493 3.493 0 0 1 2 5.5ZM11 4a3.001 3.001 0 0 1 2.22 5.018 5.01 5.01 0 0 1 2.56 3.012.749.749 0 0 1-.885.954.752.752 0 0 1-.549-.514 3.507 3.507 0 0 0-2.522-2.372.75.75 0 0 1-.574-.73v-.352a.75.75 0 0 1 .416-.672A1.5 1.5 0 0 0 11 5.5.75.75 0 0 1 11 4Z" />
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
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-primary text-sm text-base-content transition-colors"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M2.5 1.75v11.5c0 .138.112.25.25.25h3.17a.75.75 0 0 1 .75.75V16L9.4 13.571c.13-.096.289-.196.601-.196h3.249a.25.25 0 0 0 .25-.25V1.75a.25.25 0 0 0-.25-.25H2.75a.25.25 0 0 0-.25.25Zm-1.5 0C1 .784 1.784 0 2.75 0h10.5C14.216 0 15 .784 15 1.75v11.5A1.75 1.75 0 0 1 13.25 15H10l-3.573 2.573A1.458 1.458 0 0 1 4 16.543V15H2.75A1.75 1.75 0 0 1 1 13.25Z" />
            </svg>
            Tasks
          </a>
          <a
            href={~p"/projects/#{@project.id}/kanban"}
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-transparent hover:border-base-content/20 text-sm text-base-content/60 hover:text-base-content transition-colors"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M0 1.75C0 .784.784 0 1.75 0h2.5C5.216 0 6 .784 6 1.75v12.5A1.75 1.75 0 0 1 4.25 16h-2.5A1.75 1.75 0 0 1 0 14.25ZM7 1.75C7 .784 7.784 0 8.75 0h2.5C12.216 0 13 .784 13 1.75v7.5A1.75 1.75 0 0 1 11.25 11h-2.5A1.75 1.75 0 0 1 7 9.25ZM14 1.75C14 .784 14.784 0 15.75 0h.5c.966 0 1.75.784 1.75 1.75v4.5A1.75 1.75 0 0 1 16.25 8h-.5A1.75 1.75 0 0 1 14 6.25Z" />
            </svg>
            Kanban
          </a>
          <a
            href={~p"/projects/#{@project.id}/notes"}
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-transparent hover:border-base-content/20 text-sm text-base-content/60 hover:text-base-content transition-colors"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v12.5A1.75 1.75 0 0 1 14.25 16H1.75A1.75 1.75 0 0 1 0 14.25Zm1.75-.25a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h12.5a.25.25 0 0 0 .25-.25V1.75a.25.25 0 0 0-.25-.25ZM3.5 4.75A.75.75 0 0 1 4.25 4h7.5a.75.75 0 0 1 0 1.5h-7.5A.75.75 0 0 1 3.5 4.75ZM4.25 7a.75.75 0 0 0 0 1.5h7.5a.75.75 0 0 0 0-1.5ZM3.5 10.75a.75.75 0 0 1 .75-.75h7.5a.75.75 0 0 1 0 1.5h-7.5a.75.75 0 0 1-.75-.75Z" />
            </svg>
            Notes
          </a>
        </div>
      </div>
    </div>

      <!-- Search -->
      <div class="px-4 sm:px-6 lg:px-8 py-6">
        <form phx-change="search" class="w-full max-w-2xl">
          <input
            type="text"
            name="query"
            value={@search_query}
            placeholder="Search tasks by title or description..."
            class="input input-bordered w-full"
            autocomplete="off"
          />
        </form>
      </div>

      <!-- Tasks Grid -->
      <div class="px-4 sm:px-6 lg:px-8 pb-8">
        <%= if length(@tasks) > 0 do %>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            <%= for task <- @tasks do %>
              <div class="card bg-base-200 border border-base-300 hover:border-primary hover:shadow-lg transition-all cursor-pointer group">
                <div class="card-body p-5">
                  <!-- Task Header with Priority -->
                  <div class="flex items-start justify-between gap-2 mb-3">
                    <div class="flex-1 min-w-0">
                      <h3 class="text-base font-semibold text-base-content group-hover:text-primary transition-colors line-clamp-2">
                        <%= task.title %>
                      </h3>
                    </div>
                    <%= cond do %>
                      <% task.priority >= 70 -> %>
                        <span class="badge badge-error badge-sm flex-shrink-0">High</span>
                      <% task.priority >= 40 -> %>
                        <span class="badge badge-warning badge-sm flex-shrink-0">Med</span>
                      <% task.priority >= 20 -> %>
                        <span class="badge badge-info badge-sm flex-shrink-0">Low</span>
                      <% true -> %>
                        <span></span>
                    <% end %>
                  </div>

                  <!-- Description -->
                  <%= if task.description do %>
                    <p class="text-sm text-base-content/60 line-clamp-3 mb-3">
                      <%= task.description %>
                    </p>
                  <% end %>

                  <!-- Task Metadata -->
                  <div class="flex flex-wrap items-center gap-2 mt-auto pt-3 border-t border-base-300">
                    <%= if task.state do %>
                      <span class="badge badge-ghost badge-sm">
                        <%= task.state.name %>
                      </span>
                    <% end %>

                    <%= if task.due_at do %>
                      <span class="badge badge-ghost badge-sm">
                        <svg class="w-3 h-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                        <%= format_due_date(task.due_at) %>
                      </span>
                    <% end %>

                    <%= if task.tags && length(task.tags) > 0 do %>
                      <%= for tag <- Enum.take(task.tags, 2) do %>
                        <span class="badge badge-outline badge-sm">
                          <%= tag.name %>
                        </span>
                      <% end %>
                      <%= if length(task.tags) > 2 do %>
                        <span class="badge badge-ghost badge-sm">
                          +<%= length(task.tags) - 2 %>
                        </span>
                      <% end %>
                    <% end %>
                  </div>

                  <!-- Completed Indicator -->
                  <%= if task.completed_at do %>
                    <div class="absolute top-3 right-3">
                      <svg class="w-5 h-5 text-success" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                    </div>
                  <% end %>
                </div>
              </div>
            <% end %>
          </div>
        <% else %>
          <!-- Empty State -->
          <div class="text-center py-16">
            <div class="mx-auto w-24 h-24 bg-base-200 rounded-full flex items-center justify-center mb-4">
              <svg class="w-12 h-12 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-base-content mb-2">
              <%= if @search_query != "" do %>
                No tasks found
              <% else %>
                No tasks yet
              <% end %>
            </h3>
            <p class="text-sm text-base-content/60">
              <%= if @search_query != "" do %>
                Try adjusting your search query
              <% else %>
                Create a task to get started
              <% end %>
            </p>
          </div>
        <% end %>
      </div>
    """
  end

  defp format_due_date(nil), do: nil
  defp format_due_date(date_string) when is_binary(date_string) do
    case Date.from_iso8601(String.slice(date_string, 0..9)) do
      {:ok, date} ->
        today = Date.utc_today()
        cond do
          Date.compare(date, today) == :eq -> "Today"
          Date.compare(date, Date.add(today, 1)) == :eq -> "Tomorrow"
          Date.compare(date, today) == :lt -> "Overdue"
          true -> Calendar.strftime(date, "%b %d")
        end
      _ -> nil
    end
  end
end
