#!/usr/bin/env bash
# Build the IAD desktop app (Linux/macOS) and bundle the iad-agent scanner beside it.
#
# Usage: ./build-app.sh [prod|dev|both] [--skip-agent]
#   prod (default) -> iad-console        (release)
#   dev            -> iad-console-dev     (debug + devtools)
#   both           -> both consoles
#
# Produces build/bin/ with the console(s) + iad-agent + rules/, so the desktop
# app's "Run Scan" works out of the box (resolveAgentBin finds the agent next to
# the executable).
#
# Requires: Go 1.22+, Node 18+, the Wails CLI, and on Linux the native deps
# (gcc, pkg-config, libgtk-3-dev, libwebkit2gtk-4.0-dev / 4.1-dev). Run
# `wails doctor` to confirm. Wails cannot cross-compile the GUI: build the Linux
# binary on Linux, the Windows binary on Windows.
set -euo pipefail

mode="prod"
skip_agent=0
for arg in "$@"; do
  case "$arg" in
    prod|dev|both) mode="$arg" ;;
    --skip-agent) skip_agent=1 ;;
    *) echo "unknown arg: $arg" >&2; exit 2 ;;
  esac
done

desktop="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$desktop/../.." && pwd)"
agent="$repo/agent"
bindir="$desktop/build/bin"

# We avoid `-clean` so the build doesn't fail if an editor/indexer holds a handle.
build_console() { # $1 = output name, $2 = "debug" | ""
  local out="$1"
  if [[ "${2:-}" == "debug" ]]; then
    ( cd "$desktop" && wails build -debug -devtools -o "$out" )
  else
    ( cd "$desktop" && wails build -o "$out" )
  fi
}

if [[ "$mode" == "prod" || "$mode" == "both" ]]; then
  echo "==> Building iad-console (release)..."
  build_console "iad-console" ""
fi
if [[ "$mode" == "dev" || "$mode" == "both" ]]; then
  echo "==> Building iad-console-dev (debug + devtools)..."
  build_console "iad-console-dev" "debug"
fi

if [[ "$skip_agent" -eq 0 ]]; then
  echo "==> Building iad-agent (scanner) into bundle..."
  mkdir -p "$bindir"
  ( cd "$agent" && go build -o "$bindir/iad-agent" ./cmd/iad-agent )

  echo "==> Bundling rules/ ..."
  rm -rf "$bindir/rules"
  cp -R "$repo/rules" "$bindir/rules"
fi

echo "==> Done. Bundle in: $bindir"
ls -la "$bindir"
