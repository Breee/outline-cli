# Feature: CI Mode

## Summary

When `OUTLINE_CLI_CI_MODE=1` (or `true`/`yes`) is set, the CLI skips all interactive confirmations and proceeds automatically. This enables unattended usage in CI/CD pipelines without requiring `--yes` on every command.

## Behavior

- `push`: skips the confirmation prompt (equivalent to `--yes`)
- Any future interactive prompts: auto-confirmed
- Diff output (`--diff`) still prints to stderr for CI logs
- Exit codes remain meaningful for pipeline gating

## Detection

```go
func isCIMode() bool {
    v := os.Getenv("OUTLINE_CLI_CI_MODE")
    switch strings.ToLower(v) {
    case "1", "true", "yes":
        return true
    }
    return false
}
```

## Integration Points

- `confirmPush()` in `cmd/push.go`: if `isCIMode()`, return `true` without prompting
- Root command `PersistentPreRun`: log a notice when CI mode is active

## Design Decisions

- Environment variable (not flag) because CI mode is a session property, not per-command
- Name includes `CLI` to avoid collision with generic `CI` vars
- Still respects `--diff` to produce reviewable output in CI logs
- Does NOT skip `--create-collection` safety check (explicit opt-in still required)

## Priority

Medium — useful once push is stable and teams adopt CI pipelines.
