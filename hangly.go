package main

import (
	"html/template"
	"net/http"
)

func main() {
	tmpl, err := template.ParseFiles("hangly.html.tmpl")
	if err != nil {
		panic(err)
	}

	pac, err := NewPac("/etc/pacman.conf")
	if err != nil {
		panic(err)
	}

	watch, err := pac.Watch()
	if err != nil {
		panic(err)
	}

	pkgs, err := pac.GetPackages()
	if err != nil {
		panic(err)
	}

	go func() {
		for newPkgs := range watch {
			pkgs = newPkgs
		}
		panic("Watch closed unexpectedly")
	}()


	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, pkgs)
		if err != nil {
			// ???
		}
	})
	http.ListenAndServe(":9012", nil)
}
