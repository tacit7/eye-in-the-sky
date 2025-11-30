defmodule EyeInTheSkyWebWeb.AgentLive.Show do
  use EyeInTheSkyWebWeb, :live_view

  alias EyeInTheSkyWeb.{Agents, Sessions, Messages}
  alias EyeInTheSkyWeb.Claude.SessionManager

  @impl true
  def mount(_params, _session, socket) do
    {:ok, socket}
  end

  @impl true
  def handle_params(%{"id" => id} = params, _, socket) do
    # Get agent + sessions list (lightweight)
    dashboard_data = Agents.get_agent_dashboard_data(id)

    # Determine selected session from URL
    session_id = params["session_id"] || params["s"]
    active_session =
      case session_id do
        nil -> dashboard_data.active_session
        sid -> Enum.find(dashboard_data.sessions, &(&1.id == sid)) || dashboard_data.active_session
      end

    # Determine active tab from URL (default: tasks)
    active_tab =
      case params["tab"] do
        "commits" -> :commits
        "logs" -> :logs
        "context" -> :context
        "notes" -> :notes
        "messages" -> :messages
        _ -> :tasks
      end

    # Compute counts for all tabs (cheap query)
    counts =
      if active_session do
        Sessions.get_session_counts(active_session.id)
      else
        %{tasks: 0, commits: 0, logs: 0, notes: 0, messages: 0}
      end

    # Load ONLY the data for the active tab
    tab_data =
      if active_session do
        load_tab_data(active_tab, active_session.id)
      else
        %{}
      end

    # Build header info
    header = build_header(dashboard_data.agent, active_session)

    # Subscribe to Claude CLI output and message updates for this session
    if connected?(socket) && active_session do
      Phoenix.PubSub.subscribe(EyeInTheSkyWeb.PubSub, "session:#{active_session.id}:messages")
      Phoenix.PubSub.subscribe(EyeInTheSkyWeb.PubSub, "session:#{active_session.id}:claude")
    end

    socket =
      socket
      |> assign(:page_title, "Session #{active_session && String.slice(active_session.id, 0..7)}")
      |> assign(:agent_id, id)
      |> assign(:header, header)
      |> assign(:session_id, active_session && active_session.id)
      |> assign(:active_tab, active_tab)
      |> assign(:counts, counts)
      |> assign(:tasks, Map.get(tab_data, :tasks))
      |> assign(:commits, Map.get(tab_data, :commits))
      |> assign(:logs, Map.get(tab_data, :logs))
      |> assign(:context, Map.get(tab_data, :context))
      |> assign(:notes, Map.get(tab_data, :notes))
      |> assign(:messages, Map.get(tab_data, :messages))

    {:noreply, socket}
  end

  @impl true
  def handle_event("change_tab", %{"tab" => tab}, socket) do
    # Navigate to same page with different tab param
    {:noreply,
      push_patch(socket,
        to: ~p"/agents/#{socket.assigns.agent_id}/sessions/#{socket.assigns.session_id}?tab=#{tab}"
      )
    }
  end

  @impl true
  def handle_event("select_session", %{"session_id" => session_id}, socket) do
    # Navigate to new session, reset to tasks tab
    {:noreply,
      push_patch(socket,
        to: ~p"/agents/#{socket.assigns.agent_id}/sessions/#{session_id}?tab=tasks"
      )
    }
  end

  @impl true
  def handle_event("copy_session_id", _params, socket) do
    # Client-side copy handled by JS hook
    {:noreply, socket}
  end

  @impl true
  def handle_event("end_session", _params, socket) do
    # TODO: Call MCP server to end session
    {:noreply, socket}
  end

  @impl true
  def handle_event("new_task", _params, socket) do
    # TODO: open a modal or navigate to tasks tab in a future PR
    {:noreply, socket}
  end

  @impl true
  def handle_event("add_note", _params, socket) do
    # TODO: open a note input in a future PR
    {:noreply, socket}
  end

  @impl true
  def handle_event("send_message", %{"body" => body, "provider" => provider}, socket) do
    session_id = socket.assigns.session_id

    # Create outbound message
    {:ok, message} = Messages.send_message(%{
      session_id: session_id,
      sender_role: "user",
      recipient_role: "agent",
      provider: provider,
      body: body
    })

    # Spawn Claude CLI subprocess with the message
    case SessionManager.start_session(session_id, body, model: provider_to_model(provider)) do
      {:ok, _session_ref} ->
        # Reload messages for the current tab
        updated_messages = serialize_messages(Messages.list_messages_for_session(session_id))
        {:noreply, assign(socket, :messages, updated_messages)}

      {:error, reason} ->
        {:noreply, put_flash(socket, :error, "Failed to start Claude: #{inspect(reason)}")}
    end
  end

  defp provider_to_model("claude"), do: "sonnet"
  defp provider_to_model("openai"), do: "sonnet"  # For now, always use Claude
  defp provider_to_model(_), do: "sonnet"

  @impl true
  def handle_info({:new_message, _message}, socket) do
    # Message received from NATS consumer via PubSub or Claude CLI
    session_id = socket.assigns.session_id

    # Reload messages and update UI
    updated_messages = serialize_messages(Messages.list_messages_for_session(session_id))

    # Update message count
    counts = Sessions.get_session_counts(session_id)

    socket =
      socket
      |> assign(:messages, updated_messages)
      |> assign(:counts, counts)

    {:noreply, socket}
  end

  @impl true
  def handle_info({:claude_output, _session_ref, parsed}, socket) do
    # Real-time Claude CLI output streaming
    # Just trigger a reload - the SessionManager already saved it to database
    if socket.assigns.active_tab == :messages do
      session_id = socket.assigns.session_id
      updated_messages = serialize_messages(Messages.list_messages_for_session(session_id))
      {:noreply, assign(socket, :messages, updated_messages)}
    else
      {:noreply, socket}
    end
  end

  @impl true
  def handle_info({:claude_complete, _session_ref, exit_code}, socket) do
    # Claude process completed
    if exit_code == 0 do
      {:noreply, put_flash(socket, :info, "Claude completed successfully")}
    else
      {:noreply, put_flash(socket, :error, "Claude exited with code #{exit_code}")}
    end
  end

  # Lazy load tab data
  defp load_tab_data(:tasks, session_id) do
    %{tasks: serialize_tasks(Sessions.load_session_tasks(session_id))}
  end

  defp load_tab_data(:commits, session_id) do
    %{commits: serialize_commits(Sessions.load_session_commits(session_id))}
  end

  defp load_tab_data(:logs, session_id) do
    %{logs: serialize_logs(Sessions.load_session_logs(session_id, limit: 100))}
  end

  defp load_tab_data(:context, session_id) do
    %{context: serialize_context(Sessions.load_session_context(session_id))}
  end

  defp load_tab_data(:notes, session_id) do
    %{notes: serialize_notes(Sessions.load_session_notes(session_id))}
  end

  defp load_tab_data(:messages, session_id) do
    %{messages: serialize_messages(Messages.list_messages_for_session(session_id))}
  end

  # Serialization functions
  defp serialize_tasks(tasks) when is_list(tasks) do
    Enum.map(tasks, fn task ->
      %{
        id: task.id,
        title: task.title,
        description: task.description,
        priority: task.priority,
        state_name: task.state && task.state.name,
        tags: task.tags && Enum.map(task.tags, &%{id: &1.id, name: &1.name}),
        created_at: task.created_at
      }
    end)
  end
  defp serialize_tasks(_), do: []

  defp serialize_commits(commits) when is_list(commits) do
    Enum.map(commits, fn commit ->
      %{
        id: commit.id,
        commit_hash: commit.commit_hash,
        commit_message: commit.commit_message,
        created_at: commit.created_at
      }
    end)
  end
  defp serialize_commits(_), do: []

  defp serialize_logs(logs) when is_list(logs) do
    Enum.map(logs, fn log ->
      %{
        id: log.id,
        type: log.type,
        message: log.message,
        timestamp: log.timestamp
      }
    end)
  end
  defp serialize_logs(_), do: []

  defp serialize_context(nil), do: nil
  defp serialize_context(context) do
    %{
      context: context.context,
      current_phase: context.current_phase,
      progress_percentage: context.progress_percentage
    }
  end

  defp serialize_notes(notes) when is_list(notes) do
    Enum.map(notes, fn note ->
      %{
        id: note.id,
        body: note.body,
        created_at: note.created_at
      }
    end)
  end
  defp serialize_notes(_), do: []

  defp serialize_messages(messages) when is_list(messages) do
    Enum.map(messages, fn message ->
      %{
        id: message.id,
        sender_role: message.sender_role,
        recipient_role: message.recipient_role,
        direction: message.direction,
        body: message.body,
        status: message.status,
        provider: message.provider,
        inserted_at: message.inserted_at
      }
    end)
  end
  defp serialize_messages(_), do: []

  # Build header map
  defp build_header(agent, nil) do
    %{
      agent_id: agent.id,
      agent_type: "Claude Code",  # TODO: Get from agent table
      status: agent.status,
      session_id: nil,
      session_name: nil,
      project: agent.project_name,
      duration: nil,
      started: nil
    }
  end

  defp build_header(agent, session) do
    %{
      agent_id: agent.id,
      agent_type: "Claude Code",  # TODO: Get from agent table
      status: session_status(session),
      session_id: session.id,
      session_name: session.name,
      project: agent.project_name,
      duration: format_duration(session.started_at, session.ended_at),
      started: format_timestamp(session.started_at)
    }
  end

  defp session_status(session) do
    if session.ended_at && session.ended_at != "", do: "completed", else: "active"
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div style="opacity: 1 !important;" class="agent-detail-wrapper">
      <.svelte
        name="AgentDetail"
        props={%{
          header: @header,
          sessionId: @session_id,
          activeTab: Atom.to_string(@active_tab),
          counts: @counts,

          tasks: @tasks,         # loaded when tab is tasks (default)
          commits: @commits,     # loaded when tab is commits
          logs: @logs,           # loaded when tab is logs
          context: @context,     # loaded when tab is context
          notes: @notes,         # loaded when tab is notes
          messages: @messages    # loaded when tab is messages
        }}
        socket={@socket}
      />
    </div>
    """
  end

  defp format_timestamp(nil), do: nil
  defp format_timestamp(""), do: nil

  defp format_timestamp(%DateTime{} = datetime) do
    Calendar.strftime(datetime, "%Y-%m-%d %H:%M")
  end

  defp format_timestamp(timestamp) when is_binary(timestamp) do
    case String.split(timestamp, " ", parts: 3) do
      [date, time | _] -> "#{date} #{String.slice(time, 0..7)}"
      _ -> timestamp
    end
  end

  defp format_duration(_started, nil), do: "Active"
  defp format_duration(_started, ""), do: "Active"
  defp format_duration(started, ended) when is_binary(started) and is_binary(ended) do
    # TODO: Calculate actual duration from timestamps
    "Ended"
  end
  defp format_duration(_, _), do: nil
end
