#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

status=0

check_tracked_paths() {
  local label="$1"
  shift
  local matches
  matches="$(
    while IFS= read -r path; do
      if [[ -e "$path" ]]; then
        printf '%s\n' "$path"
      fi
    done < <(git ls-files "$@" || true)
  )"
  if [[ -n "$matches" ]]; then
    echo "public-audit: tracked files found in $label:" >&2
    printf '%s\n' "$matches" >&2
    status=1
  fi
}

check_grep() {
  local label="$1"
  local pattern="$2"
  local matches
  matches="$(git grep -nE "$pattern" -- . ':!src/internal/webui/dist/*' || true)"
  if [[ -n "$matches" ]]; then
    echo "public-audit: potential $label found:" >&2
    printf '%s\n' "$matches" >&2
    status=1
  fi
}

check_tracked_paths "local build/test outputs" 'build/*' 'testing/*'
check_tracked_paths "local tool settings" '.agents/*' '.claude/*' '.codex/*' '.vscode/*' '.idea/*'
check_tracked_paths "temporary generators" 'src/tmpgen/*'

check_grep "absolute local filesystem paths" '/Users/[^/]+/|/home/[^/]+/|C:\\Users\\'
check_grep "high-risk secret patterns" 'ghp_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|AKIA[0-9A-Z]{16}|-----BEGIN (RSA|OPENSSH|EC|DSA|PRIVATE) KEY-----'

echo "public-audit: manual review required for example data, screenshots, and release notes." >&2

exit "$status"
