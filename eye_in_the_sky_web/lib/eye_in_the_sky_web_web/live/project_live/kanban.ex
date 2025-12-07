defmodule EyeInTheSkyWebWeb.ProjectLive.Kanban do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Projects
  alias EyeInTheSkyWeb.Tasks
  alias EyeInTheSkyWeb.Repo

  @impl true
  def mount(%{"id" => id}, _session, socket) do
    project_id = String.to_integer(id)
    project = Projects.get_project!(project_id)
    |> Repo.preload([:agents, :commits])

    # Load workflow states
    workflow_states = Tasks.list_workflow_states()

    socket =
      socket
      |> assign(:page_title, "Kanban - #{project.name}")
      |> assign(:project, project)
      |> assign(:project_id, project_id)
      |> assign(:search_query, "")
      |> assign(:workflow_states, workflow_states)
      |> assign(:tasks, [])
      |> assign(:tasks_by_state, %{})
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

    # Group tasks by state for kanban view
    tasks_by_state = Enum.group_by(tasks, fn task ->
      if task.state, do: task.state.id, else: nil
    end)

    socket
    |> assign(:tasks, tasks)
    |> assign(:tasks_by_state, tasks_by_state)
  end

  @impl true
  def render(assigns) do
    ~H"""
    <.live_component module={EyeInTheSkyWebWeb.Components.Navbar} id="navbar" current_project={@project} />

    <EyeInTheSkyWebWeb.Components.ProjectNav.render
      project={@project}
      tasks={@tasks}
      current_tab={:kanban}
    />

    <div class="px-4 sm:px-6 lg:px-8 py-8">
      <!-- Search Input -->
      <div class="max-w-7xl mx-auto mb-6">
        <form phx-change="search" class="w-full max-w-md">
          <input
            type="text"
            name="query"
            value={@search_query}
            placeholder="Search tasks..."
            class="input input-bordered w-full input-sm"
            autocomplete="off"
          />
        </form>
      </div>

      <%= if length(@tasks) > 0 do %>
        <!-- Kanban Board -->
        <div class="overflow-x-auto pb-4">
          <div class="inline-flex gap-4 min-w-full px-4">
            <%= for state <- @workflow_states do %>
              <div class="flex-shrink-0 w-80">
                <!-- Column Header -->
                <div class="bg-base-200 rounded-t-lg px-4 py-3">
                  <div class="flex items-center justify-between">
                    <h3 class="font-semibold text-sm text-base-content">
                      <%= state.name %>
                    </h3>
                    <span class="badge badge-sm">
                      <%= length(Map.get(@tasks_by_state, state.id, [])) %>
                    </span>
                  </div>
                </div>

                <!-- Column Content -->
                <div class="bg-base-100 rounded-b-lg border border-base-300 border-t-0 p-3 min-h-[600px]">
                  <div class="space-y-3">
                    <%= for task <- Map.get(@tasks_by_state, state.id, []) do %>
                      <!-- Task Card -->
                      <div class="card bg-base-100 border border-base-300 hover:border-primary hover:shadow-md transition-all cursor-pointer">
                        <div class="card-body p-3">
                          <!-- Task Title -->
                          <div class="flex items-start gap-2 mb-2">
                            <button class={
                              "mt-0.5 flex-shrink-0 w-4 h-4 rounded-sm border-2 transition-all " <>
                              if task.completed_at do
                                "bg-success border-success"
                              else
                                "border-base-content/30 hover:border-base-content/60"
                              end
                            }>
                              <%= if task.completed_at do %>
                                <svg class="w-full h-full text-white p-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                                </svg>
                              <% end %>
                            </button>
                            <h4 class={
                              "text-sm font-medium flex-1 " <>
                              if task.completed_at do
                                "text-base-content/50 line-through"
                              else
                                "text-base-content"
                              end
                            }>
                              <%= task.title %>
                            </h4>
                            <%= if task.priority && task.priority > 0 do %>
                              <span class={priority_class(task.priority)}>
                                <%= priority_label(task.priority) %>
                              </span>
                            <% end %>
                          </div>

                          <!-- Description -->
                          <%= if task.description do %>
                            <p class="text-xs text-base-content/60 line-clamp-2 mb-2">
                              <%= task.description %>
                            </p>
                          <% end %>

                          <!-- Meta Info -->
                          <div class="flex items-center gap-2 text-xs text-base-content/50 flex-wrap">
                            <%= if task.due_at do %>
                              <span class="flex items-center gap-1">
                                <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 16 16">
                                  <path d="M3.5 0a.5.5 0 0 1 .5.5V1h8V.5a.5.5 0 0 1 1 0V1h1a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2V3a2 2 0 0 1 2-2h1V.5a.5.5 0 0 1 .5-.5zM1 4v10a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1V4H1z" />
                                </svg>
                                <%= format_date(task.due_at) %>
                              </span>
                            <% end %>
                            <%= if task.agent_id do %>
                              <span class="flex items-center gap-1 font-mono text-xs">
                                <%= String.slice(task.agent_id, 0..7) %>
                              </span>
                            <% end %>
                          </div>
                        </div>
                      </div>
                    <% end %>
                  </div>
                </div>
              </div>
            <% end %>
          </div>
        </div>
      <% else %>
        <!-- Empty State -->
        <div class="max-w-6xl mx-auto">
          <div class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-base-content/40" fill="currentColor" viewBox="0 0 16 16">
              <path d="M2.5 1.75v11.5c0 .138.112.25.25.25h3.17a.75.75 0 0 1 .75.75V16L9.4 13.571c.13-.096.289-.196.601-.196h3.249a.25.25 0 0 0 .25-.25V1.75a.25.25 0 0 0-.25-.25H2.75a.25.25 0 0 0-.25.25Zm-1.5 0C1 .784 1.784 0 2.75 0h10.5C14.216 0 15 .784 15 1.75v11.5A1.75 1.75 0 0 1 13.25 15H10l-3.573 2.573A1.458 1.458 0 0 1 4 16.543V15H2.75A1.75 1.75 0 0 1 1 13.25Z" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-base-content">No tasks yet</h3>
            <p class="mt-1 text-sm text-base-content/60">
              Tasks will appear here when agents create them for this project
            </p>
          </div>
        </div>
      <% end %>

    </div>
    """
  end

  defp priority_class(priority) do
    cond do
      priority >= 4 -> "text-xs text-error"
      priority >= 3 -> "text-xs text-warning"
      priority >= 2 -> "text-xs text-info"
      true -> "text-xs text-base-content/40"
    end
  end

  defp priority_label(priority) do
    cond do
      priority >= 4 -> "P1"
      priority >= 3 -> "P2"
      priority >= 2 -> "P3"
      true -> "P#{priority}"
    end
  end

  defp format_date(nil), do: ""
  defp format_date(datetime) do
    today = Date.utc_today()
    date = NaiveDateTime.to_date(datetime)

    cond do
      Date.compare(date, today) == :eq -> "Today"
      Date.compare(date, Date.add(today, 1)) == :eq -> "Tomorrow"
      true -> Calendar.strftime(datetime, "%b %d")
    end
  end
end
