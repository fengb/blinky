package main

import (
	"html/template"
	"net/http"
)

func Serve(packageUpdate <-chan []Package) error {
	tmpl, err := template.ParseFiles("blinky.html.tmpl")
	if err != nil {
		return err
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
