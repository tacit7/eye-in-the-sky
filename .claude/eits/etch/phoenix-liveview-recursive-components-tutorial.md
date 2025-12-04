# Fixing Phoenix LiveView Recursive Rendering Error

## The Problem

When implementing a recursive file tree using DaisyUI menu components in Phoenix LiveView, we encountered this error:

```
lists in Phoenix.HTML and templates may only contain integers representing bytes,
binaries or other lists, got invalid entry: %Phoenix.LiveView.Rendered{...}
```

## Root Cause Analysis

### What Was Happening

The original implementation used a helper function that returned a list of rendered templates:

```elixir
defp render_file_tree(items, project_id) do
  Enum.map(items, fn item ->
    case item.type do
      :directory ->
        assigns = %{item: item, project_id: project_id}
        ~H"""
        <li>
          <details>
            <summary><%= @item.name %></summary>
            <ul>
              <%= render_file_tree(@item.children, @project_id) %>
            </ul>
          </details>
        </li>
        """
      :file ->
        # Similar pattern...
    end
  end)
end
```

### Why It Failed

1. **Enum.map Returns a List**: When `Enum.map` iterates over items and each iteration returns a ~H block, you get a list like `[%Phoenix.LiveView.Rendered{}, %Phoenix.LiveView.Rendered{}, ...]`

2. **Nested Rendering Problem**: When the recursive call `<%= render_file_tree(@item.children, @project_id) %>` executes, it returns that list of `Phoenix.LiveView.Rendered` structs

3. **Phoenix Template Constraint**: Phoenix templates can only embed:
   - Integers (representing byte values)
   - Binaries (strings)
   - Lists of the above
   - **NOT** lists of `Phoenix.LiveView.Rendered` structs

4. **The Recursive Trap**: Each level of recursion compounds the problem - you're trying to embed a list of rendered structs inside another rendered struct, creating nested incompatible types

## The Solution: Function Components with `attr`

### Step 1: Define Attributes

Phoenix LiveView function components use the `attr` directive for type-safe parameters:

```elixir
attr :item, :map, required: true
attr :project_id, :integer, required: true

defp tree_item(assigns) do
  # component implementation
end
```

**Why This Matters**:
- `attr` declarations tell Phoenix LiveView what data the component expects
- Provides compile-time validation
- Makes the component self-documenting
- Required for the `:for` comprehension syntax to work properly

### Step 2: Implement the Component Function

```elixir
defp tree_item(assigns) do
  case assigns.item.type do
    :directory ->
      ~H"""
      <li>
        <details>
          <summary>
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
              <path d="M1.75 1A1.75 1.75 0 0 0 0 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0 0 16 13.25v-8.5A1.75 1.75 0 0 0 14.25 3H7.5a.25.25 0 0 1-.2-.1l-.9-1.2C6.07 1.26 5.55 1 5 1H1.75Z" />
            </svg>
            <%= @item.name %>
          </summary>
          <ul>
            <.tree_item :for={child <- @item.children} item={child} project_id={@project_id} />
          </ul>
        </details>
      </li>
      """

    :file ->
      ~H"""
      <li>
        <a href={~p"/projects/#{@project_id}/files?path=#{@item.path}"}>
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
            <path d="M4 0a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2H4zm0 1h8a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1z"/>
          </svg>
          <%= @item.name %>
          <%= if @item.size do %>
            <span class="badge badge-ghost badge-xs ml-auto"><%= @item.size %></span>
          <% end %>
        </a>
      </li>
      """
  end
end
```

**Key Differences from Before**:
- Function takes `assigns` map as single parameter (required for function components)
- No manual `assigns = %{...}` creation needed
- Uses `case` on `assigns.item.type` instead of pattern matching in function head

### Step 3: The Magic - `:for` Comprehension Syntax

The recursive call uses special LiveView syntax:

```elixir
<.tree_item :for={child <- @item.children} item={child} project_id={@project_id} />
```

**Breaking Down the Syntax**:

- `<.tree_item ... />`: Function component invocation (the `.` prefix indicates local component)
- `:for={child <- @item.children}`: Special attribute that tells LiveView to render the component once for each child
- `item={child}`: Passes each child as the `item` attribute
- `project_id={@project_id}`: Passes through the project_id

**Why This Works**:
1. Phoenix LiveView handles the `:for` comprehension internally
2. It properly manages the rendering lifecycle for each iteration
3. Each component instance is rendered in isolation
4. The results are properly flattened and merged into the parent template
5. No manual list creation or `Enum.map` needed

### Step 4: Update the Main Template

Change from:
```elixir
<ul class="menu menu-sm bg-base-200 rounded-lg">
  <%= render_file_tree(@file_tree, @project.id) %>
</ul>
```

To:
```elixir
<ul class="menu menu-sm bg-base-200 rounded-lg">
  <.tree_item :for={item <- @file_tree} item={item} project_id={@project.id} />
</ul>
```

## How Phoenix LiveView Handles This Internally

### The Rendering Pipeline

1. **Component Registration**: When you use `attr`, Phoenix registers the component with its expected parameters

