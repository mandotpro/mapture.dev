#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

usage() {
  cat <<'EOF' >&2
usage: sync-homebrew-tap.sh \
  --tap-dir <path> \
  --formula-name <name> \
  --class-name <ruby-class> \
  --formula-version <version> \
  --binary-version <version> \
  --source-url <url> \
  --source-sha256 <sha256> \
  [--repo <owner/repo>] \
  [--tap-repo <owner/homebrew-name>]
EOF
  exit 1
}

tap_dir=""
formula_name=""
class_name=""
formula_version=""
binary_version=""
source_url=""
source_sha256=""
repo="mandotpro/mapture.dev"
tap_repo="mandotpro/homebrew-mapture"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tap-dir)
      tap_dir="${2:-}"
      shift 2
      ;;
    --formula-name)
      formula_name="${2:-}"
      shift 2
      ;;
    --class-name)
      class_name="${2:-}"
      shift 2
      ;;
    --formula-version)
      formula_version="${2:-}"
      shift 2
      ;;
    --binary-version)
      binary_version="${2:-}"
      shift 2
      ;;
    --source-url)
      source_url="${2:-}"
      shift 2
      ;;
    --source-sha256)
      source_sha256="${2:-}"
      shift 2
      ;;
    --repo)
      repo="${2:-}"
      shift 2
      ;;
    --tap-repo)
      tap_repo="${2:-}"
      shift 2
      ;;
    *)
      usage
      ;;
  esac
done

if [[ -z "$tap_dir" || -z "$formula_name" || -z "$class_name" || -z "$formula_version" || -z "$binary_version" || -z "$source_url" || -z "$source_sha256" ]]; then
  usage
fi

mkdir -p "$tap_dir/Formula"

generate_args=(
  --formula-name "$formula_name"
  --class-name "$class_name"
  --formula-version "$formula_version"
  --binary-version "$binary_version"
  --source-url "$source_url"
  --source-sha256 "$source_sha256"
  --repo "$repo"
)

"$(root_dir)/scripts/generate-homebrew-formula.sh" "${generate_args[@]}" > "$tap_dir/Formula/${formula_name}.rb"

stable_install='Stable formula will appear after the first semver release is published.'
if [[ -f "$tap_dir/Formula/mapture.rb" ]]; then
  stable_install='`brew install mandotpro/mapture/mapture` for stable tagged releases.'
fi

canary_install='Canary formula will appear after the next successful `main` canary sync.'
if [[ -f "$tap_dir/Formula/mapture-canary.rb" ]]; then
  canary_install='`brew install mandotpro/mapture/mapture-canary` for the rolling canary channel.'
fi

cat > "$tap_dir/README.md" <<EOF
# homebrew-mapture

Homebrew tap for [Mapture](https://github.com/${repo}).

## Install

- ${stable_install}
- ${canary_install}

Add the tap once:

\`\`\`bash
brew tap mandotpro/mapture
\`\`\`

Then install or upgrade as needed:

\`\`\`bash
brew upgrade mapture-canary
\`\`\`

Both release channels install the \`mapture\` binary. If you switch channels, uninstall the other formula first.

This repository is updated by the release automation in [${repo}](https://github.com/${repo}).
EOF
