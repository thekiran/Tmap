#!/usr/bin/env bash
# Build the IAD desktop app for Linux (or macOS) and bundle the iad-agent
# scanner beside it.
#
# Produces a folder under build/bin/ containing both:
#   - iad-console   (Wails desktop UI)
#   - iad-agent     (the network scanner the UI invokes)
# so the desktop app's "Run Scan" works out of the box (app.go resolveAgentBin
# finds the agent next to the executable).
#
# Requires: Go 1.22+, Node 18+, the Wails CLI, and on Linux the native deps
# (gcc, pkg-config, libgtk-3-dev, libwebkit2gtk-4.0-dev / 4.1-dev). Run
# `wails doctor` to confirm. Wails cannot cross-compile the GUI: build the
# Linux binary on Linux, the Windows binary on Windows.
set -euo pipefail

desktop="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$desktop/../.." && pwd)"
agent="$repo/agent"
bindir="$desktop/build/bin"

# Build the desktop UI first (overwrites build/bin in place). We avoid `-clean`
# so the build doesn't fail if an editor/indexer holds a handle on build/bin.
echo "==> Building iad-console (desktop UI)..."
( cd "$desktop" && wails build )

echo "==> Building iad-agent (scanner) into bundle..."
mkdir -p "$bindir"
( cd "$agent" && go build -o "$bindir/iad-agent" ./cmd/iad-agent )

# The agent needs the rules/ directory for --classify / --full scans. Ship it
# next to the binary so resolveRulesDir finds it (exe-dir/rules).
echo "==> Bundling rules/ ..."
rm -rf "$bindir/rules"
cp -R "$repo/rules" "$bindir/rules"

echo "==> Done. Bundle in: $bindir"
ls -la "$bindir"
