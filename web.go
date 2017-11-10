package main

//go:generate "$GOPATH/bin/templify" -o web_index.go index.html.tmpl

import (
	"html/template"
	"net/http"
)

func Serve(packageUpdate <-chan []Package) error {
	tmpl, err := template.New("index").Parse(web_indexTemplate())
	if err != nil {
		panic(err)
	}

	var packages []Package

	go func() {
		for newPackages := range packageUpdate {
			packages = newPackages
		}
		panic("Watch closed unexpectedly")
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, packages)
		if err != nil {
			// ???
		}
	})
	http.ListenAndServe(":9012", nil)
	return nil
}
