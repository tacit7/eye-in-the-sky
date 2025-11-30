defmodule EyeInTheSkyWeb.Commits.Commit do
  use Ecto.Schema
  import Ecto.Changeset

  schema "commits" do
    field :commit_hash, :string
    field :commit_message, :string

    belongs_to :agent, EyeInTheSkyWeb.Agents.Agent, type: :string
    belongs_to :session, EyeInTheSkyWeb.Sessions.Session, type: :string
    belongs_to :project, EyeInTheSkyWeb.Projects.Project

    many_to_many :tasks, EyeInTheSkyWeb.Tasks.Task,
      join_through: "commit_tasks",
      join_keys: [commit_id: :id, task_id: :id]

    timestamps(inserted_at: :created_at, updated_at: false, type: :utc_datetime)
  end

  @doc false
  def changeset(commit, attrs) do
    commit
    |> cast(attrs, [:agent_id, :session_id, :project_id, :commit_hash, :commit_message])
    |> validate_required([:agent_id, :commit_hash])
  end
end
