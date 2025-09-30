# Add Session Note

Add a contextual note to your current session.

## Instructions

1. Use the `i-note` MCP tool to add a timestamped note
2. Specify note type: observation, decision, blocker, reminder, learning
3. Set priority: low, medium, high, critical
4. Optionally add tags for categorization

## Example

```
Use the i-note MCP tool with:
- agent_id: your agent id
- type: "{{type}}"
- content: "{{note}}"
- priority: "{{priority}}" (default: medium)
- tags: array of tags (optional)
```

## Usage

```
/i-note Need to refactor auth middleware before adding new features
```

This adds a note with type "reminder" and medium priority.

## Note Types

- **observation**: Insights discovered during work
- **decision**: Important decisions made
- **blocker**: Issues blocking progress
- **reminder**: Things to remember for later
- **learning**: Lessons learned or new knowledge
