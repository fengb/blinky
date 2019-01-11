package main

import (
	"errors"
	"log"
	"strings"
	"time"
)

type Refresh struct {
	conf   *Conf
	ticker *DailyTicker
}

func NewRefresh(conf *Conf) (Actor, error) {
	r := Refresh{ticker: NewDailyTicker(time.Now())}

	err := r.UpdateConf(conf)
	if err != nil {
		return nil, err
	}

	go func() {
		for _ = range r.ticker.C {
			log.Println("Running pacman --noconfirm -Syuwq")
			_, _, err := CmdRun("pacman", "--noconfirm", "-Syuwq")
			if err != nil {
				log.Println(err)
			}
		}

	}()

	return &r, nil
}

func (r *Refresh) UpdateConf(conf *Conf) error {
	if conf.Refresh.Enable && !allowsPacmanSync() {
		return errors.New("Cannot run pacman -S")
	}

	r.conf = conf

	r.ticker.Reset(conf.Refresh.At)
	log.Println("Next refresh:", r.ticker.NextRun())

	return nil
}

func allowsPacmanSync() bool {
	// Ideally use an exit status but all errors are 1
	_, stderr, _ := CmdRun("pacman", "-S")
	return !strings.Contains(stderr, "root")
}
