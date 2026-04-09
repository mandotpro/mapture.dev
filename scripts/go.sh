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

  echo "unknown fixture: $fixture" >&2
  echo "known fixtures: $(fixture_names | tr '\n' ' ' | sed 's/ $//')" >&2
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
    validate|scan|graph)
      ;;
    *)
      echo "unsupported all-fixtures command: $command" >&2
      echo "supported commands: validate scan graph" >&2
      exit 1
      ;;
  esac

  local fixture
  while IFS= read -r fixture; do
    printf '== %s (%s) ==\n' "$fixture" "$command"
    case "$command" in
      validate)
        "$BIN" validate "$(fixture_path "$fixture")" "$@"
        ;;
      scan)
        run_scan "$fixture"
        ;;
      graph)
        run_graph "$fixture"
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
      echo "unknown fixture: $1" >&2
      exit 1
      ;;
  esac
}

show_help() {
  local fixture
  cat <<EOF
Repo Development Commands:
  ./scripts/go.sh build
  ./scripts/go.sh init

Repo Verification Commands:
  ./scripts/go.sh validate <fixture|all>
  ./scripts/go.sh scan <fixture|all>
  ./scripts/go.sh graph <fixture|all>

Local Verification With Fixtures:
  ./scripts/go.sh fixtures
  ./scripts/go.sh serve <fixture>
  ./scripts/go.sh fixture <fixture> <command>

Paths:
  built binary       -> $BIN
  testing root       -> $TESTING_DIR
  playground         -> $PLAYGROUND_DIR
  outputs            -> $OUTPUTS_DIR
  testing examples   -> $TESTING_EXAMPLES_DIR

Fixtures:
EOF
  while IFS= read -r fixture; do
    printf '  %-10s -> %s\n' "$fixture" "$(fixture_path "$fixture")"
  done < <(fixture_names)
}

run_validate() {
  local fixture="$1"
  if run_across_examples_if_needed validate "$fixture"; then
    return
  fi
  build_binary "$BIN"
  exec "$BIN" validate "$(fixture_path "$fixture")"
}

run_scan() {
  local fixture="$1"
  if run_across_examples_if_needed scan "$fixture"; then
    return
  fi
  build_binary "$BIN"
  local target output
  target="$(fixture_path "$fixture")"
  output="$(fixture_output "$fixture" "scan.json")"
  "$BIN" scan "$target" >"$output"
  echo "wrote $output"
}

run_graph() {
  local fixture="$1"
  if run_across_examples_if_needed graph "$fixture"; then
    return
  fi
  build_binary "$BIN"
  local target output
  target="$(fixture_path "$fixture")"
  output="$(fixture_output "$fixture" "mmd")"
  "$BIN" graph "$target" -o "$output"
  echo "wrote $output"
}

run_serve() {
  local fixture="$1"
  if [[ "$fixture" == "all" ]]; then
    echo 'serve does not support fixture "all"; choose one fixture' >&2
    exit 1
  fi
  build_binary "$BIN"
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
  fixtures)
    fixture_names
    ;;
  path)
    shift
    fixture_path "${1:-demo}"
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
  serve)
    shift
    run_serve "${1:-ecommerce}"
    ;;
  fixture)
    shift
    fixture="${1:-}"
    command="${2:-validate}"
    if [[ -z "$fixture" ]]; then
      echo "usage: ./scripts/go.sh fixture <fixture> [command]" >&2
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
  demo|ecommerce|migration|playground)
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
