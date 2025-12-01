defmodule EyeInTheSkyWebWeb.AgentLive.Index do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Agents
  alias Phoenix.LiveView.JS
  import EyeInTheSkyWebWeb.Helpers.ViewHelpers

  @stale_threshold_hours 24

  @impl true
  def mount(_params, _session, socket) do
    agents = Agents.list_agents_with_sessions()

    socket =
      socket
      |> assign(:page_title, "Eye in the Sky - Agents")
      |> assign(:agents, agents)
      |> assign(:search_query, "")
      |> assign(:status_filter, "all")
      |> assign(:sort_by, "updated_desc")

    {:ok, socket}
  end

  @impl true
  def handle_params(params, _url, socket) do
    {:noreply, apply_action(socket, socket.assigns.live_action, params)}
  end

  @impl true
  def handle_event("search", %{"query" => query}, socket) do
    {:noreply, assign(socket, :search_query, query)}
  end

  @impl true
  def handle_event("filter_status", %{"status" => status}, socket) do
    {:noreply, assign(socket, :status_filter, status)}
  end

  @impl true
  def handle_event("sort", %{"by" => sort_by}, socket) do
    {:noreply, assign(socket, :sort_by, sort_by)}
  end

  defp apply_action(socket, :index, _params) do
    socket
    |> assign(:page_title, "Listing Agents")
  end

  @impl true
  def render(assigns) do
    # Apply filters and search
    assigns = assign(assigns, :filtered_agents, filter_and_sort_agents(assigns))

    ~H"""
    <div class="px-4 sm:px-6 lg:px-8">
      <div class="sm:flex sm:items-center sm:justify-between">
        <div class="sm:flex-auto">
          <h1 class="text-base font-semibold leading-6 text-gray-900 dark:text-gray-100">Agents</h1>
          <p class="mt-2 text-sm text-gray-700 dark:text-gray-400">
            Real-time overview of all Claude Code agents
          </p>
        </div>
      </div>

      <!-- Search and Filters -->
      <div class="mt-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:gap-6">
        <!-- Search -->
        <div class="flex-1 max-w-md">
          <label for="search" class="sr-only">Search sessions</label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
              <svg class="h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z" clip-rule="evenodd" />
              </svg>
            </div>
            <input
              type="text"
              name="search"
              id="search"
              phx-keyup="search"
              phx-debounce="300"
              value={@search_query}
              class="input input-bordered w-full pl-10"
              placeholder="Search sessions, projects, descriptions..."
            />
          </div>
        </div>

        <!-- Status Filter -->
        <div class="btn-group">
          <button
            phx-click="filter_status"
            phx-value-status="all"
            class={"btn btn-sm #{if @status_filter == "all", do: "btn-active"}"}
          >
            All
          </button>
          <button
            phx-click="filter_status"
            phx-value-status="active"
            class={"btn btn-sm #{if @status_filter == "active", do: "btn-active"}"}
          >
            Active
          </button>
          <button
            phx-click="filter_status"
            phx-value-status="completed"
            class={"btn btn-sm #{if @status_filter == "completed", do: "btn-active"}"}
          >
            Completed
          </button>
          <button
            phx-click="filter_status"
            phx-value-status="stale"
            class={"btn btn-sm #{if @status_filter == "stale", do: "btn-active"}"}
          >
            Stale
          </button>
        </div>
      </div>

      <div class="mt-6 flow-root">
        <div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
          <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
            <div class="overflow-hidden rounded-lg bg-white dark:bg-gray-900 ring-1 ring-gray-200 dark:ring-gray-700">
              <table class="min-w-full divide-y divide-gray-300 dark:divide-gray-700">
                <thead class="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10">
                  <tr>
                    <th scope="col" class="w-32 py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 dark:text-gray-100 sm:pl-4">
                      Status
                    </th>
                    <th scope="col" class="w-48 px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-100">
                      Session
                    </th>
                    <th scope="col" class="w-56 px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-100">
                      Project
                    </th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-100">
                      Description
                    </th>
                    <th scope="col" class="w-32 py-3.5 pl-3 pr-4 text-left text-sm font-semibold text-gray-900 dark:text-gray-100 sm:pr-4">
                      Last Active
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                  <%= if @filtered_agents == [] do %>
                    <tr>
                      <td colspan="5" class="py-12 text-center text-sm text-gray-500 dark:text-gray-400">
                        No agents found matching your criteria.
                      </td>
                    </tr>
                  <% else %>
                    <%= for agent <- @filtered_agents do %>
                      <tr
                        phx-click={JS.navigate(~p"/agents/#{agent.id}")}
                        class="cursor-pointer hover:bg-gray-50 dark:hover:bg-white/5 transition-colors group"
                      >
                        <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm sm:pl-4">
                          {render_status_badge(assigns, agent)}
                        </td>
                        <td class="px-3 py-4 text-sm">
                          <div class="flex items-center gap-2">
                            <span class="font-mono font-semibold text-gray-900 dark:text-gray-100">
                              {agent.session_id && String.slice(agent.session_id, 0..7) || "—"}
                            </span>
                            <%= if agent.session_id do %>
                              <button
                                id={"copy-btn-#{agent.session_id}"}
                                type="button"
                                phx-hook="CopySessionId"
                                data-session-id={agent.session_id}
                                class="relative text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                                aria-label="Copy full session ID"
                              >
                                <.icon name="hero-clipboard" class="h-4 w-4" />
                              </button>
                            <% end %>
                          </div>
                          {render_agent_meta(assigns, agent)}
                        </td>
                        <td class="px-3 py-4 text-sm">
                          {render_project_badge(assigns, agent.project_name)}
                        </td>
                        <td class="px-3 py-4 text-sm text-gray-600 dark:text-gray-400">
                          <div class="line-clamp-1" title={agent.description || agent.feature_description}>
                            {agent.description || agent.feature_description || "—"}
                          </div>
                        </td>
                        <td class="whitespace-nowrap py-4 pl-3 pr-4 sm:pr-4">
                          <div class="flex items-center justify-between gap-2">
                            <span class="text-sm text-gray-500 dark:text-gray-400" title={format_datetime_full(agent.last_activity_at)}>
                              {relative_time(agent.last_activity_at)}
                            </span>
                            <svg class="h-5 w-5 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity" viewBox="0 0 20 20" fill="currentColor">
                              <path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd" />
                            </svg>
                          </div>
                        </td>
                      </tr>
                    <% end %>
                  <% end %>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
    """
  end


  defp filter_and_sort_agents(assigns) do
    agents = assigns.agents
    query = String.downcase(assigns.search_query)
    status_filter = assigns.status_filter

    agents
    |> Enum.filter(fn agent ->
      # Search filter
      search_match =
        if query == "" do
          true
        else
          String.contains?(String.downcase(agent.session_id || ""), query) ||
            String.contains?(String.downcase(agent.description || ""), query) ||
            String.contains?(String.downcase(agent.feature_description || ""), query) ||
            String.contains?(String.downcase(agent.project_name || ""), query)
        end

      # Status filter
      status_match =
        case status_filter do
          "all" -> true
          "stale" -> is_stale?(agent, @stale_threshold_hours)
          status -> agent.status == status
        end

      search_match && status_match
    end)
    |> sort_agents(assigns.sort_by)
  end

  defp sort_agents(agents, "updated_desc") do
    Enum.sort_by(agents, &parse_last_activity/1, {:desc, DateTime})
  end

  defp parse_last_activity(%{last_activity_at: nil}), do: ~U[1970-01-01 00:00:00Z]
  defp parse_last_activity(%{last_activity_at: %DateTime{} = dt}), do: dt
  defp parse_last_activity(%{last_activity_at: str}) when is_binary(str) do
    case parse_datetime(str) do
      {:ok, dt} -> dt
      :error -> ~U[1970-01-01 00:00:00Z]
    end
  end

  defp render_agent_meta(assigns, agent) do
    # Count sessions and tasks from preloaded associations
    session_count = if agent.sessions, do: length(agent.sessions), else: 0
    task_count = if agent.tasks, do: length(agent.tasks), else: 0

    assigns = Map.merge(assigns, %{
      session_count: session_count,
      task_count: task_count,
      last_active: relative_time(agent.last_activity_at)
    })

    ~H"""
    <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400 font-normal">
      <span class="inline-flex items-center gap-1.5">
        <span>Session #{@session_count || "—"}</span>
        <span class="text-gray-400">·</span>
        <span>{@last_active}</span>
        <%= if @task_count > 0 do %>
          <span class="text-gray-400">·</span>
          <span>{@task_count} tasks</span>
        <% end %>
      </span>
    </div>
    """
  end
end
