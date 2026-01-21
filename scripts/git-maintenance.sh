#!/bin/bash
# Git maintenance script to prevent and fix common issues

echo "=== Git Maintenance Script ==="
echo "Running maintenance tasks..."

# 1. Remove corrupt references with spaces
echo -n "Checking for corrupt references... "
CORRUPT_REFS=$(find .git/refs -name '*[[:space:]]*' 2>/dev/null)
if [ ! -z "$CORRUPT_REFS" ]; then
    echo "Found corrupt references:"
    echo "$CORRUPT_REFS"
    find .git/refs -name '*[[:space:]]*' -delete
    echo "Corrupt references removed."
else
    echo "OK"
fi

# 2. Clean up git repository
echo -n "Cleaning up git repository... "
git gc --auto --quiet
git prune
echo "OK"

# 3. Check for staging issues
STAGED_DELETIONS=$(git diff --cached --name-status | grep -c "^D" || echo "0")
if [ "$STAGED_DELETIONS" -gt 0 ]; then
    echo "WARNING: Found $STAGED_DELETIONS staged deletions"
    echo "Run 'git status' to review and 'git reset HEAD' if unintended"
fi

# 4. Verify critical files exist
echo -n "Checking critical files... "
CRITICAL_FILES=(
    "Dockerfile"
    "server/cmd/server/main.go"
    "server/go.mod"
    "web/package.json"
    "web/app/page.js"
    "railway.json"
    ".gitignore"
)

MISSING_COUNT=0
for file in "${CRITICAL_FILES[@]}"; do
    if [ ! -f "$file" ]; then
        echo ""
        echo "  WARNING: Missing $file"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done

if [ "$MISSING_COUNT" -eq 0 ]; then
    echo "OK (all critical files present)"
else
    echo ""
    echo "Found $MISSING_COUNT missing critical files"
    echo "This might affect deployments. Consider restoring from git."
fi

# 5. Check remote tracking
echo -n "Checking remote tracking... "
REMOTE_BRANCH=$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null)
if [ -z "$REMOTE_BRANCH" ]; then
    echo "WARNING: Current branch not tracking remote"
    echo "Set upstream with: git push -u origin $(git branch --show-current)"
else
    echo "OK (tracking $REMOTE_BRANCH)"
fi

# 6. Repository size check
GIT_SIZE=$(du -sm .git | cut -f1)
echo "Repository size: ${GIT_SIZE}MB"
if [ "$GIT_SIZE" -gt 500 ]; then
    echo "  Repository is large. Consider running: git gc --aggressive"
fi

echo ""
echo "=== Maintenance Complete ==="
echo ""
echo "Tips to prevent git issues:"
echo "  • Always use 'git status' before committing"
echo "  • Avoid force operations unless necessary"
echo "  • Keep .gitignore up to date"
echo "  • Run this script weekly: ./scripts/git-maintenance.sh"