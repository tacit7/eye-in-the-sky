# Phoenix App Bookmark UI Implementation

## 1. Svelte Bookmark Button Component

**Location:** `eye_in_the_sky_web/assets/svelte/components/BookmarkButton.svelte`

```svelte
<script>
  export let bookmarkType;  // 'file' | 'note' | 'agent' | 'session' | 'task'
  export let bookmarkId = null;
  export let filePath = null;
  export let lineNumber = null;
  export let title = null;
  export let category = null;
  export let projectId = null;
  export let agentId = null;
  export let isBookmarked = false;
  export let size = 'sm';  // 'xs' | 'sm' | 'md' | 'lg'

  let loading = false;

  async function toggleBookmark() {
    loading = true;

    const payload = {
      bookmark_type: bookmarkType,
      bookmark_id: bookmarkId,
      file_path: filePath,
      line_number: lineNumber,
      title: title,
      category: category,
      project_id: projectId,
      agent_id: agentId
    };

    try {
      if (isBookmarked) {
        // Delete bookmark
        await fetch(`/api/bookmarks/${bookmarkId}`, {
          method: 'DELETE',
          headers: { 'Content-Type': 'application/json' }
        });
        isBookmarked = false;
      } else {
        // Create bookmark
        const response = await fetch('/api/bookmarks', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        const data = await response.json();
        isBookmarked = true;
        bookmarkId = data.id;
      }
    } catch (error) {
      console.error('Bookmark error:', error);
    } finally {
      loading = false;
    }
  }

  const sizeClasses = {
    xs: 'w-3 h-3',
    sm: 'w-4 h-4',
    md: 'w-5 h-5',
    lg: 'w-6 h-6'
  };
</script>

<button
  on:click={toggleBookmark}
  disabled={loading}
  class="btn btn-ghost btn-sm p-1 hover:bg-base-200 transition-colors"
  title={isBookmarked ? 'Remove bookmark' : 'Add bookmark'}
>
  {#if loading}
    <span class="loading loading-spinner {sizeClasses[size]}"></span>
  {:else if isBookmarked}
    <svg class="{sizeClasses[size]} text-warning" fill="currentColor" viewBox="0 0 20 20">
      <!-- Filled bookmark -->
      <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z"/>
    </svg>
  {:else}
    <svg class="{sizeClasses[size]} text-base-content/40 hover:text-base-content/70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <!-- Outline bookmark -->
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"/>
    </svg>
  {/if}
</button>
```

## 2. Phoenix Context - Bookmarks

**Location:** `eye_in_the_sky_web/lib/eye_in_the_sky_web/bookmarks.ex`

```elixir
defmodule EyeInTheSkyWeb.Bookmarks do
  @moduledoc """
  Context for managing bookmarks.
  """

  alias EyeInTheSkyWeb.Repo
  alias EyeInTheSkyWeb.Bookmarks.Bookmark

  def list_bookmarks(opts \\ []) do
    query = from b in Bookmark, order_by: [desc: b.priority, desc: b.created_at]

    query
    |> maybe_filter_by_type(opts[:bookmark_type])
    |> maybe_filter_by_category(opts[:category])
    |> maybe_filter_by_project(opts[:project_id])
    |> maybe_filter_by_agent(opts[:agent_id])
    |> maybe_limit(opts[:limit] || 100)
    |> Repo.all()
  end

  def get_bookmark!(id), do: Repo.get!(Bookmark, id)

  def create_bookmark(attrs \\ %{}) do
    %Bookmark{}
    |> Bookmark.changeset(attrs)
    |> Repo.insert()
  end

  def update_bookmark(%Bookmark{} = bookmark, attrs) do
    bookmark
    |> Bookmark.changeset(attrs)
    |> Repo.update()
  end

  def delete_bookmark(%Bookmark{} = bookmark) do
    Repo.delete(bookmark)
  end

  def check_if_bookmarked(bookmark_type, identifier) do
    case bookmark_type do
      "file" ->
        from(b in Bookmark, where: b.bookmark_type == "file" and b.file_path == ^identifier)
        |> Repo.exists?()

      type when type in ["note", "agent", "session", "task"] ->
        from(b in Bookmark, where: b.bookmark_type == ^type and b.bookmark_id == ^identifier)
        |> Repo.exists?()

      _ -> false
    end
  end

  # Private helpers
  defp maybe_filter_by_type(query, nil), do: query
  defp maybe_filter_by_type(query, type), do: from b in query, where: b.bookmark_type == ^type

  defp maybe_filter_by_category(query, nil), do: query
  defp maybe_filter_by_category(query, cat), do: from b in query, where: b.category == ^cat

  defp maybe_filter_by_project(query, nil), do: query
  defp maybe_filter_by_project(query, pid), do: from b in query, where: b.project_id == ^pid

  defp maybe_filter_by_agent(query, nil), do: query
  defp maybe_filter_by_agent(query, aid), do: from b in query, where: b.agent_id == ^aid

  defp maybe_limit(query, limit), do: from b in query, limit: ^limit
end
```

