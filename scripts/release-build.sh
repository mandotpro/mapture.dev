#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

if [[ $# -ne 4 ]]; then
  echo "usage: $0 <version> <goos> <goarch> <output-dir>" >&2
  exit 1
fi

VERSION="$1"
GOOS_TARGET="$2"
GOARCH_TARGET="$3"
OUTPUT_DIR="$4"

ROOT_DIR="$(root_dir)"
TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

mkdir -p "$OUTPUT_DIR"
OUTPUT_DIR="$(cd "$OUTPUT_DIR" && pwd)"

archive_version="$(printf '%s' "$VERSION" | tr '+' '_' | tr '/' '_')"
archive_base="mapture_${archive_version}_${GOOS_TARGET}_${GOARCH_TARGET}"
binary_name="mapture"
if [[ "$GOOS_TARGET" == "windows" ]]; then
  binary_name="mapture.exe"
fi

MAPTURE_VERSION="$VERSION" GOOS="$GOOS_TARGET" GOARCH="$GOARCH_TARGET" CGO_ENABLED=0 \
  build_binary "$TMP_DIR/$binary_name"

cp "$ROOT_DIR/LICENSE" "$TMP_DIR/LICENSE"
cp "$ROOT_DIR/README.md" "$TMP_DIR/README.md"

if [[ "$GOOS_TARGET" == "windows" ]]; then
  (
    cd "$TMP_DIR"
    zip -q "$OUTPUT_DIR/${archive_base}.zip" "$binary_name" LICENSE README.md
  )
  printf '%s\n' "$OUTPUT_DIR/${archive_base}.zip"
  exit 0
fi

tar -C "$TMP_DIR" -czf "$OUTPUT_DIR/${archive_base}.tar.gz" "$binary_name" LICENSE README.md
printf '%s\n' "$OUTPUT_DIR/${archive_base}.tar.gz"
