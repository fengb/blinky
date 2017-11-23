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

	conf, err := NewConf(ConfDir)
	if err != nil {
		panic(err)
	}

	refresh, err := NewRefresh(conf)
	if err != nil {
		panic(err)
	}
	http, err := NewHttp(conf)
	if err != nil {
		panic(err)
	}
	watchSignals(refresh, http)
}
