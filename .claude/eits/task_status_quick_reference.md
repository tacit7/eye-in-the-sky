# Task Status Quick Reference

## Essential Developer States (Most Common)

```bash
+ready      → Ready to start implementation
+working    → Actively coding
+testing    → Writing/running tests
+review     → Code complete, needs review
+blocked    → Can't proceed (annotate why)
+done       → task done (completes it)
```

## Quick Commands

```bash
# Start work
task 42 modify +working

# Submit for review
task 42 modify -working +review

# Mark blocked
task 42 modify +blocked
task 42 annotate "Blocked: waiting for API endpoint"

# Complete task
task 42 done
```

## Most Useful Queries

```bash
task +working list           # What am I doing now?
task +review list            # What needs review?
task +ready list             # What can I start?
task +blocked list           # What's stuck?
task +working or +testing    # All active work
```

## Bug Workflow

```bash
+investigating → +identified → +fixing → +testing → +review
```

## Feature Workflow

```bash
+draft → +ready → +working → +testing → +review → +merged → +deployed
```

## Pro Tips

1. Always annotate when blocking:
   ```bash
   task 42 modify +blocked && task 42 annotate "why blocked"
   ```

2. Chain status changes:
   ```bash
   task 42 modify -working -testing +review
   ```

3. Quick status check:
   ```bash
   task 42 info | grep Tags
   ```

4. See task history:
   ```bash
   task 42 history
   ```