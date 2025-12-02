package app

// GetActionIcon returns an icon for a given action type using nerdfonts
func GetActionIcon(actionType string) string {
	switch actionType {
	case "task_start":
		return "\uf04b" // nf-fa-play
	case "file_operation":
		return "\uf07b" // nf-fa-folder
	case "git_commit":
		return "\uf040" // nf-fa-pencil
	case "status_update":
		return "\uf080" // nf-fa-bar_chart
	case "log":
		return "\uf0ea" // nf-fa-paste
	case "note":
		return "\uf08d" // nf-fa-thumb_tack
	default:
		return "\uf0e7" // nf-fa-bolt
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
