//go:build windows

package main

import (
	"os"
	"os/exec"
)

// reexec starts a fresh copy of the process and exits the current one.
func reexec() error {
	cmd := exec.Command(executablePath(), os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
