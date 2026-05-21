<!-- Collection: test -->
<!-- Title: Feature: Doc Health & Freshness -->

# Feature: Doc Health & Freshness

## Why This Matters

The #1 problem with documentation isn't writing it — it's keeping it current. Stale docs are worse than no docs because they actively mislead. A CLI that surfaces staleness, broken links, and quality issues transforms docs from a liability into a living asset.

This is what separates a great wiki from a graveyard of outdated pages.

## Commands

### `outline health`

Generate a health report for a collection or the entire wiki.

```bash
# Health report for a collection
outline health --collection "Engineering"

# Full wiki health
outline health --all

# Machine-readable output
outline health --collection "Ops" --format json

# Only show problems
outline health --collection "Engineering" --problems-only
```

### Output

```
📊 Health Report: Engineering (42 documents)

🔴 Critical (action needed):
   Deployment Guide         — last updated 187 days ago (stale threshold: 90d)
   API v1 Reference         — last updated 342 days ago
   Onboarding Checklist     — 3 broken links detected

🟡 Warning:
   Monitoring Setup         — last updated 67 days ago (approaching 90d threshold)
   Database Migrations      — no owner assigned
   Security Policies        — 0 views in last 30 days (possibly obsolete)

🟢 Healthy: 36 documents up to date

Summary:
   Freshness score: 86% (36/42 within threshold)
   Link health:     97% (1 doc with broken links)
   Coverage:        — (no coverage rules defined)
```

### `outline health --watch`

Continuous monitoring mode — run in CI on a schedule to alert when docs go stale:

```bash
# Exit code 1 if any doc exceeds staleness threshold (for CI)
outline health --collection "Engineering" --max-stale 90d --exit-code

# Post results to Slack/webhook
outline health --all --format json | curl -X POST -d @- $SLACK_WEBHOOK
```

## Configuration

Per-collection staleness thresholds in config or metadata:

```yaml
# .outline-health.yaml (in repo root)
rules:
  default:
    max_stale_days: 90
  collections:
    Operations:
      max_stale_days: 30  # ops docs must be fresher
    Architecture:
      max_stale_days: 180  # ADRs age more gracefully
  ignore:
    - "Archive/*"
    - "Meeting Notes/*"
```

## Metrics Tracked

| Metric | Source | Why |
|--------|--------|-----|
| Last updated | Outline API `updatedAt` | Core staleness signal |
| View count (30d) | Outline API (if available) | Identifies potentially obsolete docs |
| Broken links | Parse markdown, resolve internal links | Broken UX |
| Missing metadata | Parse headers | Organizational gaps |
| Orphan docs | No parent, not in navigation | Lost/unfindable content |
| Document length | Content analysis | Very short docs may be stubs |

## Link Validation

```bash
outline links --path ./docs/          # check local files before push
outline links --collection "Eng"      # check remote docs
```

Checks:
- Internal wiki links (do target docs exist?)
- External URLs (HTTP HEAD request, report 404s)
- Image references (do attachments exist?)
- Anchor links (`#heading` — does the heading exist in target?)

## Implementation Plan

1. Add `documents.list` with metadata (updatedAt, views) to API client
2. Create `internal/health/` package:
   - `report.go` — compute health metrics
   - `rules.go` — load/apply staleness rules
   - `links.go` — parse and validate links
3. Create `cmd/health.go` and `cmd/links.go`
4. Add `.outline-health.yaml` config loader

## CI Integration

```yaml
# Run weekly to alert on stale docs
name: Doc Health Check
on:
  schedule:
    - cron: '0 9 * * 1'  # Monday 9am

jobs:
  health:
    runs-on: ubuntu-latest
    steps:
      - uses: Breee/outline-cli/action@v1
        with:
          command: health
          server-url: ${{ vars.OUTLINE_URL }}
          api-token: ${{ secrets.OUTLINE_API_TOKEN }}
          args: --all --max-stale 90d --exit-code
```

## Testing

- Unit test staleness calculation with various dates
- Unit test link parsing and resolution
- Unit test rule loading and matching
- Golden file tests for report formatting
- Integration test with mock server returning docs of various ages
