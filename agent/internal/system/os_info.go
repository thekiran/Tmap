// Package system wraps OS-specific concerns: running external commands and the
// per-platform traceroute. Keeping the platform forks here means the probe and
// detection layers stay cross-platform and testable.
package system

import "runtime"

// OSInfo describes the host the agent is running on.
type OSInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// Info returns the current host OS/arch.
func Info() OSInfo {
	return OSInfo{OS: runtime.GOOS, Arch: runtime.GOARCH}
}
