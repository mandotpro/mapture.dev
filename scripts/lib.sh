#!/usr/bin/env bash

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
    "$repo/src"
}

sha256_file() {
  local path="$1"

  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$path" | awk '{print $1}'
    return 0
  fi

  shasum -a 256 "$path" | awk '{print $1}'
}
