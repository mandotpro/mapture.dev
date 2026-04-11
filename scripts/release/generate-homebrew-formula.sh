#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF' >&2
usage: generate-homebrew-formula.sh \
  --formula-name <name> \
  --class-name <ruby-class> \
  --formula-version <version> \
  --binary-version <version> \
  --source-url <url> \
  --source-sha256 <sha256> \
  [--repo <owner/repo>] \
  [--description <text>] \
  [--license <license>]
EOF
  exit 1
}

formula_name=""
class_name=""
formula_version=""
binary_version=""
source_url=""
source_sha256=""
repo="mandotpro/mapture.dev"
description="Repo-native architecture mapping that stays close to the code"
license_name="MIT"

while [[ $# -gt 0 ]]; do
  case "$1" in
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
    --description)
      description="${2:-}"
      shift 2
      ;;
    --license)
      license_name="${2:-}"
      shift 2
      ;;
    *)
      usage
      ;;
  esac
done

if [[ -z "$formula_name" || -z "$class_name" || -z "$formula_version" || -z "$binary_version" || -z "$source_url" || -z "$source_sha256" ]]; then
  usage
fi

cat <<EOF
class ${class_name} < Formula
  desc "${description}"
  homepage "https://github.com/${repo}"
  url "${source_url}"
  sha256 "${source_sha256}"
  version "${formula_version}"
  license "${license_name}"

  livecheck do
    skip "Managed by the Mapture release automation."
  end
EOF

cat <<EOF

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s
      -w
      -X github.com/mandotpro/mapture.dev/src/cmd.version=${binary_version}
    ]

    system "go", "build", *std_go_args(ldflags: ldflags, output: bin/"mapture"), "./src"
  end

  test do
    assert_match "mapture version ${binary_version}", shell_output("#{bin}/mapture --version")
  end
end
EOF

