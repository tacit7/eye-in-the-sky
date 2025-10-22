# Task Count Display Requirements

## Problem
Currently, the Agent Overview list shows:
- Agent Status
- Agent ID
- Current Task (from database field)
- Source/Project
- Session ID

But it does NOT show TaskWarrior task counts. This means users cannot see that subagents have TaskWarrior tasks associated with them without navigating to the detail view.

## Example
For subagent `42ebc445-bc7a-446b-b6ff-33976ca6a479`:
- Has 3 TaskWarrior tasks (+subagent_42ebc445...)
- Tasks are visible in Detail View > Tasks tab
- But no indication in Agent Overview list

## Solution
Add task count to the Agent Overview display:
```
｜ ACTIVE    42ebc445-bc7a... [3 tasks]    Fix UI...    eye-in-the-sky
```

## Implementation Notes
1. Query TaskWarrior for each agent/subagent
2. Count tasks with matching tags:
   - For parent agents: `+session_<session_id>`
   - For subagents: `+subagent_<agent_id>`
3. Display count inline or as a badge
4. Consider performance - may need caching or async loading

## Benefits
- Immediate visibility of task workload
- No need to navigate to detail view to check
- Better overview of agent activity
- Especially useful for subagents