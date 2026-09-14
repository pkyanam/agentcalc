#!/usr/bin/env bash
# Install the repository-bundled agentcalc skill for a local agent.
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
agent='shared'; skill_dir=''; skill_dir_set=0; agent_set=0; binary=''; force=0; print_path=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --agent) [[ $# -ge 2 ]] || { echo 'install-skill: --agent needs shared, codex, claude, or hermes' >&2; exit 2; }; agent=$2; agent_set=1; shift 2;;
    --skill) shift;;
    --print-path) print_path=1; shift;;
    --skill-dir) [[ $# -ge 2 ]] || { echo 'install-skill: --skill-dir needs a directory' >&2; exit 2; }; skill_dir=$2; skill_dir_set=1; shift 2;;
    --binary) [[ $# -ge 2 ]] || { echo 'install-skill: --binary needs a path' >&2; exit 2; }; binary=$2; shift 2;;
    --force) force=1; shift;;
    -h|--help) echo 'Usage: install-skill.sh [--agent shared|codex|claude|hermes] [--skill-dir DIR] [--binary PATH] [--force] [--print-path]'; exit 0;;
    *) echo "install-skill: unknown option $1" >&2; exit 2;;
  esac
done
if [[ $skill_dir_set -eq 1 && $agent_set -eq 1 ]]; then
  echo 'install-skill: --agent and --skill-dir are mutually exclusive' >&2; exit 2
fi
case "$agent" in
  shared) root="${AGENTS_HOME:-$HOME/.agents}";;
  codex) root="${CODEX_HOME:-$HOME/.codex}";;
  claude) root="${CLAUDE_CONFIG_DIR:-$HOME/.claude}";;
  hermes) root="${HERMES_HOME:-$HOME/.hermes}";;
  *) echo "install-skill: unknown agent $agent (choose shared, codex, claude, or hermes)" >&2; exit 2;;
esac
[[ $skill_dir_set -eq 1 ]] || skill_dir="$root/skills/agentcalc"
if [[ $print_path -eq 1 ]]; then printf '%s\n' "$skill_dir"; exit 0; fi
if [[ -n "$binary" ]]; then
  if [[ "$binary" != */* ]]; then binary=$(command -v "$binary") || { echo 'install-skill: binary not found' >&2; exit 1; }; fi
  [[ -x "$binary" && ! -d "$binary" ]] || { echo 'install-skill: --binary must name an executable file' >&2; exit 1; }
  binary="$(cd "$(dirname "$binary")" && pwd)/$(basename "$binary")"
fi
source_dir="$repo_root/skills/agentcalc"
[[ -d "$source_dir" ]] || { echo "install-skill: bundled skill not found at $source_dir" >&2; exit 1; }
if [[ -L "$skill_dir" || ( -e "$skill_dir" && ! -d "$skill_dir" ) ]]; then echo "install-skill: $skill_dir is not a directory" >&2; exit 1; fi
if [[ -e "$skill_dir" && $force -ne 1 ]]; then echo "install-skill: $skill_dir exists; use --force to replace" >&2; exit 1; fi
mkdir -p "$(dirname "$skill_dir")" "$skill_dir"
cp -R "$source_dir/." "$skill_dir/"
if [[ -n "$binary" ]]; then
  printf '\n## Local installation\n\nThe installed binary is `%s`. Use this absolute path when `agentcalc` is not on PATH.\n' "$binary" >> "$skill_dir/SKILL.md"
fi
printf 'Installed skill files at %s\n' "$skill_dir"
