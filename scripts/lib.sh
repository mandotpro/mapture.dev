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

ensure_gotestsum() {
  local tools_dir bin version
  tools_dir="$(tools_bin_dir)"
  bin="$tools_dir/gotestsum"
  version="v1.13.0"

  mkdir -p "$tools_dir" "$(reports_dir)"

  if [[ ! -x "$bin" ]]; then
    echo "Installing gotestsum $version into $tools_dir"
    GOBIN="$tools_dir" go install "gotest.tools/gotestsum@$version"
  fi

  printf '%s\n' "$bin"
}

build_binary() {
  local output="$1"
  local repo
  repo="$(root_dir)"

  mkdir -p "$(dirname "$output")"
  go build -o "$output" "$repo/src"
}
