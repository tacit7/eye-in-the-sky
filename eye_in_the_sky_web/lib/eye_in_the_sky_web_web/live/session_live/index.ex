defmodule EyeInTheSkyWebWeb.SessionLive.Index do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Sessions

  @impl true
  def mount(_params, _session, socket) do
    sessions = Sessions.list_session_overview_rows(limit: 20)

    socket =
      socket
      |> assign(:page_title, "Session Overview")
      |> assign(:sessions, sessions)

    {:ok, socket}
  end

  @impl true
  def handle_event("start_session", %{"agent_id" => _agent_id}, socket) do
    # Mocked button - just show a message for now
    {:noreply, socket}
  end

  @impl true
  def handle_event("start_session_global", _params, socket) do
    {:noreply, push_navigate(socket, to: ~p"/")}
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div class="px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between mb-6">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900">Session Overview</h1>
          <p class="mt-1 text-sm text-gray-500">
            View all sessions across agents and projects
          </p>
        </div>

        <button
          phx-click="start_session_global"
          class="inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
        >
          Start New Session
        </button>
      </div>

      <div class="mt-8 flow-root">
        <div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
          <div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
            <div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
              <table class="min-w-full divide-y divide-gray-300">
                <thead class="bg-gray-50">
                  <tr>
                    <th
                      scope="col"
                      class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6"
                    >
                      Session ID
                    </th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Project
                    </th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Session Name
                    </th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Started
                    </th>
                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Duration
                    </th>
                    <th
                      scope="col"
                      class="relative py-3.5 pl-3 pr-4 sm:pr-6"
                    >
                      <span class="sr-only">Actions</span>
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white">
                  <%= for session <- @sessions do %>
                    <tr class="hover:bg-gray-50">
                      <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-mono sm:pl-6">
                        <.link
                          navigate={~p"/agents/#{session.agent_id}?s=#{session.session_id}"}
                          class="text-indigo-600 hover:text-indigo-900 underline decoration-dotted"
                        >
                          <%= String.slice(session.session_id, 0..11) %>...
                        </.link>
                      </td>

                      <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        <%= session.project_name || "—" %>
                      </td>

                      <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        <%= session.session_name || "—" %>
                      </td>

                      <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        <%= format_timestamp(session.started_at) %>
                      </td>

                      <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        <%= format_duration(session.started_at, session.ended_at) %>
                      </td>

                      <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                        <div class="flex gap-2 justify-end">
                          <button
                            phx-hook="CopyToClipboard"
                            id={"copy-#{session.session_id}"}
                            data-session-id={session.session_id}
                            class="inline-flex items-center rounded border border-gray-300 bg-white px-2.5 py-1.5 text-xs font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                          >
                            Copy ID
                          </button>

                          <button
                            phx-click="start_session"
                            phx-value-agent_id={session.agent_id}
                            class="inline-flex items-center rounded border border-gray-300 bg-white px-2.5 py-1.5 text-xs font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                          >
                            New Session
                          </button>
                        </div>
                      </td>
                    </tr>
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

  defp format_timestamp(nil), do: "—"

  defp format_timestamp(timestamp) when is_binary(timestamp) do
    # Handle Go time format strings
    case String.split(timestamp, " ", parts: 3) do
      [date, time | _] -> "#{date} #{String.slice(time, 0..7)}"
      _ -> timestamp
    end
  end

  defp format_duration(started_at, ended_at) when is_binary(started_at) do
    # For Go-formatted strings, we can't easily calculate duration
    # Just show if it's ended or active
    if ended_at && ended_at != "", do: "Ended", else: "Active"
  end

  defp format_duration(_, _), do: "—"
end
