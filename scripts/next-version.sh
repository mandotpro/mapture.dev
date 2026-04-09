#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <patch|minor|major>" >&2
  exit 1
fi

BUMP="$1"
INITIAL_VERSION="${INITIAL_VERSION:-v0.1.0}"

latest_tag="$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-version:refname | head -n 1)"
if [[ -z "$latest_tag" ]]; then
  printf '%s\n' "$INITIAL_VERSION"
  exit 0
fi

version_core="${latest_tag#v}"
IFS='.' read -r major minor patch <<<"$version_core"

case "$BUMP" in
  patch)
    patch=$((patch + 1))
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  *)
    echo "unknown bump type: $BUMP" >&2
    exit 1
    ;;
esac

printf 'v%s.%s.%s\n' "$major" "$minor" "$patch"
