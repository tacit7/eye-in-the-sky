defmodule EyeInTheSkyWeb.Messages do
  @moduledoc """
  The Messages context for managing agent-user messaging.
  """

  import Ecto.Query, warn: false
  alias EyeInTheSkyWeb.Repo
  alias EyeInTheSkyWeb.Messages.Message

  @doc """
  Returns the list of messages.
  """
  def list_messages do
    Repo.all(Message)
  end

  @doc """
  Returns the list of messages for a specific session.
  """
  def list_messages_for_session(session_id) do
    Message
    |> where([m], m.session_id == ^session_id)
    |> order_by([m], asc: m.inserted_at)
    |> Repo.all()
  end

  @doc """
  Returns the list of messages for a specific project.
  """
  def list_messages_for_project(project_id) do
    Message
    |> where([m], m.project_id == ^project_id)
    |> order_by([m], asc: m.inserted_at)
    |> Repo.all()
  end

  @doc """
  Gets a single message.

  Raises `Ecto.NoResultsError` if the Message does not exist.
  """
  def get_message!(id) do
    Repo.get!(Message, id)
  end

  @doc """
  Checks if a message with the given ID exists.
  """
  def message_exists?(id) do
    Message
    |> where([m], m.id == ^id)
    |> Repo.exists?()
  end

  @doc """
  Creates a message.
  """
  def create_message(attrs \\ %{}) do
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    attrs = Map.merge(attrs, %{
      inserted_at: now,
      updated_at: now
    })

    %Message{}
    |> Message.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Sends a message (creates an outbound message).
  """
  def send_message(attrs) do
    attrs
    |> Map.put(:id, Ecto.UUID.generate())
    |> Map.put(:direction, "outbound")
    |> Map.put(:status, "pending")
    |> create_message()
  end

  @doc """
  Records an incoming reply (creates an inbound message).
  """
  def record_incoming_reply(session_id, provider, body) do
    create_message(%{
      id: Ecto.UUID.generate(),
      session_id: session_id,
      sender_role: "agent",
      recipient_role: "user",
      provider: provider,
      direction: "inbound",
      body: body,
      status: "delivered",
      metadata: %{}
    })
  end

  @doc """
  Updates a message.
  """
  def update_message(%Message{} = message, attrs) do
    message
    |> Message.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Updates message status.
  """
  def update_message_status(%Message{} = message, status) do
    update_message(message, %{status: status})
  end

  @doc """
  Deletes a message.
  """
  def delete_message(%Message{} = message) do
    Repo.delete(message)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking message changes.
  """
  def change_message(%Message{} = message, attrs \\ %{}) do
    Message.changeset(message, attrs)
  end

  @doc """
  Returns recent messages for a session (default last 50).
  """
  def list_recent_messages(session_id, limit \\ 50) do
    Message
    |> where([m], m.session_id == ^session_id)
    |> order_by([m], desc: m.inserted_at)
    |> limit(^limit)
    |> Repo.all()
    |> Enum.reverse()
  end

  @doc """
  Returns conversation thread for a session with pagination.
  """
  def get_conversation_thread(session_id, opts \\ []) do
    offset = Keyword.get(opts, :offset, 0)
    limit = Keyword.get(opts, :limit, 50)

    Message
    |> where([m], m.session_id == ^session_id)
    |> order_by([m], asc: m.inserted_at)
    |> offset(^offset)
    |> limit(^limit)
    |> Repo.all()
  end

  @doc """
  Counts messages for a session.
  """
  def count_messages_for_session(session_id) do
    Message
    |> where([m], m.session_id == ^session_id)
    |> Repo.aggregate(:count)
  end

  @doc """
  Returns unread/pending messages for a session.
  """
  def list_pending_messages(session_id) do
    Message
    |> where([m], m.session_id == ^session_id and m.status == "pending")
    |> order_by([m], asc: m.inserted_at)
    |> Repo.all()
  end

  # Channel-based messaging

  @doc """
  Returns the list of messages for a specific channel.
  """
  def list_messages_for_channel(channel_id, opts \\ []) do
    offset = Keyword.get(opts, :offset, 0)
    limit = Keyword.get(opts, :limit, 100)

    Message
    |> where([m], m.channel_id == ^channel_id and is_nil(m.parent_message_id))
    |> order_by([m], asc: m.inserted_at)
    |> offset(^offset)
    |> limit(^limit)
    |> preload([:reactions, :attachments, :session])
    |> Repo.all()
  end

  @doc """
  Creates a channel message.
  """
  def create_channel_message(attrs) do
    attrs
    |> Map.put(:id, Ecto.UUID.generate())
    |> create_message()
  end

  @doc """
  Sends a message to a channel (creates an outbound message).
  """
  def send_channel_message(attrs) do
    attrs
    |> Map.put(:id, Ecto.UUID.generate())
    |> Map.put(:direction, "outbound")
    |> Map.put(:status, "pending")
    |> create_message()
  end

  # Threading support

  @doc """
  Returns thread replies for a parent message.
  """
  def list_thread_replies(parent_message_id) do
    Message
    |> where([m], m.parent_message_id == ^parent_message_id)
    |> order_by([m], asc: m.inserted_at)
    |> preload([:reactions, :attachments])
    |> Repo.all()
  end

  @doc """
  Creates a thread reply.
  """
  def create_thread_reply(parent_message_id, attrs) do
    attrs = Map.put(attrs, :parent_message_id, parent_message_id)

    with {:ok, message} <- create_channel_message(attrs) do
      increment_thread_count(parent_message_id)
      {:ok, message}
    end
  end

  @doc """
  Increments the thread reply count for a parent message.
  """
  def increment_thread_count(parent_message_id) do
    now = DateTime.utc_now() |> DateTime.truncate(:second)

    from(m in Message,
      where: m.id == ^parent_message_id
    )
    |> Repo.update_all(
      inc: [thread_reply_count: 1],
      set: [last_thread_reply_at: now]
    )
  end

  @doc """
  Gets a message with its thread replies loaded.
  """
  def get_message_with_thread!(id) do
    Message
    |> where([m], m.id == ^id)
    |> preload([:thread_replies, :reactions, :attachments])
    |> Repo.one!()
  end

  # Reactions support

  @doc """
  Adds a reaction to a message.
  """
  def add_reaction(message_id, session_id, emoji) do
    alias EyeInTheSkyWeb.Messages.MessageReaction

    now = DateTime.utc_now() |> DateTime.truncate(:second)

    attrs = %{
      message_id: message_id,
      session_id: session_id,
      emoji: emoji,
      inserted_at: now
    }

    %MessageReaction{}
    |> MessageReaction.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Removes a reaction from a message.
  """
  def remove_reaction(message_id, session_id, emoji) do
    alias EyeInTheSkyWeb.Messages.MessageReaction

    from(r in MessageReaction,
      where: r.message_id == ^message_id and r.session_id == ^session_id and r.emoji == ^emoji
    )
    |> Repo.delete_all()
  end

  @doc """
  Lists all reactions for a message, grouped by emoji.
  """
  def list_reactions_for_message(message_id) do
    alias EyeInTheSkyWeb.Messages.MessageReaction

    from(r in MessageReaction,
      where: r.message_id == ^message_id,
      order_by: [asc: r.inserted_at]
    )
    |> Repo.all()
    |> Enum.group_by(& &1.emoji)
    |> Enum.map(fn {emoji, reactions} ->
      %{
        emoji: emoji,
        count: length(reactions),
        session_ids: Enum.map(reactions, & &1.session_id)
      }
    end)
  end

  @doc """
  Toggles a reaction (adds if not present, removes if present).
  """
  def toggle_reaction(message_id, session_id, emoji) do
    alias EyeInTheSkyWeb.Messages.MessageReaction

    existing = from(r in MessageReaction,
      where: r.message_id == ^message_id and r.session_id == ^session_id and r.emoji == ^emoji
    )
    |> Repo.one()

    if existing do
      remove_reaction(message_id, session_id, emoji)
      {:ok, :removed}
    else
      case add_reaction(message_id, session_id, emoji) do
        {:ok, _reaction} -> {:ok, :added}
        error -> error
      end
    end
  end
end
