#!/usr/bin/env bash
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
version=${1:-${VERSION:-}}
if [[ -z "$version" ]]; then version=$(git -C "$repo_root" describe --tags --always 2>/dev/null || printf 'dev'); fi
version=${version#v}
dist="$repo_root/dist"
rm -rf "$dist"
mkdir -p "$dist"
for target in 'darwin amd64' 'darwin arm64' 'linux amd64' 'linux arm64' 'windows amd64' 'windows arm64'; do
  read -r goos goarch <<<"$target"
  name="agentcalc-${version}-${goos}-${goarch}"
  stage=$(mktemp -d)
  mkdir -p "$stage/$name"
  ext=''; [[ "$goos" == windows ]] && ext='.exe'
  (cd "$repo_root" && CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$stage/$name/agentcalc$ext" ./cmd/agentcalc)
  cp "$repo_root/README.md" "$repo_root/LICENSE" "$stage/$name/"
  cp -R "$repo_root/skills" "$stage/$name/skills"
  if [[ "$goos" == windows ]]; then (cd "$stage" && zip -q -r "$dist/$name.zip" "$name"); else tar -C "$stage" -czf "$dist/$name.tar.gz" "$name"; fi
  rm -rf "$stage"
done
if command -v sha256sum >/dev/null; then
  (cd "$dist" && sha256sum *.tar.gz *.zip > checksums.txt)
else
  (cd "$dist" && shasum -a 256 *.tar.gz *.zip > checksums.txt)
fi
printf 'Wrote release %s to %s\n' "$version" "$dist"
