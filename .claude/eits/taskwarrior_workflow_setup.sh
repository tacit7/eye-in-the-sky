#!/bin/bash
# Set up custom task workflow with working and review statuses

echo "Setting up custom task workflow..."

# Create a User Defined Attribute for workflow status
task config uda.workstatus.type string
task config uda.workstatus.label "Work Status"
task config uda.workstatus.values pending,working,review,completed,blocked

# Set up colors for different work statuses
task config color.uda.workstatus.pending yellow
task config color.uda.workstatus.working blue bold
task config color.uda.workstatus.review magenta
task config color.uda.workstatus.completed green
task config color.uda.workstatus.blocked red

# Add urgency coefficients for work status
task config urgency.uda.workstatus.working.coefficient 15.0
task config urgency.uda.workstatus.review.coefficient 10.0
task config urgency.uda.workstatus.pending.coefficient 5.0
task config urgency.uda.workstatus.blocked.coefficient 8.0
task config urgency.uda.workstatus.completed.coefficient -5.0

# Create custom reports that show work status
task config report.workflow.description "Tasks by workflow status"
task config report.workflow.columns id,project,workstatus,description.desc,tags
task config report.workflow.labels ID,Project,Status,Description,Tags
task config report.workflow.filter "status:pending"
task config report.workflow.sort workstatus-,urgency-

# Report for working tasks
task config report.working.description "Tasks currently being worked on"
task config report.working.columns id,project,description.desc,tags
task config report.working.labels ID,Project,Description,Tags
task config report.working.filter "status:pending workstatus:working"
task config report.working.sort urgency-

# Report for review tasks
task config report.review.description "Tasks ready for review"
task config report.review.columns id,project,description.desc,tags
task config report.review.labels ID,Project,Description,Tags
task config report.review.filter "status:pending workstatus:review"
task config report.review.sort urgency-

echo "Custom workflow setup complete!"
echo ""
echo "Usage:"
echo "  Start working: task <id> modify workstatus:working"
echo "  Ready for review: task <id> modify workstatus:review"
echo "  Mark completed: task <id> modify workstatus:completed && task <id> done"
echo "  Mark blocked: task <id> modify workstatus:blocked"
echo ""
echo "View reports:"
echo "  task workflow  - All tasks by workflow status"
echo "  task working   - Tasks currently being worked on"
echo "  task review    - Tasks ready for review"