#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"

cd "$ROOT_DIR"

./scripts/test-go.sh
go vet ./src/...
go run src/main.go --help >/dev/null
go run src/main.go validate examples/demo >/dev/null
go run src/main.go validate examples/ecommerce >/dev/null
go run src/main.go scan examples/demo >/dev/null
go run src/main.go scan examples/ecommerce >/dev/null

expect_failure() {
  local path="$1"
  if go run src/main.go validate "$path" >/dev/null 2>&1; then
    echo "expected validation to fail for $path" >&2
    exit 1
  fi
}

expect_scan_failure() {
  local path="$1"
  if go run src/main.go scan "$path" >/dev/null 2>&1; then
    echo "expected scan to fail for $path" >&2
    exit 1
  fi
}

expect_failure examples/invalid/bad-config-role
expect_failure examples/invalid/duplicate-team
expect_failure examples/invalid/unknown-domain-owner
expect_failure examples/invalid/invalid-event-status
expect_failure examples/invalid/missing-teams-file
expect_scan_failure examples/invalid/comment-missing-owner
expect_scan_failure examples/invalid/comment-bad-event-role
expect_scan_failure examples/invalid/comment-unknown-key