## 3. Phoenix Schema

**Location:** `eye_in_the_sky_web/lib/eye_in_the_sky_web/bookmarks/bookmark.ex`

```elixir
defmodule EyeInTheSkyWeb.Bookmarks.Bookmark do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id

  schema "bookmarks" do
    field :bookmark_type, :string
    field :bookmark_id, :string
    field :file_path, :string
    field :line_number, :integer
    field :url, :string
    field :title, :string
    field :description, :string
    field :category, :string
    field :priority, :integer, default: 0
    field :position, :integer
    field :project_id, :integer
    field :agent_id, :string
    field :accessed_at, :utc_datetime

    timestamps(type: :utc_datetime)
  end

  def changeset(bookmark, attrs) do
    bookmark
    |> cast(attrs, [
      :bookmark_type, :bookmark_id, :file_path, :line_number, :url,
      :title, :description, :category, :priority, :position,
      :project_id, :agent_id, :accessed_at
    ])
    |> validate_required([:bookmark_type])
    |> validate_inclusion(:bookmark_type, ~w(file note agent session task url))
    |> validate_bookmark_fields()
  end

  defp validate_bookmark_fields(changeset) do
    type = get_field(changeset, :bookmark_type)

    case type do
      "file" -> validate_required(changeset, [:file_path])
      "url" -> validate_required(changeset, [:url])
      type when type in ~w(note agent session task) -> validate_required(changeset, [:bookmark_id])
      _ -> changeset
    end
  end
end
```

## 4. API Controller

**Location:** `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/controllers/bookmark_controller.ex`

```elixir
defmodule EyeInTheSkyWebWeb.BookmarkController do
  use EyeInTheSkyWebWeb, :controller

  alias EyeInTheSkyWeb.Bookmarks
  alias EyeInTheSkyWeb.Bookmarks.Bookmark

  def index(conn, params) do
    bookmarks = Bookmarks.list_bookmarks(
      bookmark_type: params["type"],
      category: params["category"],
      project_id: params["project_id"],
      agent_id: params["agent_id"]
    )

    json(conn, %{bookmarks: bookmarks})
  end

  def create(conn, bookmark_params) do
    case Bookmarks.create_bookmark(bookmark_params) do
      {:ok, bookmark} ->
        conn
        |> put_status(:created)
        |> json(%{id: bookmark.id, bookmark: bookmark})

      {:error, changeset} ->
        conn
        |> put_status(:unprocessable_entity)
        |> json(%{errors: translate_errors(changeset)})
    end
  end

  def delete(conn, %{"id" => id}) do
    bookmark = Bookmarks.get_bookmark!(id)
    {:ok, _bookmark} = Bookmarks.delete_bookmark(bookmark)

    send_resp(conn, :no_content, "")
  end

  defp translate_errors(changeset) do
    Ecto.Changeset.traverse_errors(changeset, fn {msg, opts} ->
      Enum.reduce(opts, msg, fn {key, value}, acc ->
        String.replace(acc, "%{#{key}}", to_string(value))
      end)
    end)
  end
end
```

## 5. Router Updates

**Location:** `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/router.ex`

```elixir
# Add to api scope
scope "/api", EyeInTheSkyWebWeb do
  pipe_through :api

  resources "/bookmarks", BookmarkController, only: [:index, :create, :delete]
end
```

## 6. LiveView Bookmarks Page

**Location:** `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/live/bookmark_live/index.ex`

