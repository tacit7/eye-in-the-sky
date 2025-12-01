defmodule EyeInTheSkyWeb.Repo.Migrations.AddClaudeSessionIdToSessions do
  use Ecto.Migration

  def change do
    alter table(:sessions) do
      add :claude_session_id, :string
    end

    create index(:sessions, [:claude_session_id])
  end
end
