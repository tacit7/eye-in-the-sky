package messages

import "github.com/tacit7/eye-in-the-sky/internal/domain"

// AgentsLoadedMsg is sent when agents are loaded from the database
type AgentsLoadedMsg struct {
	Agents []domain.Agent
}
