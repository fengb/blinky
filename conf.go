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
}

func LoadConfFile(filename string) (Conf, error) {
	var conf Conf
	cfg, err := ini.Load(filename)
	if err != nil {
		return conf, err
	}

	sec := cfg.Section("http")
	if sec.HasKey("host") {
		conf.Http.Host = sec.Key("host").Value()
	}

	if !sec.HasKey("port") {
		conf.Http.Port = 9012
	} else {
		conf.Http.Port, err = sec.Key("port").Uint()
		if err != nil {
			return conf, err
		}
	}

	sec = cfg.Section("refresh")
	if sec.HasKey("enabled") {
		conf.Refresh.Enabled, err = sec.Key("enabled").Bool()
		if err != nil {
			return conf, err
		}
	}

	if !sec.HasKey("at") {
		conf.Refresh.At, err = time.Parse(hhmm, "02:30")
	} else {
		conf.Refresh.At, err = sec.Key("at").TimeFormat(hhmm)
		if err != nil {
			return conf, err
		}
	}

	return conf, err
}

func LoadConfDir(dir string) (Conf, error) {
	conf, err := LoadConfFile(dir + "/blinky.conf")
	if err != nil {
		return conf, err
	}

	conf.Http.Index, err = template.ParseFiles(dir + "/index.html.tmpl")
	if err != nil {
		return conf, err
	}

	conf.Pac, err = NewPac("/etc/pacman.conf")
	if err != nil {
		return conf, err
	}

	return conf, err
}
