defmodule EyeInTheSkyWeb.Logs.SessionLog do
  use Ecto.Schema
  import Ecto.Changeset

  schema "session_logs" do
    field :log_level, :string
    field :category, :string
    field :message, :string
    field :details, :string
    field :timestamp, :string

    belongs_to :session, EyeInTheSkyWeb.Sessions.Session, type: :integer
  end

  @doc false
  def changeset(session_log, attrs) do
    session_log
    |> cast(attrs, [:session_id, :log_level, :category, :message, :details, :timestamp])
    |> validate_required([:session_id, :log_level, :category, :message, :timestamp])
  end
end
