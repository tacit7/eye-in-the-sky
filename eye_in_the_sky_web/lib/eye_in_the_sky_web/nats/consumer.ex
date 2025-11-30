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

    # Subscribe to agent reply topics
    {:ok, _sub} = Gnat.sub(conn, self(), "events.chat")
    {:ok, _sub} = Gnat.sub(conn, self(), "events.protocol")

    Logger.info("NATS Consumer started, subscribed to events.chat and events.protocol")

    {:ok, %{conn: conn}}
  end

  @impl true
  def handle_info({:msg, %{topic: topic, body: body}}, state) do
    Logger.debug("Received NATS message on #{topic}: #{body}")

    case Jason.decode(body) do
      {:ok, envelope} ->
        handle_envelope(envelope, topic)

      {:error, reason} ->
        Logger.error("Failed to decode NATS message: #{inspect(reason)}")
    end

    {:noreply, state}
  end

  defp handle_envelope(%{"op" => "msg", "channel" => "chat"} = envelope, _topic) do
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
