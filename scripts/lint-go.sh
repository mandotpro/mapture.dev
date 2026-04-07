#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"
GOLANGCI_LINT_BIN="$(ensure_golangci_lint)"

cd "$ROOT_DIR"

"$GOLANGCI_LINT_BIN" run ./src/...
