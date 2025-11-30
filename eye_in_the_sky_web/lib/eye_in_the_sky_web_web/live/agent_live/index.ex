defmodule EyeInTheSkyWebWeb.AgentLive.Index do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Agents
  alias Phoenix.LiveView.JS

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
          <label for="search" class="sr-only">Search agents</label>
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
              class="block w-full rounded-md border-0 py-1.5 pl-10 pr-3 text-gray-900 dark:text-gray-100 bg-white dark:bg-gray-800 ring-1 ring-inset ring-gray-300 dark:ring-gray-600 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
              placeholder="Search agents, projects, descriptions..."
            />
          </div>
        </div>

        <!-- Status Filter -->
        <div class="flex gap-2">
          <button
            phx-click="filter_status"
            phx-value-status="all"
            class={filter_button_class(@status_filter == "all")}
          >
            All
          </button>
          <button
            phx-click="filter_status"
            phx-value-status="active"
            class={filter_button_class(@status_filter == "active")}
          >
            Active
          </button>
          <button
            phx-click="filter_status"
            phx-value-status="completed"
            class={filter_button_class(@status_filter == "completed")}
          >
            Completed
          </button>
          <button
            phx-click="filter_status"
            phx-value-status="stale"
            class={filter_button_class(@status_filter == "stale")}
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
                      Agent
                    </th>
                    <th scope="col" class="w-56 px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-100">
                      Project
                    </th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-gray-100">
                      Description
                    </th>
                    <th scope="col" class="w-32 py-3.5 pl-3 pr-4 text-left text-sm font-semibold text-gray-900 dark:text-gray-100 sm:pr-4">
                      Updated
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
                          <div class="font-mono font-semibold text-gray-900 dark:text-gray-100">
                            {String.slice(agent.id, 0..7)}
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
                            <span class="text-sm text-gray-500 dark:text-gray-400" title={format_datetime_full(agent.updated_at)}>
                              {relative_time(agent.updated_at)}
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

  defp filter_button_class(active) do
    base = "px-3 py-1.5 text-sm font-medium rounded-md transition-colors"
    if active do
      "#{base} bg-indigo-600 text-white"
    else
      "#{base} bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700"
    end
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
          String.contains?(String.downcase(agent.id || ""), query) ||
            String.contains?(String.downcase(agent.description || ""), query) ||
            String.contains?(String.downcase(agent.feature_description || ""), query) ||
            String.contains?(String.downcase(agent.project_name || ""), query)
        end

      # Status filter
      status_match =
        case status_filter do
          "all" -> true
          "stale" -> is_stale?(agent)
          status -> agent.status == status
        end

      search_match && status_match
    end)
    |> sort_agents(assigns.sort_by)
  end

  defp sort_agents(agents, "updated_desc") do
    Enum.sort_by(agents, &parse_updated_at/1, {:desc, DateTime})
  end

  defp parse_updated_at(%{updated_at: nil}), do: ~U[1970-01-01 00:00:00Z]
  defp parse_updated_at(%{updated_at: %DateTime{} = dt}), do: dt
  defp parse_updated_at(%{updated_at: str}) when is_binary(str) do
    case parse_datetime(str) do
      {:ok, dt} -> dt
      :error -> ~U[1970-01-01 00:00:00Z]
    end
  end

  defp render_status_badge(assigns, agent) do
    status = derive_display_status(agent)
    assigns = Map.put(assigns, :status, status)

    ~H"""
    <span class={"inline-flex items-center gap-x-1.5 rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset #{status_badge_class(@status)}"}>
      <svg class={"h-1.5 w-1.5 #{status_dot_class(@status)}"} viewBox="0 0 6 6" aria-hidden="true">
        <circle cx="3" cy="3" r="3" />
      </svg>
      {@status}
    </span>
    """
  end

  defp render_agent_meta(assigns, agent) do
    # TODO: Fetch actual session metadata from database
    # For now, placeholder
    session_count = 0
    task_count = 0
    log_count = 0

    assigns = Map.merge(assigns, %{
      session_count: session_count,
      task_count: task_count,
      log_count: log_count,
      last_active: relative_time(agent.updated_at)
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

  defp render_project_badge(_assigns, nil), do: render_no_project()
  defp render_project_badge(_assigns, ""), do: render_no_project()
  defp render_project_badge(_assigns, "-"), do: render_no_project()

  defp render_project_badge(_assigns, project_name) do
    assigns = %{project_name: project_name}

    ~H"""
    <span class="inline-flex items-center rounded-md bg-indigo-50 dark:bg-indigo-900/30 px-2.5 py-1 text-xs font-semibold text-indigo-700 dark:text-indigo-300 ring-1 ring-inset ring-indigo-700/10 dark:ring-indigo-400/30">
      {@project_name}
    </span>
    """
  end

  defp render_no_project do
    assigns = %{}

    ~H"""
    <span class="inline-flex items-center rounded-md bg-gray-50 dark:bg-gray-800 px-2.5 py-1 text-xs font-medium text-gray-600 dark:text-gray-400 ring-1 ring-inset ring-gray-500/10 dark:ring-gray-600">
      Unassigned
    </span>
    """
  end

  # Derive display status considering staleness
  defp derive_display_status(agent) do
    if agent.status in ["active", "working"] && is_stale?(agent) do
      "stale"
    else
      agent.status
    end
  end

  defp is_stale?(agent) do
    case parse_updated_at(agent) do
      %DateTime{} = updated_at ->
        diff_hours = DateTime.diff(DateTime.utc_now(), updated_at, :hour)
        diff_hours >= @stale_threshold_hours && agent.status in ["active", "working"]
      _ ->
        false
    end
  end

  # Status badge classes with dark mode support
  defp status_badge_class("active"), do: "text-green-700 dark:text-green-400 bg-green-50 dark:bg-green-900/30 ring-green-600/20 dark:ring-green-400/30"
  defp status_badge_class("working"), do: "text-yellow-700 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900/30 ring-yellow-600/20 dark:ring-yellow-400/30"
  defp status_badge_class("idle"), do: "text-blue-700 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30 ring-blue-600/20 dark:ring-blue-400/30"
  defp status_badge_class("stale"), do: "text-orange-700 dark:text-orange-400 bg-orange-50 dark:bg-orange-900/30 ring-orange-600/20 dark:ring-orange-400/30"
  defp status_badge_class("completed"), do: "text-gray-700 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 ring-gray-600/20 dark:ring-gray-500/30"
  defp status_badge_class("failed"), do: "text-red-700 dark:text-red-400 bg-red-50 dark:bg-red-900/30 ring-red-600/20 dark:ring-red-400/30"
  defp status_badge_class(_), do: "text-gray-700 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 ring-gray-600/20 dark:ring-gray-500/30"

  defp status_dot_class("active"), do: "fill-green-500"
  defp status_dot_class("working"), do: "fill-yellow-500"
  defp status_dot_class("idle"), do: "fill-blue-500"
  defp status_dot_class("stale"), do: "fill-orange-500"
  defp status_dot_class("completed"), do: "fill-gray-400"
  defp status_dot_class("failed"), do: "fill-red-500"
  defp status_dot_class(_), do: "fill-gray-400"

  # Relative time formatting
  defp relative_time(nil), do: "—"

  defp relative_time(datetime) when is_binary(datetime) do
    case parse_datetime(datetime) do
      {:ok, dt} -> relative_time(dt)
      :error -> format_datetime_short(datetime)
    end
  end

  defp relative_time(%DateTime{} = datetime) do
    now = DateTime.utc_now()
    diff_seconds = DateTime.diff(now, datetime, :second)

    cond do
      diff_seconds < 60 -> "just now"
      diff_seconds < 3600 -> "#{div(diff_seconds, 60)}m ago"
      diff_seconds < 86400 -> "#{div(diff_seconds, 3600)}h ago"
      diff_seconds < 172_800 -> "yesterday"
      diff_seconds < 604_800 -> "#{div(diff_seconds, 86400)}d ago"
      diff_seconds < 2_592_000 -> "#{div(diff_seconds, 604_800)}w ago"
      true -> format_datetime_short(datetime)
    end
  end

  defp relative_time(_), do: "—"

  # Parse Go datetime strings
  defp parse_datetime(datetime) when is_binary(datetime) do
    # Go format: "2025-01-15 10:30:45.123456789 -0700 MST"
    case String.split(datetime, " ", parts: 3) do
      [date, time | _] ->
        time_clean = String.slice(time, 0..7)
        case DateTime.from_iso8601("#{date}T#{time_clean}Z") do
          {:ok, dt, _} -> {:ok, dt}
          _ -> :error
        end
      _ -> :error
    end
  end

  # Full datetime for tooltips
  defp format_datetime_full(nil), do: ""

  defp format_datetime_full(%DateTime{} = datetime) do
    Calendar.strftime(datetime, "%Y-%m-%d %H:%M:%S UTC")
  end

  defp format_datetime_full(datetime) when is_binary(datetime) do
    case String.split(datetime, " ", parts: 3) do
      [date, time | _] -> "#{date} #{String.slice(time, 0..7)}"
      _ -> datetime
    end
  end

  # Short datetime fallback
  defp format_datetime_short(%DateTime{} = datetime) do
    Calendar.strftime(datetime, "%b %d")
  end

  defp format_datetime_short(datetime) when is_binary(datetime) do
    case String.split(datetime, "-") do
      [_year, month, rest] ->
        day = String.slice(rest, 0..1)
        month_name = month_abbrev(month)
        "#{month_name} #{day}"
      _ -> datetime
    end
  end

  defp format_datetime_short(_), do: "—"

  defp month_abbrev("01"), do: "Jan"
  defp month_abbrev("02"), do: "Feb"
  defp month_abbrev("03"), do: "Mar"
  defp month_abbrev("04"), do: "Apr"
  defp month_abbrev("05"), do: "May"
  defp month_abbrev("06"), do: "Jun"
  defp month_abbrev("07"), do: "Jul"
  defp month_abbrev("08"), do: "Aug"
  defp month_abbrev("09"), do: "Sep"
  defp month_abbrev("10"), do: "Oct"
  defp month_abbrev("11"), do: "Nov"
  defp month_abbrev("12"), do: "Dec"
  defp month_abbrev(_), do: "?"
end
