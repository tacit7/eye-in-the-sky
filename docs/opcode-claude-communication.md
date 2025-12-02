# opcode - Claude Code Communication Architecture

**Date**: 2025-11-30
**Source**: Analysis of opcode codebase

---

## Overview

**opcode** is a desktop/web application that wraps Claude Code CLI by spawning it as a subprocess. There is **no API or SDK involved**—all communication happens through:

1. **Process spawning** - Launching the `claude` binary as a subprocess
2. **Command-line arguments** - Passing messages via `-p` flag
3. **Stdout parsing** - Reading JSON output line-by-line
4. **Filesystem access** - Reading/writing `~/.claude/` files

---

## Architecture

```
┌─────────────────────┐
│   opcode (Tauri)    │
│  ┌──────────────┐   │
│  │  React UI    │   │
│  └──────┬───────┘   │
│         │           │
│  ┌──────▼───────┐   │
│  │ Rust Backend │   │
│  └──────┬───────┘   │
└─────────┼───────────┘
          │ subprocess spawn
          │ with -p "message"
          ▼
┌─────────────────────┐
│  claude binary      │
│  (Node.js app)      │
│  - npm/homebrew     │
│  - PATH detection   │
└─────────┬───────────┘
          │
          ▼ writes to
┌─────────────────────┐
│  ~/.claude/         │
│  ├─ projects/       │
│  ├─ settings.json   │
│  └─ CLAUDE.md       │
└─────────────────────┘
```

---

## How Messages Are Sent to Claude Code

### Command-Line Arguments

Messages are sent to Claude Code via the **`-p` (prompt) flag** when spawning the subprocess:

```bash
claude -p "your message here" \
       --model sonnet \
       --output-format stream-json \
       --verbose \
       --dangerously-skip-permissions
```

### Three Message Types

#### 1. New Session
**File**: `src-tauri/src/commands/claude.rs:920-948`

```rust
let args = vec![
    "-p".to_string(),
    prompt.clone(),              // User's message
    "--model".to_string(),
    model.clone(),               // "sonnet" or "opus"
    "--output-format".to_string(),
    "stream-json".to_string(),
    "--verbose".to_string(),
    "--dangerously-skip-permissions".to_string(),
];

let cmd = create_system_command(&claude_path, args, &project_path);
spawn_claude_process(app, cmd, prompt, model, project_path).await
```

**Executed Command**:
```bash
claude -p "fix the authentication bug" \
       --model sonnet \
       --output-format stream-json \
       --verbose \
       --dangerously-skip-permissions
```

#### 2. Continue Session
**File**: `src-tauri/src/commands/claude.rs:952-980`

```rust
let args = vec![
    "-c".to_string(),            // Continue flag
    "-p".to_string(),
    prompt.clone(),              // Follow-up message
    "--model".to_string(),
    model.clone(),
    "--output-format".to_string(),
    "stream-json".to_string(),
    "--verbose".to_string(),
    "--dangerously-skip-permissions".to_string(),
];
```

**Executed Command**:
```bash
claude -c \
       -p "now add unit tests" \
       --model sonnet \
       --output-format stream-json \
       --verbose \
       --dangerously-skip-permissions
```

**How it works**: Claude Code reads the last session file in `~/.claude/projects/<encoded-path>/` and appends the new prompt.

#### 3. Resume Specific Session
**File**: `src-tauri/src/commands/claude.rs:984-1015`

```rust
let args = vec![
    "--resume".to_string(),
    session_id.clone(),          // Session UUID
    "-p".to_string(),
    prompt.clone(),
    "--model".to_string(),
    model.clone(),
    "--output-format".to_string(),
    "stream-json".to_string(),
    "--verbose".to_string(),
    "--dangerously-skip-permissions".to_string(),
];
```

**Executed Command**:
```bash
claude --resume abc123-def456-session-uuid \
       -p "revert the last change" \
       --model sonnet \
       --output-format stream-json \
       --verbose \
       --dangerously-skip-permissions
```

