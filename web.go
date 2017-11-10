package main

//go:generate "$GOPATH/bin/templify" -o web_index.go index.html.tmpl

import "net/http"

func Serve(conf *Conf, pac *Pac) error {
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
	http.ListenAndServe(":9012", nil)
	return nil
}
