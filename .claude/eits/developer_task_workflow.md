# Developer Task Workflow with TaskWarrior Tags

## Core Workflow States

### 📝 Planning & Design
- **`+draft`** - Task needs spec/design work
- **`+ready`** - Spec complete, ready to implement

### 💻 Active Development
- **`+working`** - Actively coding/implementing
- **`+testing`** - Writing/running tests
- **`+debugging`** - Fixing issues found

### 🔍 Review & Quality
- **`+review`** - Code complete, needs review
- **`+revision`** - Review feedback needs addressing
- **`+qa`** - Needs QA testing

### 🚧 Blocked States
- **`+blocked`** - Can't proceed (add annotation why)
- **`+waiting`** - Waiting on external dependency
- **`+hold`** - Paused for business reasons

### 🚀 Completion States
- **`+merged`** - Code merged to main
- **`+deployed`** - Deployed to production
- **`+verified`** - Confirmed working in production

## Developer Commands

### Starting a new feature:
```bash
# Planning phase
task add "Implement user authentication" project:backend +feature +draft
task <id> annotate "Need to decide: JWT or session-based"

# Ready to implement
task <id> modify -draft +ready
```

### Active development:
```bash
# Start coding
task <id> start
task <id> modify +working
task <id> annotate "Implementing JWT authentication"

# Testing phase
task <id> modify -working +testing
task <id> annotate "Writing unit tests for auth middleware"

# Debugging
task <id> modify -testing +debugging
task <id> annotate "Found race condition in token refresh"
```

### Review process:
```bash
# Submit for review
task <id> modify -debugging +review
task <id> annotate "PR #234 ready for review"

# After review feedback
task <id> modify -review +revision
task <id> annotate "Addressing feedback: improve error messages"

# Resubmit
task <id> modify -revision +review
```

### Handling blocks:
```bash
# Blocked by dependency
task <id> modify +blocked
task <id> annotate "Blocked: waiting for database schema from DBA team"

# When unblocked
task <id> modify -blocked +working
```

### Completion:
```bash
# After merge
task <id> modify +merged
task <id> annotate "Merged in PR #234"

# After deployment
task <id> modify +deployed
task <id> annotate "Deployed to prod in release v2.3.0"

# Mark done
task <id> done
```

## Useful Queries

### What am I working on?
```bash
task +working list
```

### What needs review?
```bash
task +review list
```

### What's blocked?
```bash
task +blocked list
```

### What's in testing?
```bash
task +testing list
```

### What needs revision after review?
```bash
task +revision list
```

### What's ready to start?
```bash
task +ready list
```

### Everything in active development:
```bash
task +working or +testing or +debugging list
```

### High priority blocked items:
```bash
task +blocked priority:H list
```

## Bug-specific Workflow

For bugs, add `+bug` tag and use:
- **`+investigating`** - Reproducing/understanding the bug
- **`+identified`** - Root cause found
- **`+fixing`** - Implementing fix
- **`+testing`** - Verifying fix works
- **`+review`** - Fix needs review

Example:
```bash
task add "Users can't login after password reset" +bug +investigating priority:H
task <id> modify -investigating +identified
task <id> annotate "Root cause: JWT token not invalidated after password change"
task <id> modify -identified +fixing
```

## Integration with Git

Include task IDs in commits:
```bash
git commit -m "feat: Implement JWT authentication [task #42] +working"
git commit -m "test: Add auth middleware tests [task #42] +testing"
git commit -m "fix: Resolve token refresh race condition [task #42] +debugging"
```

## Priority + Status Combinations

High-value queries combining priority and status:
```bash
# Critical bugs being fixed
task +bug +fixing priority:H list

# High priority items ready to start
task +ready priority:H list

# Everything in review sorted by priority
task +review rc.report.list.sort=priority-,urgency-
```

## Team Collaboration Tags

When working with others:
- **`+assigned:<name>`** - Who's working on it
- **`+reviewer:<name>`** - Who's reviewing
- **`+pair`** - Needs pair programming
- **`+mob`** - Mob programming session

Example:
```bash
task <id> modify +working +assigned:uriel
task <id> modify +review +reviewer:alice
```

## Automation Ideas

Create bash aliases for common transitions:
```bash
# Add to ~/.bashrc or ~/.zshrc
alias tw-start='task $1 modify +working'
alias tw-review='task $1 modify -working -testing -debugging +review'
alias tw-block='task $1 modify +blocked'
alias tw-unblock='task $1 modify -blocked +working'
```

Then use: `tw-start 42`, `tw-review 42`, etc.