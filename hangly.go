package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

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

	for pkgs := range watch {
		err = tmpl.Execute(os.Stdout, pkgs)
		if err != nil {
			panic(err)
		}
	}

	panic("Watch closed unexpectedly")

	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":8080", nil)
}
