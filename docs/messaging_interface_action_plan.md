# Messaging Interface Action Plan

## Phase 0 · Scope and Assumptions

You are building a messaging interface for communication between user
and agents. Using Phoenix, LiveView, LiveSvelte, DaisyUI, and NATS.

## Phase 1 · Data Model and Migration

### Create messages table

``` elixir
create table(:messages, primary_key: false) do
  add :id, :binary_id, primary_key: true
  add :project_id, references(:projects, type: :binary_id)
  add :session_id, references(:sessions, type: :binary_id)
  add :sender_role, :string
  add :recipient_role, :string
  add :provider, :string
  add :provider_session_id, :string
  add :direction, :string
  add :body, :text
  add :status, :string, default: "sent"
  add :metadata, :map, default: %{}
  timestamps()
end
```

## Phase 2 · Messaging Context

Implement:

-   `list_messages/1`
-   `send_message/1`
-   `record_incoming_reply/3`

## Phase 3 · NATS Architecture

Use NATS for asynchronous agent requests.

### Topics

-   `agents.request`
-   `agents.reply`

### Worker Responsibilities

-   Subscribe to `agents.request`
-   Fetch message
-   Call Claude/OpenAI
-   Insert incoming message
-   Publish `agents.reply`

## Phase 4 · LiveView + LiveSvelte

`AgentInboxLive` loads initial messages and handles UI updates.

### Handle send_message event

-   Insert outgoing message
-   Publish NATS job
-   Update `@messages`

### Subscribe to updates

Use PubSub or `agents.reply` to update LiveView in real time.

## Phase 5 · Svelte UI

Implement DaisyUI chat UI.

Props:

-   `messages`
-   `projectId`
-   `sessionId`
-   `live`

Use:

``` ts
live.pushEvent("send_message", { body, provider })
```

## Phase 6 · Provider Integration Layer

Implement provider clients:

-   `ClaudeClient.chat/1`
-   `OpenAIClient.chat/1`

Worker inserts replies with `record_incoming_reply/3`.

## Phase 7 · Polish

-   Pagination
-   Per-project session navigator
-   Observability
-   Tests
-   Permissions
