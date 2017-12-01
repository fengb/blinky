package main

import (
	"log"
	"net"
	"net/http"
)

type Http struct {
	conf     *Conf
	srv      *http.Server
	listener net.Listener
}

func NewHttp(conf *Conf) (Actor, error) {
	h := &Http{conf: conf, srv: &http.Server{}}

	http.HandleFunc("/", h.Index)

	err := h.UpdateConf(conf)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Http) Index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" || r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	snapshot, err := h.conf.Pac.GetSnapshot()
	if err != nil {
		// 500
		return
	}

	err = h.conf.Http.Index.Execute(w, snapshot)
	if err != nil {
		log.Println(err)
	}
}

func (h *Http) UpdateConf(conf *Conf) error {
	if h.listener == nil || h.conf.Http.Addr != conf.Http.Addr {
		listener, err := net.Listen("tcp", conf.Http.Addr)
		if err != nil {
			return err
		}

		if h.listener != nil {
			err = h.listener.Close()
			if err != nil {
				log.Println(err)
			}
		}

		h.listener = listener
		log.Println("Listening to", conf.Http.Addr)
		go func() {
			err := h.srv.Serve(listener)
			log.Println(err)
		}()
	}

	h.conf = conf

	return nil
}
