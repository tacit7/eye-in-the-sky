defmodule EyeInTheSkyWeb.NATS.Consumer do
  @moduledoc """
  NATS consumer GenServer that subscribes to agent reply messages.
  """

  use GenServer
  require Logger

  alias EyeInTheSkyWeb.Messages

  def start_link(opts) do
    GenServer.start_link(__MODULE__, opts, name: __MODULE__)
  end

  @impl true
  def init(_opts) do
    # Connect to NATS
    {:ok, conn} = Gnat.start_link(%{
      host: "localhost",
      port: 4222
    })

    # Register connection globally
    Process.register(conn, :gnat)

    # Subscribe to all events.* topics with wildcard
    {:ok, _sub} = Gnat.sub(conn, self(), "events.>")

    Logger.info("NATS Consumer started, subscribed to events.>")

    {:ok, %{conn: conn}}
  end

  @impl true
  def handle_info({:msg, %{topic: topic, body: body}}, state) do
    Logger.debug("Received NATS message on #{topic}: #{body}")

    case Jason.decode(body) do
      {:ok, envelope} ->
        # Broadcast ALL messages to NATS viewer
        Phoenix.PubSub.broadcast(
          EyeInTheSkyWeb.PubSub,
          "nats:events",
          {:nats_message, topic, envelope}
        )

        # Also handle specific message types
        handle_envelope(envelope, topic)

      {:error, reason} ->
        Logger.error("Failed to decode NATS message: #{inspect(reason)}")

        # Still broadcast raw messages to viewer
        Phoenix.PubSub.broadcast(
          EyeInTheSkyWeb.PubSub,
          "nats:events",
          {:nats_message, topic, body}
        )
    end

    {:noreply, state}
  end

  defp handle_envelope(%{"op" => "msg", "channel" => "chat", "version" => "eits-messaging-v2"} = envelope, _topic) do
    # Handle v2 channel messages
    message_id = get_in(envelope, ["meta", "message_id"])

    if message_id && Messages.message_exists?(message_id) do
      Logger.debug("Skipping duplicate channel message #{message_id}")
      :ok
    else
      channel_id = envelope["channel_id"]
      parent_message_id = envelope["parent_message_id"]
      sender_session_id = get_in(envelope, ["meta", "sender_session_id"])
      provider = get_in(envelope, ["meta", "provider"]) || "unknown"
      message_body = envelope["msg"]

      # Create incoming channel message
      attrs = %{
        id: message_id || Ecto.UUID.generate(),
        channel_id: channel_id,
        parent_message_id: parent_message_id,
        session_id: sender_session_id,
        sender_role: "agent",
        recipient_role: "user",
        provider: provider,
        direction: "inbound",
        body: message_body,
        status: "delivered",
        metadata: %{}
      }

      case Messages.create_message(attrs) do
        {:ok, message} ->
          Logger.info("Recorded incoming channel message #{message.id} for channel #{channel_id}")

          # Broadcast to Phoenix PubSub for LiveView updates (channel-specific)
          Phoenix.PubSub.broadcast(
            EyeInTheSkyWeb.PubSub,
            "channel:#{channel_id}:messages",
            {:new_message, message}
          )

        {:error, reason} ->
          Logger.error("Failed to record incoming channel message: #{inspect(reason)}")
      end
    end
  end

  defp handle_envelope(%{"op" => "msg", "channel" => "chat"} = envelope, _topic) do
    # Handle v1 session-based messages (backward compatibility)
    message_id = get_in(envelope, ["meta", "message_id"])

    if message_id && Messages.message_exists?(message_id) do
      Logger.debug("Skipping duplicate message #{message_id}")
      :ok
    else
      # This is an agent reply message
      session_id = envelope["reply_to"]
      provider = get_in(envelope, ["meta", "provider"]) || "unknown"
      message_body = envelope["msg"]

      case Messages.record_incoming_reply(session_id, provider, message_body) do
        {:ok, message} ->
          Logger.info("Recorded incoming message #{message.id} for session #{session_id}")

          # Broadcast to Phoenix PubSub for LiveView updates
          Phoenix.PubSub.broadcast(
            EyeInTheSkyWeb.PubSub,
            "session:#{session_id}:messages",
            {:new_message, message}
          )

        {:error, reason} ->
          Logger.error("Failed to record incoming message: #{inspect(reason)}")
      end
    end
  end

  defp handle_envelope(%{"op" => "ack"} = envelope, _topic) do
    Logger.debug("Received ACK: #{inspect(envelope)}")
    # Handle acknowledgments if needed
  end

  defp handle_envelope(envelope, topic) do
    Logger.debug("Unhandled envelope on #{topic}: #{inspect(envelope)}")
  end

  @impl true
  def terminate(_reason, %{conn: conn}) do
    Gnat.stop(conn)
    :ok
  end
end
