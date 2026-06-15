package system

import (
	"bytes"
	"context"
	"os/exec"
)

// Run executes an external command bounded by ctx and returns its combined
// stdout+stderr. Both streams are merged because the parsers downstream only
// care about the human-readable text, and some tools (tracert) print partial
// useful output even on a non-zero exit.
func Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}
