defmodule EyeInTheSkyWeb.Logs do
  @moduledoc """
  The Logs context for managing logs and session logs.
  """

  import Ecto.Query, warn: false
  alias EyeInTheSkyWeb.Repo
  alias EyeInTheSkyWeb.Logs.{Log, SessionLog}

  # Log functions

  @doc """
  Returns the list of logs.
  """
  def list_logs do
    Repo.all(Log)
  end

  @doc """
  Returns the list of logs for a specific session.
  """
  def list_logs_for_session(session_id, opts \\ []) do
    limit = Keyword.get(opts, :limit)

    query =
      Log
      |> where([l], l.session_id == ^session_id)
      |> order_by([l], desc: l.timestamp)

    query = if limit, do: limit(query, ^limit), else: query

    Repo.all(query)
  end

  @doc """
  Returns recent logs for a session with a limit.
  """
  def list_recent_logs(session_id, limit \\ 50) do
    list_logs_for_session(session_id, limit: limit)
  end

  @doc """
  Counts logs for a specific session.
  """
  def count_logs_for_session(session_id) do
    Log
    |> where([l], l.session_id == ^session_id)
    |> select([l], count(l.id))
    |> Repo.one() || 0
  end

  @doc """
  Filters logs by type.
  """
  def filter_logs_by_type(session_id, type) do
    Log
    |> where([l], l.session_id == ^session_id and l.type == ^type)
    |> order_by([l], desc: l.timestamp)
    |> Repo.all()
  end

  @doc """
  Gets a single log.
  """
  def get_log!(id) do
    Repo.get!(Log, id)
  end

  @doc """
  Creates a log.
  """
  def create_log(attrs \\ %{}) do
    %Log{}
    |> Log.changeset(attrs)
    |> Repo.insert()
  end

  # SessionLog functions

  @doc """
  Returns the list of session logs.
  """
  def list_session_logs do
    Repo.all(SessionLog)
  end

  @doc """
  Returns the list of session logs for a specific session.
  """
  def list_session_logs_for_session(session_id) do
    SessionLog
    |> where([sl], sl.session_id == ^session_id)
    |> order_by([sl], desc: sl.timestamp)
    |> Repo.all()
  end

  @doc """
  Filters session logs by log level.
  """
  def filter_session_logs_by_level(session_id, log_level) do
    SessionLog
    |> where([sl], sl.session_id == ^session_id and sl.log_level == ^log_level)
    |> order_by([sl], desc: sl.timestamp)
    |> Repo.all()
  end

  @doc """
  Creates a session log.
  """
  def create_session_log(attrs \\ %{}) do
    %SessionLog{}
    |> SessionLog.changeset(attrs)
    |> Repo.insert()
  end
end
