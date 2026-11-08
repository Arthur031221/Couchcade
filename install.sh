#!/bin/sh
# Download the latest Couchcade release for this machine into ~/.local/bin.
set -eu
repo="Arthur031221/Couchcade"
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "unsupported CPU: $(uname -m)" >&2; exit 1 ;;
esac
tag=$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)
[ -n "$tag" ] || { echo "no release found, build from source instead" >&2; exit 1; }
url="https://github.com/$repo/releases/download/$tag/couchcade_${tag#v}_${os}_${arch}.tar.gz"
dir="${HOME}/.local/bin"
mkdir -p "$dir"
curl -fsSL "$url" | tar -xz -C "$dir" couchcade
echo "installed $dir/couchcade $tag"
