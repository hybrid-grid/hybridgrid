#!/bin/bash
# scripts/changelog.sh - Generate changelog draft from git tags
# This script extracts commit messages between git tags and outputs a draft changelog.
# It requires manual review and editing before release.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

# Get list of tags
TAGS=$(cd "$REPO_ROOT" && git tag -l 'v0.*' --sort=-version:refname || echo "")

if [[ -z "$TAGS" ]]; then
  echo "No version tags found. Use 'git tag v0.2.0' to create initial release tags."
  exit 0
fi

echo "# Changelog Draft"
echo ""
echo "Auto-generated from git tags. Manual review and editing required before release."
echo ""

# Generate entries for each consecutive pair of tags
prev_tag=""
for tag in $TAGS; do
  if [[ -n "$prev_tag" ]]; then
    echo "## Changes between $prev_tag and $tag"
    echo ""
    cd "$REPO_ROOT" && git log --oneline "$tag..$prev_tag" | sed 's/^/- /' || true
    echo ""
  fi
  prev_tag="$tag"
done

# Show current commits since last tag
if [[ -n "$prev_tag" ]]; then
  echo "## Unreleased (since $prev_tag)"
  echo ""
  cd "$REPO_ROOT" && git log --oneline "$prev_tag..HEAD" | sed 's/^/- /' || true
  echo ""
fi

exit 0
