package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// actionsStore implements SQL operations for actions
type actionsStore struct {
	db *sql.DB
}

// NewActionsStore creates a new actions store
func NewActionsStore(db *sql.DB) *actionsStore {
	return &actionsStore{db: db}
}

// LoadByAgent loads actions for a specific agent
func (s *actionsStore) LoadByAgent(ctx context.Context, agentID domain.AgentID) ([]domain.Action, error) {
	query := `
		SELECT id, agent_id, action_type, description, details, timestamp
		FROM actions
		WHERE agent_id = ?
		ORDER BY timestamp DESC
	`

	rows, err := s.db.QueryContext(ctx, query, string(agentID))
	if err != nil {
		return nil, fmt.Errorf("query actions: %w", err)
	}
	defer rows.Close()

	var actions []domain.Action
	for rows.Next() {
		var a domain.Action
		var details sql.NullString

		err := rows.Scan(&a.ID, &a.AgentID, &a.ActionType, &a.Description, &details, &a.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}

		if details.Valid {
			a.Details = details.String
		}

		actions = append(actions, a)
	}

	return actions, nil
}

// Create creates a new action
func (s *actionsStore) Create(ctx context.Context, action domain.Action) error {
	query := `
		INSERT INTO actions (agent_id, action_type, description, details, timestamp)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := s.db.ExecContext(ctx, query,
		string(action.AgentID),
		action.ActionType,
		action.Description,
		action.Details,
	)
	if err != nil {
		return fmt.Errorf("insert action: %w", err)
	}

	return nil
}