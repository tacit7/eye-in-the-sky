defmodule EyeInTheSkyWeb.Messages.Message do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :string, autogenerate: false}
  @foreign_key_type :string

  schema "messages" do
    field :sender_role, :string
    field :recipient_role, :string
    field :provider, :string
    field :provider_session_id, :string
    field :direction, :string
    field :body, :string
    field :status, :string, default: "sent"
    field :metadata, :map

    belongs_to :project, EyeInTheSkyWeb.Projects.Project, type: :integer
    belongs_to :session, EyeInTheSkyWeb.Sessions.Session,
      define_field: false,
      foreign_key: :session_id,
      type: :string

    field :session_id, :string
    field :inserted_at, :utc_datetime
    field :updated_at, :utc_datetime
  end

  @doc false
  def changeset(message, attrs) do
    message
    |> cast(attrs, [
      :id,
      :project_id,
      :session_id,
      :sender_role,
      :recipient_role,
      :provider,
      :provider_session_id,
      :direction,
      :body,
      :status,
      :metadata,
      :inserted_at,
      :updated_at
    ])
    |> validate_required([:id, :sender_role, :direction, :body])
    |> validate_inclusion(:direction, ["inbound", "outbound"])
    |> validate_inclusion(:status, ["sent", "delivered", "failed", "pending"])
  end
end
