#!/usr/bin/env bash

mapture_color_mode() {
  local mode="${MAPTURE_COLOR:-auto}"
  case "$mode" in
    auto|always|never) ;;
    *) mode="auto" ;;
  esac
  printf '%s\n' "$mode"
}

mapture_color_enabled() {
  local mode
  mode="$(mapture_color_mode)"
  case "$mode" in
    always) return 0 ;;
    never) return 1 ;;
  esac

  if [[ -n "${NO_COLOR:-}" ]]; then
    return 1
  fi

  [[ -t 1 ]]
}

mapture_style() {
  local code="$1"
  shift
  if mapture_color_enabled; then
    printf '\033[%sm%s\033[0m' "$code" "$*"
  else
    printf '%s' "$*"
  fi
}

mapture_accent() { mapture_style "36" "$*"; }
mapture_success() { mapture_style "32" "$*"; }
mapture_warning() { mapture_style "33" "$*"; }
mapture_error() { mapture_style "31" "$*"; }
mapture_strong() { mapture_style "1" "$*"; }
mapture_muted() { mapture_style "90" "$*"; }

mapture_print_section() {
  printf '\n%s\n' "$(mapture_strong "$1")"
}

mapture_print_kv() {
  printf '  %s %s\n' "$(mapture_accent "$(printf '%-18s' "$1")")" "$2"
}

root_dir() {
  cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd
}

testing_dir() {
  printf '%s/testing\n' "$(root_dir)"
}

tools_bin_dir() {
  printf '%s/tools/bin\n' "$(testing_dir)"
}

reports_dir() {
  printf '%s/reports\n' "$(testing_dir)"
}

playground_dir() {
  printf '%s/playground\n' "$(testing_dir)"
}

outputs_dir() {
  printf '%s/outputs\n' "$(testing_dir)"
}

testing_examples_dir() {
  printf '%s/examples\n' "$(testing_dir)"
}

ensure_gotestsum() {
  local tools_dir bin version
  tools_dir="$(tools_bin_dir)"
  bin="$tools_dir/gotestsum"
  version="v1.13.0"

  mkdir -p "$tools_dir" "$(reports_dir)"

  if [[ ! -x "$bin" ]]; then
    echo "Installing gotestsum $version into $tools_dir" >&2
    GOBIN="$tools_dir" go install "gotest.tools/gotestsum@$version"
  fi

  printf '%s\n' "$bin"
}

ensure_golangci_lint() {
  local tools_dir bin version
  tools_dir="$(tools_bin_dir)"
  bin="$tools_dir/golangci-lint"
  version="v2.11.4"

  mkdir -p "$tools_dir"

  if [[ ! -x "$bin" ]] || ! "$bin" version 2>/dev/null | grep -q "version $version"; then
    echo "Installing golangci-lint $version into $tools_dir" >&2
    rm -f "$bin"
    GOBIN="$tools_dir" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$version"
  fi

  printf '%s\n' "$bin"
}

build_binary() {
  local output="$1"
  local repo
  local version
  repo="$(root_dir)"
  version="${MAPTURE_VERSION:-0.0.0-dev}"

  mkdir -p "$(dirname "$output")"
  CGO_ENABLED="${CGO_ENABLED:-0}" \
  go build \
    -trimpath \
    -ldflags "-X github.com/mandotpro/mapture.dev/src/cmd.version=$version" \
    -o "$output" \
    "$repo/cmd/mapture"
}

sha256_file() {
  local path="$1"

  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$path" | awk '{print $1}'
    return 0
  fi

  shasum -a 256 "$path" | awk '{print $1}'
}
