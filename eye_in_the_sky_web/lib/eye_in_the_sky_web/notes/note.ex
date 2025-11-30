defmodule EyeInTheSkyWeb.Notes.Note do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :id, autogenerate: true}
  schema "notes" do
    field :parent_type, :string
    field :parent_id, :string
    field :body, :string
    field :created_at, :string
    field :inserted_at, :string
    field :updated_at, :string
  end

  @doc false
  def changeset(note, attrs) do
    note
    |> cast(attrs, [:parent_type, :parent_id, :body])
    |> validate_required([:parent_type, :parent_id, :body])
    |> validate_inclusion(:parent_type, ["session", "task", "agent"])
  end
end
