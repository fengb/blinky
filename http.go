package main

import (
	"log"
	"net/http"
)

func Serve(conf *Conf) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		conf.Lock()
		pac := conf.Pac
		index := conf.Http.Index
		conf.Unlock()

		snapshot, err := pac.GetSnapshot()
		if err != nil {
			// 500
			return
		}

		err = index.Execute(w, snapshot)
		if err != nil {
			// ???
		}
	})

	confWatch := conf.Watch()

	listenServer := func(addr string) http.Server {
		srv := http.Server{Addr: addr}
		go func() {
			log.Println("Listening on", addr)
			err := srv.ListenAndServe()
			if err != nil {
				panic(err)
			}
		}()
		return srv
	}

	srv := listenServer(conf.Http.Listen)
	for _ = range confWatch {
		conf.Lock()
		newAddr := conf.Http.Listen
		conf.Unlock()

		if newAddr != srv.Addr {
			// TODO: why doesn't this stop listening?
			err := srv.Shutdown(nil)
			if err != nil {
				panic(err)
			}

			srv = listenServer(newAddr)
		}
	}
}
