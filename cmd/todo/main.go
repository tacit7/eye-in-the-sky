package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tacit7/eye-in-the-sky/internal/todo"
	"github.com/tacit7/eye-in-the-sky/internal/todo/db"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Initialize database
	database, err := db.OpenDB()
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Create service
	svc := todo.NewService(database)
	defer svc.Close()

	// Route command
	switch command {
	case "project":
		handleProject(svc, os.Args[2:])
	case "task":
		handleTask(svc, os.Args[2:])
	case "note":
		handleNote(svc, os.Args[2:])
	case "search":
		handleSearch(svc, os.Args[2:])
	case "db":
		handleDatabase(svc, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Todo CLI - Manage tasks using SQLite backend

Usage:
  todo <command> [options]

Commands:
  project     Manage projects
  task        Manage tasks
  note        Manage notes
  search      Search tasks and notes
  db          Database maintenance

Examples:
  todo project create "My Project"
  todo task create <project_id> "Task description"
  todo task list <project_id>
  todo task add-tag <task_id> "important"
  todo note add <task_id> "This is a note"
  todo search <project_id> "search query"
  todo db vacuum
  todo db reindex

For more information, use:
  todo <command> help
`)
}

func handleProject(svc *todo.Service, args []string) {
	if len(args) == 0 {
		fmt.Println("project subcommands: create, list, update, delete")
		return
	}

	subcommand := args[0]

	switch subcommand {
	case "create":
		if len(args) < 2 {
			fmt.Println("Usage: todo project create <name>")
			return
		}
		name := args[1]
		uuid := fmt.Sprintf("project-%d", os.Getpid()) // Placeholder UUID
		project, err := svc.CreateProject(uuid, name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created project: %v\n", project)

	case "list":
		projects, err := svc.ListProjects()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Projects:\n")
		for _, p := range projects {
			fmt.Printf("  [%d] %s\n", p.ID, p.Name)
		}

	default:
		fmt.Printf("Unknown project subcommand: %s\n", subcommand)
	}
}

func handleTask(svc *todo.Service, args []string) {
	if len(args) == 0 {
		fmt.Println("task subcommands: create, list, get, delete, add-tag, add-note")
		return
	}

	subcommand := args[0]

	switch subcommand {
	case "create":
		if len(args) < 3 {
			fmt.Println("Usage: todo task create <project_id> <description>")
			return
		}
		projectID := 0
		fmt.Sscanf(args[1], "%d", &projectID)
		description := args[2]

		task, err := svc.CreateTask(projectID, description, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created task: %v\n", task)

	case "list":
		if len(args) < 2 {
			fmt.Println("Usage: todo task list <project_id>")
			return
		}
		projectID := 0
		fmt.Sscanf(args[1], "%d", &projectID)

		tasks, err := svc.ListTasks(projectID, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Tasks for project %d:\n", projectID)
		for _, t := range tasks {
			fmt.Printf("  [%d] %s\n", t.ID, t.Description)
		}

	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: todo task get <task_id>")
			return
		}
		taskID := 0
		fmt.Sscanf(args[1], "%d", &taskID)

		task, err := svc.GetTask(taskID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Task: %v\n", task)

	case "add-tag":
		if len(args) < 3 {
			fmt.Println("Usage: todo task add-tag <task_id> <tag_name>")
			return
		}
		taskID := 0
		fmt.Sscanf(args[1], "%d", &taskID)
		tagName := args[2]

		tag, err := svc.AddTaskTag(taskID, tagName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Added tag: %v\n", tag)

	case "add-note":
		if len(args) < 3 {
			fmt.Println("Usage: todo task add-note <task_id> <note_text>")
			return
		}
		taskID := 0
		fmt.Sscanf(args[1], "%d", &taskID)
		noteText := args[2]

		note, err := svc.AddTaskNote(taskID, noteText)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Added note: %v\n", note)

	default:
		fmt.Printf("Unknown task subcommand: %s\n", subcommand)
	}
}

func handleNote(svc *todo.Service, args []string) {
	if len(args) == 0 {
		fmt.Println("note subcommands: get, add, delete")
		return
	}

	subcommand := args[0]

	switch subcommand {
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: todo note get <task_id>")
			return
		}
		taskID := 0
		fmt.Sscanf(args[1], "%d", &taskID)

		notes, err := svc.GetNotesByTask(taskID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Notes for task %d:\n", taskID)
		for _, n := range notes {
			fmt.Printf("  [%d] %s\n", n.ID, n.BodyMarkdown)
		}

	default:
		fmt.Printf("Unknown note subcommand: %s\n", subcommand)
	}
}

func handleSearch(svc *todo.Service, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: todo search <project_id> <query>")
		return
	}

	projectID := 0
	fmt.Sscanf(args[0], "%d", &projectID)
	query := args[1]

	results, err := svc.SearchTasks(projectID, query, 10, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Search results for '%s':\n", query)
	for _, result := range results {
		fmt.Printf("  [%d] %s (rank: %.2f)\n", result.Task.ID, result.Task.Description, result.Rank)
	}
}

func handleDatabase(svc *todo.Service, args []string) {
	if len(args) == 0 {
		fmt.Println("db subcommands: vacuum, reindex, check-reindex")
		return
	}

	subcommand := args[0]

	switch subcommand {
	case "vacuum":
		if err := svc.Vacuum(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Database vacuumed")

	case "reindex":
		if err := svc.Reindex(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Search index reindexed")

	case "check-reindex":
		if err := svc.CheckAndMaybeReindex(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Reindex check completed")

	default:
		fmt.Printf("Unknown db subcommand: %s\n", subcommand)
	}
}
