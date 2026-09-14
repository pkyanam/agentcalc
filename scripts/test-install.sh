#!/usr/bin/env bash
# Installer integration test against locally built release assets.
set -euo pipefail
test_version=${1:-0.3.0}
export AGENTCALC_TEST_VERSION="$test_version"
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
test_root=$(mktemp -d)
trap 'rm -rf "$test_root"' EXIT
mkdir -p "$test_root/mock" "$test_root/assets"
cp "$repo_root"/dist/* "$test_root/assets/"
export AGENTCALC_TEST_ASSETS="$test_root/assets"
cat > "$test_root/mock/gh" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1 $2" == 'release view' ]]; then echo "v$AGENTCALC_TEST_VERSION"; exit 0; fi
patterns=(); destination=''
while [[ $# -gt 0 ]]; do
  case "$1" in
    --pattern) patterns+=("$2"); shift 2;;
    --dir) destination=$2; shift 2;;
    *) shift;;
  esac
done
for pattern in "${patterns[@]}"; do cp "$AGENTCALC_TEST_ASSETS/$pattern" "$destination/"; done
MOCK
chmod +x "$test_root/mock/gh"
export PATH="$test_root/mock:$PATH"
bash "$repo_root/scripts/install.sh" --bin-dir "$test_root/bin" --skill-dir "$test_root/skill"
"$test_root/bin/agentcalc" version | grep -Fq "$test_version"
test -f "$test_root/skill/SKILL.md"
grep -Fq "$test_root/bin/agentcalc" "$test_root/skill/SKILL.md"
if bash "$repo_root/scripts/install.sh" --bin-dir "$test_root/bin" --skill-dir "$test_root/skill" 2>/dev/null; then
  echo 'installer unexpectedly overwrote existing files' >&2; exit 1
fi
bash "$repo_root/scripts/install.sh" --bin-dir "$test_root/bin" --skill-dir "$test_root/skill" --force
test ! -e "$test_root/skill/agentcalc"
bash "$repo_root/scripts/install.sh" --bin-dir "$test_root/only-bin"
# Corrupt all archives, preserving the original checksums.
for archive in "$test_root/assets"/*.tar.gz "$test_root/assets"/*.zip; do printf 'corrupt' >> "$archive"; done
if bash "$repo_root/scripts/install.sh" --bin-dir "$test_root/corrupt" 2>/dev/null; then
  echo 'installer accepted a checksum mismatch' >&2; exit 1
fi
test ! -e "$test_root/corrupt/agentcalc"

# Agent aliases select their documented home roots without touching the real home.
AGENTS_HOME="$test_root/agents" bash "$repo_root/scripts/install-skill.sh" --agent shared --binary "$test_root/bin/agentcalc"
test -f "$test_root/agents/skills/agentcalc/SKILL.md"
grep -Fq "$test_root/bin/agentcalc" "$test_root/agents/skills/agentcalc/SKILL.md"
test "$(bash "$repo_root/scripts/install-skill.sh" --agent codex --print-path)" = "${CODEX_HOME:-$HOME/.codex}/skills/agentcalc"
CLAUDE_CONFIG_DIR="$test_root/claude" bash "$repo_root/scripts/install-skill.sh" --agent claude
HERMES_HOME="$test_root/hermes" bash "$repo_root/scripts/install-skill.sh" --agent hermes
for root in agents claude hermes; do test -f "$test_root/$root/skills/agentcalc/SKILL.md"; done
mkdir -p "$test_root/real-skill"
ln -s "$test_root/real-skill" "$test_root/link-skill"
if bash "$repo_root/scripts/install-skill.sh" --skill-dir "$test_root/link-skill" 2>/dev/null; then
  echo 'install-skill unexpectedly followed a skill symlink' >&2; exit 1
fi
if bash "$repo_root/scripts/install-skill.sh" --agent shared --skill-dir "$test_root/ambiguous" 2>/dev/null; then
  echo 'install-skill unexpectedly accepted --agent with --skill-dir' >&2; exit 1
fi
if bash "$repo_root/scripts/install.sh" --agent shared --skill-dir "$test_root/ambiguous" --bin-dir "$test_root/ambiguous-bin" --version "$test_version" 2>/dev/null; then
  echo 'installer unexpectedly accepted --agent with --skill-dir' >&2; exit 1
fi
echo 'Installer integration checks passed.'
