defmodule EyeInTheSkyWeb.Commits do
  @moduledoc """
  The Commits context for managing git commits.
  """

  import Ecto.Query, warn: false
  alias EyeInTheSkyWeb.Repo
  alias EyeInTheSkyWeb.Commits.Commit

  @doc """
  Returns the list of commits.
  """
  def list_commits do
    Repo.all(Commit)
  end

  @doc """
  Returns the list of commits for a specific agent.
  """
  def list_commits_for_agent(agent_id) do
    Commit
    |> where([c], c.agent_id == ^agent_id)
    |> order_by([c], desc: c.created_at)
    |> Repo.all()
  end

  @doc """
  Returns recent commits for an agent with a limit.
  """
  def list_recent_commits(agent_id, limit \\ 10) do
    Commit
    |> where([c], c.agent_id == ^agent_id)
    |> order_by([c], desc: c.created_at)
    |> limit(^limit)
    |> Repo.all()
  end

  @doc """
  Returns commits for a specific session.
  """
  def list_commits_for_session(session_id, opts \\ []) do
    limit = Keyword.get(opts, :limit)

    query =
      Commit
      |> where([c], c.session_id == ^session_id)
      |> order_by([c], desc: c.created_at)

    query = if limit, do: limit(query, ^limit), else: query

    Repo.all(query)
  end

  @doc """
  Counts commits for a specific session.
  """
  def count_commits_for_session(session_id) do
    Commit
    |> where([c], c.session_id == ^session_id)
    |> select([c], count(c.id))
    |> Repo.one() || 0
  end

  @doc """
  Gets a single commit.

  Raises `Ecto.NoResultsError` if the Commit does not exist.
  """
  def get_commit!(id) do
    Repo.get!(Commit, id)
  end

  @doc """
  Gets a commit by hash.
  """
  def get_commit_by_hash(hash) do
    Repo.get_by(Commit, commit_hash: hash)
  end

  @doc """
  Creates a commit.
  """
  def create_commit(attrs \\ %{}) do
    %Commit{}
    |> Commit.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates a commit.
  """
  def update_commit(%Commit{} = commit, attrs) do
    commit
    |> Commit.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Deletes a commit.
  """
  def delete_commit(%Commit{} = commit) do
    Repo.delete(commit)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking commit changes.
  """
  def change_commit(%Commit{} = commit, attrs \\ %{}) do
    Commit.changeset(commit, attrs)
  end
end
