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
go run src/main.go graph examples/demo >/dev/null

graph_output="$(mktemp)"
release_output_dir="$(mktemp -d)"
formula_output="$(mktemp)"
tap_output_dir="$(mktemp -d)"
trap 'rm -f "$graph_output" "$formula_output"; rm -rf "$release_output_dir" "$tap_output_dir"' EXIT
go run src/main.go graph examples/ecommerce --domain billing -o "$graph_output"
test -s "$graph_output"

./scripts/build.sh >/dev/null
./build/mapture --version | grep -q "0.0.0-dev"

./scripts/release-build.sh "v0.0.0-test" "linux" "amd64" "$release_output_dir" >/dev/null
test -f "$release_output_dir/mapture_v0.0.0-test_linux_amd64.tar.gz"
./scripts/release-build.sh "v0.0.0-test" "windows" "amd64" "$release_output_dir" >/dev/null
test -f "$release_output_dir/mapture_v0.0.0-test_windows_amd64.zip"

./scripts/generate-homebrew-formula.sh \
  --formula-name mapture-canary \
  --class-name MaptureCanary \
  --formula-version 0.0.0-canary.20260409153413.c649dd6 \
  --binary-version 0.0.0-canary.20260409+sha.c649dd6 \
  --source-url https://github.com/mandotpro/mapture.dev/archive/c649dd651a22bb6c12509b13ef66d2a4dc4f552a.tar.gz \
  --source-sha256 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef \
  > "$formula_output"
grep -q 'class MaptureCanary < Formula' "$formula_output"
! grep -q 'conflicts_with' "$formula_output"
grep -q 'assert_match "mapture version 0.0.0-canary.20260409+sha.c649dd6"' "$formula_output"

./scripts/sync-homebrew-tap.sh \
  --tap-dir "$tap_output_dir" \
  --tap-repo mandotpro/homebrew-mapture \
  --formula-name mapture-canary \
  --class-name MaptureCanary \
  --formula-version 0.0.0-canary.20260409153413.c649dd6 \
  --binary-version 0.0.0-canary.20260409+sha.c649dd6 \
  --source-url https://github.com/mandotpro/mapture.dev/archive/c649dd651a22bb6c12509b13ef66d2a4dc4f552a.tar.gz \
  --source-sha256 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef >/dev/null
test -f "$tap_output_dir/Formula/mapture-canary.rb"
! grep -q 'conflicts_with' "$tap_output_dir/Formula/mapture-canary.rb"
grep -q 'brew install mandotpro/mapture/mapture-canary' "$tap_output_dir/README.md"

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
expect_failure examples/invalid/comment-unknown-domain-ref
expect_failure examples/invalid/comment-event-domain-mismatch
expect_failure examples/invalid/comment-unknown-node-target
expect_failure examples/invalid/comment-missing-owner
expect_failure examples/invalid/comment-bad-event-role
expect_failure examples/invalid/comment-unknown-key
expect_scan_failure examples/invalid/comment-missing-owner
expect_scan_failure examples/invalid/comment-bad-event-role
expect_scan_failure examples/invalid/comment-unknown-key
