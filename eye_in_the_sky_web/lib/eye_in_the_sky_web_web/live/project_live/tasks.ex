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

    # Load workflow states
    workflow_states = Tasks.list_workflow_states()

    socket =
      socket
      |> assign(:page_title, "Tasks - #{project.name}")
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
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-primary text-sm font-medium text-base-content"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M2.5 1.75v11.5c0 .138.112.25.25.25h3.17a.75.75 0 0 1 .75.75V16L9.4 13.571c.13-.096.289-.196.601-.196h3.249a.25.25 0 0 0 .25-.25V1.75a.25.25 0 0 0-.25-.25H2.75a.25.25 0 0 0-.25.25Zm-1.5 0C1 .784 1.784 0 2.75 0h10.5C14.216 0 15 .784 15 1.75v11.5A1.75 1.75 0 0 1 13.25 15H10l-3.573 2.573A1.458 1.458 0 0 1 4 16.543V15H2.75A1.75 1.75 0 0 1 1 13.25Z" />
            </svg>
            Tasks
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
