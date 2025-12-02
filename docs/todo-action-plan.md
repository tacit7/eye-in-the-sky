Backend Implementation Plan – To-Do System (SQLite + Go)

This document describes the exact steps for building the backend of the to-do system. It assumes Go 1.22+, SQLite (with FTS5 enabled), and the standard database/sql package.

⸻

1. Goals
	•	Local-only persistent task manager using SQLite.
	•	Supports Projects, Workflows, Tasks, Subtasks, Tags, and Notes.
	•	Full-text search (FTS5) on descriptions and notes.
	•	Weekly FTS reindex every Sunday.
	•	Safe ordering, recursive subtasks, soft deletes.
	•	Clean repository layer for the CLI/TUI.

⸻

2. Folder Structure

cmd/
  todo/
    main.go
internal/
  db/
    db.go
    migrate.go
    maintenance.go
  models/
    project.go
    task.go
    note.go
    tag.go
  repository/
    project_repo.go
    task_repo.go
    note_repo.go
  util/
    yaml_workflow.go


⸻

3. Schema Summary

Tables
	•	projects – contains UUID, name, and soft-delete flag.
	•	workflow_states – project-scoped state codes and user labels.
	•	tasks – main table with description, priority, weight, parent_id, position, etc.
	•	task_notes – append-only markdown notes.
	•	tags – global tag names.
	•	task_tags – join table for many-to-many relation.
	•	task_events – optional audit trail.
	•	task_search – FTS5 virtual table for text search.
	•	meta – stores maintenance metadata (e.g., last reindex date).

Triggers
	•	Update updated_at on each task change.
	•	Sync task_search on insert/update/delete.
	•	Sync latest note to FTS index.

Weekly Reindex
	•	Reindex runs automatically on Sundays during OpenDB().
	•	Last run tracked in meta (key='last_reindex_at').

⸻

4. Repository Methods (Core CRUD)

ProjectRepo
	•	CreateOrGetProjectByGitRepo(repoSlug string) – find or insert by repo name.
	•	SoftDeleteProject(id int) – set archived_at.
	•	SyncWorkflowFromYAML(projectID int, yaml WorkflowYAML) – upsert states.

TaskRepo
	•	CreateTask(projectID int, desc string, parentID *int, stateCode string)
	•	UpdateDescription(taskID int, desc string)
	•	AddNote(taskID int, bodyMD string)
	•	DeleteNote(noteID int)
	•	AddTag(taskID int, tag string)
	•	RemoveTag(taskID int, tag string)
	•	MoveToState(taskID int, stateCode string)
	•	SetDue(taskID int, due *time.Time)
	•	SetPriority(taskID int, p *int)
	•	SetWeight(taskID int, w *int)
	•	SetParent(childID, parentID *int) – prevents cycles, depth > 32.
	•	Reorder(taskID int, newPos int)
	•	List(projectID int, filters Filters, order SortOrder)
	•	FindByID(taskID int)
	•	Search(projectID int, query string, limit, offset int)

Notes
	•	Append-only, deletable through API.
	•	New notes automatically update the search snapshot.

⸻

5. Query & Search Behavior
	•	FTS5 search: SELECT t.*, rank FROM task_search s JOIN tasks t ON t.id=s.rowid WHERE t.project_id=? AND s MATCH ? ORDER BY rank ASC LIMIT ?;
	•	Project-scoped by default.
	•	Includes latest note text.
	•	Filters combine with AND semantics (status, tags, due).

⸻

6. Maintenance

Automatic on startup
	1.	Open database in WAL mode:

PRAGMA journal_mode=WAL;


	2.	Run maybeReindex():

if time.Now().Weekday()==time.Sunday && sevenDaysSince(lastReindex) {
    db.Exec(`REINDEX task_search;`)
    db.Exec(`UPDATE meta SET value=date('now') WHERE key='last_reindex_at';`)
}


	3.	Optional: VACUUM + ANALYZE once weekly after reindex.

Manual Commands
	•	todo db vacuum → runs VACUUM; ANALYZE;
	•	todo db reindex → full rebuild of task_search.

⸻

7. Data Rules & Validation
	•	Description cannot be empty.
	•	Tags trimmed; max length 64.
	•	Priority 1–5; weight integer, optional.
	•	Parent cannot equal child; depth limited to 32.
	•	Subtasks deleted recursively.
	•	Projects and tasks use soft delete (archived_at).

⸻

8. YAML Workflow Sync

Example YAML:

workflow:
  - code: todo
    label: To do
  - code: doing
    label: Doing
  - code: review
    label: Needs review
  - code: done
    label: Done

Sync logic:
	•	Parse YAML; upsert each state (match by code).
	•	Update display_name and position.
	•	Warn if unknown state codes appear.

⸻

9. Development Checklist
	1.	Setup: create todo.db, enable FTS5.
	2.	Migrations: build migrate.go with CREATE TABLE statements.
	3.	Triggers: copy the provided trigger SQL.
	4.	Repositories: implement ProjectRepo, then TaskRepo, then NoteRepo.
	5.	Validation: add checks for empty desc, recursion, etc.
	6.	Maintenance: implement maybeReindex() and meta tracking.
	7.	CLI integration: wire repository calls to Bubble Tea commands.
	8.	Test: add fixtures for project creation, task creation, tagging, search, and reindex.

⸻

10. Testing Scenarios
	•	Create project from GitHub slug.
	•	Add tasks, subtasks, and notes.
	•	Add/remove tags.
	•	Search with FTS (confirm ranking).
	•	Trigger Sunday reindex.
	•	Validate recursive delete works.
	•	Validate YAML sync adds new workflow states.

⸻

11. Deliverables
	•	todo.db schema and triggers verified.
	•	Go repositories with full CRUD + search.
	•	CLI/TUI command bindings.
	•	Working Sunday reindex and manual todo db vacuum.

⸻

Done right, this backend stays tiny, fast, and predictable; no dependencies beyond the Go stdlib and SQLite.