```elixir
defmodule EyeInTheSkyWebWeb.BookmarkLive.Index do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.Bookmarks

  @impl true
  def mount(_params, _session, socket) do
    bookmarks = Bookmarks.list_bookmarks()

    socket =
      socket
      |> assign(:page_title, "Bookmarks")
      |> assign(:bookmarks, bookmarks)
      |> assign(:filter_type, nil)
      |> assign(:filter_category, nil)

    {:ok, socket}
  end

  @impl true
  def handle_event("filter_type", %{"type" => type}, socket) do
    filter_type = if type == "", do: nil, else: type
    bookmarks = Bookmarks.list_bookmarks(bookmark_type: filter_type)

    {:noreply, assign(socket, bookmarks: bookmarks, filter_type: filter_type)}
  end

  @impl true
  def handle_event("delete", %{"id" => id}, socket) do
    bookmark = Bookmarks.get_bookmark!(id)
    {:ok, _} = Bookmarks.delete_bookmark(bookmark)

    bookmarks = Bookmarks.list_bookmarks(
      bookmark_type: socket.assigns.filter_type,
      category: socket.assigns.filter_category
    )

    {:noreply, assign(socket, :bookmarks, bookmarks)}
  end

  @impl true
  def render(assigns) do
    ~H"""
    <.live_component module={EyeInTheSkyWebWeb.Components.Navbar} id="navbar" />

    <div class="px-4 sm:px-6 lg:px-8 py-4">
      <div class="max-w-7xl mx-auto">
        <div class="flex items-center justify-between mb-4">
          <h1 class="text-2xl font-bold">Bookmarks</h1>

          <select class="select select-bordered select-sm" phx-change="filter_type">
            <option value="">All Types</option>
            <option value="file">Files</option>
            <option value="note">Notes</option>
            <option value="agent">Agents</option>
            <option value="session">Sessions</option>
            <option value="task">Tasks</option>
          </select>
        </div>

        <div class="space-y-2">
          <%= for bookmark <- @bookmarks do %>
            <div class="card bg-base-100 shadow-sm">
              <div class="card-body p-4">
                <div class="flex items-start justify-between">
                  <div class="flex-1">
                    <div class="flex items-center gap-2">
                      <span class="badge badge-sm"><%= bookmark.bookmark_type %></span>
                      <%= if bookmark.category do %>
                        <span class="badge badge-sm badge-outline"><%= bookmark.category %></span>
                      <% end %>
                      <%= if bookmark.priority && bookmark.priority > 0 do %>
                        <span class="badge badge-sm badge-warning">P<%= bookmark.priority %></span>
                      <% end %>
                    </div>

                    <h3 class="font-semibold mt-2">
                      <%= bookmark.title || bookmark_display_text(bookmark) %>
                    </h3>

                    <%= if bookmark.description do %>
                      <p class="text-sm text-base-content/60 mt-1"><%= bookmark.description %></p>
                    <% end %>

                    <div class="text-xs text-base-content/40 mt-2">
                      <%= bookmark_details(bookmark) %>
                    </div>
                  </div>

                  <button
                    phx-click="delete"
                    phx-value-id={bookmark.id}
                    class="btn btn-ghost btn-sm"
                  >
                    <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          <% end %>
        </div>
      </div>
    </div>
    """
  end

  defp bookmark_display_text(bookmark) do
    case bookmark.bookmark_type do
      "file" -> Path.basename(bookmark.file_path || "")
      "url" -> bookmark.url
      _ -> bookmark.bookmark_id
    end
  end

  defp bookmark_details(bookmark) do
    case bookmark.bookmark_type do
      "file" ->
        if bookmark.line_number do
          "#{bookmark.file_path}:#{bookmark.line_number}"
        else
          bookmark.file_path
        end

      "url" -> bookmark.url
      _ -> "ID: #{bookmark.bookmark_id}"
    end
  end
end
```

## 7. Usage in File Tree

Update `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/live/project_live/files.ex`:

```elixir
# Add to file list rendering
<div class="flex items-center gap-2">
  <a href={~p"/projects/#{@project.id}/files?path=#{file.path}"}>
    <%= file.name %>
  </a>

  <!-- Bookmark button -->
  <BookmarkButton
    bookmarkType="file"
    filePath={file.path}
    projectId={@project.id}
    isBookmarked={check_bookmarked("file", file.path)}
    size="xs"
  />
</div>
```

## Implementation Checklist

- [ ] Add `bookmarks` table migration
- [ ] Create Bookmark schema and context
- [ ] Add BookmarkController API endpoints
- [ ] Create BookmarkButton Svelte component
- [ ] Add BookmarkLive.Index page
- [ ] Update router with bookmark routes
- [ ] Add bookmark buttons to:
  - [ ] File tree view
  - [ ] Agent detail page
  - [ ] Task list
  - [ ] Notes display
  - [ ] Session view
- [ ] Add "Bookmarks" to project navigation
- [ ] Add global bookmarks page link to navbar
