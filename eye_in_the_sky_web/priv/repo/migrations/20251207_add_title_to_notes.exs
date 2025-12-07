defmodule EyeInTheSkyWeb.Repo.Migrations.AddTitleToNotes do
  use Ecto.Migration

  def change do
    alter table(:notes) do
      add :title, :string, default: ""
    end
  end
end
