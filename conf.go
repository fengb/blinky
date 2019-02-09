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
	Pacman struct {
		Refresh Clock
		Conf    struct {
			DBPath  string
			LogFile string
		}
	}

	Http struct {
		Addr  string
		Index *template.Template
	}

	Multicast struct {
		Listen bool
		Send   time.Duration
		Addr   string
		Secret []byte
	}

	dir string
}

func NewConf(dir string) (*Conf, error) {
	conf := Conf{dir: dir}
	cfg, err := ini.Load(conf.dir + "/blinky.conf")
	if err != nil {
		return nil, err
	}

	err = Parallel(
		func() error { return conf.parsePacman(cfg.Section("pacman")) },
		func() error { return conf.parseHttp(cfg.Section("http")) },
		func() error { return conf.parseMulticast(cfg.Section("multicast")) },
	)
	if err != nil {
		return nil, err
	}

	return &conf, err
}

func (c *Conf) parsePacman(sec *ini.Section) (err error) {
	if sec.HasKey("refresh") {
		c.Pacman.Refresh, err = sec.Key("refresh").TimeFormat(hhmm)
		if err != nil {
			return err
		}
	}

	conf := "/etc/pacman.conf"
	if sec.HasKey("conf") {
		conf = sec.Key("conf").Value()
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{SkipUnrecognizableLines: true}, conf)
	if err != nil {
		return err
	}

	opts := cfg.Section("options")
	if opts.HasKey("DBPath") {
		c.Pacman.Conf.DBPath = opts.Key("DBPath").Value()
	} else {
		c.Pacman.Conf.DBPath = "/var/lib/pacman"
	}
	if opts.HasKey("LogFile") {
		c.Pacman.Conf.LogFile = opts.Key("LogFile").Value()
	} else {
		c.Pacman.Conf.LogFile = "/var/log/pacman.log"
	}

	return nil
}

func (c *Conf) parseHttp(sec *ini.Section) (err error) {
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

func (c *Conf) parseMulticast(sec *ini.Section) (err error) {
	if sec.HasKey("listen") {
		c.Multicast.Listen, err = sec.Key("listen").Bool()
		if err != nil {
			return err
		}
	}

	if sec.HasKey("send") {
		c.Multicast.Send, err = sec.Key("send").Duration()
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

	if sec.HasKey("secret") {
		c.Multicast.Secret = []byte(sec.Key("secret").Value())
	}

	return nil
}