**How it works**: Claude Code loads the specific session file `~/.claude/projects/<project>/<session-id>.jsonl`.

---

## Critical Flags

| Flag | Purpose |
|------|---------|
| `-p "message"` | The actual user prompt/message |
| `--output-format stream-json` | Makes Claude output JSON lines instead of human-readable format |
| `--dangerously-skip-permissions` | Bypasses interactive permission prompts (required for automation) |
| `--verbose` | Provides more detailed output for debugging |
| `-c` | Continue last session in current directory |
| `--resume <uuid>` | Resume specific session by ID |
| `--model <name>` | Specify model (sonnet, opus, haiku) |

---

## Message Flow Diagram

```
┌─────────────────────┐
│   User types in UI  │
│   "fix the bug"     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────────────────────┐
│  Frontend (React)                   │
│  executeClaudeCode(project, prompt) │
└──────────┬──────────────────────────┘
           │ Tauri IPC / WebSocket
           ▼
┌─────────────────────────────────────┐
│  Rust Backend                       │
│  execute_claude_code()              │
│  - Finds claude binary              │
│  - Builds args: ["-p", prompt, ...] │
│  - Spawns subprocess                │
└──────────┬──────────────────────────┘
           │ subprocess spawn
           ▼
┌─────────────────────────────────────┐
│  Claude Code Binary (Node.js)       │
│  $ claude -p "fix the bug" \        │
│           --model sonnet \           │
│           --output-format stream-json│
└──────────┬──────────────────────────┘
           │
           ▼
┌─────────────────────────────────────┐
│  Claude reads:                      │
│  - Project files (via cwd)          │
│  - CLAUDE.md system prompt          │
│  - Session history (if -c/--resume) │
│  - User message from -p flag        │
└──────────┬──────────────────────────┘
           │
           ▼
┌─────────────────────────────────────┐
│  Claude processes, writes response  │
│  to stdout as JSON lines            │
└──────────┬──────────────────────────┘
           │ stdout pipe
           ▼
┌─────────────────────────────────────┐
│  Rust reads stdout line-by-line     │
│  Emits: claude-output:{sessionId}   │
└──────────┬──────────────────────────┘
           │
           ▼
┌─────────────────────────────────────┐
│  Frontend receives JSON events      │
│  Renders to UI                      │
└─────────────────────────────────────┘
```

---

## Output Streaming

### Desktop Mode (Tauri)
**File**: `src-tauri/src/commands/claude.rs:1174-1340`

```rust
// Spawn the process
let mut child = cmd.spawn()?;
let stdout = child.stdout.take()?;
let stderr = child.stderr.take()?;

// Read stdout line by line
let stdout_reader = BufReader::new(stdout);
let mut lines = stdout_reader.lines();

while let Ok(Some(line)) = lines.next_line().await {
    // Parse session ID from init message
    if let Ok(msg) = serde_json::from_str::<Value>(&line) {
        if msg["type"] == "system" && msg["subtype"] == "init" {
            let session_id = msg["session_id"].as_str();
            // Register in ProcessRegistry
        }
    }

    // Store live output in registry
    registry.append_live_output(run_id, &line);

    // Emit to frontend (session-specific)
    app.emit(&format!("claude-output:{}", session_id), &line);

    // Emit to frontend (generic for backward compatibility)
    app.emit("claude-output", &line);
}
```

**Events Emitted**:
- `claude-output:{sessionId}` - Session-specific output
- `claude-output` - Generic output (backward compatibility)
- `claude-error:{sessionId}` - Session-specific errors
- `claude-complete:{sessionId}` - Process completion
- `claude-cancelled:{sessionId}` - User cancellation

### Web Mode (Axum WebSocket)
**File**: `src-tauri/src/web_server.rs:445-568`

```rust
// Spawn Claude process
let mut child = cmd.spawn()?;
let stdout = child.stdout.take()?;
let stdout_reader = BufReader::new(stdout);

// Stream output line by line
let mut lines = stdout_reader.lines();
while let Ok(Some(line)) = lines.next_line().await {
    // Send each line to WebSocket
    let message = json!({
        "type": "output",
        "content": line
    }).to_string();

    send_to_session(&state, &session_id, message).await;
}
```

