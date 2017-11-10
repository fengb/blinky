package main

import "html/template"

//go:generate "$GOPATH/bin/templify" -o index_template.go index.html.tmpl

type Conf struct {
	index *template.Template
}

func tryTemplates(filename string, genTemplate func() string) (*template.Template, error) {
	tmpl, err := template.ParseFiles(filename)
	if err == nil {
		return tmpl, err
	}

	return template.New("index").Parse(genTemplate())
}

func DefaultConf() (*Conf, error) {
	tmpl, err := tryTemplates("/etc/blinky/index.html.tmpl", index_templateTemplate)
	if err != nil {
		return nil, err
	}
	return &Conf{tmpl}, nil
}
