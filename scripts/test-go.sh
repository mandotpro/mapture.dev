#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"
GOTESTSUM_BIN="$(ensure_gotestsum)"
REPORTS_DIR="$(reports_dir)"

if [[ "${1:-}" == "--install-only" ]]; then
  ensure_golangci_lint >/dev/null
  echo "Installed Go dev tools into $(tools_bin_dir)"
  exit 0
fi

cd "$ROOT_DIR"

mapfile -t packages < <(mapture_go_packages)

"$GOTESTSUM_BIN" \
  --format pkgname \
  --junitfile "$REPORTS_DIR/gotestsum-junit.xml" \
  -- \
  -count=1 \
  "${packages[@]}"
