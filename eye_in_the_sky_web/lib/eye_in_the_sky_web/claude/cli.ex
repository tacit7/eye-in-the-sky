defmodule EyeInTheSkyWeb.Claude.CLI do
  @moduledoc """
  Claude CLI subprocess spawner - spawns fresh Claude Code instances like opcode does.

  Spawns `claude` binary as a subprocess with `-p "message"` flag and streams
  stdout/stderr back to the caller via message passing.
  """

  require Logger

  @doc """
  Spawns a new Claude Code session with a user prompt.

  ## Options
    * `:model` - Model to use ("sonnet", "opus", "haiku"). Default: "sonnet"
    * `:project_path` - Working directory for Claude. Default: current directory
    * `:output_format` - Output format. Default: "stream-json"
    * `:skip_permissions` - Skip permission prompts. Default: true
    * `:caller` - PID to send output to. Default: self()

  ## Returns
    * `{:ok, port, session_ref}` - Port handle and session reference
    * `{:error, reason}` - If Claude binary not found or spawn failed

  ## Messages Sent to Caller
    * `{:claude_output, session_ref, line}` - Each line of stdout
    * `{:claude_error, session_ref, line}` - Each line of stderr
    * `{:claude_exit, session_ref, exit_code}` - Process exited
  """
  def spawn_new_session(prompt, opts \\ []) do
    model = Keyword.get(opts, :model, "sonnet")
    project_path = Keyword.get(opts, :project_path, File.cwd!())
    output_format = Keyword.get(opts, :output_format, "stream-json")
    skip_permissions = Keyword.get(opts, :skip_permissions, true)
    caller = Keyword.get(opts, :caller, self())

    # Find claude binary
    case find_claude_binary() do
      {:ok, claude_path} ->
        # Build command args like opcode does
        args = build_args(prompt, model, output_format, skip_permissions)

        # DEBUG: Log the command being executed
        require Logger
        Logger.info("🚀 CLAUDE COMMAND: cd #{project_path} && #{claude_path} #{Enum.join(args, " ")}")
        Logger.info("Spawning Claude: #{claude_path} #{Enum.join(args, " ")}")
        Logger.info("Project path: #{project_path}")

        session_ref = make_ref()

        # Spawn output handler first
        handler_pid = spawn_link(fn ->
          receive do
            {:port, port} ->
              handle_port_output(port, session_ref, caller)
          end
        end)

        # Spawn the process using 'script' to provide a pseudo-TTY
        # macOS: script -q /dev/null command args...
        script_args = ["-q", "/dev/null", claude_path] ++ args

        port = Port.open(
          {:spawn_executable, "/usr/bin/script"},
          [
            :binary,
            :exit_status,
            :use_stdio,
            :stderr_to_stdout,
            {:args, script_args},
            {:cd, project_path},
            {:env, build_env()}
          ]
        )

        # Connect port to handler
        Port.connect(port, handler_pid)
        send(handler_pid, {:port, port})

        {:ok, port, session_ref}

      {:error, reason} ->
        {:error, reason}
    end
  end

  @doc """
  Continues an existing session (uses `-c` flag).
  """
  def continue_session(prompt, opts \\ []) do
    opts = Keyword.put(opts, :continue, true)
    spawn_with_flag(prompt, "-c", opts)
  end

  @doc """
  Resumes a specific session by UUID (uses `--resume` flag).
  """
  def resume_session(session_id, prompt, opts \\ []) do
    opts = Keyword.put(opts, :session_id, session_id)
    spawn_with_flag(prompt, "--resume", opts)
  end

  @doc """
  Cancels a running Claude process.
  """
  def cancel(port) when is_port(port) do
    Port.close(port)
    :ok
  end

  # Private functions

  defp spawn_with_flag(prompt, flag, opts) do
    model = Keyword.get(opts, :model, "sonnet")
    project_path = Keyword.get(opts, :project_path, File.cwd!())
    output_format = Keyword.get(opts, :output_format, "stream-json")
    skip_permissions = Keyword.get(opts, :skip_permissions, true)
    caller = Keyword.get(opts, :caller, self())
    session_id = Keyword.get(opts, :session_id)

    case find_claude_binary() do
      {:ok, claude_path} ->
        # Build args with continue or resume flag
        base_args = build_args(prompt, model, output_format, skip_permissions)
        args =
          if flag == "--resume" && session_id do
            [flag, session_id] ++ base_args
          else
            [flag] ++ base_args
          end

        # DEBUG: Log the command being executed
        require Logger
        Logger.info("🚀 CLAUDE COMMAND: cd #{project_path} && #{claude_path} #{Enum.join(args, " ")}")
        Logger.info("Spawning Claude with #{flag}: #{claude_path} #{Enum.join(args, " ")}")
        Logger.info("Project path: #{project_path}")

        session_ref = make_ref()

        # Spawn output handler first
        handler_pid = spawn_link(fn ->
          receive do
            {:port, port} ->
              handle_port_output(port, session_ref, caller)
          end
        end)

        # Spawn the process using 'script' to provide a pseudo-TTY
        script_args = ["-q", "/dev/null", claude_path] ++ args

        port = Port.open(
          {:spawn_executable, "/usr/bin/script"},
          [
            :binary,
            :exit_status,
            :use_stdio,
            :stderr_to_stdout,
            {:args, script_args},
            {:cd, project_path},
            {:env, build_env()}
          ]
        )

        # Connect port to handler
        Port.connect(port, handler_pid)
        send(handler_pid, {:port, port})

        {:ok, port, session_ref}

      {:error, reason} ->
        {:error, reason}
    end
  end

  defp build_args(prompt, model, output_format, skip_permissions) do
    base = [
      "-p", prompt,
      "--output-format", output_format,
      "--verbose"
    ]

    if skip_permissions do
      base ++ ["--dangerously-skip-permissions"]
    else
      base
    end
  end

  defp build_env do
    # Pass ALL environment variables to subprocess
    base_env = for {key, value} <- System.get_env() do
      {String.to_charlist(key), String.to_charlist(value)}
    end

    # Force non-interactive mode for Claude (disable TTY requirements)
    [
      {'CI', 'true'},  # Tell Claude it's running in CI (no TTY)
      {'TERM', 'dumb'} # Disable terminal features
      | base_env
    ]
  end

  defp handle_port_output(port, session_ref, caller) do
    require Logger

    receive do
      {^port, {:data, data}} ->
        Logger.debug("Claude output received: #{byte_size(data)} bytes")
        # Split by newlines and send each line
        data
        |> String.split("\n", trim: true)
        |> Enum.each(fn line ->
          Logger.debug("Claude line: #{line}")
          send(caller, {:claude_output, session_ref, line})
        end)

        handle_port_output(port, session_ref, caller)

      {^port, {:exit_status, status}} ->
        Logger.info("Claude process exited with status #{status}")
        send(caller, {:claude_exit, session_ref, status})
        :ok

      after
        300_000 ->
          Logger.warning("No output from Claude after 5 minutes, timing out")
          Port.close(port)  # Kill the subprocess to prevent zombie process
          send(caller, {:claude_exit, session_ref, :timeout})
          :ok
    end
  end

  defp find_claude_binary do
    # Try multiple detection strategies like opcode
    cond do
      # 1. Check which/where command
      path = System.find_executable("claude") ->
        {:ok, path}

      # 2. Check standard locations
      path = find_in_standard_paths() ->
        {:ok, path}

      # 3. Check NVM installations
      path = find_in_nvm() ->
        {:ok, path}

      true ->
        {:error, "Claude binary not found in PATH, standard locations, or NVM"}
    end
  end

  defp find_in_standard_paths do
    standard_paths = [
      "/usr/local/bin/claude",
      "/opt/homebrew/bin/claude",
      Path.expand("~/.local/bin/claude")
    ]

    Enum.find(standard_paths, &File.exists?/1)
  end

  defp find_in_nvm do
    nvm_dir = System.get_env("NVM_DIR") || Path.expand("~/.nvm")
    versions_dir = Path.join(nvm_dir, "versions/node")

    if File.dir?(versions_dir) do
      versions_dir
      |> File.ls!()
      |> Enum.map(&Path.join([versions_dir, &1, "bin", "claude"]))
      |> Enum.filter(&File.exists?/1)
      |> List.first()
    else
      nil
    end
  end
end
