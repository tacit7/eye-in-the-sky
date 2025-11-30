# Eye in the Sky — Messaging Protocol v1 (eits-messaging-v1)

This document defines conventions for communicating between agents via the i-nats-send / i-nats-listen tools.

## Goals
- Reliable first-try delivery on a single JetStream stream (EVENTS)
- Simple subjects under one namespace and a small, structured envelope
- Clear ack/confirm flow and retry guidance

## Stream & Subjects
- Stream: `EVENTS` (subjects: `events.*`)
- Subjects:
  - `events.chat` — conversational messages
  - `events.protocol` — control/handshakes (acks/confirms)
  - `events.task` — actionable requests or task routing
  - Fallback: leave `subject` empty to default to `events.test`

Notes:
- Always publish under `events.*`. If unsure, use the default by leaving `subject` empty.
- The MCP server should ensure stream creation and normalize subjects to the `events.*` pattern.

## Targeting
- `receiver_id` must be a session_id (not an agent_id)
- Use `receiver_id: ""` for broadcast announcements

## Envelope (JSON string in `message`)
```
{
  "op": "propose|ack|confirm|msg|ping|pong",
  "channel": "protocol|chat|task",
  "version": "eits-messaging-v1",
  "reply_to": "<session_id>",
  "msg": "free-form text",
  "meta": { "any": "additional context" }
}
```

## Acknowledgments & Reliability
- Ack: receiver replies within 30s with `{op: "ack", channel: "protocol"}`
- Confirm: sender may follow with `{op: "confirm", channel: "protocol"}` upon agreement
- Retry: up to 3 attempts with exponential backoff if no ack
- Ordering/idempotency: track `last_sequence` in `i-nats-listen` to avoid duplicate processing

## Usage Examples

### Send (targeted)
```
i-nats-send({
  "sender_id": "<agent_uuid>",
  "receiver_id": "<target_session_id>",
  "subject": "",  // defaults to events.test
  "message": "{\"op\":\"msg\",\"channel\":\"chat\",\"version\":\"eits-messaging-v1\",\"reply_to\":\"<your_session_id>\",\"msg\":\"hello\"}"
})
```

### Listen
```
i-nats-listen({
  "session_id": "<your_session_id>",
  "last_sequence": 0,
  "max_messages": 10
})
```

### Broadcast (announce)
```
i-nats-send({
  "sender_id": "<agent_uuid>",
  "receiver_id": "",
  "subject": "",  // falls back to events.test
  "message": "{\"op\":\"msg\",\"channel\":\"protocol\",\"version\":\"eits-messaging-v1\",\"msg\":\"Protocol v1 available\"}"
})
```

## Implementation Notes
- Server should:
  - Ensure `EVENTS` stream exists at startup
  - Normalize subjects to `events.*` pattern when possible
  - (Optional) Support `target_agent_id` → resolve active `session_id`
- Clients should:
  - Prefer targeted messages via `receiver_id=session_id`
  - Use `events.chat` for conversation and `events.protocol` for control

## Appendix — Common Payloads
- Ack: `{ "op": "ack", "channel": "protocol", "version": "eits-messaging-v1" }`
- Confirm: `{ "op": "confirm", "channel": "protocol", "version": "eits-messaging-v1" }`
- Chat: `{ "op": "msg", "channel": "chat", "version": "eits-messaging-v1", "msg": "..." }`
