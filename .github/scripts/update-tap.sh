#!/usr/bin/env bash
set -euo pipefail

REPO="PapaDanielVi/jamshid"
TAP_REPO="PapaDanielVi/homebrew-tap"
CASK_PATH="Casks/jamshid.rb"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE="$SCRIPT_DIR/../../homebrew/jamshid.rb.tmpl"
# shellcheck disable=SC2016
ENVSUBST_VARS='${VERSION} ${SHA256_DARWIN_ARM64} ${SHA256_DARWIN_X86_64} ${SHA256_LINUX_ARM64} ${SHA256_LINUX_X86_64}'

TAG="${1:?usage: update-tap.sh <tag>}"
VERSION="${TAG#v}"

: "${TAP_GITHUB_TOKEN:?TAP_GITHUB_TOKEN must be set}"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

gh release download "$TAG" --repo "$REPO" --pattern '*_checksums.txt' --dir "$WORKDIR"
CHECKSUMS="$WORKDIR/jamshid_${VERSION}_checksums.txt"
if [[ ! -f "$CHECKSUMS" ]]; then
  echo "checksum file jamshid_${VERSION}_checksums.txt not found in release $TAG" >&2
  exit 1
fi

checksum() {
  local sum
  sum="$(awk -v f="$1" '$2 == f { print $1 }' "$CHECKSUMS")"
  if [[ -z "$sum" ]]; then
    echo "no checksum for $1 in release $TAG" >&2
    exit 1
  fi
  printf '%s' "$sum"
}

SHA256_DARWIN_ARM64="$(checksum jamshid_Darwin_arm64.tar.gz)"
SHA256_DARWIN_X86_64="$(checksum jamshid_Darwin_x86_64.tar.gz)"
SHA256_LINUX_ARM64="$(checksum jamshid_Linux_arm64.tar.gz)"
SHA256_LINUX_X86_64="$(checksum jamshid_Linux_x86_64.tar.gz)"
export VERSION SHA256_DARWIN_ARM64 SHA256_DARWIN_X86_64 SHA256_LINUX_ARM64 SHA256_LINUX_X86_64

envsubst "$ENVSUBST_VARS" < "$TEMPLATE" > "$WORKDIR/jamshid.rb"

git clone --depth 1 "https://x-access-token:${TAP_GITHUB_TOKEN}@github.com/${TAP_REPO}.git" "$WORKDIR/tap"

if diff -q "$WORKDIR/jamshid.rb" "$WORKDIR/tap/$CASK_PATH" >/dev/null 2>&1; then
  echo "$CASK_PATH already up to date for $TAG"
  exit 0
fi

cp "$WORKDIR/jamshid.rb" "$WORKDIR/tap/$CASK_PATH"
git -C "$WORKDIR/tap" config user.name "github-actions[bot]"
git -C "$WORKDIR/tap" config user.email "41898282+github-actions[bot]@users.noreply.github.com"
git -C "$WORKDIR/tap" add "$CASK_PATH"
git -C "$WORKDIR/tap" commit -m "Brew cask update for jamshid version $TAG"
git -C "$WORKDIR/tap" push
