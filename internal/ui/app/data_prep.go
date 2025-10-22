package app

import (
	"sort"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// AgentProvider provides access to agent data
type AgentProvider interface {
	Agents() []domain.Agent
}

// FilterConfig defines filtering options for agent data
type FilterConfig struct {
	ShowAll      bool
	StatusFilter []string // Empty means show all statuses
	SourceFilter []string // Empty means show all sources
}

// SortConfig defines sorting options for agent data
type SortConfig struct {
	SortBy    string // "created", "status", "name", "hierarchy"
	Ascending bool
}

// PrepareAgentData filters, sorts, and groups agent data
func PrepareAgentData(provider AgentProvider, filter FilterConfig, sort SortConfig) []domain.Agent {
	agents := provider.Agents()

	// Apply filters
	filtered := filterAgents(agents, filter)

	// Apply sorting
	sorted := sortAgents(filtered, sort)

	return sorted
}

// filterAgents applies filters to the agent list
func filterAgents(agents []domain.Agent, filter FilterConfig) []domain.Agent {
	if !filter.ShowAll && len(filter.StatusFilter) == 0 {
		// Default: show only active agents
		filter.StatusFilter = []string{"active", "working", "idle"}
	}

	var filtered []domain.Agent
	for _, agent := range agents {
		if shouldIncludeAgent(agent, filter) {
			filtered = append(filtered, agent)
		}
	}
	return filtered
}

// shouldIncludeAgent determines if an agent passes the filters
func shouldIncludeAgent(agent domain.Agent, filter FilterConfig) bool {
	// Check status filter
	if len(filter.StatusFilter) > 0 {
		statusMatch := false
		for _, status := range filter.StatusFilter {
			if string(agent.Status) == status {
				statusMatch = true
				break
			}
		}
		if !statusMatch {
			return false
		}
	}

	// Check source filter
	if len(filter.SourceFilter) > 0 {
		sourceMatch := false
		for _, source := range filter.SourceFilter {
			if string(agent.Source) == source {
				sourceMatch = true
				break
			}
		}
		if !sourceMatch {
			return false
		}
	}

	return true
}

// sortAgents sorts the agent list according to the sort configuration
func sortAgents(agents []domain.Agent, config SortConfig) []domain.Agent {
	sorted := make([]domain.Agent, len(agents))
	copy(sorted, agents)

	switch config.SortBy {
	case "hierarchy":
		sortByHierarchy(sorted)
	case "status":
		sortByStatus(sorted, config.Ascending)
	case "name":
		sortByName(sorted, config.Ascending)
	default: // "created" or default
		sortByCreated(sorted, config.Ascending)
	}

	return sorted
}

// sortByHierarchy sorts agents with parent-child relationships grouped
func sortByHierarchy(agents []domain.Agent) {
	sort.Slice(agents, func(i, j int) bool {
		// Parents (no ParentSessionID) come first
		if agents[i].ParentSessionID == "" && agents[j].ParentSessionID != "" {
			return true
		}
		if agents[i].ParentSessionID != "" && agents[j].ParentSessionID == "" {
			return false
		}

		// If both are parents or both are children, sort by creation time
		return agents[i].CreatedAt.Before(agents[j].CreatedAt)
	})
}

// sortByStatus sorts agents by their status
func sortByStatus(agents []domain.Agent, ascending bool) {
	statusPriority := map[string]int{
		"active":    1,
		"working":   2,
		"idle":      3,
		"completed": 4,
		"failed":    5,
	}

	sort.Slice(agents, func(i, j int) bool {
		iPri := statusPriority[string(agents[i].Status)]
		jPri := statusPriority[string(agents[j].Status)]

		if iPri == 0 {
			iPri = 99
		}
		if jPri == 0 {
			jPri = 99
		}

		if ascending {
			return iPri < jPri
		}
		return iPri > jPri
	})
}

// sortByName sorts agents by their ID
func sortByName(agents []domain.Agent, ascending bool) {
	sort.Slice(agents, func(i, j int) bool {
		if ascending {
			return agents[i].ID < agents[j].ID
		}
		return agents[i].ID > agents[j].ID
	})
}

// sortByCreated sorts agents by creation time
func sortByCreated(agents []domain.Agent, ascending bool) {
	sort.Slice(agents, func(i, j int) bool {
		if ascending {
			return agents[i].CreatedAt.Before(agents[j].CreatedAt)
		}
		return agents[i].CreatedAt.After(agents[j].CreatedAt)
	})
}

// BuildAgentHierarchy creates a hierarchical structure of agents
func BuildAgentHierarchy(agents []domain.Agent) map[string][]domain.Agent {
	hierarchy := make(map[string][]domain.Agent)

	for _, agent := range agents {
		if agent.ParentSessionID == "" {
			// This is a parent agent
			hierarchy[agent.SessionID] = []domain.Agent{agent}
		} else {
			// This is a subagent
			hierarchy[agent.ParentSessionID] = append(hierarchy[agent.ParentSessionID], agent)
		}
	}

	return hierarchy
}

// RenderSubagentGroup prepares agent data for hierarchical display
func RenderSubagentGroup(agents []domain.Agent) []domain.Agent {
	var result []domain.Agent

	// First, sort by hierarchy
	sortByHierarchy(agents)

	// Track which agents have been added
	added := make(map[string]bool)

	// Build parent-child groups
	for _, agent := range agents {
		// Skip if already added
		if added[agent.SessionID] {
			continue
		}

		// If this is a parent, add it and its children
		if agent.ParentSessionID == "" {
			result = append(result, agent)
			added[agent.SessionID] = true

			// Find and add all children
			for _, child := range agents {
				if child.ParentSessionID == agent.SessionID && !added[child.SessionID] {
					result = append(result, child)
					added[child.SessionID] = true
				}
			}
		}
	}

	// Add any orphaned children
	for _, agent := range agents {
		if !added[agent.SessionID] {
			result = append(result, agent)
		}
	}

	return result
}

// GetVisibleAgents returns agents that should be visible based on current view settings
func GetVisibleAgents(m *Model) []domain.Agent {
	filter := FilterConfig{
		ShowAll: m.showAll,
	}

	sort := SortConfig{
		SortBy:    "hierarchy",
		Ascending: false,
	}

	// Use the model as the provider
	provider := modelAgentProvider{m: m}
	return PrepareAgentData(provider, filter, sort)
}

// modelAgentProvider implements AgentProvider for the Model
type modelAgentProvider struct {
	m *Model
}

func (p modelAgentProvider) Agents() []domain.Agent {
	return p.m.agents
}