**WebSocket Message Format**:
```json
// Output
{
  "type": "output",
  "content": "{\"type\":\"assistant\",\"message\":\"...\"}"
}

// Completion
{
  "type": "completion",
  "status": "success"
}

// Error
{
  "type": "error",
  "message": "Failed to spawn Claude"
}
```

---

## Claude Binary Detection

**File**: `src-tauri/src/claude_binary.rs`

opcode detects the `claude` binary through multiple strategies (in order of priority):

1. **Database-stored path**: Checks SQLite for user's saved preference
2. **`which` command**: `which claude` (Unix) or `where claude` (Windows)
3. **NVM installations**: Scans `~/.nvm/versions/node/*/bin/claude`
4. **Standard paths**:
   - `/usr/local/bin/claude`
   - `/opt/homebrew/bin/claude`
   - `~/.local/bin/claude`
5. **Version detection**: Runs `claude --version` to verify

**Priority System**:
- User-stored preference (highest)
- Latest version (semantic versioning)
- Source preference: `which` > homebrew > system > nvm

---

## Session Data Storage

Claude Code stores sessions in `~/.claude/projects/`:

```
~/.claude/
├── projects/
│   └── -Users-foo-myproject/    # Encoded project path
│       ├── session-id-1.jsonl   # Session 1 history
│       └── session-id-2.jsonl   # Session 2 history
├── settings.json                 # Claude settings
├── CLAUDE.md                     # Global system prompt
└── todos/                        # Todo data per session
```

### JSONL Format

Each line in a session file is a JSON object:

```jsonl
{"type":"system","subtype":"init","session_id":"abc-123","cwd":"/Users/foo/project"}
{"type":"message","role":"user","content":"fix the bug","timestamp":"2025-11-30T10:00:00Z"}
{"type":"message","role":"assistant","content":"I'll help fix that bug...","timestamp":"2025-11-30T10:00:05Z"}
{"type":"tool_use","tool":"Edit","parameters":{"file_path":"/path/to/file.ts",...}}
{"type":"tool_result","tool":"Edit","result":"Success"}
```

**Important**: opcode **only reads** these files—it never writes to them. Only Claude Code writes session data.

---

## Process Management

### ProcessRegistry
**File**: `src-tauri/src/process/registry.rs`

Tracks all running Claude processes:

```rust
pub struct ProcessInfo {
    pub run_id: i64,              // Internal opcode ID
    pub session_id: String,       // Claude session UUID
    pub pid: u32,                 // OS process ID
    pub project_path: String,
    pub prompt: String,
    pub model: String,
    pub started_at: String,
}
```

**Features**:
- Maps session IDs to PIDs
- Captures live output
- Enables process termination
- Cleans up finished processes

### ClaudeProcessState
**File**: `src-tauri/src/commands/claude.rs:14-24`

Global state for current Claude process:

```rust
pub struct ClaudeProcessState {
    pub current_process: Arc<Mutex<Option<Child>>>,
}
```

Used for backward compatibility and quick cancellation.

---

## Process Cancellation

**File**: `src-tauri/src/commands/claude.rs:1018-1149`

Three methods attempted (in order):

1. **ProcessRegistry lookup**: Find by session ID, kill by run_id
2. **ClaudeProcessState**: Kill current process handle
3. **System kill**: Fallback using `kill -KILL <pid>` (Unix) or `taskkill /F /PID <pid>` (Windows)

```rust
// Method 1: ProcessRegistry
if let Some(process_info) = registry.get_claude_session_by_id(&session_id) {
    registry.kill_process(process_info.run_id).await;
}

// Method 2: ClaudeProcessState
if let Some(mut child) = current_process.take() {
    child.kill().await;
}

// Method 3: System kill
std::process::Command::new("kill")
    .args(["-KILL", &pid.to_string()])
    .output()
```

---

## Web vs Desktop Mode

