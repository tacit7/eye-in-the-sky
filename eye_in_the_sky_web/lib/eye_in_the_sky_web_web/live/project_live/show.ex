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

    # Load project files
    files = load_project_files(project.path)

    socket =
      socket
      |> assign(:page_title, "Project: #{project.name}")
      |> assign(:project, project)
      |> assign(:tasks, tasks)
      |> assign(:files, files)

    {:ok, socket}
  end

  defp load_project_files(nil), do: []
  defp load_project_files(path) do
    case File.ls(path) do
      {:ok, files} ->
        files
        |> Enum.reject(&String.starts_with?(&1, "."))
        |> Enum.map(fn file ->
          file_path = Path.join(path, file)
          %{
            name: file,
            path: file,
            is_dir: File.dir?(file_path),
            size: get_file_size(file_path)
          }
        end)
        |> Enum.sort_by(&{!&1.is_dir, &1.name})

      {:error, _} ->
        []
    end
  end

  defp get_file_size(path) do
    if File.dir?(path) do
      ""
    else
      case File.stat(path) do
        {:ok, %{size: size}} -> format_size(size)
        _ -> ""
      end
    end
  end

  defp format_size(size) when size < 1024, do: "#{size} B"
  defp format_size(size) when size < 1024 * 1024, do: "#{Float.round(size / 1024, 1)} KB"
  defp format_size(size), do: "#{Float.round(size / (1024 * 1024), 1)} MB"

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
            class="flex items-center gap-2 px-4 py-2 border-b-2 border-primary text-sm font-medium text-base-content"
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

        <!-- Project Details -->
        <div class="card bg-base-100 shadow-sm mb-8">
          <div class="card-body">
            <h2 class="card-title mb-4">Project Details</h2>
            <div class="space-y-3">
              <%= if @project.path do %>
                <div>
                  <span class="text-sm font-medium text-base-content/70">Path:</span>
                  <span class="ml-2 text-sm font-mono text-base-content"><%= @project.path %></span>
                </div>
              <% end %>
              <%= if @project.remote_url do %>
                <div>
                  <span class="text-sm font-medium text-base-content/70">Remote URL:</span>
                  <a href={@project.remote_url} target="_blank" class="ml-2 text-sm link link-primary">
                    <%= @project.remote_url %>
                  </a>
                </div>
              <% end %>
            </div>
          </div>
        </div>

        <!-- Project Files Quick Links -->
        <div class="card bg-base-100 shadow-sm mb-8">
          <div class="card-body">
            <h2 class="card-title mb-4">Quick Access</h2>
            <div class="space-y-2">
              <a href={~p"/projects/#{@project.id}/files?path=CLAUDE.md"} class="flex items-center gap-3 p-3 rounded-lg hover:bg-base-200 transition-colors">
                <svg class="w-5 h-5 text-primary" fill="currentColor" viewBox="0 0 16 16">
                  <path d="M4 0a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2H4zm0 1h8a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1z"/>
                </svg>
                <div class="flex-1">
                  <p class="text-sm font-medium text-base-content">CLAUDE.md</p>
                  <p class="text-xs text-base-content/60">Project-specific Claude Code instructions</p>
                </div>
                <svg class="w-4 h-4 text-base-content/40" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
              </a>
              <a href={~p"/projects/#{@project.id}/files?path=.claude/hooks"} class="flex items-center gap-3 p-3 rounded-lg hover:bg-base-200 transition-colors">
                <svg class="w-5 h-5 text-secondary" fill="currentColor" viewBox="0 0 16 16">
                  <path d="M1.75 1A1.75 1.75 0 0 0 0 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0 0 16 13.25v-8.5A1.75 1.75 0 0 0 14.25 3H7.5a.25.25 0 0 1-.2-.1l-.9-1.2C6.07 1.26 5.55 1 5 1H1.75Z" />
                </svg>
                <div class="flex-1">
                  <p class="text-sm font-medium text-base-content">.claude/hooks/</p>
                  <p class="text-xs text-base-content/60">Claude Code hooks configuration</p>
                </div>
                <svg class="w-4 h-4 text-base-content/40" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
              </a>
            </div>
          </div>
        </div>

        <!-- File Browser -->
        <%= if length(@files) > 0 do %>
          <div class="card bg-base-100 shadow-sm mb-8">
            <div class="card-body">
              <h2 class="card-title mb-4">Repository Files</h2>
              <div class="border border-base-300 rounded-lg overflow-hidden">
                <%= for file <- @files do %>
                  <a href={~p"/projects/#{@project.id}/files?path=#{file.path}"} class="flex items-center gap-3 px-4 py-3 border-b border-base-300 last:border-b-0 hover:bg-base-200 transition-colors">
                    <%= if file.is_dir do %>
                      <svg class="w-4 h-4 text-primary flex-shrink-0" fill="currentColor" viewBox="0 0 16 16">
                        <path d="M1.75 1A1.75 1.75 0 0 0 0 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0 0 16 13.25v-8.5A1.75 1.75 0 0 0 14.25 3H7.5a.25.25 0 0 1-.2-.1l-.9-1.2C6.07 1.26 5.55 1 5 1H1.75Z" />
                      </svg>
                    <% else %>
                      <svg class="w-4 h-4 text-base-content/60 flex-shrink-0" fill="currentColor" viewBox="0 0 16 16">
                        <path d="M4 0a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2H4zm0 1h8a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1z"/>
                      </svg>
                    <% end %>
                    <span class="flex-1 text-sm font-mono text-base-content"><%= file.name %></span>
                    <%= if !file.is_dir && file.size != "" do %>
                      <span class="text-xs text-base-content/60"><%= file.size %></span>
                    <% end %>
                    <svg class="w-4 h-4 text-base-content/40 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                    </svg>
                  </a>
                <% end %>
              </div>
            </div>
          </div>
        <% end %>

        <!-- Recent Agents -->
        <%= if length(@project.agents) > 0 do %>
          <div class="card bg-base-100 shadow-sm mb-8">
            <div class="card-body">
              <h2 class="card-title mb-4">Recent Agents</h2>
              <div class="overflow-x-auto">
                <table class="table table-zebra">
                  <thead>
                    <tr>
                      <th>Agent ID</th>
                      <th>Status</th>
                      <th>Description</th>
                    </tr>
                  </thead>
                  <tbody>
                    <%= for agent <- Enum.take(@project.agents, 10) do %>
                      <tr>
                        <td>
                          <a href={"/agents/#{agent.id}"} class="link link-primary font-mono text-sm">
                            <%= String.slice(agent.id, 0..7) %>
                          </a>
                        </td>
                        <td>
                          <span class={"badge badge-sm #{status_badge_class(agent.status)}"}>
                            <%= agent.status %>
                          </span>
                        </td>
                        <td class="text-sm text-base-content/70">
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
            <div class="card-body">
              <h2 class="card-title mb-4">Recent Tasks</h2>
              <div class="space-y-2">
                <%= for task <- Enum.take(@tasks, 10) do %>
                  <div class="flex items-center gap-3 p-3 rounded-lg hover:bg-base-200 transition-colors">
                    <div class="flex-shrink-0">
                      <%= if task.completed_at do %>
                        <svg class="w-5 h-5 text-success" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                        </svg>
                      <% else %>
                        <div class="w-5 h-5 rounded-full border-2 border-base-content/30"></div>
                      <% end %>
                    </div>
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-medium text-base-content"><%= task.title %></p>
                      <%= if task.description do %>
                        <p class="text-xs text-base-content/60 truncate"><%= task.description %></p>
                      <% end %>
                    </div>
                    <%= if task.priority && task.priority > 0 do %>
                      <span class="badge badge-sm">P<%= task.priority %></span>
                    <% end %>
                  </div>
                <% end %>
              </div>
            </div>
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
