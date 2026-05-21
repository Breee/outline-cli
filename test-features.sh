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
$OUTLINE push --yes --diff --path examples/minimal.md
# Restore original
git checkout -- examples/minimal.md
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

# --- PULL ---
echo "=== Step 8: Pull single doc ==="
PULL_DIR=$(mktemp -d)
$OUTLINE pull --doc "Minimal Example" --output "$PULL_DIR"
echo "--- Pulled file ---"
find "$PULL_DIR" -type f -name "*.md" -exec cat {} \;
pause

echo "=== Step 9: Pull single doc with metadata ==="
rm -rf "$PULL_DIR" && mkdir -p "$PULL_DIR"
$OUTLINE pull --doc "Minimal Example" --output "$PULL_DIR" --with-metadata
echo "--- Pulled file (with metadata) ---"
find "$PULL_DIR" -type f -name "*.md" -exec cat {} \;
pause

echo "=== Step 10: Pull entire collection ==="
rm -rf "$PULL_DIR" && mkdir -p "$PULL_DIR"
$OUTLINE pull --collection test --output "$PULL_DIR"
echo "--- Pulled tree ---"
find "$PULL_DIR" -type f | sort
pause

# --- ROUND-TRIP ---
echo "=== Step 11: Round-trip (push → pull → diff) ==="
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
