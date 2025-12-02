# How Local Agents Receive Messages (Action Plan)

This document explains how **local agents running on your machine**
(such as the Claude Code CLI) actually receive messages from your
messaging interface. These agents are not servers and do not subscribe
to NATS; they respond only when your backend **executes their binary**
with a prompt.

## 1. High-Level Principle

Local agents "receive" messages when your backend:

1.  **Launches the agent process** (e.g., `claude`)
2.  **Passes the user message as a prompt**
3.  **Captures the JSON output**
4.  **Stores the reply**
5.  **Updates the UI**

There is no persistent connection. No socket. No push mechanism.\
Agents only react when executed.

## 2. Messaging Flow

### Step 1: User sends a message in the UI

The Svelte component calls:

    live.pushEvent("send_message", { body, agentId })

### Step 2: Phoenix stores the message

LiveView:

-   Inserts an outgoing message into the `messages` table\
-   Enqueues a job via NATS or a worker queue

### Step 3: Worker picks up the job

The worker process is local and subscribed to your job queue.

Payload:

    { "message_id": "uuid" }

### Step 4: Worker executes the agent CLI

Based on the Claude Code cheat sheet:

-   First call:

```{=html}
<!-- -->
```
    claude -p "User message" --output-format json

-   Continuing an agent session:

```{=html}
<!-- -->
```
    claude -r "<session_id>" -p "Next message" --output-format json

Parse the JSON output:

    { "text": "Agent reply", "session_id": "abc123" }

Store the returned `session_id` in your message metadata.

### Step 5: Worker writes the agent reply to the DB

A new incoming message is inserted with:

-   `direction: "incoming"`
-   `body: reply text`

### Step 6: LiveView updates the interface

-   LiveView receives a PubSub event\
-   Updates `@messages`\
-   Svelte receives new props and rerenders

## 3. Why NATS Is Not Used for Agent Delivery

NATS is used **only** as a job queue:

-   Phoenix → publish job
-   Worker → consumes job and runs agent process

Agents themselves do **not** subscribe to NATS; they do not process
messages directly.

## 4. Components Involved

### Frontend

-   Svelte chat UI
-   LiveSvelte event bridge
-   DaisyUI components

### Phoenix backend

-   LiveView state
-   `messages` table
-   PubSub to push updates
-   Job enqueue (`jobs.agent_request`)

### Worker process

-   Listens for job requests
-   Executes agent binaries
-   Parses output
-   Saves replies

### Local agents (Claude Code CLI)

-   Run as command line processes
-   Receive prompts via execution
-   Maintain their own session ID
-   Output JSON

## 5. Example Worker (Elixir)

``` elixir
defmodule MyApp.AgentWorker do
  def handle_job(%{"message_id" => id}) do
    msg = Messaging.get_message!(id)

    sid = msg.metadata["claude_session_id"]

    args =
      if sid,
        do: ["-r", sid, "-p", msg.body, "--output-format", "json"],
        else: ["-p", msg.body, "--output-format", "json"]

    {output, 0} = System.cmd("claude", args, stderr_to_stdout: true)

    {:ok, %{"text" => reply, "session_id" => new_sid}} = Jason.decode(output)

    Messaging.record_incoming_reply(msg, reply, %{"claude_session_id" => new_sid})
  end
end
```

## 6. Summary

Local agents receive messages when:

1.  The user sends a message\
2.  Phoenix stores it and queues a job\
3.  A worker executes the agent binary (e.g., `claude`)\
4.  The agent returns JSON output\
5.  The worker stores the reply\
6.  LiveView updates the UI

This is the actual and correct delivery mechanism for local machine
agents.
