defmodule EyeInTheSkyWeb.Agents.Agent do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}
  @foreign_key_type :string

  schema "agents" do
    field :persona_id, :string
    field :status, :string
    field :source, :string
    field :git_worktree_path, :string
    field :feature_description, :string
    field :current_task, :string
    field :window_id, :string
    field :project_name, :string
    field :session_id, :string
    field :description, :string
    field :terminal_application, :string
    field :parent_agent_id, :string
    field :parent_session_id, :string
    field :bookmarked, :boolean, default: false
    field :last_activity_at, :naive_datetime
    field :completed_at, :naive_datetime

    belongs_to :project, EyeInTheSkyWeb.Projects.Project, type: :integer

    has_many :sessions, EyeInTheSkyWeb.Sessions.Session, foreign_key: :agent_id
    has_many :commits, EyeInTheSkyWeb.Commits.Commit, foreign_key: :agent_id
    has_many :tasks, EyeInTheSkyWeb.Tasks.Task, foreign_key: :agent_id

    field :created_at, :naive_datetime
    field :updated_at, :naive_datetime
  end

  @doc false
  def changeset(agent, attrs) do
    agent
    |> cast(attrs, [
      :id,
      :persona_id,
      :project_id,
      :status,
      :source,
      :git_worktree_path,
      :feature_description,
      :current_task,
      :window_id,
      :project_name,
      :session_id,
      :description,
      :terminal_application,
      :parent_agent_id,
      :parent_session_id,
      :bookmarked,
      :last_activity_at,
      :completed_at
    ])
    |> validate_required([:id, :status])
    |> validate_inclusion(:status, ["active", "idle", "working", "completed", "failed"])
  end
end
