package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Mock data generators
var (
	statuses = []string{"active", "working", "idle", "completed", "failed"}
	features = []string{
		"user authentication system",
		"payment processing integration",
		"real-time notifications",
		"data analytics dashboard",
		"API rate limiting",
		"search functionality",
		"file upload system",
		"email verification",
		"password reset flow",
		"admin panel",
		"user profile management",
		"chat system",
		"reporting engine",
		"export functionality",
		"multi-language support",
	}
	tasks = []string{
		"implementing database schema",
		"writing unit tests",
		"fixing authentication bug",
		"optimizing query performance",
		"updating documentation",
		"refactoring legacy code",
		"adding validation logic",
		"integrating third-party API",
		"debugging production issue",
		"reviewing pull requests",
	}
	projects = []string{
		"web-app", "api-server", "mobile-backend", "data-pipeline",
		"microservice", "cli-tool", "dashboard", "automation-tool",
	}
	noteContents = []string{
		"Initial session started for TUI refactoring work",
		"Discovered N+1 query pattern in agent loading - needs optimization",
		"Removed FTS5 triggers for better performance",
		"Fixed viewport scroll bounds issue",
		"Added Bubble Tea viewport migration",
		"Refactored overview tabs to use proper MVC pattern",
		"Implemented keybinding system with YAML configuration",
		"Added title field to notes table",
		"Updated UI to display timestamp, title, and content",
		"Created migration for database schema changes",
	}
	actionTypes = []string{"task_start", "file_operation", "git_commit", "status_update"}
	actionDescriptions = []string{
		"Started working on authentication feature",
		"Modified database schema for notes table",
		"Committed viewport migration changes",
		"Updated agent status to working",
		"Fixed import cycle in overview package",
		"Refactored tab rendering logic",
		"Added scroll indicators to viewports",
	}
	commitMessages = []string{
		"feat: Add user authentication",
		"fix: Resolve viewport scroll issue",
		"refactor: Extract shared types to common package",
		"docs: Update API documentation",
		"test: Add unit tests for agents",
		"chore: Update dependencies",
		"perf: Optimize database queries",
	}
)

