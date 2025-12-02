-- Eye in the Sky Database Schema

CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    git_worktree_path TEXT,
    feature_description TEXT,
    current_task TEXT,
    last_activity_at TIMESTAMP,
    source TEXT NOT NULL DEFAULT 'worktree',
    window_id TEXT,
    project_name TEXT,
    project_id TEXT REFERENCES projects(id),
    session_id TEXT,
    persona_id TEXT REFERENCES personas(id),
    description TEXT,
    terminal_application TEXT,
    parent_agent_id TEXT DEFAULT NULL,
    parent_session_id TEXT,
    bookmarked INTEGER NOT NULL DEFAULT 0,
    completed_at DATETIME
);

CREATE TABLE IF NOT EXISTS actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    action_type TEXT NOT NULL,
    description TEXT NOT NULL,
    details TEXT,
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE TABLE IF NOT EXISTS commits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    commit_hash TEXT NOT NULL,
    commit_message TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    project_id TEXT REFERENCES projects(id),
    session_id TEXT REFERENCES sessions(id),
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

-- Simplified session_context table using markdown format
CREATE TABLE IF NOT EXISTS session_context (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    context TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE TABLE IF NOT EXISTS session_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    note_type TEXT NOT NULL,
    content TEXT NOT NULL,
    priority TEXT NOT NULL,
    tags TEXT,
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE TABLE IF NOT EXISTS session_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    log_level TEXT NOT NULL,
    category TEXT NOT NULL,
    message TEXT NOT NULL,
    details TEXT,
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE TABLE IF NOT EXISTS personas (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    expertise TEXT NOT NULL,
    initial_context TEXT NOT NULL,
    preferred_tools TEXT,
    specialization TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    name TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE TABLE IF NOT EXISTS notes (
    id TEXT PRIMARY KEY,
    parent_id TEXT NOT NULL,
    parent_type TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS context (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(session_id, key),
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE TABLE IF NOT EXISTS compactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    old_session_id TEXT,
    new_session_id TEXT NOT NULL,
    compacted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    summary TEXT,
    jsonl_file_path TEXT,
    jsonl_file_size INTEGER,
    message_count INTEGER,
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE TABLE IF NOT EXISTS session_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL,
    session_id TEXT,
    tokens_used INTEGER NOT NULL,
    tokens_budget INTEGER NOT NULL,
    tokens_remaining INTEGER NOT NULL,
    input_tokens INTEGER,
    output_tokens INTEGER,
    estimated_cost_usd REAL,
    model_name TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT,
    path TEXT UNIQUE,
    remote_url TEXT,
    git_remote TEXT,
    repo_url TEXT,
    branch TEXT,
    commit_hash TEXT,
    subpath TEXT,
    module TEXT,
    salt TEXT,
    id_algorithm TEXT DEFAULT 'uuidv5',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_commit TEXT,
    active BOOLEAN DEFAULT 1
);

CREATE TABLE IF NOT EXISTS workflow_states (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    position INTEGER,
    color TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    state_id INTEGER,
    project_id TEXT,
    priority INTEGER DEFAULT 0,
    due_at DATETIME,
    completed_at DATETIME,
    agent_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    archived BOOLEAN DEFAULT 0,
    FOREIGN KEY (state_id) REFERENCES workflow_states(id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE TABLE IF NOT EXISTS task_notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT,
    author TEXT,
    body TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);

CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    color TEXT
);

CREATE TABLE IF NOT EXISTS task_tags (
    task_id TEXT,
    tag_id INTEGER,
    PRIMARY KEY (task_id, tag_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

CREATE TABLE IF NOT EXISTS task_sessions (
    task_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (task_id, session_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS task_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT,
    event_type TEXT,
    payload TEXT,
    actor TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);

CREATE TABLE IF NOT EXISTS meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS commit_tasks (
    commit_id INTEGER NOT NULL,
    task_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (commit_id, task_id),
    FOREIGN KEY (commit_id) REFERENCES commits(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE VIRTUAL TABLE IF NOT EXISTS task_search USING fts5(
    task_id UNINDEXED,
    title,
    description,
    tokenize='porter'
);

CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_updated_at ON agents(updated_at);
CREATE INDEX IF NOT EXISTS idx_agents_source ON agents(source);
CREATE INDEX IF NOT EXISTS idx_agents_window_id ON agents(window_id);
CREATE INDEX IF NOT EXISTS idx_agents_description ON agents(description);
CREATE INDEX IF NOT EXISTS idx_agents_parent_agent_id ON agents(parent_agent_id);
CREATE INDEX IF NOT EXISTS idx_agents_session_id ON agents(session_id);
CREATE INDEX IF NOT EXISTS idx_agents_parent_session_id ON agents(parent_session_id);
CREATE INDEX IF NOT EXISTS idx_agents_bookmarked ON agents(bookmarked);
CREATE INDEX IF NOT EXISTS idx_actions_agent_id ON actions(agent_id);
CREATE INDEX IF NOT EXISTS idx_actions_timestamp ON actions(timestamp);
CREATE INDEX IF NOT EXISTS idx_commits_agent_id ON commits(agent_id);
CREATE INDEX IF NOT EXISTS idx_commits_project ON commits(project_id);
CREATE INDEX IF NOT EXISTS idx_commits_session ON commits(session_id);
CREATE INDEX IF NOT EXISTS idx_session_context_agent_id ON session_context(agent_id);
CREATE INDEX IF NOT EXISTS idx_session_context_session_id ON session_context(session_id);
CREATE INDEX IF NOT EXISTS idx_session_context_updated_at ON session_context(updated_at);
CREATE INDEX IF NOT EXISTS idx_session_notes_agent_id ON session_notes(agent_id);
CREATE INDEX IF NOT EXISTS idx_session_notes_timestamp ON session_notes(timestamp);
CREATE INDEX IF NOT EXISTS idx_session_notes_priority ON session_notes(priority);
CREATE INDEX IF NOT EXISTS idx_session_notes_note_type ON session_notes(note_type);
CREATE INDEX IF NOT EXISTS idx_session_logs_agent_id ON session_logs(agent_id);
CREATE INDEX IF NOT EXISTS idx_session_logs_timestamp ON session_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_session_logs_log_level ON session_logs(log_level);
CREATE INDEX IF NOT EXISTS idx_session_logs_category ON session_logs(category);
CREATE INDEX IF NOT EXISTS idx_personas_specialization ON personas(specialization);
CREATE INDEX IF NOT EXISTS idx_sessions_agent_id ON sessions(agent_id);
CREATE INDEX IF NOT EXISTS idx_logs_session_id ON logs(session_id);
CREATE INDEX IF NOT EXISTS idx_notes_parent ON notes(parent_type, parent_id);
CREATE INDEX IF NOT EXISTS idx_context_session_id ON context(session_id);
CREATE INDEX IF NOT EXISTS idx_compactions_agent_id ON compactions(agent_id);
CREATE INDEX IF NOT EXISTS idx_compactions_new_session_id ON compactions(new_session_id);
CREATE INDEX IF NOT EXISTS idx_compactions_compacted_at ON compactions(compacted_at);
CREATE INDEX IF NOT EXISTS idx_session_metrics_agent_id ON session_metrics(agent_id);
CREATE INDEX IF NOT EXISTS idx_session_metrics_session_id ON session_metrics(session_id);
CREATE INDEX IF NOT EXISTS idx_session_metrics_timestamp ON session_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_session_metrics_session_ts ON session_metrics(session_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state_id);
CREATE INDEX IF NOT EXISTS idx_tasks_due ON tasks(due_at);
CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent_id);
CREATE INDEX IF NOT EXISTS idx_task_notes_task_id ON task_notes(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tags_task ON task_tags(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tags_tag ON task_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_task_sessions_task ON task_sessions(task_id);
CREATE INDEX IF NOT EXISTS idx_task_sessions_session ON task_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_task_events_task ON task_events(task_id);
CREATE INDEX IF NOT EXISTS idx_commit_tasks_commit ON commit_tasks(commit_id);
CREATE INDEX IF NOT EXISTS idx_commit_tasks_task ON commit_tasks(task_id);

CREATE TRIGGER IF NOT EXISTS update_agents_updated_at
AFTER UPDATE ON agents
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE agents SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_session_context_updated_at
AFTER UPDATE ON session_context
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE session_context SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_tasks_updated_at
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
    UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_projects_updated_at
AFTER UPDATE ON projects
FOR EACH ROW
BEGIN
    UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_workflow_states_updated_at
AFTER UPDATE ON workflow_states
FOR EACH ROW
BEGIN
    UPDATE workflow_states SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
CREATE TABLE IF NOT EXISTS action_plans (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id),
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);

CREATE INDEX IF NOT EXISTS idx_action_plans_session ON action_plans(session_id);
CREATE INDEX IF NOT EXISTS idx_action_plans_agent ON action_plans(agent_id);

-- Subagent Prompts: Reusable prompt templates for spawning subagents
CREATE TABLE IF NOT EXISTS subagent_prompts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT,
    prompt_text TEXT NOT NULL,
    project_id TEXT,
    active BOOLEAN DEFAULT 1,
    version INTEGER DEFAULT 1,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CHECK (slug GLOB '[a-z][a-z0-9-]*')
);

-- Partial unique indexes: Allow same slug for global + per-project
CREATE UNIQUE INDEX IF NOT EXISTS idx_subagent_prompts_slug_global
  ON subagent_prompts(slug) WHERE project_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_subagent_prompts_slug_project
  ON subagent_prompts(slug, project_id) WHERE project_id IS NOT NULL;

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_subagent_prompts_project_id ON subagent_prompts(project_id);
CREATE INDEX IF NOT EXISTS idx_subagent_prompts_active ON subagent_prompts(active);

-- BEFORE UPDATE trigger to auto-update timestamps and version
CREATE TRIGGER IF NOT EXISTS update_subagent_prompts_timestamp
BEFORE UPDATE ON subagent_prompts
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    SELECT CASE
        WHEN NEW.prompt_text != OLD.prompt_text THEN
            (SELECT NEW.version + 1)
        ELSE
            OLD.version
    END;
    SELECT CURRENT_TIMESTAMP;
END;
