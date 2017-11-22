package main

import (
	"gopkg.in/ini.v1"
	"html/template"
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
		Host  string
		Port  uint
		Index *template.Template
	}

	Refresh struct {
		Enabled bool
		At      Clock
	}

	Pac *Pac

	dir string
}

func (c *Conf) parseHttp(sec *ini.Section) error {
	var err error
	if sec.HasKey("host") {
		c.Http.Host = sec.Key("host").Value()
	}

	if !sec.HasKey("port") {
		c.Http.Port = 9012
	} else {
		c.Http.Port, err = sec.Key("port").Uint()
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

	conf.Pac, err = NewPac("/etc/pacman.conf")
	if err != nil {
		return nil, err
	}

	return &conf, err
}
