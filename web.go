package main

//go:generate "$GOPATH/bin/templify" -o index_template.go index.html.tmpl

import (
	"fmt"
	"net/http"
)

func Serve(conf Conf, pac *Pac) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		snapshot, err := pac.GetSnapshot()
		if err != nil {
			// 500
			return
		}

		err = conf.index.Execute(w, snapshot)
		if err != nil {
			// ???
		}
	})

	listenString := fmt.Sprintf("%s:%d", conf.host, conf.port)
	fmt.Println("Listening on", listenString)
	http.ListenAndServe(listenString, nil)
	return nil
}
