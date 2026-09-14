#!/usr/bin/env bash
set -euo pipefail
repo='pkyanam/agentcalc'; version=''; bin_dir="${AGENTCALC_BIN_DIR:-$HOME/.local/bin}"; skill_dir="${CODEX_HOME:-$HOME/.codex}/skills/agentcalc"; install_skill=0; force=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) [[ $# -ge 2 ]] || { echo 'install: --version needs a value' >&2; exit 2; }; version=$2; shift 2;;
    --bin-dir) [[ $# -ge 2 ]] || { echo 'install: --bin-dir needs a directory' >&2; exit 2; }; bin_dir=$2; shift 2;;
    --skill) install_skill=1; shift;;
    --skill-dir) [[ $# -ge 2 ]] || { echo 'install: --skill-dir needs a directory' >&2; exit 2; }; skill_dir=$2; install_skill=1; shift 2;;
    --force) force=1; shift;;
    -h|--help) echo 'Usage: install.sh [--version VERSION] [--bin-dir DIR] [--skill] [--skill-dir DIR] [--force]'; exit 0;;
    *) echo "install: unknown option $1" >&2; exit 2;;
  esac
done
os=$(uname -s | tr '[:upper:]' '[:lower:]'); arch=$(uname -m)
case "$os" in darwin) os=darwin;; linux) os=linux;; mingw*|msys*|cygwin*) os=windows;; *) echo "install: unsupported OS $(uname -s)" >&2; exit 1;; esac
case "$arch" in x86_64|amd64) arch=amd64;; arm64|aarch64) arch=arm64;; *) echo "install: unsupported architecture $(uname -m)" >&2; exit 1;; esac
command -v gh >/dev/null || { echo 'install: gh (GitHub CLI) is required' >&2; exit 1; }
[[ -n "$version" ]] || version=$(gh release view -R "$repo" --json tagName --jq .tagName); version=${version#v}
asset="agentcalc-${version}-${os}-${arch}"; [[ "$os" == windows ]] && asset+='.zip' || asset+='.tar.gz'
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
gh release download "v$version" -R "$repo" --pattern "$asset" --pattern checksums.txt --dir "$tmp"
line=$(awk -v f="$asset" '$2 == f {print; exit}' "$tmp/checksums.txt"); [[ -n "$line" ]] || { echo "install: checksum missing for $asset" >&2; exit 1; }
(cd "$tmp" && printf '%s\n' "$line" | (if command -v sha256sum >/dev/null; then sha256sum -c -; else shasum -a 256 -c -; fi))
mkdir -p "$bin_dir"; bin_dir=$(cd "$bin_dir" && pwd); binary="$bin_dir/agentcalc"; [[ "$os" == windows ]] && binary+='.exe'
if [[ -e "$binary" && $force -ne 1 ]]; then echo "install: $binary exists; use --force to replace" >&2; exit 1; fi
if [[ $install_skill -eq 1 && -e "$skill_dir" && $force -ne 1 ]]; then echo "install: $skill_dir exists; use --force to replace" >&2; exit 1; fi
mkdir -p "$tmp/unpack"
if [[ "$os" == windows ]]; then unzip -q "$tmp/$asset" -d "$tmp/unpack"; else tar -xzf "$tmp/$asset" -C "$tmp/unpack"; fi
package=$(find "$tmp/unpack" -mindepth 1 -maxdepth 1 -type d -print -quit)
install -m 0755 "$package/agentcalc${binary##*agentcalc}" "$binary"
if [[ $install_skill -eq 1 ]]; then
  mkdir -p "$(dirname "$skill_dir")"
  mkdir -p "$skill_dir"
  cp -R "$package/skills/agentcalc/." "$skill_dir/"
  printf '\n## Local installation\n\nThe installed binary is `%s`. Use this absolute path when `agentcalc` is not on PATH.\n' "$binary" >> "$skill_dir/SKILL.md"
  printf 'Installed %s and skill files at %s\n' "$binary" "$skill_dir"
else
  printf 'Installed %s\n' "$binary"
fi
