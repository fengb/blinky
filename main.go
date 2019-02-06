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
	Close() error
}

func InitApp(confDir string, snapshotState *SnapshotState) ([]Actor, error) {
	conf, err := NewConf(ConfDir)
	if err != nil {
		panic(err)
	}

	var pacman, multicast, http Actor
	err = Parallel(
		func() (err error) {
			pacman, err = NewPacman(conf, snapshotState)
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
		return nil, err
	}

	return []Actor{pacman, multicast, http}, nil
}

func main() {
	if ConfDir == "" {
		ConfDir = "static_files/etc/blinky"
	}

	snapshotState := NewSnapshotState()

	actors, err := InitApp(ConfDir, snapshotState)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	for _ = range c {
		log.Printf("HUP received - reloading")

		newActors, err := InitApp(ConfDir, snapshotState)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, actor := range actors {
			actor.Close()
		}

		actors = newActors
	}
}
