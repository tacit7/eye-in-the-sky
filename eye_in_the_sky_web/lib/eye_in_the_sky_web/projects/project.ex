defmodule EyeInTheSkyWeb.Projects.Project do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :id, autogenerate: true}
  schema "projects" do
    field :name, :string
    field :path, :string
    field :remote_url, :string

    has_many :agents, EyeInTheSkyWeb.Agents.Agent
    has_many :commits, EyeInTheSkyWeb.Commits.Commit
    # Note: tasks.project_id is TEXT but projects.id is INTEGER
    # Manual loading required - see Projects.get_project_tasks/1
  end

  # Note: created_at and updated_at fields are stored by Go in a format that Ecto can't parse
  # They are omitted from the schema to avoid type casting errors

  @doc false
  def changeset(project, attrs) do
    project
    |> cast(attrs, [:name, :path, :remote_url])
    |> validate_required([:name])
    |> unique_constraint(:path)
  end
end
