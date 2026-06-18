package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound backend for the IAD console.
//
// SAFETY CONTRACT
//   - This type NEVER parses, interprets, or mutates scan evidence. Validation
//     and normalization happen on the TypeScript side (Zod). Here we only move
//     bytes: read a file the user picked, write an export the user requested,
//     or invoke the external iad-agent binary and hand back its raw stdout.
//   - The scanner ("agent" module) is not imported. It is preserved as-is and
//     run as a child process when present.
type App struct {
	ctx context.Context
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// OpenScanFile shows a native open dialog and returns the raw file contents as
// a string. The frontend validates it with Zod before doing anything with it.
func (a *App) OpenScanFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import IAD scan report",
		Filters: []runtime.FileFilter{
			{DisplayName: "Scan JSON (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil // user cancelled
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}
	return string(b), nil
}

// SaveExport shows a native save dialog and writes the provided content
// verbatim. The frontend decides what to write (full report, summary, etc.).
func (a *App) SaveExport(suggestedName string, content string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export",
		DefaultFilename: suggestedName,
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// ListInterfaces invokes `iad-agent interfaces --json` and returns its raw JSON
// array of network interfaces. The Wireshark-style launch screen uses this to
// let the user pick which adapter to scan before starting.
func (a *App) ListInterfaces() (string, error) {
	out, err := a.runAgent(30*time.Second, "interfaces", "--json")
	if err != nil {
		return "", err
	}
	return out, nil
}

// RunScan invokes the EXTERNAL iad-agent binary (the preserved Go scanner) and
// returns its raw JSON stdout. This is a passthrough — the console never
// reimplements detection. mode selects the scan profile; iface (optional) pins
// the scan to a specific network interface chosen on the launch screen.
func (a *App) RunScan(mode string, iface string) (string, error) {
	args, err := agentScanArgs(mode, iface)
	if err != nil {
		return "", err
	}
	return a.runAgent(4*time.Minute, args...)
}

// runAgent locates and executes the bundled iad-agent with the given args,
// returning its stdout. It runs the agent in its own directory (so relative
// lookups like the bundled rules/ resolve), hides the console window on
// Windows, and surfaces stderr in the error.
func (a *App) runAgent(timeout time.Duration, args ...string) (string, error) {
	bin, err := a.resolveAgentBin()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(a.ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = filepath.Dir(bin)
	hideConsole(cmd)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", fmt.Errorf("iad-agent failed: %w: %s", err, msg)
		}
		return "", fmt.Errorf("iad-agent failed: %w", err)
	}
	return string(out), nil
}

// Platform is a tiny helper the UI uses for OS-specific affordances.
func (a *App) Platform() string { return goruntime.GOOS }

// resolveAgentBin locates the external iad-agent binary, in priority order:
//  1. IAD_AGENT_BIN env var (explicit override),
//  2. next to the desktop executable (the shipped/bundled layout),
//  3. on PATH.
//
// This lets a packaged install "just work" when the agent binary is placed
// beside the console, while still honoring an explicit override or a
// developer's PATH.
func (a *App) resolveAgentBin() (string, error) {
	name := "iad-agent"
	if goruntime.GOOS == "windows" {
		name = "iad-agent.exe"
	}

	if bin := strings.TrimSpace(os.Getenv("IAD_AGENT_BIN")); bin != "" {
		if _, err := os.Stat(bin); err != nil {
			return "", fmt.Errorf("IAD_AGENT_BIN=%q not found: %w", bin, err)
		}
		return bin, nil
	}

	if exe, err := os.Executable(); err == nil {
		cand := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(cand); err == nil {
			return cand, nil
		}
	}

	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}

	return "", fmt.Errorf("iad-agent binary not found: place %s next to the app, set IAD_AGENT_BIN, or add it to PATH; otherwise use Import instead", name)
}

func agentScanArgs(mode string, iface string) ([]string, error) {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == "" {
		mode = "standard"
	}

	var args []string
	switch mode {
	case "full":
		args = []string{"scan", "--full"}
	case "quick", "standard", "deep":
		args = []string{"scan", "--cidr", "auto", "--profile", mode}
		if mode != "quick" {
			args = append(args, "--classify")
		}
	default:
		return nil, fmt.Errorf("invalid scan mode %q (use quick, standard, deep, or full)", mode)
	}

	// Pin the scan to the interface chosen on the launch screen (Wireshark-style).
	if iface = strings.TrimSpace(iface); iface != "" {
		args = append(args, "--interface", iface)
	}
	return args, nil
}
