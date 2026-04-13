#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"

discover_example_fixtures() {
  find "$ROOT_DIR/examples" -mindepth 2 -maxdepth 2 -name mapture.yaml -exec dirname {} \; | xargs -n1 basename | sort
}

print_help() {
  mapture_print_section "Repo Development Commands"
  mapture_print_kv "install-dev-tools" "Install local Go dev tools into testing/tools/bin"
  mapture_print_kv "install-git-hooks" "Configure git to use the repo-managed hooks"
  mapture_print_kv "init-hooks" "Configure git to use the repo-managed hooks"
  mapture_print_kv "build" "Build the local mapture binary into build/"
  mapture_print_kv "web" "Rebuild the frontend bundle under src/internal/webui/dist/"

  mapture_print_section "Repo Verification Commands"
  mapture_print_kv "test-go" "Run Go tests through gotestsum"
  mapture_print_kv "test" "Run the full local verification suite"
  mapture_print_kv "lint" "Run golangci-lint against src/"
  mapture_print_kv "vet" "Run go vet against src/"
  mapture_print_kv "fmt" "Format Go source files under src/"
  mapture_print_kv "audit-public" "Run public-release hygiene checks against tracked files"
  mapture_print_kv "cli-help" "Show CLI help from the current source tree"

  mapture_print_section "Local Verification With Fixtures"
  mapture_print_kv "fixtures" "List discovered fixtures"
  mapture_print_kv "testing-help" "Show the testing-first wrapper commands and fixture paths"
  mapture_print_kv "testing-build" "Build the current source into testing/bin/mapture"
  mapture_print_kv "testing-init" "Run init against testing/playground"
  mapture_print_kv "playground-init" "Run init against the gitignored testing playground"
  mapture_print_kv "validate" "Validate a fixture: make validate FIXTURE=<fixture|all>"
  mapture_print_kv "scan" "Scan a fixture: make scan FIXTURE=<fixture|all>"
  mapture_print_kv "graph" "Export Mermaid for a fixture: make graph FIXTURE=<fixture|all>"
  mapture_print_kv "serve" "Run the local server against a fixture: make serve FIXTURE=<fixture>"
  mapture_print_kv "run" "Run any CLI command for a fixture: make run FIXTURE=<fixture> CMD=<cli-command>"

  mapture_print_section "Fixtures"
  while IFS= read -r fixture; do
    printf '  %s\n' "$fixture"
  done < <(discover_example_fixtures)
  printf '  playground\n'

  mapture_print_section "Fixture Aliases"
  while IFS= read -r fixture; do
    printf '  validate.%s  scan.%s  graph.%s  serve.%s\n' "$fixture" "$fixture" "$fixture" "$fixture"
  done < <(discover_example_fixtures)
  printf '  validate.playground  scan.playground  graph.playground  serve.playground\n'
}

print_help
