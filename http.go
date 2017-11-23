package main

import (
	"log"
	"net/http"
)

type Http struct {
	conf *Conf
}

func NewHttp(conf *Conf) (Actor, error) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

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

	go func() {
		log.Println("Listening on", conf.Http.Listen)
		http.ListenAndServe(conf.Http.Listen, nil)
	}()
	return &Http{conf}, nil
}

func (h *Http) UpdateConf(conf *Conf) error {
	return nil
}
