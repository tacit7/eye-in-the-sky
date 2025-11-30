defmodule EyeInTheSkyWeb.Projects.Project do
  use Ecto.Schema
  import Ecto.Changeset

  schema "projects" do
    field :name, :string
    field :path, :string
    field :remote_url, :string

    has_many :agents, EyeInTheSkyWeb.Agents.Agent
    has_many :commits, EyeInTheSkyWeb.Commits.Commit
    has_many :tasks, EyeInTheSkyWeb.Tasks.Task

    field :created_at, :utc_datetime
    field :updated_at, :utc_datetime
  end

  @doc false
  def changeset(project, attrs) do
    project
    |> cast(attrs, [:name, :path, :remote_url])
    |> validate_required([:name])
    |> unique_constraint(:path)
  end
end
