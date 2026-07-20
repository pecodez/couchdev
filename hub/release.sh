#!/usr/bin/env bash
set -euo pipefail

BUMP="${1:-patch}"

case "$BUMP" in
  patch|minor|major) ;;
  *) echo "Usage: release.sh [patch|minor|major]" >&2; exit 1 ;;
esac

REPO_ROOT="$(git rev-parse --show-toplevel)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Require a clean working tree
if [[ -n "$(git -C "$REPO_ROOT" status --porcelain)" ]]; then
  echo "error: working tree is not clean — commit or stash changes first" >&2
  exit 1
fi

# Resolve current version
CURRENT="$(git -C "$REPO_ROOT" tag -l 'v*' --sort=-v:refname | head -1)"
[[ -z "$CURRENT" ]] && CURRENT="v0.0.0"

# Parse semver
V="${CURRENT#v}"
MAJOR="${V%%.*}"
MINOR="${V#*.}"; MINOR="${MINOR%%.*}"
PATCH="${V##*.}"

# Bump
case "$BUMP" in
  major) MAJOR=$((MAJOR+1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR+1)); PATCH=0 ;;
  patch) PATCH=$((PATCH+1)) ;;
esac

NEW="v${MAJOR}.${MINOR}.${PATCH}"

echo "Current: ${CURRENT}"
echo "New:     ${NEW}"
echo ""

if [[ ! -t 0 ]]; then
  echo "error: confirmation requires an interactive terminal (stdin is not a TTY)" >&2
  echo "       refusing to guess — run this from an interactive shell" >&2
  exit 1
fi

read -rp "Tag and publish ${NEW}? [y/N] " confirm
[[ "$confirm" =~ ^[Yy]$ ]] || { echo "Aborted."; exit 0; }

# Tag and push
git -C "$REPO_ROOT" tag -a "${NEW}" -m "Release ${NEW}"
echo "Tagged ${NEW}"
git -C "$REPO_ROOT" push origin "${NEW}"
echo "Pushed tag"

# Build release binaries with version stamped
cd "$SCRIPT_DIR"
make build-release VERSION="${NEW}"
echo "Built release binaries"

# Publish GitHub release
gh release create "${NEW}" \
  bin/couchdev-linux-amd64 \
  bin/couchdev-linux-arm64 \
  --title "${NEW}" \
  --generate-notes

echo ""
echo "Published: https://github.com/pecodez/couchdev/releases/tag/${NEW}"