2. **`:for` Expansion**: When LiveView encounters `:for={child <- list}`, it:
   - Iterates over the list
   - Creates a new component instance for each item
   - Passes the item as the specified attribute
   - Collects all rendered results

3. **Recursive Resolution**: For nested components:
   - Each level of recursion is handled independently
   - Children are rendered before parents
   - Results bubble up through the component tree
   - Final output is a single, flat HTML structure

4. **Memory Management**: LiveView properly manages the temporary `Phoenix.LiveView.Rendered` structs and converts them to the final HTML at the right time

## Why Enum.map Doesn't Work

When you use `Enum.map` with ~H blocks:

```elixir
Enum.map(items, fn item -> ~H"""<li>...</li>""" end)
# Returns: [%Phoenix.LiveView.Rendered{}, %Phoenix.LiveView.Rendered{}, ...]
```

This creates a raw list that Phoenix doesn't know how to handle when you try to embed it with `<%= ... %>`.

The `:for` comprehension, however, is parsed and handled by LiveView's template compiler, which knows exactly how to manage the rendering lifecycle.

## Alternative Approaches (Not Recommended)

### 1. Using Phoenix.HTML.raw (Dangerous)
```elixir
<%= Phoenix.HTML.raw(render_file_tree(@file_tree, @project.id)) %>
```
- Requires converting to strings manually
- Loses LiveView reactivity
- Security risk with user content
- Performance overhead

### 2. Building HTML Strings (Ugly)
```elixir
defp render_file_tree(items, _) do
  Enum.map_join(items, "", fn item ->
    "<li>#{item.name}</li>"
  end)
end
```
- Loses template safety
- No XSS protection
- Harder to maintain
- Mixing concerns

### 3. Flattening Lists (Hacky)
```elixir
<%= List.flatten(render_file_tree(@file_tree, @project.id)) %>
```
- Still doesn't solve the type mismatch
- Doesn't work with nested structures
- Unclear intent

## Best Practices for Recursive LiveView Components

### 1. Always Use Function Components for Recursion
```elixir
attr :node, :map, required: true
defp recursive_component(assigns) do
  ~H"""
  <div>
    <%= @node.value %>
    <.recursive_component :for={child <- @node.children} node={child} />
  </div>
  """
end
```

### 2. Declare All Attributes Explicitly
```elixir
attr :item, :map, required: true
attr :depth, :integer, default: 0
attr :max_depth, :integer, default: 5
attr :project_id, :integer, required: true
```

### 3. Handle Edge Cases in Data Preparation
```elixir
def mount(_params, _session, socket) do
  tree = build_tree()
    |> limit_depth(max_depth: 5)  # Prevent infinite recursion
    |> sort_items()                # Prepare data upfront

  {:ok, assign(socket, :tree, tree)}
end
```

### 4. Use Guards for Safety
```elixir
defp tree_item(assigns) when is_map(assigns.item) do
  # Implementation
end
```

### 5. Consider Performance with Large Trees
```elixir
# Lazy load deep levels
defp tree_item(assigns) do
  ~H"""
  <li>
    <%= if @depth < 3 do %>
      <.tree_item :for={child <- @item.children} item={child} depth={@depth + 1} />
    <% else %>
      <button phx-click="load_children" phx-value-id={@item.id}>Load more...</button>
    <% end %>
  </li>
  """
end
```

## Common Pitfalls to Avoid

1. **Forgetting `attr` declarations**: Without them, `:for` won't work
2. **Using `Enum.map` in templates**: Always use `:for` comprehension instead
3. **Manual assigns creation**: Let Phoenix handle it via function component pattern
4. **Mixing patterns**: Choose either components or helpers, don't mix both
5. **Deep recursion without limits**: Always implement max depth protection

## Testing Recursive Components

```elixir
defmodule MyAppWeb.TreeComponentTest do
  use MyAppWeb.ConnCase, async: true
  import Phoenix.LiveViewTest

  test "renders nested tree structure" do
    tree = [
      %{name: "folder", type: :directory, children: [
        %{name: "file.txt", type: :file, size: "1KB", path: "folder/file.txt"}
      ]}
    ]

    html = render_component(&tree_item/1, item: tree, project_id: 1)

    assert html =~ "folder"
    assert html =~ "file.txt"
    assert html =~ "<details>"
  end
end
```

## Summary

**Problem**: `Enum.map` with ~H blocks creates incompatible list types for Phoenix templates

**Solution**: Use Phoenix LiveView function components with `attr` declarations and `:for` comprehension syntax

**Key Insight**: Phoenix LiveView's template compiler handles `:for` specially, managing the rendering lifecycle properly where manual list operations fail

**File Modified**: `eye_in_the_sky_web/lib/eye_in_the_sky_web_web/live/project_live/files.ex`

**Lines Changed**: 193-230, 314

## References

- [Phoenix LiveView Components Documentation](https://hexdocs.pm/phoenix_live_view/Phoenix.Component.html)
- [HEEx Template Syntax](https://hexdocs.pm/phoenix_live_view/Phoenix.Component.html#sigil_H/2)
- [Function Components Guide](https://hexdocs.pm/phoenix_live_view/Phoenix.Component.html#module-function-components)
