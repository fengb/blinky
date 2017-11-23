package main

import (
	"gopkg.in/ini.v1"
	"html/template"
	"net"
	"time"
)

const hhmm = "15:04"

type Clock interface {
	Hour() int
	Minute() int
	Second() int
}

type ConfHttp struct {
	Listen string
	Index  *template.Template
}

type ConfRefresh struct {
	Enabled bool
	At      Clock
}

type Conf struct {
	Http    ConfHttp
	Refresh ConfRefresh
	Pac     *Pac

	dir     string
	watches []chan struct{}
}

func NewConf(dir string) (*Conf, error) {
	var err error
	conf := Conf{dir: dir}

	conf.Pac, err = NewPac("/etc/pacman.conf")
	if err != nil {
		return nil, err
	}

	err = conf.Reload()

	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func (c *Conf) Watch() <-chan struct{} {
	watch := make(chan struct{}, 1)
	c.watches = append(c.watches, watch)
	return watch
}

func (c *Conf) Reload() error {
	cfg, err := ini.Load(c.dir + "/blinky.conf")
	if err != nil {
		return err
	}

	c.Http, err = parseHttp(cfg.Section("http"), c.dir)
	if err != nil {
		return err
	}

	c.Refresh, err = parseRefresh(cfg.Section("refresh"))
	if err != nil {
		return err
	}

	for _, watch := range c.watches {
		watch <- struct{}{}
	}

	return nil
}

func parseHttp(sec *ini.Section, dir string) (ConfHttp, error) {
	c := ConfHttp{}
	var err error
	if sec.HasKey("listen") {
		c.Listen = sec.Key("listen").Value()
	}

	_, _, err = net.SplitHostPort(c.Listen)
	if err != nil {
		return c, err
	}

	c.Index, err = template.ParseFiles(dir + "/index.html.tmpl")
	if err != nil {
		return c, err
	}

	return c, nil
}

func parseRefresh(sec *ini.Section) (ConfRefresh, error) {
	c := ConfRefresh{}
	var err error
	if sec.HasKey("enabled") {
		c.Enabled, err = sec.Key("enabled").Bool()
		if err != nil {
			return c, err
		}
	}

	if !sec.HasKey("at") {
		c.At, err = time.Parse(hhmm, "02:30")
	} else {
		c.At, err = sec.Key("at").TimeFormat(hhmm)
		if err != nil {
			return c, err
		}
	}

	return c, nil
}
