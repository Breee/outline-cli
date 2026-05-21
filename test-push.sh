#!/usr/bin/env bash
set -euo pipefail

SERVER_URL="${OUTLINE_URL:-http://outline.localhost}"
OUTLINE="./outline --server-url $SERVER_URL"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== Building ==="
make build

echo ""
echo "=== Step 0: Verify auth (auto-authenticates if needed) ==="
$OUTLINE auth check || $OUTLINE auth oidc-login

echo ""
echo "=== Step 1: Push single file (expect 'created' or 'updated') ==="
$OUTLINE push --path examples/minimal.md

echo ""
echo "=== Step 2: Push same file again (expect 'updated') ==="
$OUTLINE push --path examples/minimal.md

echo ""
echo "=== Step 3: Push markdown-reference (has images) ==="
$OUTLINE push --path examples/markdown-reference.md

echo ""
echo "=== Step 4: Push entire examples directory ==="
$OUTLINE push --collection-id test --path examples/

echo ""
echo "=== Step 5: Set up subdoc hierarchy ==="
mkdir -p examples/guide
cat > examples/guide/index.md << 'EOF'
<!-- Collection: test -->
<!-- Title: User Guide -->

# User Guide

This is the top-level guide document.
EOF

cat > examples/guide/getting-started.md << 'EOF'
<!-- Collection: test -->
<!-- Title: Getting Started -->

# Getting Started

This is a child document of the User Guide.
EOF

echo "--- Push guide directory (parent + child) ---"
$OUTLINE push --path examples/guide/

echo ""
echo "=== Step 6: Push guide again (expect both 'updated') ==="
$OUTLINE push --path examples/guide/

echo ""
echo "=== Step 7: Explicit parent header ==="
cat > /tmp/explicit-parent.md << 'EOF'
<!-- Collection: test -->
<!-- Title: Advanced Topics -->
<!-- Parent: User Guide -->

# Advanced Topics

This doc is explicitly parented under "User Guide".
EOF

$OUTLINE push --path /tmp/explicit-parent.md

echo ""
echo "=== All tests passed ==="
