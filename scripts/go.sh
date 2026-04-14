#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"
TESTING_DIR="$(testing_dir)"
TESTING_EXAMPLES_DIR="$(testing_examples_dir)"
PLAYGROUND_DIR="$(playground_dir)"
OUTPUTS_DIR="$(outputs_dir)"
BIN_DIR="$TESTING_DIR/bin"
BIN="$BIN_DIR/mapture"

mkdir -p "$PLAYGROUND_DIR" "$BIN_DIR" "$OUTPUTS_DIR" "$TESTING_EXAMPLES_DIR"

discover_example_fixtures() {
  local path fixture
  while IFS= read -r path; do
    fixture="$(basename "$(dirname "$path")")"
    printf '%s\n' "$fixture"
  done < <(find "$ROOT_DIR/examples" -mindepth 2 -maxdepth 2 -name mapture.yaml | sort)
}

fixture_names() {
  discover_example_fixtures
  printf '%s\n' "playground"
}

example_fixture_names() {
  discover_example_fixtures
}

fixture_path() {
  local fixture="$1"
  local example_dir="$ROOT_DIR/examples/$fixture"

  if [[ "$fixture" == "playground" ]]; then
    printf '%s\n' "$PLAYGROUND_DIR"
    return
  fi

  if [[ -f "$example_dir/mapture.yaml" ]]; then
    printf '%s\n' "$example_dir"
    return
  fi

  printf '%s\n' "$(mapture_error "unknown fixture: $fixture")" >&2
  printf '%s\n' "$(mapture_muted "known fixtures: $(fixture_names | tr '\n' ' ' | sed 's/ $//')")" >&2
  exit 1
}

testing_fixture_path() {
  local fixture="$1"

  if [[ "$fixture" == "playground" ]]; then
    printf '%s\n' "$PLAYGROUND_DIR"
    return
  fi

  printf '%s/%s\n' "$TESTING_EXAMPLES_DIR" "$fixture"
}

fixture_output() {
  local fixture="$1"
  local kind="$2"
  printf '%s/%s.%s\n' "$OUTPUTS_DIR" "$fixture" "$kind"
}

ensure_testing_binary() {
  local command="${1:-}"
  if [[ "$command" == "serve" ]]; then
    rebuild_embedded_web_bundle
  else
    ensure_embedded_web_bundle
  fi
  build_binary "$BIN"
}

sync_example_fixtures() {
  local requested="${1:-all}"
  local fixture source target

  mkdir -p "$TESTING_EXAMPLES_DIR"

  while IFS= read -r fixture; do
    if [[ "$requested" != "all" && "$fixture" != "$requested" ]]; then
      continue
    fi
    source="$ROOT_DIR/examples/$fixture"
    target="$TESTING_EXAMPLES_DIR/$fixture"
    rm -rf "$target"
    mkdir -p "$(dirname "$target")"
    cp -R "$source" "$target"
    printf 'copied %s -> %s\n' "$source" "$target"
  done < <(example_fixture_names)
}

run_for_all_examples() {
  local command="$1"
  shift || true

  case "$command" in
    validate|scan|export-json-graph|export-json-visualisation)
      ;;
    *)
      printf '%s\n' "$(mapture_error "unsupported all-fixtures command: $command")" >&2
      printf '%s\n' "$(mapture_muted "supported commands: validate scan export-json-graph export-json-visualisation")" >&2
      exit 1
      ;;
  esac

  ensure_testing_binary "$command"

  local fixture target output
  while IFS= read -r fixture; do
    printf '%s\n' "$(mapture_strong "== $fixture ($command) ==")"
    case "$command" in
      validate)
        "$BIN" validate "$(fixture_path "$fixture")" "$@"
        ;;
      scan)
        target="$(fixture_path "$fixture")"
        output="$(fixture_output "$fixture" "scan.json")"
        "$BIN" scan "$target" >"$output"
        printf '%s\n' "$(mapture_success "wrote $(mapture_muted "$output")")"
        ;;
      export-json-graph)
        target="$(fixture_path "$fixture")"
        output="$(fixture_output "$fixture" "graph.json")"
        "$BIN" export-json-graph "$target" -o "$output"
        printf '%s\n' "$(mapture_success "wrote $(mapture_muted "$output")")"
        ;;
      export-json-visualisation)
        target="$(fixture_path "$fixture")"
        output="$(fixture_output "$fixture" "visualisation.json")"
        "$BIN" export-json-visualisation "$target" -o "$output"
        printf '%s\n' "$(mapture_success "wrote $(mapture_muted "$output")")"
        ;;
    esac
  done < <(example_fixture_names)
}

