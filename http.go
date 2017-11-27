package main

import (
	"log"
	"net/http"
	"time"
)

type Http struct {
	conf *Conf
	srv  *http.Server
}

func NewHttp(conf *Conf) (Actor, error) {
	h := &Http{conf: conf}

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
		// ???
	}
}

func (h *Http) UpdateConf(conf *Conf) error {
	if h.srv == nil || h.srv.Addr != conf.Http.Addr {
		srv, err := startServer(conf.Http.Addr)
		if err != nil {
			return err
		}

		if h.srv != nil {
			go func() {
				// TODO: this doesn't stop listening...
				err := h.srv.Shutdown(nil)
				if err != nil {
					log.Println(err)
				}
			}()
		}
		h.srv = srv
	}

	h.conf = conf

	return nil
}

func startServer(addr string) (*http.Server, error) {
	srv := &http.Server{Addr: addr}
	log.Println("Listening on", addr)
	listenError := make(chan error)
	go func() {
		listenError <- srv.ListenAndServe()
	}()

	select {
	case err := <-listenError:
		return nil, err
	case <-time.After(1 * time.Second):
		// TODO: listen on actual success
		return srv, nil
	}
}
