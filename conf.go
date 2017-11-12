package main

import (
	"gopkg.in/ini.v1"
	"html/template"
)

type Conf struct {
	Host  string
	Port  uint
	Index *template.Template
}

func LoadConfFile(filename string) (Conf, error) {
	var conf Conf
	cfg, err := ini.Load(filename)
	if err != nil {
		return conf, err
	}

	sec := cfg.Section("web")
	if sec.HasKey("host") {
		conf.Host = sec.Key("host").Value()
	}

	if sec.HasKey("port") {
		conf.Port, err = sec.Key("port").Uint()
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

	conf.Index, err = template.ParseFiles(dir + "/index.html.tmpl")
	if err != nil {
		return conf, err
	}

	return conf, err
}
