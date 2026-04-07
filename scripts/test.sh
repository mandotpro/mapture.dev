#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"

go test ./src/...
go vet ./src/...
go run src/main.go --help >/dev/null
go run src/main.go validate examples/demo >/dev/null
