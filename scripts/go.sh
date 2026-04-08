#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"
TESTING_DIR="$(testing_dir)"
PLAYGROUND_DIR="$(playground_dir)"
OUTPUTS_DIR="$(outputs_dir)"
BIN_DIR="$TESTING_DIR/bin"
BIN="$BIN_DIR/mapture"
DEMO_DIR="$ROOT_DIR/examples/demo"
ECOMMERCE_DIR="$ROOT_DIR/examples/ecommerce"

mkdir -p "$PLAYGROUND_DIR" "$BIN_DIR" "$OUTPUTS_DIR"
build_binary "$BIN"

fixture_path() {
  case "$1" in
    demo) printf '%s\n' "$DEMO_DIR" ;;
    ecommerce) printf '%s\n' "$ECOMMERCE_DIR" ;;
    playground) printf '%s\n' "$PLAYGROUND_DIR" ;;
    *)
      echo "unknown fixture: $1" >&2
      exit 1
      ;;
  esac
}

fixture_output() {
  local fixture="$1"
  local kind="$2"
  printf '%s/%s.%s\n' "$OUTPUTS_DIR" "$fixture" "$kind"
}

serve_port() {
  case "$1" in
    demo) printf '%s\n' "127.0.0.1:8766" ;;
    ecommerce) printf '%s\n' "127.0.0.1:8765" ;;
    playground) printf '%s\n' "127.0.0.1:8767" ;;
    *)
      echo "unknown fixture: $1" >&2
      exit 1
      ;;
  esac
}

show_help() {
  cat <<EOF
Built: $BIN
Testing root: $TESTING_DIR
Playground: $PLAYGROUND_DIR
Outputs: $OUTPUTS_DIR

Fixtures:
  demo       -> $DEMO_DIR
  ecommerce  -> $ECOMMERCE_DIR
  playground -> $PLAYGROUND_DIR

Common commands:
  ./scripts/go.sh build
  ./scripts/go.sh init
  ./scripts/go.sh validate demo
  ./scripts/go.sh scan ecommerce
  ./scripts/go.sh graph demo
  ./scripts/go.sh web ecommerce
  ./scripts/go.sh web playground

Artifact commands:
  ./scripts/go.sh scan <fixture>      # writes testing/outputs/<fixture>.scan.json
  ./scripts/go.sh graph <fixture>     # writes testing/outputs/<fixture>.mmd

Direct execution:
  ./scripts/go.sh run validate "$DEMO_DIR"
  ./scripts/go.sh fixture ecommerce validate
EOF
}

run_validate() {
  local fixture="$1"
  exec "$BIN" validate "$(fixture_path "$fixture")"
}

run_scan() {
  local fixture="$1"
  local target output
  target="$(fixture_path "$fixture")"
  output="$(fixture_output "$fixture" "scan.json")"
  "$BIN" scan "$target" >"$output"
  echo "wrote $output"
}

run_graph() {
  local fixture="$1"
  local target output
  target="$(fixture_path "$fixture")"
  output="$(fixture_output "$fixture" "mmd")"
  "$BIN" graph "$target" -o "$output"
  echo "wrote $output"
}

run_web() {
  local fixture="$1"
  local target addr
  target="$(fixture_path "$fixture")"
  addr="$(serve_port "$fixture")"
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
  build)
    echo "built $BIN"
    ;;
  init)
    shift
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
  graph)
    shift
    run_graph "${1:-demo}"
    ;;
  web)
    shift
    run_web "${1:-ecommerce}"
    ;;
  validate-demo)
    run_validate demo
    ;;
  validate-ecommerce)
    run_validate ecommerce
    ;;
  validate-playground)
    run_validate playground
    ;;
  scan-demo)
    run_scan demo
    ;;
  scan-ecommerce)
    run_scan ecommerce
    ;;
  scan-playground)
    run_scan playground
    ;;
  graph-demo)
    run_graph demo
    ;;
  graph-ecommerce)
    run_graph ecommerce
    ;;
  graph-playground)
    run_graph playground
    ;;
  web-demo)
    run_web demo
    ;;
  web-ecommerce)
    run_web ecommerce
    ;;
  web-playground)
    run_web playground
    ;;
  fixture)
    shift
    fixture="${1:-}"
    command="${2:-validate}"
    if [[ -z "$fixture" ]]; then
      echo "usage: ./scripts/go.sh fixture <demo|ecommerce|playground> [command]" >&2
      exit 1
    fi
    shift
    shift || true
    exec "$BIN" "$command" "$(fixture_path "$fixture")" "$@"
    ;;
  run)
    shift
    exec "$BIN" "$@"
    ;;
  demo|ecommerce|playground)
    fixture="$1"
    shift
    if [[ $# -eq 0 ]]; then
      set -- validate
    fi
    exec "$BIN" "$1" "$(fixture_path "$fixture")" "${@:2}"
    ;;
  *)
    exec "$BIN" "$@"
    ;;
esac
