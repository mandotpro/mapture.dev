#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

ROOT_DIR="$(root_dir)"
OUTPUT="$ROOT_DIR/build/mapture"

mapture_print_section "Local Build"
ensure_embedded_web_bundle
build_binary "$OUTPUT"
print_binary_summary "local build" "$OUTPUT"
printf '%s\n' "$(mapture_success "build complete")"
