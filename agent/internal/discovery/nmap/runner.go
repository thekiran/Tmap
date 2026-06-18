package nmap

import (
	"context"
	"os/exec"
)

// Runner executes Nmap. The zero value is usable and looks up "nmap" in PATH.
type Runner struct {
	Binary string

	lookPath func(string) (string, error)
	run      func(ctx context.Context, bin string, args ...string) ([]byte, error)
}

// Available reports whether Nmap is installed. Callers should skip Nmap
// gracefully when false; Nmap is always optional.
func (r Runner) Available() bool {
	_, err := r.lookup()
	return err == nil
}

// Detect returns the binary path and version when available.
func (r Runner) Detect(ctx context.Context) Detection {
	return Detector{Binary: r.Binary, LookPath: r.lookPath, Run: r.run}.Detect(ctx)
}

// Scan runs Nmap against target with the given profile, requests XML on stdout,
// and parses only that XML.
func (r Runner) Scan(ctx context.Context, target, profile string) ([]Host, error) {
	bin, err := r.lookup()
	if err != nil {
		return nil, ErrNotInstalled
	}
	args := append(ArgsForProfile(profile), target)
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
	lookPath := r.lookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	return lookPath(bin)
}

func (r Runner) exec(ctx context.Context, bin string, args ...string) ([]byte, error) {
	if r.run != nil {
		return r.run(ctx, bin, args...)
	}
	return exec.CommandContext(ctx, bin, args...).Output()
}
