package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

type CmdError struct {
	*exec.ExitError
}

func (e *CmdError) Error() string {
	return fmt.Sprintf("%s\n%s", e.ExitError.Error(), e.Stderr)
}

func CmdRun(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	outbuff := &bytes.Buffer{}
	errbuff := &bytes.Buffer{}
	cmd.Stdout = outbuff
	cmd.Stderr = errbuff
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return outbuff.String(), errbuff.String(), err
	}
	exitErr.Stderr = errbuff.Bytes()
	return outbuff.String(), string(exitErr.Stderr), &CmdError{exitErr}
}
