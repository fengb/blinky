package main

import (
	"log"
	"net/http"
)

func Serve(conf *Conf) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		snapshot, err := conf.Pac.GetSnapshot()
		if err != nil {
			// 500
			return
		}

		err = conf.Http.Index.Execute(w, snapshot)
		if err != nil {
			// ???
		}
	})

	log.Println("Listening on", conf.Http.Listen)
	http.ListenAndServe(conf.Http.Listen, nil)
	return nil
}
