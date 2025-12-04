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
    <div class="min-h-screen bg-base-100">
      <!-- Header -->
      <div class="border-b border-base-300 bg-base-200">
        <div class="px-4 sm:px-6 lg:px-8 py-4">
          <div class="flex items-center justify-between">
            <div>
              <h1 class="text-2xl font-bold text-base-content"><%= @project.name %> Tasks</h1>
              <p class="text-sm text-base-content/60 mt-1"><%= length(@tasks) %> tasks</p>
            </div>
            <.link navigate={~p"/projects/#{@project.id}"} class="btn btn-ghost btn-sm">
              <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              Back to Project
            </.link>
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
                    <%= if task.priority >= 70 do %>
                      <span class="badge badge-error badge-sm flex-shrink-0">High</span>
                    <% else if task.priority >= 40 do %>
                      <span class="badge badge-warning badge-sm flex-shrink-0">Med</span>
                    <% else if task.priority >= 20 do %>
                      <span class="badge badge-info badge-sm flex-shrink-0">Low</span>
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
