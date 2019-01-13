package main

import (
	"bytes"
	"os/exec"
	"strings"
)

type CmdError struct {
	*exec.ExitError
}

func (e *CmdError) Error() string {
	return e.ExitError.Error() + "\n" + string(e.Stderr)
}

func CmdRun(name string, args ...string) (string, string, error) {
	stdout, stderr, err := CmdRunBytes(name, args...)
	// Posix programs output a newline at the end. This should not be a part of the typical API.
	return strings.TrimSuffix(string(stdout), "\n"), strings.TrimSuffix(string(stderr), "\n"), err
}

func CmdRunBytes(name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.Command(name, args...)
	outbuff := &bytes.Buffer{}
	errbuff := &bytes.Buffer{}
	cmd.Stdout = outbuff
	cmd.Stderr = errbuff
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return outbuff.Bytes(), errbuff.Bytes(), err
	}
	exitErr.Stderr = errbuff.Bytes()
	return outbuff.Bytes(), exitErr.Stderr, &CmdError{exitErr}
}
