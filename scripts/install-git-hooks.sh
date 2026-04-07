#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"

git -C "$ROOT_DIR" config core.hooksPath .githooks
echo "Configured git hooks path: $ROOT_DIR/.githooks"
