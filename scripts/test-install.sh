#!/usr/bin/env bash
# Installer integration test against locally built release assets.
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
test_root=$(mktemp -d)
trap 'rm -rf "$test_root"' EXIT
mkdir -p "$test_root/mock" "$test_root/assets"
cp "$repo_root"/dist/* "$test_root/assets/"
export AGENTCALC_TEST_ASSETS="$test_root/assets"
cat > "$test_root/mock/gh" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1 $2" == 'release view' ]]; then echo v0.1.0; exit 0; fi
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
"$test_root/bin/agentcalc" version | grep -q '0.1.0'
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
echo 'Installer integration checks passed.'
