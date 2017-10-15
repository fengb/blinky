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

	outdated, err := PacOutdated("/etc/pacman.conf")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(os.Stdout, outdated)
	if err != nil {
		panic(err)
	}

	testFunc := NewDebounced(100, func() {
		fmt.Println("Hello!")
	})
	testFunc.Call()
	testFunc.Call()
	testFunc.Call()
	testFunc.Call()
	<-testFunc.Drain()
	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":8080", nil)
}
