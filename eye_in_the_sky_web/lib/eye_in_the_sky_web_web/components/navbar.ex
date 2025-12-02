defmodule EyeInTheSkyWebWeb.Components.Navbar do
  use EyeInTheSkyWebWeb, :live_component

  alias EyeInTheSkyWeb.Projects

  @impl true
  def mount(socket) do
    projects = Projects.list_projects()
    {:ok, assign(socket, projects: projects, current_project: nil)}
  end

  @impl true
  def update(assigns, socket) do
    {:ok, assign(socket, assigns)}
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div>
      <div class="navbar bg-base-100 shadow-sm">
        <div class="navbar-start">
          <a href="/" class="btn btn-ghost text-xl">
            <img src="/images/logo.svg" width="36" />
            Eye in the Sky
          </a>
        </div>
        <div class="navbar-center hidden lg:flex">
          <ul class="menu menu-horizontal px-1">
            <li><a href="/">Overview</a></li>
            <li><a href="/prompts">Prompts</a></li>
            <li><a href="/chat">Chat</a></li>
            <li><a href="/nats">NATS</a></li>
          </ul>
        </div>
        <div class="navbar-end gap-2">
          <EyeInTheSkyWebWeb.Layouts.project_switcher projects={@projects} current_project={@current_project} />
          <div class="card relative flex flex-row items-center border-2 border-base-300 bg-base-300 rounded-full">
            <div class="absolute w-1/3 h-full rounded-full border-1 border-base-200 bg-base-100 brightness-200 left-0 [[data-theme=light]_&]:left-1/3 [[data-theme=dark]_&]:left-2/3 transition-[left]" />

            <button
              class="flex p-2 cursor-pointer w-1/3"
              onclick="setTheme('system')"
              data-phx-theme="system"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="size-4 opacity-75 hover:opacity-100">
                <path fill-rule="evenodd" d="M1 11.5a.5.5 0 0 0 .5.5h.5a.5.5 0 0 0 .5-.5v-1a.5.5 0 0 0-.5-.5h-.5a.5.5 0 0 0-.5.5v1ZM4.5 10A.5.5 0 0 0 4 10.5v1a.5.5 0 0 0 .5.5h.5a.5.5 0 0 0 .5-.5v-1A.5.5 0 0 0 5 10h-.5ZM7 10.5a.5.5 0 0 1 .5-.5h.5a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-.5a.5.5 0 0 1-.5-.5v-1Zm3-.5a.5.5 0 0 0-.5.5v1a.5.5 0 0 0 .5.5h.5a.5.5 0 0 0 .5-.5v-1a.5.5 0 0 0-.5-.5h-.5ZM1 8a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v5a1 1 0 0 1-1 1H2a1 1 0 0 1-1-1V8Zm1-.5a.5.5 0 0 0-.5.5v1.5a.5.5 0 0 0 .5.5h12a.5.5 0 0 0 .5-.5V8a.5.5 0 0 0-.5-.5H2Z" clip-rule="evenodd" />
                <path d="M2 3.5A1.5 1.5 0 0 1 3.5 2h9A1.5 1.5 0 0 1 14 3.5V6H2V3.5ZM3.5 3a.5.5 0 0 0-.5.5V5h10V3.5a.5.5 0 0 0-.5-.5h-9Z" />
              </svg>
            </button>

            <button
              class="flex p-2 cursor-pointer w-1/3"
              onclick="setTheme('light')"
              data-phx-theme="light"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="size-4 opacity-75 hover:opacity-100">
                <path d="M8 1a.75.75 0 0 1 .75.75v1.5a.75.75 0 0 1-1.5 0v-1.5A.75.75 0 0 1 8 1ZM10.5 8a2.5 2.5 0 1 1-5 0 2.5 2.5 0 0 1 5 0ZM12.95 4.11a.75.75 0 1 0-1.06-1.06l-1.062 1.06a.75.75 0 0 0 1.061 1.062l1.06-1.061ZM15 8a.75.75 0 0 1-.75.75h-1.5a.75.75 0 0 1 0-1.5h1.5A.75.75 0 0 1 15 8ZM11.89 12.95a.75.75 0 0 0 1.06-1.06l-1.06-1.062a.75.75 0 0 0-1.062 1.061l1.061 1.06ZM8 12a.75.75 0 0 1 .75.75v1.5a.75.75 0 0 1-1.5 0v-1.5A.75.75 0 0 1 8 12ZM5.172 11.89a.75.75 0 0 0-1.061-1.062L3.05 11.89a.75.75 0 1 0 1.06 1.06l1.06-1.06ZM4 8a.75.75 0 0 1-.75.75h-1.5a.75.75 0 0 1 0-1.5h1.5A.75.75 0 0 1 4 8ZM4.11 5.172A.75.75 0 0 0 5.173 4.11L4.11 3.05a.75.75 0 1 0-1.06 1.06l1.06 1.06Z" />
              </svg>
            </button>

            <button
              class="flex p-2 cursor-pointer w-1/3"
              onclick="setTheme('dark')"
              data-phx-theme="dark"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="size-4 opacity-75 hover:opacity-100">
                <path d="M14.438 10.148c.19-.425-.321-.787-.748-.601A5.5 5.5 0 0 1 6.453 2.31c.186-.427-.176-.938-.6-.748a6.501 6.501 0 1 0 8.585 8.586Z" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <script>
        function setTheme(theme) {
          if (theme === 'system') {
            const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
            document.documentElement.setAttribute('data-theme', systemTheme);
            localStorage.setItem('theme', 'system');
          } else {
            document.documentElement.setAttribute('data-theme', theme);
            localStorage.setItem('theme', theme);
          }
        }
      </script>
    </div>
    """
  end
end
