package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// linker constants
var (
	ConfDir string
	Version string
)

type Actor interface {
	UpdateConf(conf *Conf) error
}

func watchSignals(actors ...Actor) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	for _ = range c {
		log.Printf("HUP received - reloading")

		conf, err := NewConf(ConfDir)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, actor := range actors {
			err = actor.UpdateConf(conf)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func main() {
	if ConfDir == "" {
		ConfDir = "etc"
	}

	snapshotState := NewSnapshotState()

	conf, err := NewConf(ConfDir)
	if err != nil {
		panic(err)
	}

	var pac, refresh, multicast, http Actor
	err = Parallel(
		func() (err error) {
			pac, err = NewPac("/etc/pacman.conf", snapshotState)
			return err
		},
		func() (err error) {
			refresh, err = NewRefresh(conf)
			return err
		},
		func() (err error) {
			multicast, err = NewMulticast(conf, snapshotState)
			return err
		},
		func() (err error) {
			http, err = NewHttp(conf, snapshotState)
			return err
		},
	)

	if err != nil {
		panic(err)
	}

	watchSignals(pac, refresh, http, multicast)
}
