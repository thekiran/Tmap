# IAD Desktop Console

Wails + React desktop app for Internet Access Detector. Runs natively on
**Windows** and **Linux** (and macOS).

This app is a UI host only. It does not import or reimplement the scanner. It
loads validated JSON reports from disk, or invokes the external `iad-agent`
binary through `RunScan` and renders the returned JSON.

## Layout

```text
apps/desktop/
  app.go             Wails bindings for import, export, and agent execution
  main.go            Wails application bootstrap
  build-app.ps1      One-shot Windows build (console + bundled agent)
  build-app.sh       One-shot Linux/macOS build (console + bundled agent)
  frontend/          Vite + React console
  build/bin/         Build output (gitignored)
```

## Prerequisites

- Go 1.22+
- Node 18+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- A C compiler:
  - **Windows** — a recent MinGW-w64 `gcc` (WebView2 runtime ships with Windows 10/11).
  - **Linux** — `gcc pkg-config libgtk-3-dev libwebkit2gtk-4.0-dev` (or `4.1-dev`).

Run `wails doctor` to confirm your toolchain.

## Development

```sh
npm install            # from the repo root (workspaces)
npm run desktop:dev    # hot-reloading dev server
```

## Production build

The build scripts compile the desktop UI **and** the `iad-agent` scanner and
place both binaries — plus the `rules/` directory the scanner needs for
classification — side by side in `build/bin/`, so the app's **Run Scan**
button works without any PATH or rules setup. The agent is spawned without a
console window (no terminal flashes on Windows).

```sh
# Windows (PowerShell)
apps/desktop/build-app.ps1

# Linux / macOS
apps/desktop/build-app.sh
```

> Wails cannot cross-compile the GUI (it links native WebView libraries), so
> build the **Windows** binary on Windows and the **Linux** binary on Linux.
> The Go agent itself *is* cross-platform and is built per-target by the scripts.

## How the app finds the scanner

`app.go`'s `resolveAgentBin` looks for the agent in this order:

1. `IAD_AGENT_BIN` environment variable (explicit override),
2. next to the desktop executable (the bundled layout the build scripts produce),
3. on `PATH`.

If none is found, the app reports a clear error and you can still use **Import**
to load a previously generated scan JSON.
