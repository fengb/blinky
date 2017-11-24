package main

import (
	"bytes"
	"log"
	"os/exec"
	"time"
)

var root = []byte("root")

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

func Refresher(conf *Conf) {
	if !allowsPacmanSync() {
		log.Println("Cannot run pacman -S. Skipping refresh")
		return
	}

	confUpdate := conf.Watch()
	timer := time.NewTimer(1 * time.Hour)

	setupTimer := func() {
		conf.Lock()
		defer conf.Unlock()

		if conf.Refresh.Enabled {
			next := nextExecution(conf.Refresh.At)
			log.Println("Next refresh:", next)
			timer.Reset(time.Until(next))
		} else {
			timer.Stop()
		}
	}

	setupTimer()

	for {
		select {
		case <-confUpdate:
			timer.Stop()
			setupTimer()
		case <-timer.C:
			timer.Stop()
			runRefresh()
			setupTimer()
		}
	}
}
