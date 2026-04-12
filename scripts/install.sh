#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF' >&2
usage: install.sh [--channel stable|canary] [--dir <path>]

Downloads the latest Mapture release archive for the current platform and
installs the `mapture` binary into the target directory.

Environment overrides:
  MAPTURE_CHANNEL       stable|canary (default: stable)
  MAPTURE_INSTALL_DIR   install directory (default: ~/.local/bin)
EOF
  exit 1
}

channel="${MAPTURE_CHANNEL:-stable}"
install_dir="${MAPTURE_INSTALL_DIR:-$HOME/.local/bin}"
repo="mandotpro/mapture.dev"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --channel)
      channel="${2:-}"
      shift 2
      ;;
    --dir)
      install_dir="${2:-}"
      shift 2
      ;;
    --help|-h)
      usage
      ;;
    *)
      usage
      ;;
  esac
done

case "$channel" in
  stable|canary) ;;
  *)
    echo "unsupported channel: $channel" >&2
    exit 1
    ;;
esac

uname_s="$(uname -s)"
uname_m="$(uname -m)"

case "$uname_s" in
  Darwin) goos="darwin" ;;
  Linux) goos="linux" ;;
  *)
    echo "install.sh currently supports macOS and Linux. Download a release archive manually for $uname_s." >&2
    exit 1
    ;;
esac

case "$uname_m" in
  x86_64|amd64) goarch="amd64" ;;
  arm64|aarch64) goarch="arm64" ;;
  *)
    echo "unsupported CPU architecture: $uname_m" >&2
    exit 1
    ;;
esac

api_url="https://api.github.com/repos/${repo}/releases/latest"
if [[ "$channel" == "canary" ]]; then
  api_url="https://api.github.com/repos/${repo}/releases/tags/canary"
fi

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

release_json="$(curl -fsSL \
  -H 'Accept: application/vnd.github+json' \
  -H 'User-Agent: mapture-install-script' \
  "$api_url")"
python_bin=""
if command -v python3 >/dev/null 2>&1; then
  python_bin="python3"
elif command -v python >/dev/null 2>&1; then
  python_bin="python"
else
  echo "install.sh requires python3 or python for GitHub release metadata parsing" >&2
  exit 1
fi

release_tag=""
asset_url=""
while IFS= read -r line; do
  if [[ -z "$release_tag" ]]; then
    release_tag="$line"
    continue
  fi
  asset_url="$line"
  break
done < <(RELEASE_JSON="$release_json" "$python_bin" - "$goos" "$goarch" <<'PY'
import json
import os
import sys

goos = sys.argv[1]
goarch = sys.argv[2]
release = json.loads(os.environ["RELEASE_JSON"])
print(release.get("tag_name", ""))
for asset in release.get("assets", []):
    name = asset.get("name", "")
    if not name.startswith("mapture_"):
        continue
    if not (name.endswith(".tar.gz") or name.endswith(".zip")):
        continue
    if f"_{goos}_{goarch}" not in name:
        continue
    print(asset.get("browser_download_url", ""))
    break
PY
)

if [[ -z "$asset_url" ]]; then
  if [[ "$channel" == "stable" ]]; then
    echo "latest stable release ${release_tag:-unknown} does not publish prebuilt ${goos}/${goarch} assets yet; use Homebrew, source install, or the canary channel until the next stable release is cut" >&2
  else
    echo "no installable asset found for ${goos}/${goarch} on channel ${channel}" >&2
  fi
  exit 1
fi

asset_name="$(basename "$asset_url")"
release_version="${asset_name#mapture_}"
release_version="${release_version%_${goos}_${goarch}.tar.gz}"
release_version="${release_version//_/+}"

mkdir -p "$install_dir"
curl -fsSL "$asset_url" -o "$tmp_dir/mapture.tar.gz"
tar -xzf "$tmp_dir/mapture.tar.gz" -C "$tmp_dir" mapture
install -m 0755 "$tmp_dir/mapture" "$install_dir/mapture"

printf 'Installed mapture %s to %s/mapture\n' "$release_version" "$install_dir"
printf 'Run `mapture --version` to verify the install.\n'

case ":$PATH:" in
  *":$install_dir:"*) ;;
  *)
    printf 'Add %s to PATH to run `mapture` directly.\n' "$install_dir"
    ;;
esac
