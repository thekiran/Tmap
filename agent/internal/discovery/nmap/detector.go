package nmap

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

// ErrNotInstalled is returned when the Nmap binary cannot be found.
var ErrNotInstalled = errors.New("nmap: binary not found")

// Detector locates the Nmap binary and optionally reads its version.
type Detector struct {
	Binary   string
	LookPath func(string) (string, error)
	Run      func(ctx context.Context, bin string, args ...string) ([]byte, error)
}

func (d Detector) Detect(ctx context.Context) Detection {
	path, err := d.lookup()
	if err != nil {
		return Detection{}
	}
	version := ""
	if out, err := d.exec(ctx, path, "--version"); err == nil {
		version = firstLine(string(out))
	}
	return Detection{Found: true, Path: path, Version: version}
}

func (d Detector) lookup() (string, error) {
	bin := d.Binary
	if bin == "" {
		bin = "nmap"
	}
	lookPath := d.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	return lookPath(bin)
}

func (d Detector) exec(ctx context.Context, bin string, args ...string) ([]byte, error) {
	if d.Run != nil {
		return d.Run(ctx, bin, args...)
	}
	return exec.CommandContext(ctx, bin, args...).CombinedOutput()
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
