package main

import (
	"log"
	"os"
	"os/exec"
	"time"
)

func runRefresh() error {
	cmd := exec.Command("pacman", "--noconfirm", "-Syuwq")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func nextExecution(target time.Time) time.Time {
	now := time.Now()
	todayWithTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		target.Hour(),
		target.Minute(),
		0,
		0,
		now.Location(),
	)
	return todayWithTime.Add(24 * time.Hour)
}

func Refresher(conf Conf) {
	if !conf.Refresh.Enabled {
		return
	}
	err := runRefresh()
	if err != nil {
		log.Print(err)
	}

	for {
		next := nextExecution(conf.Refresh.At)
		time.Sleep(time.Until(next))

		err := runRefresh()
		if err != nil {
			log.Print(err)
		}
	}
}
