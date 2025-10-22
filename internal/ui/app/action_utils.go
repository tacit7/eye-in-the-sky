package app

// GetActionIcon returns an icon/emoji for a given action type
func GetActionIcon(actionType string) string {
	switch actionType {
	case "task_start":
		return "▶️"
	case "file_operation":
		return "📁"
	case "git_commit":
		return "📝"
	case "status_update":
		return "📊"
	case "log":
		return "📋"
	case "note":
		return "📌"
	default:
		return "⚡"
	}
}

// GetPriorityWeight returns a numeric weight for task priority sorting
func GetPriorityWeight(priority string) int {
	switch priority {
	case "H":
		return 3
	case "M":
		return 2
	case "L":
		return 1
	default:
		return 0
	}
}
