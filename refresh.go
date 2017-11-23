package main

import (
	"bytes"
	"log"
	"os/exec"
	"time"
)

var root = []byte("root")

type Refresh struct {
	conf *Conf
}

func NewRefresh(conf *Conf) (Actor, error) {
	if !conf.Refresh.Enabled {
		//return
	}

	if !allowsPacmanSync() {
		log.Println("Cannot run pacman -S. Skipping refresh")
		//return
	}

	go func() {
		for {
			next := nextExecution(conf.Refresh.At)
			log.Println("Next refresh:", next)
			time.Sleep(time.Until(next))
			runRefresh()
		}
	}()

	return &Refresh{conf}, nil
}

func (r *Refresh) UpdateConf(conf *Conf) error {
	return nil
}

func allowsPacmanSync() bool {
	// Ideally use an exit status but all errors are 1
	cmd := exec.Command("pacman", "-S")
	buf := &bytes.Buffer{}
	cmd.Stderr = buf
	cmd.Run()
	return !bytes.Contains(buf.Bytes(), root)
}

func runRefresh() {
	log.Println("Running pacman --noconfirm -Syuwq")
	cmd := exec.Command("pacman", "--noconfirm", "-Syuwq")
	buf := &bytes.Buffer{}
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		log.Println(buf.String())
	}
}

func nextExecution(target Clock) time.Time {
	now := time.Now()
	targetTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		target.Hour(),
		target.Minute(),
		0,
		0,
		now.Location(),
	)
	if now.Before(targetTime) {
		return targetTime
	} else {
		return targetTime.Add(24 * time.Hour)
	}
}
