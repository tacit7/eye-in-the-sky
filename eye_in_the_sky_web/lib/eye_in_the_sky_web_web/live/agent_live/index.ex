defmodule EyeInTheSkyWebWeb.AgentLive.Index do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Sessions
  alias Phoenix.LiveView.JS
  import EyeInTheSkyWebWeb.Helpers.ViewHelpers

  @stale_threshold_hours 24

  @impl true
  def mount(_params, _session, socket) do
    # Subscribe to agent updates if connected
    if connected?(socket) do
      Phoenix.PubSub.subscribe(EyeInTheSkyWeb.PubSub, "agents")
      # Refresh agents list every 30 seconds (less aggressive)
      :timer.send_interval(30_000, self(), :refresh_agents)
    end

    sessions = Sessions.list_sessions_with_agent()

    socket =
      socket
      |> assign(:page_title, "Eye in the Sky - Sessions")
      |> assign(:sessions, sessions)
      |> assign(:search_query, "")
      |> assign(:status_filter, "all")
      |> assign(:sort_by, "recent")
      |> assign(:filtered_sessions, sessions)  # Initialize filtered_sessions

    {:ok, socket}
  end

  @impl true
  def handle_params(params, _url, socket) do
    {:noreply, apply_action(socket, socket.assigns.live_action, params)}
  end

  @impl true
  def handle_event("search", %{"query" => query}, socket) do
    socket = socket
      |> assign(:search_query, query)
      |> update_filtered_sessions()
    {:noreply, socket}
  end

  @impl true
  def handle_event("filter_status", %{"status" => status}, socket) do
    IO.puts("Filter clicked: #{status}")
    socket = socket
      |> assign(:status_filter, status)
      |> update_filtered_sessions()
    {:noreply, socket}
  end

  @impl true
  def handle_event("sort", %{"by" => sort_by}, socket) do
    socket = socket
      |> assign(:sort_by, sort_by)
      |> update_filtered_sessions()
    {:noreply, socket}
  end

  defp update_filtered_sessions(socket) do
    filtered = filter_and_sort_sessions(socket.assigns)
    assign(socket, :filtered_sessions, filtered)
  end

  @impl true
  def handle_info(:refresh_agents, socket) do
    sessions = Sessions.list_sessions_with_agent()
    socket = socket
      |> assign(:sessions, sessions)
      |> update_filtered_sessions()
    {:noreply, socket}
  end

  @impl true
  def handle_info({:agent_updated, _agent}, socket) do
    # Reload sessions when we receive PubSub notifications
    sessions = Sessions.list_sessions_with_agent()
    socket = socket
      |> assign(:sessions, sessions)
      |> update_filtered_sessions()
    {:noreply, socket}
  end

  defp apply_action(socket, :index, _params) do
    socket
    |> assign(:page_title, "Listing Agents")
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div class="px-4 sm:px-6 lg:px-8">
      <div class="sm:flex sm:items-center sm:justify-between">
        <div class="sm:flex-auto">
          <h1 class="text-base font-semibold leading-6 text-gray-900 dark:text-gray-100">Agents</h1>
          <p class="mt-2 text-sm text-gray-700 dark:text-gray-400">
            Real-time overview of all Claude Code agents
          </p>
        </div>
        <div class="mt-4 sm:mt-0">
          <label class="swap swap-rotate btn btn-ghost btn-sm btn-circle">
            <input type="checkbox" class="theme-controller" value="dark" />
            <!-- sun icon -->
            <svg class="swap-on h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            <!-- moon icon -->
            <svg class="swap-off h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
            </svg>
          </label>
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

      <div class="mt-6 overflow-x-auto">
        <table class="table table-zebra table-pin-rows">
          <thead>
            <tr>
              <th>Status</th>
              <th>Session</th>
              <th>Project</th>
              <th>Description</th>
              <th>Last Active</th>
            </tr>
          </thead>
          <tbody>
            <%= if @filtered_sessions == [] do %>
              <tr>
                <td colspan="5" class="text-center">
                  No sessions found matching your criteria.
                </td>
              </tr>
            <% else %>
              <%= for session <- @filtered_sessions do %>
                <tr
                  phx-click={JS.navigate(~p"/agents/#{session.agent.id}")}
                  class="hover cursor-pointer group"
                >
                  <td>
                    <%= if is_nil(session.ended_at) do %>
                      <span class="badge badge-success badge-sm">Active</span>
                    <% else %>
                      <span class="badge badge-ghost badge-sm">Completed</span>
                    <% end %>
                  </td>
                  <td>
                    <div class="flex items-center gap-2">
                      <span class="font-mono font-semibold">
                        {session.id && String.slice(session.id, 0..7) || "—"}
                      </span>
                      <%= if session.id do %>
                        <button
                          id={"copy-btn-#{session.id}"}
                          type="button"
                          phx-hook="CopySessionId"
                          data-session-id={session.id}
                          class="relative text-base-content/40 hover:text-base-content/70 transition-colors"
                          aria-label="Copy full session ID"
                        >
                          <.icon name="hero-clipboard" class="h-4 w-4" />
                        </button>
                      <% end %>
                    </div>
                    <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400 font-normal">
                      <span>{session.name || "Unnamed session"}</span>
                    </div>
                  </td>
                  <td>
                    {render_project_badge(assigns, session.agent.project_name)}
                  </td>
                  <td>
                    <div class="line-clamp-1" title={session.agent.description}>
                      {session.agent.description || "—"}
                    </div>
                  </td>
                  <td>
                    <div class="flex items-center justify-between gap-2">
                      <span class="text-sm" title={format_datetime_full(session.started_at)}>
                        {relative_time(session.started_at)}
                      </span>
                      <svg class="h-5 w-5 text-base-content/40 opacity-0 group-hover:opacity-100 transition-opacity" viewBox="0 0 20 20" fill="currentColor">
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
    """
  end


  defp filter_and_sort_sessions(assigns) do
    sessions = assigns.sessions
    query = String.downcase(assigns.search_query)
    status_filter = assigns.status_filter

    IO.puts("Filtering sessions: status_filter=#{status_filter}, total_sessions=#{length(sessions)}")

    sessions
    |> Enum.filter(fn session ->
      # Search filter
      search_match =
        if query == "" do
          true
        else
          String.contains?(String.downcase(session.id || ""), query) ||
            String.contains?(String.downcase(session.name || ""), query) ||
            String.contains?(String.downcase(session.agent.description || ""), query) ||
            String.contains?(String.downcase(session.agent.project_name || ""), query)
        end

      # Status filter - based on session ended_at
      status_match =
        case status_filter do
          "all" -> true
          "active" -> is_nil(session.ended_at)  # Session hasn't ended
          "completed" -> not is_nil(session.ended_at)  # Session has ended
          "stale" -> is_session_stale?(session, @stale_threshold_hours)
          _ -> true
        end

      search_match && status_match
    end)
    |> tap(fn filtered ->
      IO.puts("After filtering: #{length(filtered)} sessions (filter: #{status_filter})")
    end)
    |> sort_sessions(assigns.sort_by)
  end

  defp sort_sessions(sessions, "recent") do
    Enum.sort_by(sessions, &parse_started_at/1, {:desc, DateTime})
  end

  defp parse_started_at(%{started_at: nil}), do: ~U[1970-01-01 00:00:00Z]
  defp parse_started_at(%{started_at: %DateTime{} = dt}), do: dt
  defp parse_started_at(%{started_at: str}) when is_binary(str) do
    case parse_datetime(str) do
      {:ok, dt} -> dt
      :error -> ~U[1970-01-01 00:00:00Z]
    end
  end

  defp is_session_stale?(%{ended_at: ended_at}, _hours) when not is_nil(ended_at), do: false
  defp is_session_stale?(%{started_at: started_at}, hours) do
    case parse_datetime(started_at) do
      {:ok, dt} ->
        now = DateTime.utc_now()
        diff_hours = DateTime.diff(now, dt, :hour)
        diff_hours > hours
      :error -> false
    end
  end

end
