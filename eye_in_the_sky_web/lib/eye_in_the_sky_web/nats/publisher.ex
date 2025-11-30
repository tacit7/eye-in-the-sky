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

  defp get_connection do
    case Process.whereis(:gnat) do
      nil ->
        Logger.error("NATS connection not found")
        nil
      pid -> pid
    end
  end
end
