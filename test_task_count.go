package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	// Test the task counting logic directly
	cmd := exec.Command("task", "export")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to export tasks: %v\n", err)
		return
	}

	var tasks []map[string]interface{}
	if err := json.Unmarshal(output, &tasks); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		return
	}

	fmt.Printf("Total tasks exported: %d\n\n", len(tasks))

	// Look for the specific subagent
	targetAgentID := "42ebc445-bc7a-446b-b6ff-33976ca6a479"
	searchTag := "subagent_" + strings.ReplaceAll(targetAgentID, "-", "_")

	fmt.Printf("Looking for tag: %s\n", searchTag)

	count := 0
	for _, task := range tasks {
		if status, ok := task["status"].(string); ok && status == "pending" {
			if tags, ok := task["tags"].([]interface{}); ok {
				for _, tag := range tags {
					if tagStr, ok := tag.(string); ok {
						if tagStr == searchTag {
							fmt.Printf("Found matching task:\n")
							fmt.Printf("  Description: %v\n", task["description"])
							fmt.Printf("  Tags: %v\n", tags)
							fmt.Printf("  Status: %v\n", status)
							count++
							break
						}
					}
				}
			}
		}
	}

	fmt.Printf("\nTotal matching tasks for agent %s: %d\n", targetAgentID, count)
}