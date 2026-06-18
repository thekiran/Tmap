//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// createNoWindow is the Windows CREATE_NO_WINDOW process-creation flag. It tells
// the OS not to allocate a console for the child process, so spawning the
// console-mode iad-agent from the GUI app does not flash a terminal window.
const createNoWindow = 0x08000000

// hideConsole hides the console window of the spawned child process on Windows.
func hideConsole(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}
