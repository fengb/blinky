package main

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"time"
)

var root = []byte("root")

type Refresh struct {
	conf  *Conf
	timer *time.Timer
}

func NewRefresh(conf *Conf) (Actor, error) {
	r := Refresh{timer: time.NewTimer(1 * time.Hour)}

	err := r.UpdateConf(conf)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			<-r.timer.C
			runRefresh()
			r.scheduleNext()
		}
	}()

	return &r, nil
}

func (r *Refresh) UpdateConf(conf *Conf) error {
	if conf.Refresh.Enable && !allowsPacmanSync() {
		return errors.New("Cannot run pacman -S")
	}

	r.conf = conf
	r.scheduleNext()

	return nil
}

func (r *Refresh) scheduleNext() {
	r.timer.Stop()
	if r.conf.Refresh.Enable {
		next := nextExecution(r.conf.Refresh.At)
		log.Println("Next refresh:", next)
		r.timer.Reset(time.Until(next))
	}
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
