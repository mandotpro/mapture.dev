#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <release-branch> <patch|minor>" >&2
  exit 1
fi

RELEASE_BRANCH="$1"
BUMP="$2"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

NEXT_VERSION="$(./scripts/release/next-version.sh "$BUMP" "$RELEASE_BRANCH")"

cat <<EOF
Release branch:  $RELEASE_BRANCH
Version bump:    $BUMP
Next version:    $NEXT_VERSION

Suggested release PR title:
  release: $BUMP

Explicit version PR title:
  release: $NEXT_VERSION

Manual GitHub fallback:
  Workflow: Stable Release
  Branch:   $RELEASE_BRANCH
  Bump:     $BUMP

Maintainer flow:
  1. Merge the intended changes into $RELEASE_BRANCH.
  2. Merge a PR into $RELEASE_BRANCH with one of the titles above.
  3. The merge creates tag $NEXT_VERSION.
  4. The Release workflow publishes the binaries and updates the Homebrew tap.
EOF
