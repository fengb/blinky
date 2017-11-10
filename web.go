package main

//go:generate "$GOPATH/bin/templify" -o web_index.go index.html.tmpl

import "net/http"

func Serve(conf *Conf, packageUpdate <-chan []Package) error {
	var packages []Package

	go func() {
		for newPackages := range packageUpdate {
			packages = newPackages
		}
		panic("Watch closed unexpectedly")
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := conf.index.Execute(w, packages)
		if err != nil {
			// ???
		}
	})
	http.ListenAndServe(":9012", nil)
	return nil
}
