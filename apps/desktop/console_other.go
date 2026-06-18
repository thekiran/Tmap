//go:build !windows

package main

import "os/exec"

// hideConsole is a no-op on non-Windows platforms, where spawning a child
// process does not create a separate terminal window.
func hideConsole(cmd *exec.Cmd) {}
