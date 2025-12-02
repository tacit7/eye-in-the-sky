defmodule EyeInTheSkyWeb.NATS.Publisher do
  @moduledoc """
  NATS publisher for sending messages to agents via NATS JetStream.
  """

  require Logger

  @doc """
  Publishes a message to NATS for agent processing.

  Message format follows the eits-messaging-v1 protocol:
  - Subject: events.chat for user messages
  - Payload: JSON envelope with op, channel, version, msg fields
  """
  def publish_message(message, opts \\ []) do
    connection = Keyword.get(opts, :connection, get_connection())

    # Build envelope following eits-messaging-v1 protocol
    envelope = %{
      op: "msg",
      channel: "chat",
      version: "eits-messaging-v1",
      reply_to: message.session_id,
      msg: message.body,
      meta: %{
        message_id: message.id,
        provider: message.provider,
        timestamp: DateTime.to_iso8601(message.inserted_at)
      }
    }

    payload = Jason.encode!(envelope)

    case Gnat.pub(connection, "events.chat", payload) do
      :ok ->
        Logger.info("Published message #{message.id} to NATS events.chat")
        {:ok, message}

      {:error, reason} ->
        Logger.error("Failed to publish message #{message.id}: #{inspect(reason)}")
        {:error, reason}
    end
  end

  @doc """
  Publishes a message to a channel for multi-agent consumption.

  Message format follows the eits-messaging-v2 protocol:
  - Subject: events.channel.{channel_id}
  - Payload: JSON envelope with channel_id, parent_message_id (optional for threads)
  """
  def publish_channel_message(message, channel_id, opts \\ []) do
    connection = Keyword.get(opts, :connection, get_connection())

    # Build envelope following eits-messaging-v2 protocol
    envelope = %{
      op: "msg",
      channel: "chat",
      version: "eits-messaging-v2",
      channel_id: channel_id,
      parent_message_id: message.parent_message_id,
      msg: message.body,
      meta: %{
        message_id: message.id,
        sender_session_id: message.session_id,
        provider: message.provider,
        timestamp: DateTime.to_iso8601(message.inserted_at),
        attachments: format_attachments(message)
      }
    }

    payload = Jason.encode!(envelope)
    subject = "events.channel.#{channel_id}"

    case Gnat.pub(connection, subject, payload) do
      :ok ->
        Logger.info("Published channel message #{message.id} to #{subject}")
        {:ok, message}

      {:error, reason} ->
        Logger.error("Failed to publish channel message #{message.id}: #{inspect(reason)}")
        {:error, reason}
    end
  end

  @doc """
  Broadcasts a message to all agents (empty receiver_id).
  """
  def broadcast_message(body, opts \\ []) do
    connection = Keyword.get(opts, :connection, get_connection())

    envelope = %{
      op: "msg",
      channel: "protocol",
      version: "eits-messaging-v1",
      msg: body
    }

    payload = Jason.encode!(envelope)

    case Gnat.pub(connection, "events.protocol", payload) do
      :ok ->
        Logger.info("Broadcast message to NATS events.protocol")
        :ok

      {:error, reason} ->
        Logger.error("Failed to broadcast message: #{inspect(reason)}")
        {:error, reason}
    end
  end

  defp format_attachments(message) do
    if Ecto.assoc_loaded?(message.attachments) do
      Enum.map(message.attachments, fn att ->
        %{
          id: att.id,
          filename: att.original_filename,
          size: att.size_bytes,
          content_type: att.content_type
        }
      end)
    else
      []
    end
  end

  defp get_connection do
    case Process.whereis(:gnat) do
      nil ->
        Logger.error("NATS connection not found")
        nil
      pid -> pid
    end
  end
end
