defmodule EyeInTheSkyWebWeb.ProjectLive.Files do
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

    # Build file tree
    file_tree = if project.path do
      build_file_tree(project.path, project.path)
    else
      []
    end

    socket =
      socket
      |> assign(:page_title, "Files - #{project.name}")
      |> assign(:project, project)
      |> assign(:tasks, tasks)
      |> assign(:file_path, nil)
      |> assign(:file_content, nil)
      |> assign(:rendered_content, nil)
      |> assign(:file_type, nil)
      |> assign(:file_tree, file_tree)
      |> assign(:error, nil)

    {:ok, socket}
  end

  @impl true
  def handle_params(%{"path" => path}, _uri, socket) do
    project = socket.assigns.project

    if project.path do
      full_path = Path.join(project.path, path)

      if File.regular?(full_path) do
        # Read file contents
        case File.read(full_path) do
          {:ok, content} ->
            file_type = detect_file_type(path)
            rendered_content = render_content(content, file_type)

            {:noreply,
             socket
             |> assign(:file_path, path)
             |> assign(:file_content, content)
             |> assign(:rendered_content, rendered_content)
             |> assign(:file_type, file_type)
             |> assign(:error, nil)}

          {:error, reason} ->
            {:noreply,
             socket
             |> assign(:error, "Failed to read file: #{reason}")
             |> assign(:file_content, nil)
             |> assign(:rendered_content, nil)}
        end
      else
        {:noreply,
         socket
         |> assign(:error, "File not found: #{path}")
         |> assign(:file_content, nil)}
      end
    else
      {:noreply,
       socket
       |> assign(:error, "Project path not configured")
       |> assign(:file_content, nil)}
    end
  end

  def handle_params(_params, _uri, socket) do
    {:noreply, socket}
  end

  defp build_file_tree(base_path, current_path, max_depth \\ 5, current_depth \\ 0) do
    if current_depth >= max_depth do
      []
    else
      case File.ls(current_path) do
        {:ok, files} ->
          files
          |> Enum.filter(fn file ->
            # Filter out common ignored directories/files
            !String.starts_with?(file, ".") or file in [".claude", ".git"]
          end)
          |> Enum.map(fn file ->
            full_path = Path.join(current_path, file)
            relative_path = Path.relative_to(full_path, base_path)

            if File.dir?(full_path) do
              children = build_file_tree(base_path, full_path, max_depth, current_depth + 1)
              %{
                name: file,
                path: relative_path,
                type: :directory,
                children: Enum.sort_by(children, &{&1.type != :directory, &1.name})
              }
            else
              %{
                name: file,
                path: relative_path,
                type: :file,
                size: get_file_size(full_path)
              }
            end
          end)
          |> Enum.sort_by(&{&1.type != :directory, &1.name})

        {:error, _reason} ->
          []
      end
    end
  end

  defp get_file_size(path) do
    case File.stat(path) do
      {:ok, %{size: size}} -> format_size(size)
      _ -> ""
    end
  end

  defp format_size(size) when size < 1024, do: "#{size} B"
  defp format_size(size) when size < 1024 * 1024, do: "#{Float.round(size / 1024, 1)} KB"
  defp format_size(size), do: "#{Float.round(size / (1024 * 1024), 1)} MB"

  defp detect_file_type(path) do
    extension = path |> Path.extname() |> String.downcase()

    case extension do
      ".md" -> :markdown
      ".markdown" -> :markdown
      ".ex" -> :elixir
      ".exs" -> :elixir
      ".js" -> :javascript
      ".ts" -> :typescript
      ".jsx" -> :javascript
      ".tsx" -> :typescript
      ".json" -> :json
      ".yml" -> :yaml
      ".yaml" -> :yaml
      ".html" -> :html
      ".css" -> :css
      ".py" -> :python
      ".rb" -> :ruby
      ".go" -> :go
      ".rs" -> :rust
      ".java" -> :java
      ".c" -> :c
      ".cpp" -> :cpp
      ".sh" -> :bash
      ".sql" -> :sql
      ".xml" -> :xml
      _ -> :text
    end
  end

  defp render_content(_content, _), do: nil

  defp language_class(file_type) do
    case file_type do
      :markdown -> "markdown"
      :elixir -> "elixir"
      :javascript -> "javascript"
      :typescript -> "typescript"
      :json -> "json"
      :yaml -> "yaml"
      :html -> "html"
      :css -> "css"
      :python -> "python"
      :ruby -> "ruby"
      :go -> "go"
      :rust -> "rust"
      :java -> "java"
      :c -> "c"
      :cpp -> "cpp"
      :bash -> "bash"
      :sql -> "sql"
      :xml -> "xml"
      _ -> "plaintext"
    end
  end

  defp render_file_tree(items, project_id) do
    for item <- items do
      case item.type do
        :directory ->
          ~H"""
          <li>
            <details>
              <summary>
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
                  <path d="M1.75 1A1.75 1.75 0 0 0 0 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0 0 16 13.25v-8.5A1.75 1.75 0 0 0 14.25 3H7.5a.25.25 0 0 1-.2-.1l-.9-1.2C6.07 1.26 5.55 1 5 1H1.75Z" />
                </svg>
                <%= item.name %>
              </summary>
              <ul>
                <%= render_file_tree(item.children, project_id) %>
              </ul>
            </details>
          </li>
          """

        :file ->
          ~H"""
          <li>
            <a href={~p"/projects/#{project_id}/files?path=#{item.path}"}>
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
                <path d="M4 0a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2H4zm0 1h8a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1z"/>
              </svg>
              <%= item.name %>
              <%= if item.size do %>
                <span class="badge badge-ghost badge-xs ml-auto"><%= item.size %></span>
              <% end %>
            </a>
          </li>
          """
      end
    end
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

    <div class="h-[calc(100vh-8rem)] flex">
      <!-- File Tree Sidebar -->
      <div class="w-80 border-r border-base-300 bg-base-100 overflow-y-auto">
        <div class="p-4">
          <h2 class="text-sm font-semibold text-base-content/80 mb-2">Files</h2>
          <ul class="menu menu-sm bg-base-200 rounded-lg">
            <%= render_file_tree(@file_tree, @project.id) %>
          </ul>
        </div>
      </div>

      <!-- File Content Viewer -->
      <div class="flex-1 overflow-y-auto">
        <%= if @error do %>
          <!-- Error Message -->
          <div class="p-4">
            <div class="alert alert-error">
              <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span><%= @error %></span>
            </div>
          </div>
        <% end %>

        <%= if @file_content do %>
          <!-- File Content -->
          <div class="p-6">
            <div class="mb-4">
              <h2 class="text-lg font-semibold text-base-content"><%= Path.basename(@file_path) %></h2>
              <p class="text-sm text-base-content/60"><%= @file_path %></p>
            </div>
            <%= if @rendered_content do %>
              <!-- Rendered Markdown -->
              <div class="prose prose-sm max-w-none dark:prose-invert bg-base-200 rounded-lg p-6 overflow-x-auto">
                <%= @rendered_content %>
              </div>
            <% else %>
              <!-- Syntax Highlighted Code -->
              <div class="bg-base-200 rounded-lg overflow-x-auto">
                <pre class="text-sm"><code id="code-viewer" class={"language-#{language_class(@file_type)}"} phx-hook="Highlight"><%= @file_content %></code></pre>
              </div>
            <% end %>
          </div>
        <% else %>
          <!-- Empty State -->
          <div class="flex items-center justify-center h-full">
            <div class="text-center">
              <svg class="w-16 h-16 mx-auto text-base-content/20 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <h3 class="text-lg font-semibold text-base-content/60 mb-2">Select a file</h3>
              <p class="text-sm text-base-content/40">Choose a file from the tree to view its contents</p>
            </div>
          </div>
        <% end %>
      </div>
    </div>
    """
  end
end
