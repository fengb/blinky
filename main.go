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

func Signal(conf *Conf) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	for _ = range c {
		log.Printf("HUP received - reloading")
		err := conf.Reload()
		if err != nil {
			log.Println(err)
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

	go Signal(conf)
	go Refresher(conf)
	go Serve(conf)
	select {}
}
