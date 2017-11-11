package main

import (
	"html/template"
	"os"
	"strconv"
)

//go:generate "$GOPATH/bin/templify" -o index_template.go index.html.tmpl

type Conf struct {
	host  string
	port  int
	index *template.Template
}

func tryTemplates(filename string, genTemplate func() string) (*template.Template, error) {
	tmpl, err := template.ParseFiles(filename)
	if err == nil {
		return tmpl, err
	}

	return template.New("index").Parse(genTemplate())
}

func DefaultConf() (Conf, error) {
	var err error
	conf := Conf{host: os.Getenv("HOST"), port: 9012}

	port := os.Getenv("PORT")
	if port != "" {
		conf.port, err = strconv.Atoi(port)
		if err != nil {
			return conf, err
		}
	}

	conf.index, err = tryTemplates("/etc/blinky/index.html.tmpl", index_templateTemplate)

	return conf, err
}
