defmodule EyeInTheSkyWeb.Contexts do
  @moduledoc """
  The Contexts context for managing session and agent contexts.
  """

  import Ecto.Query, warn: false
  alias EyeInTheSkyWeb.Repo
  alias EyeInTheSkyWeb.Contexts.{SessionContext, AgentContext}

  # SessionContext functions

  @doc """
  Gets session context for a specific session.
  """
  def get_session_context(session_id) do
    SessionContext
    |> where([sc], sc.session_id == ^session_id)
    |> order_by([sc], desc: sc.updated_at)
    |> limit(1)
    |> Repo.one()
  end

  @doc """
  Creates or updates session context.
  """
  def upsert_session_context(attrs) do
    case get_session_context(attrs.session_id) do
      nil ->
        %SessionContext{}
        |> SessionContext.changeset(attrs)
        |> Repo.insert()

      context ->
        context
        |> SessionContext.changeset(attrs)
        |> Repo.update()
    end
  end

  # AgentContext functions

  @doc """
  Gets agent context for a specific agent and project.
  """
  def get_agent_context(agent_id, project_id) do
    Repo.get_by(AgentContext, agent_id: agent_id, project_id: project_id)
  end

  @doc """
  Creates or updates agent context.
  """
  def upsert_agent_context(attrs) do
    case get_agent_context(attrs.agent_id, attrs.project_id) do
      nil ->
        %AgentContext{}
        |> AgentContext.changeset(attrs)
        |> Repo.insert()

      context ->
        context
        |> AgentContext.changeset(attrs)
        |> Repo.update()
    end
  end
end
