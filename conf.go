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

type Conf struct {
	Http struct {
		Addr  string
		Index *template.Template
	}

	Refresh struct {
		Enabled bool
		At      Clock
	}

	Multicast struct {
		Addr   string
		Listen bool
		Ping   time.Duration
	}

	Pac *Pac

	dir string
}

func NewConf(dir string) (*Conf, error) {
	conf := Conf{dir: dir}
	cfg, err := ini.Load(conf.dir + "/blinky.conf")
	if err != nil {
		return nil, err
	}

	err = conf.parseHttp(cfg.Section("http"))
	if err != nil {
		return nil, err
	}

	err = conf.parseRefresh(cfg.Section("refresh"))
	if err != nil {
		return nil, err
	}

	err = conf.parseMulticast(cfg.Section("multicast"))
	if err != nil {
		return nil, err
	}

	conf.Pac, err = NewPac("/etc/pacman.conf")
	if err != nil {
		return nil, err
	}

	return &conf, err
}

func (c *Conf) Close() error {
	return c.Pac.Close()
}

func (c *Conf) parseHttp(sec *ini.Section) error {
	var err error
	if sec.HasKey("addr") {
		c.Http.Addr = sec.Key("addr").String()

		_, _, err = net.SplitHostPort(c.Http.Addr)
		if err != nil {
			return err
		}
	}

	c.Http.Index, err = template.ParseFiles(c.dir + "/index.html.tmpl")
	if err != nil {
		return err
	}

	return nil
}

func (c *Conf) parseRefresh(sec *ini.Section) error {
	var err error
	if sec.HasKey("enabled") {
		c.Refresh.Enabled, err = sec.Key("enabled").Bool()
		if err != nil {
			return err
		}
	}

	if !sec.HasKey("at") {
		c.Refresh.At, err = time.Parse(hhmm, "02:30")
	} else {
		c.Refresh.At, err = sec.Key("at").TimeFormat(hhmm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Conf) parseMulticast(sec *ini.Section) error {
	var err error
	if sec.HasKey("listen") {
		c.Multicast.Listen, err = sec.Key("listen").Bool()
		if err != nil {
			return err
		}
	}

	if sec.HasKey("ping") {
		c.Multicast.Ping, err = sec.Key("ping").Duration()
		if err != nil {
			return err
		}
	}

	if sec.HasKey("addr") {
		c.Multicast.Addr = sec.Key("addr").String()

		_, _, err = net.SplitHostPort(c.Multicast.Addr)
		if err != nil {
			return err
		}
	}

	return nil
}
