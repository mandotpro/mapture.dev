#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"
TESTING_DIR="$(testing_dir)"
PLAYGROUND_DIR="$(playground_dir)"
BIN_DIR="$TESTING_DIR/bin"
BIN="$BIN_DIR/mapture"
DEMO_DIR="$ROOT_DIR/examples/demo"

mkdir -p "$PLAYGROUND_DIR" "$BIN_DIR"

build_binary "$BIN"

if [[ $# -eq 0 ]]; then
  cat <<EOF
Built: $BIN
Playground: $PLAYGROUND_DIR
Demo fixture: $DEMO_DIR

Examples:
  ./scripts/go.sh help
  ./scripts/go.sh init
  ./scripts/go.sh validate-demo
  ./scripts/go.sh validate-playground
  ./scripts/go.sh run graph "$DEMO_DIR"
EOF
  exit 0
fi

case "$1" in
  help|-h|--help)
    exec "$BIN" --help
    ;;
  init)
    shift
    exec "$BIN" init "${1:-$PLAYGROUND_DIR}"
    ;;
  validate-demo)
    shift
    exec "$BIN" validate "$DEMO_DIR" "$@"
    ;;
  validate-playground)
    shift
    exec "$BIN" validate "$PLAYGROUND_DIR" "$@"
    ;;
  run)
    shift
    exec "$BIN" "$@"
    ;;
  demo)
    shift
    if [[ $# -eq 0 ]]; then
      set -- validate
    fi
    exec "$BIN" "$1" "$DEMO_DIR" "${@:2}"
    ;;
  playground)
    shift
    if [[ $# -eq 0 ]]; then
      set -- validate
    fi
    exec "$BIN" "$1" "$PLAYGROUND_DIR" "${@:2}"
    ;;
  *)
    exec "$BIN" "$@"
    ;;
esac