### Desktop (Tauri)
- ✅ Full-featured, direct process access
- ✅ Session-scoped events
- ✅ Process cancellation works
- ✅ stderr handling
- ✅ Process registry integration

### Web (Axum Server)
- ✅ Basic functionality
- ❌ **Session-scoped events not fully implemented** (only generic events work)
- ❌ **Cancel button doesn't kill processes**
- ❌ **stderr handling broken** (errors don't show up)
- ❌ **Session ID mapping issues** (WebSocket handler generates own IDs)

**Source**: `web_server.design.md` (design doc)

---

## Environment Setup

### Command Creation
**File**: `src-tauri/src/commands/claude.rs:232-290`

```rust
fn create_command_with_env(program: &str) -> Command {
    let mut tokio_cmd = Command::new(program);

    // Copy critical environment variables
    for (key, value) in std::env::vars() {
        if key == "PATH" || key == "HOME" || key == "USER"
           || key == "NODE_PATH" || key == "NVM_DIR"
           || key == "NVM_BIN" || key.starts_with("HOMEBREW_") {
            tokio_cmd.env(&key, &value);
        }
    }

    // Add NVM support
    if program.contains("/.nvm/versions/node/") {
        let node_bin_dir = Path::new(program).parent();
        let new_path = format!("{}:{}", node_bin_dir, current_path);
        tokio_cmd.env("PATH", new_path);
    }

    // Add Homebrew support
    if program.contains("/homebrew/") || program.contains("/opt/homebrew/") {
        let homebrew_bin = Path::new(program).parent();
        let new_path = format!("{}:{}", homebrew_bin, current_path);
        tokio_cmd.env("PATH", new_path);
    }

    tokio_cmd
}
```

Sets working directory:
```rust
cmd.current_dir(&project_path)
```

---

## Key Insights

### No Interactive stdin
opcode does **NOT** use interactive stdin. Each message:
1. Passed as complete string via `-p` flag
2. Spawned as new subprocess
3. Process terminates when Claude finishes

### Multi-turn Conversations
For conversations with history:
- **Continue mode** (`-c`): Claude reads last session in current directory
- **Resume mode** (`--resume`): Claude loads specific session by UUID

Both modes read from `~/.claude/projects/` to get conversation history.

### No Persistent Connection
There is no persistent connection to Claude. Each user message spawns a fresh process:

```
User message 1 → Spawn claude → Read history → Process → Write → Exit
User message 2 → Spawn claude → Read history → Process → Write → Exit
User message 3 → Spawn claude → Read history → Process → Write → Exit
```

### Filesystem is Source of Truth
The session JSONL files in `~/.claude/projects/` are the single source of truth. opcode reads them to:
- Display session history
- Extract first user message
- Get project path from `cwd` field
- Track usage/tokens
- Create checkpoints

---

## Code References

| Component | File | Lines |
|-----------|------|-------|
| New session execution | `src-tauri/src/commands/claude.rs` | 920-948 |
| Continue session | `src-tauri/src/commands/claude.rs` | 952-980 |
| Resume session | `src-tauri/src/commands/claude.rs` | 984-1015 |
| Process spawning | `src-tauri/src/commands/claude.rs` | 1174-1340 |
| Process cancellation | `src-tauri/src/commands/claude.rs` | 1018-1149 |
| Binary detection | `src-tauri/src/claude_binary.rs` | 173-484 |
| Web mode execution | `src-tauri/src/web_server.rs` | 445-699 |
| Environment setup | `src-tauri/src/commands/claude.rs` | 232-306 |

---

## Summary

opcode communicates with Claude Code by:

1. **Detecting** the `claude` binary from PATH, NVM, or Homebrew installations
2. **Spawning** it as a subprocess with proper environment variables
3. **Passing** user messages via `-p "message"` command-line flag
4. **Streaming** stdout/stderr through Tauri events or WebSocket
5. **Reading** session files from `~/.claude/projects/` for history
6. **Managing** processes via ProcessRegistry for cancellation

**No API, no SDK, no persistent connection**—just subprocess spawning with CLI arguments and filesystem access.
