defmodule EyeInTheSkyWeb.Claude.SessionManager do
  @moduledoc """
  Manages multiple Claude CLI subprocess sessions.

  Tracks running processes, routes output to correct LiveView sessions,
  and handles process cleanup.
  """

  use GenServer
  require Logger

  alias EyeInTheSkyWeb.Claude.CLI
  alias EyeInTheSkyWeb.Messages
  alias EyeInTheSkyWeb.NATS.Publisher
  alias EyeInTheSkyWeb.Sessions

  # Client API

  def start_link(opts) do
    GenServer.start_link(__MODULE__, opts, name: __MODULE__)
  end

  @doc """
  Starts a new Claude CLI session.

  Returns `{:ok, session_ref}` which should be used to identify this session.
  """
  def start_session(session_id, prompt, opts \\ []) do
    GenServer.call(__MODULE__, {:start_session, session_id, prompt, opts})
  end

  @doc """
  Continues an existing Claude session.
  """
  def continue_session(session_id, prompt, opts \\ []) do
    GenServer.call(__MODULE__, {:continue_session, session_id, prompt, opts})
  end

  @doc """
  Resumes a specific Claude session by UUID.
  """
  def resume_session(claude_session_id, prompt, opts \\ []) do
    GenServer.call(__MODULE__, {:resume_session, claude_session_id, prompt, opts})
  end

  @doc """
  Cancels a running Claude session.
  """
  def cancel_session(session_ref) do
    GenServer.call(__MODULE__, {:cancel_session, session_ref})
  end

  @doc """
  Lists all active sessions.
  """
  def list_sessions do
    GenServer.call(__MODULE__, :list_sessions)
  end

  # Server Callbacks

  @impl true
  def init(_opts) do
    # State: %{session_ref => %{port, session_id, started_at, output_buffer}}
    {:ok, %{}}
  end

  @impl true
  def handle_call({:start_session, session_id, prompt, opts}, _from, state) do
    # Add this GenServer as the caller
    opts = Keyword.put(opts, :caller, self())

    case CLI.spawn_new_session(prompt, opts) do
      {:ok, port, session_ref} ->
        session_info = %{
          port: port,
          session_id: session_id,
          started_at: DateTime.utc_now(),
          output_buffer: [],
          claude_session_id: nil  # Will be extracted from init message
        }

        new_state = Map.put(state, session_ref, session_info)

        Logger.info("Started Claude CLI session #{inspect(session_ref)} for #{session_id}")

        {:reply, {:ok, session_ref}, new_state}

      {:error, reason} ->
        Logger.error("Failed to start Claude CLI: #{inspect(reason)}")
        {:reply, {:error, reason}, state}
    end
  end

  @impl true
  def handle_call({:continue_session, session_id, prompt, opts}, _from, state) do
    opts = Keyword.put(opts, :caller, self())

    case CLI.continue_session(prompt, opts) do
      {:ok, port, session_ref} ->
        session_info = %{
          port: port,
          session_id: session_id,
          started_at: DateTime.utc_now(),
          output_buffer: [],
          claude_session_id: nil
        }

        new_state = Map.put(state, session_ref, session_info)

        {:reply, {:ok, session_ref}, new_state}

      {:error, reason} ->
        {:reply, {:error, reason}, state}
    end
  end

  @impl true
  def handle_call({:resume_session, claude_session_id, prompt, opts}, _from, state) do
    # Note: session_id here is the Eye in the Sky session_id, extracted from opts if provided
    session_id = Keyword.get(opts, :session_id, claude_session_id)
    opts = Keyword.put(opts, :caller, self())

    case CLI.resume_session(claude_session_id, prompt, opts) do
      {:ok, port, session_ref} ->
        session_info = %{
          port: port,
          session_id: session_id,
          started_at: DateTime.utc_now(),
          output_buffer: [],
          claude_session_id: claude_session_id
        }

        new_state = Map.put(state, session_ref, session_info)

        Logger.info("Resumed Claude CLI session #{claude_session_id} (ref: #{inspect(session_ref)})")

        {:reply, {:ok, session_ref}, new_state}

      {:error, reason} ->
        Logger.error("Failed to resume Claude CLI session #{claude_session_id}: #{inspect(reason)}")
        {:reply, {:error, reason}, state}
    end
  end

  @impl true
  def handle_call({:cancel_session, session_ref}, _from, state) do
    case Map.get(state, session_ref) do
      nil ->
        {:reply, {:error, :not_found}, state}

      session_info ->
        CLI.cancel(session_info.port)
        new_state = Map.delete(state, session_ref)

        Logger.info("Cancelled Claude CLI session #{inspect(session_ref)}")

        {:reply, :ok, new_state}
    end
  end

  @impl true
  def handle_call(:list_sessions, _from, state) do
    sessions =
      Enum.map(state, fn {ref, info} ->
        %{
          session_ref: ref,
          session_id: info.session_id,
          claude_session_id: info.claude_session_id,
          started_at: info.started_at,
          output_lines: length(info.output_buffer)
        }
      end)

    {:reply, sessions, state}
  end

  @impl true
  def handle_info({:claude_output, session_ref, line}, state) do
    case Map.get(state, session_ref) do
      nil ->
        {:noreply, state}

      session_info ->
        # Parse the JSON line
        case Jason.decode(line) do
          {:ok, parsed} ->
            handle_parsed_output(session_ref, session_info, parsed, state)

          {:error, _} ->
            # Not JSON, just buffer it
            updated_info = update_in(session_info.output_buffer, &[line | &1])
            {:noreply, Map.put(state, session_ref, updated_info)}
        end
    end
  end

  @impl true
  def handle_info({:claude_exit, session_ref, exit_code}, state) do
    case Map.get(state, session_ref) do
      nil ->
        {:noreply, state}

      session_info ->
        Logger.info("Claude CLI session #{inspect(session_ref)} exited with code #{exit_code}")

        # Broadcast completion to LiveView
        Phoenix.PubSub.broadcast(
          EyeInTheSkyWeb.PubSub,
          "session:#{session_info.session_id}:claude",
          {:claude_complete, session_ref, exit_code}
        )

        # Clean up
        new_state = Map.delete(state, session_ref)
        {:noreply, new_state}
    end
  end

  # Private Helpers

  defp handle_parsed_output(session_ref, session_info, parsed, state) do
    # Extract Claude session ID from init message
    updated_info =
      if parsed["type"] == "system" && parsed["subtype"] == "init" do
        claude_session_id = parsed["session_id"]
        Logger.info("Claude session ID: #{claude_session_id}")

        # Store Claude session ID in database
        case Sessions.get_session!(session_info.session_id) do
          session when not is_nil(session) ->
            Sessions.update_claude_session_id(session, claude_session_id)
            Logger.info("Stored Claude session ID #{claude_session_id} for session #{session_info.session_id}")
          _ ->
            Logger.warning("Could not find session #{session_info.session_id} to store Claude session ID")
        end

        %{session_info | claude_session_id: claude_session_id}
      else
        session_info
      end

    # Handle assistant messages
    updated_info =
      if parsed["type"] == "assistant" do
        content = parsed["content"] || parsed["message"]

        # Create inbound message in database
        if content do
          {:ok, message} = Messages.record_incoming_reply(
            session_info.session_id,
            "claude",  # provider
            content
          )

          # Publish agent reply to NATS for agent-to-agent communication
          Publisher.publish_message(message)
        end

        updated_info
      else
        updated_info
      end

    # Buffer the output
    updated_info = update_in(updated_info.output_buffer, &[parsed | &1])

    # Broadcast to LiveView
    Phoenix.PubSub.broadcast(
      EyeInTheSkyWeb.PubSub,
      "session:#{session_info.session_id}:claude",
      {:claude_output, session_ref, parsed}
    )

    {:noreply, Map.put(state, session_ref, updated_info)}
  end
end
