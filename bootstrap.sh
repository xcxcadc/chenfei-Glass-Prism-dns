#!/usr/bin/env bash
set -euo pipefail

repository='xcxcadc/chenfei-Glass-Prism-dns'
release_url="https://github.com/$repository/releases/latest"
tag_prefix="https://github.com/$repository/releases/tag/"

fail() { printf 'Prism DNS: %s\n' "$1" >&2; exit 1; }

command -v curl >/dev/null 2>&1 || fail '缺少 curl，请先安装。'
resolved=$(curl -qfsSL --proto '=https' --proto-redir '=https' --connect-timeout 8 --max-time 20 \
  --output /dev/null --write-out '%{url_effective}' "$release_url") || fail '无法查询 GitHub 的最新正式发布。'
tag=${resolved#"$tag_prefix"}
[[ "$resolved" == "$tag_prefix"* && "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || fail '未找到有效的正式发布版本。'

control=$(mktemp "${TMPDIR:-/tmp}/prism-bootstrap.XXXXXX")
trap 'rm -f -- "$control"' EXIT
curl -qfsSL --proto '=https' --proto-redir '=https' --connect-timeout 8 --max-time 60 \
  --output "$control" "https://raw.githubusercontent.com/$repository/refs/tags/$tag/dns.sh" \
  || fail "无法下载 $tag 的控制脚本。"
grep -Fxq "PRISMCTL_VERSION=\"${tag#v}\"" "$control" && bash -n "$control" \
  || fail '控制脚本版本或语法与发布不符。'
PRISM_SCRIPT_REPO="$repository" PRISM_SCRIPT_REF="$tag" bash "$control" "$@"