run_across_examples_if_needed() {
  local command="$1"
  local fixture="$2"
  shift 2 || true

  if [[ "$fixture" != "all" ]]; then
    return 1
  fi

  run_for_all_examples "$command" "$@"
  return 0
}

serve_port() {
  case "$1" in
    demo) printf '%s\n' "127.0.0.1:8766" ;;
    ecommerce) printf '%s\n' "127.0.0.1:8765" ;;
    migration) printf '%s\n' "127.0.0.1:8768" ;;
    playground) printf '%s\n' "127.0.0.1:8767" ;;
    *)
      printf '%s\n' "$(mapture_error "unknown fixture: $1")" >&2
      exit 1
      ;;
  esac
}

show_help() {
  local fixture
  mapture_print_section "Repo Development Commands"
  mapture_print_kv "./scripts/go.sh build" "Build the latest local app into $(mapture_muted "$BIN")"
  mapture_print_kv "./scripts/go.sh init" "Run init against the testing playground"

  mapture_print_section "Repo Verification Commands"
  mapture_print_kv "./scripts/go.sh validate <fixture|all>" "Validate one fixture or every example fixture"
  mapture_print_kv "./scripts/go.sh scan <fixture|all>" "Write normalized scan output into $(mapture_muted "$OUTPUTS_DIR")"
  mapture_print_kv "./scripts/go.sh export-json-graph <fixture|all>" "Write JGF output into $(mapture_muted "$OUTPUTS_DIR")"
  mapture_print_kv "./scripts/go.sh export-json-visualisation <fixture|all>" "Write explorer JSON into $(mapture_muted "$OUTPUTS_DIR")"

  mapture_print_section "Local Verification With Fixtures"
  mapture_print_kv "./scripts/go.sh fixtures" "List known fixtures"
  mapture_print_kv "./scripts/go.sh serve <fixture>" "Run the latest local explorer for a fixture"
  mapture_print_kv "./scripts/go.sh fixture <fixture> <command>" "Run any CLI command against a fixture path"

  mapture_print_section "Paths"
  printf '  %s -> %s\n' "$(mapture_accent "built binary")" "$(mapture_muted "$BIN")"
  printf '  %s -> %s\n' "$(mapture_accent "testing root")" "$(mapture_muted "$TESTING_DIR")"
  printf '  %s -> %s\n' "$(mapture_accent "playground")" "$(mapture_muted "$PLAYGROUND_DIR")"
  printf '  %s -> %s\n' "$(mapture_accent "outputs")" "$(mapture_muted "$OUTPUTS_DIR")"
  printf '  %s -> %s\n' "$(mapture_accent "testing examples")" "$(mapture_muted "$TESTING_EXAMPLES_DIR")"

  mapture_print_section "Fixtures"
  while IFS= read -r fixture; do
    printf '  %s -> %s\n' "$(mapture_accent "$(printf '%-10s' "$fixture")")" "$(fixture_path "$fixture")"
  done < <(fixture_names)
}

run_validate() {
  local fixture="$1"
  if run_across_examples_if_needed validate "$fixture"; then
    return
  fi
  ensure_testing_binary validate
  exec "$BIN" validate "$(fixture_path "$fixture")"
}

