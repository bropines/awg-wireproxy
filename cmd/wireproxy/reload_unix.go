//go:build !windows

package main

import (
	"os"
	"syscall"
)

// reexec replaces the running process with a fresh copy of itself, which
// re-reads the configuration. Listening sockets are close-on-exec, so the new
// process can bind the same ports.
func reexec() error {
	return syscall.Exec(executablePath(), os.Args, os.Environ())
}
