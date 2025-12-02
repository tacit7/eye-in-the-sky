defmodule EyeInTheSkyWeb.Sessions do
  @moduledoc """
  The Sessions context for managing agent sessions.
  """

  import Ecto.Query, warn: false
  alias EyeInTheSkyWeb.Repo
  alias EyeInTheSkyWeb.Sessions.Session

  @doc """
  Returns the list of sessions.
  """
  def list_sessions do
    Repo.all(Session)
  end

  @doc """
  Returns the list of sessions for a specific agent.
  """
  def list_sessions_for_agent(agent_id) do
    Session
    |> where([s], s.agent_id == ^agent_id)
    |> order_by([s], desc: s.started_at)
    |> Repo.all()
  end

  @doc """
  Gets a single session.

  Raises `Ecto.NoResultsError` if the Session does not exist.
  """
  def get_session!(id) do
    Repo.get!(Session, id)
  end

  @doc """
  Gets a single session, returning {:ok, session} or {:error, :not_found}.

  This is the safe version that doesn't raise exceptions.
  """
  def get_session(id) do
    case Repo.get(Session, id) do
      nil -> {:error, :not_found}
      session -> {:ok, session}
    end
  end

  @doc """
  Gets a single session with logs preloaded.
  """
  def get_session_with_logs!(id) do
    Session
    |> preload(:logs)
    |> Repo.get!(id)
  end

  @doc """
  Creates a session.
  """
  def create_session(attrs \\ %{}) do
    %Session{}
    |> Session.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates a session.
  """
  def update_session(%Session{} = session, attrs) do
    session
    |> Session.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Ends a session by setting ended_at timestamp.
  """
  def end_session(%Session{} = session) do
    update_session(session, %{ended_at: DateTime.utc_now()})
  end

  @doc """
  Updates the claude_session_id for a session.
  """
  def update_claude_session_id(%Session{} = session, claude_session_id) do
    session
    |> Session.changeset(%{claude_session_id: claude_session_id})
    |> Repo.update()
  end

  @doc """
  Deletes a session.
  """
  def delete_session(%Session{} = session) do
    Repo.delete(session)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking session changes.
  """
  def change_session(%Session{} = session, attrs \\ %{}) do
    Session.changeset(session, attrs)
  end

  @doc """
  Lists active sessions (not ended).
  """
  def list_active_sessions do
    Session
    |> where([s], is_nil(s.ended_at))
    |> order_by([s], desc: s.started_at)
    |> Repo.all()
  end

  @doc """
  Returns session overview rows for the sessions table.
  Joins sessions with agents and projects to get complete information.
  """
  def list_session_overview_rows(opts \\ []) do
    limit = Keyword.get(opts, :limit, 20)

    from(s in Session,
      join: a in assoc(s, :agent),
      left_join: p in EyeInTheSkyWeb.Projects.Project,
      on: p.id == a.project_id,
      order_by: [desc: s.started_at],
      limit: ^limit,
      select: %{
        session_id: s.id,
        session_name: s.name,
        agent_id: a.id,
        project_name: p.name,
        started_at: s.started_at,
        ended_at: s.ended_at
      }
    )
    |> Repo.all()
  end

  @doc """
  Loads all data for a specific session.
  Returns tasks, commits, logs, notes, context, and metrics.
  """
  def load_session_data(session_id) do
    alias EyeInTheSkyWeb.{Tasks, Commits, Logs, Contexts}

    %{
      tasks: Tasks.list_tasks_for_session(session_id),
      commits: Commits.list_commits_for_session(session_id),
      logs: Logs.list_logs_for_session(session_id),
      notes: [], # TODO: Fix parent_id type mismatch (INTEGER vs TEXT)
      session_context: Contexts.get_session_context(session_id),
      metrics: nil # TODO: Add metrics when table exists
    }
  end

  @doc """
  Gets counts for all tabs (cheap aggregate queries).
  """
  def get_session_counts(session_id) do
    alias EyeInTheSkyWeb.{Tasks, Commits, Logs, Notes, Messages}

    %{
      tasks: Tasks.count_tasks_for_session(session_id),
      commits: Commits.count_commits_for_session(session_id),
      logs: Logs.count_logs_for_session(session_id),
      notes: Notes.count_notes_for_session(session_id),
      messages: Messages.count_messages_for_session(session_id)
    }
  end

  @doc """
  Lazy load: tasks only
  """
  def load_session_tasks(session_id) do
    EyeInTheSkyWeb.Tasks.list_tasks_for_session(session_id)
  end

  @doc """
  Lazy load: commits only
  """
  def load_session_commits(session_id, opts \\ []) do
    limit = Keyword.get(opts, :limit, 50)
    EyeInTheSkyWeb.Commits.list_commits_for_session(session_id, limit: limit)
  end

  @doc """
  Lazy load: logs only
  """
  def load_session_logs(session_id, opts \\ []) do
    limit = Keyword.get(opts, :limit, 100)
    EyeInTheSkyWeb.Logs.list_logs_for_session(session_id, limit: limit)
  end

  @doc """
  Lazy load: context only
  """
  def load_session_context(session_id) do
    EyeInTheSkyWeb.Contexts.get_session_context(session_id)
  end

  @doc """
  Lazy load: notes only
  """
  def load_session_notes(session_id) do
    EyeInTheSkyWeb.Notes.list_notes_for_session(session_id)
  end
end
