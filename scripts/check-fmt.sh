#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"

cd "$ROOT_DIR"

mapfile -d '' go_files < <(find ./src -name '*.go' -print0)

if [[ ${#go_files[@]} -eq 0 ]]; then
  exit 0
fi

mapfile -t unformatted_files < <(gofmt -l "${go_files[@]}")

if [[ ${#unformatted_files[@]} -gt 0 ]]; then
  gofmt -w "${unformatted_files[@]}"
  echo "Go files under src/ were reformatted. Review and re-stage them before committing." >&2
  printf '%s\n' "${unformatted_files[@]}" >&2
  exit 1
fi
