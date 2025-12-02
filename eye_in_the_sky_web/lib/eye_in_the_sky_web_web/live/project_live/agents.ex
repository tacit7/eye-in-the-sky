defmodule EyeInTheSkyWebWeb.ProjectLive.Agents do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Projects
  alias EyeInTheSkyWeb.Repo

  @impl true
  def mount(%{"id" => id}, _session, socket) do
    project_id = String.to_integer(id)
    project = Projects.get_project!(project_id)
    |> Repo.preload([:agents, :commits])

    # Load tasks manually due to type mismatch
    tasks = Projects.get_project_tasks(project_id)

    socket =
      socket
      |> assign(:page_title, "Agents - #{project.name}")
      |> assign(:project, project)
      |> assign(:tasks, tasks)
      |> assign(:search_query, "")
      |> assign(:status_filter, "all")
      |> assign(:filtered_agents, project.agents)

    {:ok, socket}
  end

  @impl true
  def handle_event("search", %{"query" => query}, socket) do
    socket = socket
      |> assign(:search_query, query)
      |> update_filtered_agents()
    {:noreply, socket}
  end

  @impl true
  def handle_event("filter_status", %{"status" => status}, socket) do
    socket = socket
      |> assign(:status_filter, status)
      |> update_filtered_agents()
    {:noreply, socket}
  end

  defp update_filtered_agents(socket) do
    filtered = filter_agents(socket.assigns)
    assign(socket, :filtered_agents, filtered)
  end

  defp filter_agents(assigns) do
    agents = assigns.project.agents
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
            String.contains?(String.downcase(agent.session_id || ""), query)
        end

      # Status filter
      status_match =
        case status_filter do
          "all" -> true
          status -> agent.status == status
        end

      search_match && status_match
    end)
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
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-primary text-sm font-medium text-base-content"
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
      <div class="max-w-6xl mx-auto">

        <!-- Search and Filters -->
        <div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:gap-6">
          <!-- Search -->
          <div class="flex-1 max-w-md">
            <form phx-change="search">
              <label for="search" class="sr-only">Search agents</label>
              <div class="relative">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
                  <svg class="h-5 w-5 text-base-content/40" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z" clip-rule="evenodd" />
                  </svg>
                </div>
                <input
                  type="text"
                  name="query"
                  id="search"
                  phx-debounce="300"
                  value={@search_query}
                  class="input input-bordered w-full pl-10"
                  placeholder="Search agents, sessions, descriptions..."
                />
              </div>
            </form>
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
              phx-value-status="working"
              class={"btn btn-sm #{if @status_filter == "working", do: "btn-active"}"}
            >
              Working
            </button>
            <button
              phx-click="filter_status"
              phx-value-status="idle"
              class={"btn btn-sm #{if @status_filter == "idle", do: "btn-active"}"}
            >
              Idle
            </button>
            <button
              phx-click="filter_status"
              phx-value-status="completed"
              class={"btn btn-sm #{if @status_filter == "completed", do: "btn-active"}"}
            >
              Completed
            </button>
          </div>
        </div>

        <%= if length(@filtered_agents) > 0 do %>
          <!-- Agents List -->
          <div class="space-y-4">
            <%= for agent <- @filtered_agents do %>
              <a href={"/agents/#{agent.id}"} class="block">
                <div class="card bg-base-100 border border-base-300 hover:border-primary hover:shadow-md transition-all">
                  <div class="card-body">
                    <div class="flex items-start justify-between">
                      <div class="flex-1 min-w-0">
                        <!-- Agent ID and Status -->
                        <div class="flex items-center gap-3 mb-2">
                          <code class="text-sm font-mono text-base-content font-semibold">
                            <%= String.slice(agent.id, 0..7) %>
                          </code>
                          <span class={"badge badge-sm #{status_badge_class(agent.status)}"}>
                            <%= agent.status %>
                          </span>
                        </div>

                        <!-- Description -->
                        <%= if agent.description || agent.feature_description do %>
                          <p class="text-sm text-base-content/80 mb-2">
                            <%= agent.description || agent.feature_description %>
                          </p>
                        <% end %>

                        <!-- Meta Information -->
                        <div class="flex items-center gap-4 text-xs text-base-content/60">
                          <%= if agent.session_id do %>
                            <span class="flex items-center gap-1">
                              <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 16 16">
                                <path d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0ZM1.5 8a6.5 6.5 0 1 0 13 0 6.5 6.5 0 0 0-13 0Z" />
                              </svg>
                              Session: <%= String.slice(agent.session_id, 0..7) %>
                            </span>
                          <% end %>
                          <%= if agent.git_worktree_path do %>
                            <span class="flex items-center gap-1 font-mono">
                              <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 16 16">
                                <path d="M1.75 1A1.75 1.75 0 0 0 0 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0 0 16 13.25v-8.5A1.75 1.75 0 0 0 14.25 3H7.5a.25.25 0 0 1-.2-.1l-.9-1.2C6.07 1.26 5.55 1 5 1H1.75Z" />
                              </svg>
                              <%= Path.basename(agent.git_worktree_path) %>
                            </span>
                          <% end %>
                        </div>
                      </div>

                      <!-- Chevron -->
                      <svg class="w-5 h-5 text-base-content/40 flex-shrink-0 mt-1" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                      </svg>
                    </div>
                  </div>
                </div>
              </a>
            <% end %>
          </div>
        <% else %>
          <!-- Empty State -->
          <div class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-base-content/40" fill="currentColor" viewBox="0 0 16 16">
              <path d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0ZM1.5 8a6.5 6.5 0 1 0 13 0 6.5 6.5 0 0 0-13 0Zm7-3.25v2.992l2.028.812a.75.75 0 0 1-.557 1.392l-2.5-1A.751.751 0 0 1 7 8.25v-3.5a.75.75 0 0 1 1.5 0Z" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-base-content">
              <%= if @search_query != "" || @status_filter != "all" do %>
                No agents match your filters
              <% else %>
                No agents yet
              <% end %>
            </h3>
            <p class="mt-1 text-sm text-base-content/60">
              <%= if @search_query != "" || @status_filter != "all" do %>
                Try adjusting your search or filters
              <% else %>
                Agents will appear here when they start working on this project
              <% end %>
            </p>
          </div>
        <% end %>

      </div>
    </div>
    """
  end

  defp status_badge_class(status) do
    case status do
      "active" -> "badge-success"
      "working" -> "badge-warning"
      "idle" -> "badge-info"
      "completed" -> "badge-ghost"
      "failed" -> "badge-error"
      _ -> "badge-ghost"
    end
  end
end
