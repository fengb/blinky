package main

import (
	"gopkg.in/ini.v1"
	"html/template"
)

type Conf struct {
	Web struct {
		Host string
		Port uint
	}

	Templates struct {
		Index *template.Template
	}
}

func LoadConfFile(filename string) (Conf, error) {
	var conf Conf
	cfg, err := ini.Load(filename)
	if err != nil {
		return conf, err
	}

	sec := cfg.Section("web")
	if sec.HasKey("host") {
		conf.Web.Host = sec.Key("host").Value()
	}

	if sec.HasKey("port") {
		conf.Web.Port, err = sec.Key("port").Uint()
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

	conf.Templates.Index, err = template.ParseFiles(dir + "/index.html.tmpl")
	if err != nil {
		return conf, err
	}

	return conf, err
}
