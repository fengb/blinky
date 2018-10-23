package main

import (
	"bytes"
	"os/exec"
)

func CmdRun(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	outbuff := &bytes.Buffer{}
	errbuff := &bytes.Buffer{}
	cmd.Stdout = outbuff
	cmd.Stderr = errbuff
	err := cmd.Run()
	return outbuff.String(), errbuff.String(), err
}
