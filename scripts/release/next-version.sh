#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $0 <patch|minor> [<release-branch>]" >&2
  exit 1
fi

BUMP="$1"
RELEASE_BRANCH="${2:-}"
INITIAL_VERSION="${INITIAL_VERSION:-}"

release_major=""
if [[ -n "$RELEASE_BRANCH" ]]; then
  if [[ ! "$RELEASE_BRANCH" =~ ^([0-9]+)\.x$ ]]; then
    echo "release branch must look like <major>.x, got: $RELEASE_BRANCH" >&2
    exit 1
  fi
  release_major="${BASH_REMATCH[1]}"
fi

tag_glob='v[0-9]*.[0-9]*.[0-9]*'
if [[ -n "$release_major" ]]; then
  tag_glob="v${release_major}.*.*"
fi

latest_tag="$(git tag --list "$tag_glob" --sort=-version:refname | head -n 1)"
if [[ -z "$latest_tag" ]]; then
  if [[ -n "$INITIAL_VERSION" ]]; then
    printf '%s\n' "$INITIAL_VERSION"
    exit 0
  fi
  if [[ -n "$release_major" ]]; then
    if [[ "$release_major" == "0" ]]; then
      printf 'v0.1.0\n'
    else
      printf 'v%s.0.0\n' "$release_major"
    fi
    exit 0
  fi
  printf 'v0.1.0\n'
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
  *)
    echo "unknown bump type: $BUMP" >&2
    exit 1
    ;;
esac

if [[ -n "$release_major" && "$major" != "$release_major" ]]; then
  echo "latest tag $latest_tag does not belong to release branch $RELEASE_BRANCH" >&2
  exit 1
fi

printf 'v%s.%s.%s\n' "$major" "$minor" "$patch"
