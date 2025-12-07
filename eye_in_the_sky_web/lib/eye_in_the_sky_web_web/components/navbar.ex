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
          <div class="dropdown dropdown-end">
            <div tabindex="0" role="button" class="avatar cursor-pointer">
              <div class="w-10 rounded-full">
                <img src="https://img.daisyui.com/images/stock/photo-1534528741775-53994a69daeb.webp" alt="Avatar" />
              </div>
            </div>
            <ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-[1] w-52 p-2 shadow">
              <li><a href="/settings">Settings</a></li>
              <li>
                <button class="flex items-center justify-between" onclick="toggleDarkMode()">
                  <span>Dark Mode</span>
                  <span id="theme-toggle-icon" class="text-lg">🌙</span>
                </button>
              </li>
            </ul>
          </div>
          <EyeInTheSkyWebWeb.Layouts.project_switcher projects={@projects} current_project={@current_project} />
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

        function toggleDarkMode() {
          const html = document.documentElement;
          const currentTheme = html.getAttribute('data-theme');
          const newTheme = currentTheme === 'dark' ? 'light' : 'dark';

          html.setAttribute('data-theme', newTheme);
          localStorage.setItem('theme', newTheme);

          const icon = document.getElementById('theme-toggle-icon');
          icon.textContent = newTheme === 'dark' ? '🌙' : '☀️';
        }

        // Initialize icon on page load
        window.addEventListener('load', function() {
          const theme = localStorage.getItem('theme') || 'light';
          const icon = document.getElementById('theme-toggle-icon');
          const currentTheme = document.documentElement.getAttribute('data-theme') || theme;
          icon.textContent = currentTheme === 'dark' ? '🌙' : '☀️';
        });
      </script>
    </div>
    """
  end
end
