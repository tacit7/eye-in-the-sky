defmodule EyeInTheSkyWebWeb.ProjectLive.Show do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Projects
  alias EyeInTheSkyWeb.Repo

  @impl true
  def mount(%{"id" => id}, _session, socket) do
    project_id = String.to_integer(id)
    project = Projects.get_project!(project_id)
    |> Repo.preload([:agents, :commits])

    # Load tasks manually due to type mismatch (projects.id is INT, tasks.project_id is TEXT)
    tasks = Projects.get_project_tasks(project_id)

    socket =
      socket
      |> assign(:page_title, "Project: #{project.name}")
      |> assign(:project, project)
      |> assign(:tasks, tasks)

    {:ok, socket}
  end

  @impl true
  def render(assigns) do
    ~H"""
    <.live_component module={EyeInTheSkyWebWeb.Components.Navbar} id="navbar" />

    <EyeInTheSkyWebWeb.Components.ProjectNav.render
      project={@project}
      tasks={@tasks}
      current_tab={:overview}
    />

    <div class="px-4 sm:px-6 lg:px-8 py-4">
      <div class="max-w-7xl mx-auto">
        <!-- Responsive Grid Layout -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <!-- Quick Access -->
          <div class="card bg-base-100 shadow-sm">
            <div class="card-body p-4">
              <h2 class="card-title text-base mb-2">Quick Access</h2>
              <%= if @project.path do %>
                <div class="mb-2 pb-2 border-b border-base-300">
                  <p class="text-xs text-base-content/60 mb-1">Project Path</p>
                  <p class="text-xs font-mono text-base-content/90 break-all"><%= @project.path %></p>
                </div>
              <% end %>
              <div class="space-y-1">
                <a href={~p"/projects/#{@project.id}/files?path=CLAUDE.md"} class="flex items-center gap-2 p-2 rounded-lg hover:bg-base-200 transition-colors">
                  <svg class="w-4 h-4 text-primary flex-shrink-0" fill="currentColor" viewBox="0 0 16 16">
                    <path d="M4 0a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2H4zm0 1h8a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1z"/>
                  </svg>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-base-content">CLAUDE.md</p>
                    <p class="text-xs text-base-content/60">Project instructions</p>
                  </div>
                </a>
                <a href={~p"/projects/#{@project.id}/files?path=.claude/hooks"} class="flex items-center gap-2 p-2 rounded-lg hover:bg-base-200 transition-colors">
                  <svg class="w-4 h-4 text-secondary flex-shrink-0" fill="currentColor" viewBox="0 0 16 16">
                    <path d="M1.75 1A1.75 1.75 0 0 0 0 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0 0 16 13.25v-8.5A1.75 1.75 0 0 0 14.25 3H7.5a.25.25 0 0 1-.2-.1l-.9-1.2C6.07 1.26 5.55 1 5 1H1.75Z" />
                  </svg>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-base-content">.claude/hooks/</p>
                    <p class="text-xs text-base-content/60">Hooks configuration</p>
                  </div>
                </a>
              </div>
            </div>
          </div>

          <!-- Recent Agents -->
          <%= if length(@project.agents) > 0 do %>
            <div class="card bg-base-100 shadow-sm">
              <div class="card-body p-4">
                <h2 class="card-title text-base mb-2">Recent Agents</h2>
                <div class="overflow-x-auto">
                  <table class="table table-xs">
                    <thead>
                      <tr>
                        <th class="text-xs">Agent ID</th>
                        <th class="text-xs">Status</th>
                        <th class="text-xs">Description</th>
                      </tr>
                    </thead>
                    <tbody>
                      <%= for agent <- @project.agents |> Enum.sort_by(& &1.created_at, :desc) |> Enum.take(5) do %>
                        <tr>
                          <td>
                            <a href={"/agents/#{agent.id}"} class="link link-primary font-mono text-xs">
                              <%= String.slice(agent.id, 0..7) %>
                            </a>
                          </td>
                          <td>
                            <span class={"badge badge-xs #{status_badge_class(agent.status)}"}>
                              <%= agent.status %>
                            </span>
                          </td>
                          <td class="text-xs text-base-content/70 truncate max-w-xs">
                            <%= agent.description || agent.feature_description || "—" %>
                          </td>
                        </tr>
                      <% end %>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          <% end %>

          <!-- Recent Tasks -->
          <%= if length(@tasks) > 0 do %>
            <div class="card bg-base-100 shadow-sm">
              <div class="card-body p-4">
                <h2 class="card-title text-base mb-2">Recent Tasks</h2>
                <div class="space-y-1">
                  <%= for task <- @tasks |> Enum.sort_by(& &1.created_at, :desc) |> Enum.take(5) do %>
                    <div class="flex items-center gap-2 p-2 rounded-lg hover:bg-base-200 transition-colors">
                      <div class="flex-shrink-0">
                        <%= if task.completed_at do %>
                          <svg class="w-4 h-4 text-success" fill="currentColor" viewBox="0 0 20 20">
                            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                          </svg>
                        <% else %>
                          <div class="w-4 h-4 rounded-full border-2 border-base-content/30"></div>
                        <% end %>
                      </div>
                      <div class="flex-1 min-w-0">
                        <p class="text-sm font-medium text-base-content truncate"><%= task.title %></p>
                        <%= if task.description do %>
                          <p class="text-xs text-base-content/60 truncate"><%= task.description %></p>
                        <% end %>
                      </div>
                      <%= if task.priority && task.priority > 0 do %>
                        <span class="badge badge-xs">P<%= task.priority %></span>
                      <% end %>
                    </div>
                  <% end %>
                </div>
              </div>
            </div>
          <% end %>
        </div>
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
