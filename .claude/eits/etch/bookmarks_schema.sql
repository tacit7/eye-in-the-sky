-- Polymorphic bookmarks table supporting multiple entity types
CREATE TABLE bookmarks (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),

  -- What's being bookmarked
  bookmark_type TEXT NOT NULL CHECK (bookmark_type IN ('file', 'note', 'agent', 'session', 'task', 'url')),
  bookmark_id TEXT,  -- NULL for file/url types, references ID for others

  -- File-specific fields (when bookmark_type='file')
  file_path TEXT,
  line_number INTEGER,

  -- URL-specific fields (when bookmark_type='url')
  url TEXT,

  -- Common metadata
  title TEXT,
  description TEXT,

  -- Organization
  category TEXT,  -- e.g., 'important', 'review-later', 'bugs', 'ideas'
  priority INTEGER DEFAULT 0,  -- Higher = more important
  position INTEGER,  -- For manual ordering within category

  -- Context
  project_id INTEGER REFERENCES projects(id),
  agent_id TEXT REFERENCES agents(id),  -- Who created it

  -- Timestamps
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  accessed_at TIMESTAMP,  -- Last time bookmark was accessed

  -- Validation
  CHECK (
    (bookmark_type = 'file' AND file_path IS NOT NULL) OR
    (bookmark_type = 'url' AND url IS NOT NULL) OR
    (bookmark_type IN ('note', 'agent', 'session', 'task') AND bookmark_id IS NOT NULL)
  )
);

-- Indexes for common queries
CREATE INDEX idx_bookmarks_type ON bookmarks(bookmark_type);
CREATE INDEX idx_bookmarks_project ON bookmarks(project_id);
CREATE INDEX idx_bookmarks_agent ON bookmarks(agent_id);
CREATE INDEX idx_bookmarks_category ON bookmarks(category);
CREATE INDEX idx_bookmarks_priority ON bookmarks(priority DESC);

-- Examples:
-- File bookmark: bookmark_type='file', file_path='/path/to/file.go', line_number=123
-- Note bookmark: bookmark_type='note', bookmark_id='note-uuid'
-- Agent bookmark: bookmark_type='agent', bookmark_id='agent-uuid'
-- URL bookmark: bookmark_type='url', url='https://docs.example.com'