run_scan() {
  local fixture="$1"
  if run_across_examples_if_needed scan "$fixture"; then
    return
  fi
  ensure_testing_binary scan
  local target output
  target="$(fixture_path "$fixture")"
  output="$(fixture_output "$fixture" "scan.json")"
  "$BIN" scan "$target" >"$output"
  printf '%s\n' "$(mapture_success "wrote $(mapture_muted "$output")")"
}

run_export_json_graph() {
  local fixture="$1"
  if run_across_examples_if_needed export-json-graph "$fixture"; then
    return
  fi
  ensure_testing_binary export-json-graph
  local target output
  target="$(fixture_path "$fixture")"
  output="$(fixture_output "$fixture" "graph.json")"
  "$BIN" export-json-graph "$target" -o "$output"
  printf '%s\n' "$(mapture_success "wrote $(mapture_muted "$output")")"
}

run_export_json_visualisation() {
  local fixture="$1"
  if run_across_examples_if_needed export-json-visualisation "$fixture"; then
    return
  fi
  ensure_testing_binary export-json-visualisation
  local target output
  target="$(fixture_path "$fixture")"
  output="$(fixture_output "$fixture" "visualisation.json")"
  "$BIN" export-json-visualisation "$target" -o "$output"
  printf '%s\n' "$(mapture_success "wrote $(mapture_muted "$output")")"
}

run_serve() {
  local fixture="$1"
  if [[ "$fixture" == "all" ]]; then
    printf '%s\n' "$(mapture_error 'serve does not support fixture "all"; choose one fixture')" >&2
    exit 1
  fi
  ensure_testing_binary serve
  local target addr
  target="$(fixture_path "$fixture")"
  addr="$(serve_port "$fixture")"
  print_binary_summary "local testing binary" "$BIN"
  printf '  %s %s\n' "$(mapture_accent "fixture")" "$(mapture_muted "$target")"
  printf '  %s %s\n' "$(mapture_accent "serve")" "$(mapture_muted "http://$addr")"
  exec "$BIN" serve "$target" --addr "$addr" --open
}

if [[ $# -eq 0 ]]; then
  show_help
  exit 0
fi

case "$1" in
  help|-h|--help)
    show_help
    ;;
  fixtures)
    fixture_names
    ;;
  path)
    shift
    fixture_path "${1:-demo}"
    ;;
  build)
    mapture_print_section "Testing Build"
    ensure_testing_binary serve
    print_binary_summary "testing build" "$BIN"
    printf '%s\n' "$(mapture_success "build complete")"
    ;;
  init)
    shift
    ensure_testing_binary init
    exec "$BIN" init "${1:-$PLAYGROUND_DIR}"
    ;;
  validate)
    shift
    run_validate "${1:-demo}"
    ;;
  scan)
    shift
    run_scan "${1:-demo}"
    ;;
  export-json-graph)
    shift
    run_export_json_graph "${1:-demo}"
    ;;
  export-json-visualisation)
    shift
    run_export_json_visualisation "${1:-demo}"
    ;;
  serve)
    shift
    run_serve "${1:-ecommerce}"
    ;;
  fixture)
    shift
    fixture="${1:-}"
    command="${2:-validate}"
    if [[ -z "$fixture" ]]; then
      printf '%s\n' "$(mapture_error "usage: ./scripts/go.sh fixture <fixture> [command]")" >&2
      exit 1
    fi
    shift
    shift || true
    ensure_testing_binary "$command"
    exec "$BIN" "$command" "$(fixture_path "$fixture")" "$@"
    ;;
  run)
    shift
    ensure_testing_binary "${1:-}"
    exec "$BIN" "$@"
    ;;
  demo|ecommerce|migration|playground)
    fixture="$1"
    shift
    if [[ $# -eq 0 ]]; then
      set -- validate
    fi
    ensure_testing_binary "${1:-validate}"
    exec "$BIN" "$1" "$(fixture_path "$fixture")" "${@:2}"
    ;;
  *)
    exec "$BIN" "$@"
    ;;
esac