func main() {
	count := flag.Int("count", 5, "Number of agents to create")
	dbPath := flag.String("db", "", "Path to database (default: ~/.config/eye-in-the-sky/agents.db)")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	// Determine database path
	var finalDBPath string
	if *dbPath != "" {
		finalDBPath = *dbPath
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		finalDBPath = filepath.Join(home, ".config", "eye-in-the-sky", "agents.db")
	}

	// Open database
	db, err := sql.Open("sqlite3", finalDBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	fmt.Printf("Creating %d test agents in %s...\n", *count, finalDBPath)

	for i := 0; i < *count; i++ {
		agent := createMockAgent()
		session := createMockSession(agent.ID)

		// Insert agent
		if err := insertAgent(ctx, db, agent); err != nil {
			log.Printf("Failed to insert agent: %v", err)
			continue
		}

		// Insert session
		if err := insertSession(ctx, db, session); err != nil {
			log.Printf("Failed to insert session: %v", err)
			continue
		}

		// Create random number of notes (0-5)
		noteCount := rand.Intn(6)
		for j := 0; j < noteCount; j++ {
			note := createMockNote(session.ID)
			if err := insertNote(ctx, db, note); err != nil {
				log.Printf("Failed to insert note: %v", err)
			}
		}

		// Create random number of actions (1-10)
		actionCount := rand.Intn(10) + 1
		for j := 0; j < actionCount; j++ {
			action := createMockAction(agent.ID)
			if err := insertAction(ctx, db, action); err != nil {
				log.Printf("Failed to insert action: %v", err)
			}
		}

		// Create random number of commits (0-5)
		commitCount := rand.Intn(6)
		for j := 0; j < commitCount; j++ {
			commit := createMockCommit(agent.ID)
			if err := insertCommit(ctx, db, commit); err != nil {
				log.Printf("Failed to insert commit: %v", err)
			}
		}

		fmt.Printf("✓ Created agent %s (session: %s) - %s: %s\n",
			agent.ID[:8], session.ID[:8], agent.Status, agent.FeatureDesc)
	}

	fmt.Printf("\n✅ Successfully created %d test agents\n", *count)
}

type Agent struct {
	ID              string
	Status          string
	Source          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	GitWorktreePath string
	FeatureDesc     string
	CurrentTask     string
	LastActivityAt  time.Time
	SessionID       string
	ProjectName     string
}

type Session struct {
	ID        string
	AgentID   string
	Name      string
	StartedAt time.Time
	EndedAt   *time.Time
}

type Note struct {
	SessionID string
	Title     string
	Content   string
	CreatedAt time.Time
}

type Action struct {
	AgentID     string
	ActionType  string
	Description string
	Timestamp   time.Time
}

type Commit struct {
	AgentID      string
	CommitHash   string
	CommitMsg    string
	Timestamp    time.Time
}

func createMockAgent() Agent {
	agentID := uuid.New().String()
	sessionID := uuid.New().String()
	status := statuses[rand.Intn(len(statuses))]
	feature := features[rand.Intn(len(features))]
	task := tasks[rand.Intn(len(tasks))]
	project := projects[rand.Intn(len(projects))]

	createdAt := time.Now().Add(-time.Duration(rand.Intn(72)) * time.Hour)
	lastActivity := createdAt.Add(time.Duration(rand.Intn(3600)) * time.Minute)

	return Agent{
		ID:              agentID,
		Status:          status,
		Source:          "worktree",
		CreatedAt:       createdAt,
		UpdatedAt:       lastActivity,
		GitWorktreePath: fmt.Sprintf("/tmp/worktree-%s", agentID[:8]),
		FeatureDesc:     feature,
		CurrentTask:     task,
		LastActivityAt:  lastActivity,
		SessionID:       sessionID,
		ProjectName:     project,
	}
}

func createMockSession(agentID string) Session {
	sessionID := uuid.New().String()
	startedAt := time.Now().Add(-time.Duration(rand.Intn(48)) * time.Hour)

	return Session{
		ID:        sessionID,
		AgentID:   agentID,
		Name:      fmt.Sprintf("Session-%s", sessionID[:8]),
		StartedAt: startedAt,
		EndedAt:   nil,
	}
}

func createMockNote(sessionID string) Note {
	content := noteContents[rand.Intn(len(noteContents))]
	title := extractTitle(content)

	return Note{
		SessionID: sessionID,
		Title:     title,
		Content:   content,
		CreatedAt: time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
	}
}

func createMockAction(agentID string) Action {
	return Action{
		AgentID:     agentID,
		ActionType:  actionTypes[rand.Intn(len(actionTypes))],
		Description: actionDescriptions[rand.Intn(len(actionDescriptions))],
		Timestamp:   time.Now().Add(-time.Duration(rand.Intn(12)) * time.Hour),
	}
}

func createMockCommit(agentID string) Commit {
	return Commit{
		AgentID:      agentID,
		CommitHash:   generateGitHash(),
		CommitMsg:    commitMessages[rand.Intn(len(commitMessages))],
		Timestamp:    time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
	}
}

func extractTitle(content string) string {
	words := []rune(content)
	spaceCount := 0
	for i, r := range words {
		if r == ' ' {
			spaceCount++
			if spaceCount == 2 {
				return string(words[:i])
			}
		}
	}
	if len(content) > 20 {
		return content[:20]
	}
	return content
}

func generateGitHash() string {
	const chars = "0123456789abcdef"
	hash := make([]byte, 40)
	for i := range hash {
		hash[i] = chars[rand.Intn(len(chars))]
	}
	return string(hash)
}

func insertAgent(ctx context.Context, db *sql.DB, a Agent) error {
	query := `
		INSERT INTO agents (
			id, status, source, created_at, updated_at,
			git_worktree_path, feature_description, current_task,
			last_activity_at, session_id, project_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.ExecContext(ctx, query,
		a.ID, a.Status, a.Source, a.CreatedAt, a.UpdatedAt,
		a.GitWorktreePath, a.FeatureDesc, a.CurrentTask,
		a.LastActivityAt, a.SessionID, a.ProjectName,
	)
	return err
}

func insertSession(ctx context.Context, db *sql.DB, s Session) error {
	query := `INSERT INTO sessions (id, agent_id, name, started_at, ended_at) VALUES (?, ?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, query, s.ID, s.AgentID, s.Name, s.StartedAt, s.EndedAt)
	return err
}

func insertNote(ctx context.Context, db *sql.DB, n Note) error {
	query := `INSERT INTO notes (session_id, title, content, created_at) VALUES (?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, query, n.SessionID, n.Title, n.Content, n.CreatedAt)
	return err
}

func insertAction(ctx context.Context, db *sql.DB, a Action) error {
	query := `INSERT INTO actions (agent_id, action_type, description, timestamp) VALUES (?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, query, a.AgentID, a.ActionType, a.Description, a.Timestamp)
	return err
}

func insertCommit(ctx context.Context, db *sql.DB, c Commit) error {
	query := `INSERT INTO commits (agent_id, commit_hash, commit_message, timestamp) VALUES (?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, query, c.AgentID, c.CommitHash, c.CommitMsg, c.Timestamp)
	return err
}
