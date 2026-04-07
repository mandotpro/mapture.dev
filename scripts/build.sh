#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

mkdir -p "$ROOT_DIR/build"
go build -o "$ROOT_DIR/build/mapture" "$ROOT_DIR/src"
