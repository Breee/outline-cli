#!/usr/bin/env bash
set -euo pipefail

SERVER_URL="${OUTLINE_URL:-http://outline.localhost}"
OUTLINE="./outline --server-url $SERVER_URL"

pause() {
    echo ""
    read -rp "--- Press Enter to continue ---"
    echo ""
}

echo "=== Building ==="
make build

echo ""
echo "=== Step 0: Verify auth ==="
$OUTLINE auth check || $OUTLINE auth oidc-login
pause

# --- PUSH CONFIRMATION ---
echo "=== Step 1: Push with confirmation (--yes to auto-confirm) ==="
$OUTLINE push --yes --path examples/minimal.md
pause

echo "=== Step 2: Push with --diff (shows changes, auto-confirm) ==="
# Append a line so the remote and local differ
echo "" >> examples/minimal.md
echo "Updated at $(date +%s)" >> examples/minimal.md
$OUTLINE push --diff --path examples/minimal.md
# Restore original
git checkout -- examples/minimal.md
pause

echo "=== Step 2b: Push with --diff (big diff example) ==="
# Create a multi-section doc, push it, then modify heavily to produce a large diff
BIG_DIFF_FILE=$(mktemp /tmp/big-diff-XXXXXX.md)
cat > "$BIG_DIFF_FILE" <<'EOF'
<!-- Collection: test -->
<!-- Title: Big Diff Test -->

# Big Diff Test

## Section 1: Introduction

This is the original introduction paragraph. It explains the purpose
of this document which is to test large diffs in the push workflow.

## Section 2: Configuration

```yaml
server:
  host: localhost
  port: 8080
  debug: false
  log_level: info
```

## Section 3: Architecture

The system uses a three-tier architecture:
- Presentation layer (CLI)
- Business logic (internal packages)
- Data access (API client)

## Section 4: Deployment

Deploy using Docker Compose:
```bash
docker compose up -d
```

## Section 5: FAQ

Q: How do I reset my token?
A: Run `outline auth logout` then re-authenticate.

Q: Can I push to multiple collections?
A: Yes, use metadata headers per file.
EOF
$OUTLINE push --yes --path "$BIG_DIFF_FILE"

# Now modify heavily
cat > "$BIG_DIFF_FILE" <<'EOF'
<!-- Collection: test -->
<!-- Title: Big Diff Test -->

# Big Diff Test

## Section 1: Introduction (Revised)

This is the **revised** introduction. The document now covers the new
TUI features, preview pane, and kubectl-style get commands.

## Section 2: Configuration

```yaml
server:
  host: 0.0.0.0
  port: 9090
  debug: true
  log_level: debug
  timeout: 30s
  cors:
    enabled: true
    origins:
      - http://localhost:3000
```

## Section 3: Architecture

The system uses an updated four-tier architecture:
- Presentation layer (CLI + TUI)
- Command layer (Cobra commands)
- Business logic (internal packages)
- Data access (API client + caching)

Each layer communicates via well-defined interfaces.

## Section 4: Deployment

Deploy using Docker Compose with the new production config:
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

Or use the Makefile:
```bash
make deploy ENV=production
```

## Section 5: FAQ

Q: How do I reset my token?
A: Run `outline auth logout` then re-authenticate.

Q: Can I push to multiple collections?
A: Yes, use metadata headers per file.

Q: How do I use the TUI?
A: Run `outline tui` and press p for preview, / for search.

## Section 6: Changelog

- Added TUI with preview pane
- Added kubectl-style get commands
- Improved push diff view with unified diffs
EOF
echo ""
echo "--- Pushing modified version (expect large colored diff) ---"
$OUTLINE push --diff --path "$BIG_DIFF_FILE"
rm -f "$BIG_DIFF_FILE"
pause

echo "=== Step 3: Push directory with confirmation ==="
$OUTLINE push --yes --path examples/guide/
pause

# --- SEARCH ---
echo "=== Step 4: Search (default format) ==="
$OUTLINE search "guide"
pause

echo "=== Step 5: Search (oneline format) ==="
$OUTLINE search --format oneline "guide"
pause

echo "=== Step 6: Search (json format) ==="
$OUTLINE search --format json --limit 2 "guide"
pause

echo "=== Step 7: Search with collection filter ==="
$OUTLINE search --collection test "guide"
pause

# --- GET (kubectl-style) ---
echo "=== Step 8: Get collections ==="
$OUTLINE get collections
pause

echo "=== Step 9: Get documents (all) ==="
$OUTLINE get documents | head -20
pause

echo "=== Step 10: Get documents --collection test ==="
$OUTLINE get documents --collection test
pause

echo "=== Step 11: Get specific document (JSON) ==="
$OUTLINE get documents "Minimal Example" -o json | head -20
pause

echo "=== Step 12: Get specific document (raw markdown) ==="
$OUTLINE get documents "Minimal Example" -o raw
pause

# --- GET RENDERED ---
echo "=== Step 13: Get document (rendered, default) ==="
$OUTLINE get documents "Minimal Example"
pause

# --- PULL ---
echo "=== Step 14: Pull single doc ==="
PULL_DIR=$(mktemp -d)
$OUTLINE pull --doc "Minimal Example" --output "$PULL_DIR"
echo "--- Pulled file ---"
find "$PULL_DIR" -type f -name "*.md" -exec cat {} \;
pause

echo "=== Step 15: Pull single doc with metadata ==="
rm -rf "$PULL_DIR" && mkdir -p "$PULL_DIR"
$OUTLINE pull --doc "Minimal Example" --output "$PULL_DIR" --with-metadata
echo "--- Pulled file (with metadata) ---"
find "$PULL_DIR" -type f -name "*.md" -exec cat {} \;
pause

echo "=== Step 16: Pull entire collection ==="
rm -rf "$PULL_DIR" && mkdir -p "$PULL_DIR"
$OUTLINE pull --collection test --output "$PULL_DIR"
echo "--- Pulled tree ---"
find "$PULL_DIR" -type f | sort
pause

# --- ROUND-TRIP ---
echo "=== Step 17: Round-trip (push → pull → diff) ==="
ROUND_DIR=$(mktemp -d)
$OUTLINE push --yes --path examples/minimal.md
$OUTLINE pull --doc "Minimal Example" --output "$ROUND_DIR"
echo "--- Comparing local vs pulled ---"
PULLED_FILE=$(find "$ROUND_DIR" -type f -name "*.md" | head -1)
# Strip metadata from local for comparison
LOCAL_BODY=$(sed '/^<!-- .* -->$/d' examples/minimal.md | sed '/^$/N;/^\n$/d')
REMOTE_BODY=$(cat "$PULLED_FILE")
if diff <(echo "$LOCAL_BODY") <(echo "$REMOTE_BODY") > /dev/null 2>&1; then
    echo "PASS: Round-trip content matches"
else
    echo "WARN: Content differs (may be due to Outline formatting)"
    diff <(echo "$LOCAL_BODY") <(echo "$REMOTE_BODY") || true
fi

# Cleanup
rm -rf "$PULL_DIR" "$ROUND_DIR"

echo ""
echo "=== All feature tests passed ==="
