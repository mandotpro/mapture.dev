#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"

cd "$ROOT_DIR"

mapfile -t packages < <(mapture_go_packages)

mapture_print_section "Go and CLI checks"
./scripts/test-go.sh
go vet "${packages[@]}"
bash -n ./scripts/install.sh
bash -n ./scripts/help.sh
go run src/main.go --help >/dev/null
go run src/main.go validate examples/demo >/dev/null
go run src/main.go validate examples/ecommerce >/dev/null
go run src/main.go scan examples/demo >/dev/null
go run src/main.go scan examples/ecommerce >/dev/null
go run src/main.go export-json-graph examples/demo >/dev/null

graph_output="$(mktemp)"
visualisation_output="$(mktemp)"
release_output_dir="$(mktemp -d)"
formula_output="$(mktemp)"
tap_output_dir="$(mktemp -d)"
install_output_dir="$(mktemp -d)"
help_plain_output="$(mktemp)"
help_color_output="$(mktemp)"
go_help_color_output="$(mktemp)"
build_output="$(mktemp)"
testing_build_output="$(mktemp)"
stale_build_output="$(mktemp)"
fresh_build_output="$(mktemp)"
web_probe="$ROOT_DIR/src/internal/webui/frontend/src/.mapture-web-stale-probe"
trap 'rm -f "$graph_output" "$visualisation_output" "$formula_output" "$help_plain_output" "$help_color_output" "$go_help_color_output" "$build_output" "$testing_build_output" "$stale_build_output" "$fresh_build_output" "$web_probe"; rm -rf "$release_output_dir" "$tap_output_dir" "$install_output_dir"' EXIT
go run src/main.go export-json-graph examples/ecommerce -o "$graph_output"
test -s "$graph_output"
go run src/main.go export-json-visualisation examples/ecommerce -o "$visualisation_output"
test -s "$visualisation_output"

mapture_print_section "Release helper checks"
./scripts/build.sh >"$build_output"
grep -q 'embedded web bundle' "$build_output"
grep -q 'build/mapture' "$build_output"
grep -q 'mapture.dev - 0.0.0-dev' "$build_output"
build_version="$(./build/mapture --version)"
[[ "$build_version" == *"0.0.0-dev"* ]]
./scripts/go.sh build >"$testing_build_output"
grep -q 'embedded web bundle' "$testing_build_output"
grep -q 'testing/bin/mapture' "$testing_build_output"
grep -q 'mapture.dev - 0.0.0-dev' "$testing_build_output"

touch "$web_probe"
./scripts/build.sh >"$stale_build_output"
grep -q 'embedded web bundle is stale; rebuilding' "$stale_build_output"
rm -f "$web_probe"
./scripts/build.sh >"$fresh_build_output"
grep -q 'embedded web bundle is current' "$fresh_build_output"
GOBIN="$install_output_dir" go install ./cmd/mapture
test -x "$install_output_dir/mapture"

./scripts/release/release-build.sh "v0.0.0-test" "linux" "amd64" "$release_output_dir" >/dev/null
test -f "$release_output_dir/mapture_v0.0.0-test_linux_amd64.tar.gz"
./scripts/release/release-build.sh "v0.0.0-test" "windows" "amd64" "$release_output_dir" >/dev/null
test -f "$release_output_dir/mapture_v0.0.0-test_windows_amd64.zip"
[[ "$(./scripts/release/next-version.sh patch 0.x)" == v0.* ]]
[[ "$(./scripts/release/next-version.sh minor 1.x)" == v1.* ]]
./scripts/release/plan-release.sh 0.x patch >/dev/null

./scripts/release/generate-homebrew-formula.sh \
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

./scripts/release/sync-homebrew-tap.sh \
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

mapture_print_section "Script color policy"
NO_COLOR=1 make help >"$help_plain_output"
! grep -q $'\033' "$help_plain_output"
MAPTURE_COLOR=always make help >"$help_color_output"
grep -q $'\033' "$help_color_output"
MAPTURE_COLOR=always ./scripts/go.sh help >"$go_help_color_output"
grep -q $'\033' "$go_help_color_output"
! grep -q 'make graph' "$help_plain_output"
! grep -q './scripts/go.sh graph' "$go_help_color_output"

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
