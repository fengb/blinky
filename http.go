package main

import (
	"context"
	"log"
	"net/http"
)

type Http struct {
	conf          *Conf
	snapshotState *SnapshotState
	srv           *http.Server
}

func NewHttp(conf *Conf, snapshotState *SnapshotState) (*Http, error) {
	h := Http{conf: conf, snapshotState: snapshotState, srv: &http.Server{Addr: conf.Http.Addr}}

	http.HandleFunc("/", h.Index)

	go func() {
		log.Println("Listening to", conf.Http.Addr)
		err := h.srv.ListenAndServe()
		log.Println(err)
	}()

	return &h, nil
}

func (h *Http) Close() error {
	return h.srv.Shutdown(context.TODO())
}

func (h *Http) Index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" || r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	err := h.conf.Http.Index.Execute(w, h.snapshotState)
	if err != nil {
		log.Println(err)
	}
}
