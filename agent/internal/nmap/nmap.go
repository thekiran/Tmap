package nmap

import (
	"context"
	"errors"
	"os/exec"
)

// ErrNotInstalled is returned when the Nmap binary cannot be found in PATH.
var ErrNotInstalled = errors.New("nmap: binary not found in PATH")

// Runner executes Nmap. The zero value is usable (it looks up "nmap" in PATH).
type Runner struct {
	// Binary overrides the executable name/path (default "nmap").
	Binary string
	// lookPath and run are injectable for tests; nil → real implementations.
	lookPath func(string) (string, error)
	run      func(ctx context.Context, bin string, args ...string) ([]byte, error)
}

// Available reports whether Nmap is installed. Callers should check this and skip
// Nmap gracefully when false — Nmap is always optional.
func (r Runner) Available() bool {
	_, err := r.lookup()
	return err == nil
}

// Scan runs Nmap against target (a CIDR or host) with the given profile and
// returns parsed hosts. It requests XML on stdout and parses only that XML.
// Returns ErrNotInstalled when Nmap is absent, or any exec/parse error otherwise.
func (r Runner) Scan(ctx context.Context, target, profile string) ([]Host, error) {
	bin, err := r.lookup()
	if err != nil {
		return nil, ErrNotInstalled
	}
	args := append(profileArgs(profile), target)
	out, err := r.exec(ctx, bin, args...)
	if err != nil {
		return nil, err
	}
	return Parse(out)
}

func (r Runner) lookup() (string, error) {
	bin := r.Binary
	if bin == "" {
		bin = "nmap"
	}
	lp := r.lookPath
	if lp == nil {
		lp = exec.LookPath
	}
	return lp(bin)
}

func (r Runner) exec(ctx context.Context, bin string, args ...string) ([]byte, error) {
	if r.run != nil {
		return r.run(ctx, bin, args...)
	}
	// Output() captures stdout (the XML). Nmap exits non-zero on some partial
	// conditions but still emits valid XML; we surface exec errors to the caller.
	return exec.CommandContext(ctx, bin, args...).Output()
}
