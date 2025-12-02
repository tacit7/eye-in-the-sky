defmodule EyeInTheSkyWeb.Sessions.Session do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}
  @foreign_key_type :string

  schema "sessions" do
    field :agent_id, :string
    field :name, :string
    field :started_at, :string
    field :ended_at, :string
    field :claude_session_id, :string

    belongs_to :agent, EyeInTheSkyWeb.Agents.Agent,
      define_field: false,
      foreign_key: :agent_id,
      type: :string

    has_many :logs, EyeInTheSkyWeb.Logs.Log, foreign_key: :session_id
    has_many :commits, EyeInTheSkyWeb.Commits.Commit, foreign_key: :session_id

    many_to_many :tasks, EyeInTheSkyWeb.Tasks.Task,
      join_through: "task_sessions",
      join_keys: [session_id: :id, task_id: :id]
  end

  @doc false
  def changeset(session, attrs) do
    session
    |> cast(attrs, [:id, :agent_id, :name, :started_at, :ended_at, :claude_session_id])
    |> validate_required([:id, :agent_id, :started_at])
  end
end
