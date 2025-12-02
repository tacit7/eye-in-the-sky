defmodule EyeInTheSkyWeb.Agents do
  @moduledoc """
  The Agents context for managing agents and their lifecycle.
  """

  import Ecto.Query, warn: false
  alias EyeInTheSkyWeb.Repo
  alias EyeInTheSkyWeb.Agents.Agent

  @doc """
  Returns the list of agents.
  """
  def list_agents do
    Repo.all(Agent)
  end

  @doc """
  Returns the list of agents with their sessions and tasks preloaded for display.
  Sorted by last activity (most recent first).
  """
  def list_agents_with_sessions do
    Agent
    |> preload([:sessions, :tasks])
    |> order_by([a], desc: a.last_activity_at)
    |> Repo.all()
  end

  @doc """
  Returns the list of active agents (not completed or failed).
  """
  def list_active_agents do
    Agent
    |> where([a], a.status not in ["completed", "failed"])
    |> order_by([a], desc: a.last_activity_at)
    |> Repo.all()
  end

  @doc """
  Gets a single agent.

  Raises `Ecto.NoResultsError` if the Agent does not exist.
  """
  def get_agent!(id) do
    Repo.get!(Agent, id)
  end

  @doc """
  Gets a single agent, returning {:ok, agent} or {:error, :not_found}.

  This is the safe version that doesn't raise exceptions.
  """
  def get_agent(id) do
    case Repo.get(Agent, id) do
      nil -> {:error, :not_found}
      agent -> {:ok, agent}
    end
  end

  @doc """
  Gets a single agent with all associations preloaded.
  """
  def get_agent_with_associations!(id) do
    Agent
    |> preload([:sessions, :commits, :tasks, :project])
    |> Repo.get!(id)
  end

  @doc """
  Creates an agent.
  """
  def create_agent(attrs \\ %{}) do
    %Agent{}
    |> Agent.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates an agent.
  """
  def update_agent(%Agent{} = agent, attrs) do
    agent
    |> Agent.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Updates an agent's status.
  """
  def update_agent_status(%Agent{} = agent, status) do
    update_agent(agent, %{status: status, last_activity_at: DateTime.utc_now()})
  end

  @doc """
  Deletes an agent.
  """
  def delete_agent(%Agent{} = agent) do
    Repo.delete(agent)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking agent changes.
  """
  def change_agent(%Agent{} = agent, attrs \\ %{}) do
    Agent.changeset(agent, attrs)
  end

  @doc """
  Lists agents by parent agent ID (subagents).
  """
  def list_subagents(parent_agent_id) do
    Agent
    |> where([a], a.parent_agent_id == ^parent_agent_id)
    |> order_by([a], asc: a.created_at)
    |> Repo.all()
  end

  @doc """
  Lists bookmarked agents.
  """
  def list_bookmarked_agents do
    Agent
    |> where([a], a.bookmarked == true)
    |> order_by([a], desc: a.updated_at)
    |> Repo.all()
  end

  @doc """
  Gets complete dashboard data for an agent.
  Returns agent, sessions, and active session.
  """
  def get_agent_dashboard_data(agent_id) do
    alias EyeInTheSkyWeb.Sessions

    agent = get_agent!(agent_id)
    sessions = Sessions.list_sessions_for_agent(agent_id)
    active_session = List.first(sessions) # Most recent session

    %{
      agent: agent,
      sessions: sessions,
      active_session: active_session
    }
  end
end
