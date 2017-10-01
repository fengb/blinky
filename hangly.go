package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	err := PacOutdated()
	if err != nil {
		panic(err)
	}
	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":8080", nil)
}
