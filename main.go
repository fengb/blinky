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
		ConfDir = "static_files/etc/blinky"
	}

	conf, err := NewConf(ConfDir)
	if err != nil {
		panic(err)
	}

	local, err := NewLocal(conf)
	if err != nil {
		panic(err)
	}

	multicast, err := NewMulticast(conf)
	if err != nil {
		panic(err)
	}

	http, err := NewHttp(conf, local, multicast)
	if err != nil {
		panic(err)
	}

	watchSignals(local, multicast, http)
}
