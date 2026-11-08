#!/bin/sh
# Assemble the GitHub Pages demo: the web pages, the engine as WebAssembly and the landing page.
set -eu
out="${1:-_site}"
root="$(cd "$(dirname "$0")/.." && pwd)"
rm -rf "$out"
mkdir -p "$out/app"
cp -r "$root/web/index.html" "$root/web/play.html" "$root/web/logo.svg" "$root/web/css" "$root/web/js" "$root/web/locales" "$out/app/"
cp "$root/site/index.html" "$root/site/wasm_exec.js" "$out/"
cp "$root/assets/social-card.png" "$out/social-card.png" 2>/dev/null || true
(cd "$root" && GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o "$out/couchcade.wasm" ./cmd/couchcade-wasm)
echo "site written to $out"
