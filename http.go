package main

import (
	"fmt"
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

	listenString := fmt.Sprintf("%s:%d", conf.Http.Host, conf.Http.Port)
	log.Println("Listening on", listenString)
	http.ListenAndServe(listenString, nil)
	return nil
}