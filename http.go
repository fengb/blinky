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

	confWatch := conf.Watch()

	listenServer := func(addr string) http.Server {
		srv := http.Server{Addr: addr}
		go func() {
			log.Println("Listening on", conf.Http.Listen)
			err := srv.ListenAndServe()
			if err != nil {
				panic(err)
			}
		}()
		return srv
	}

	srv := listenServer(conf.Http.Listen)
	for _ = range confWatch {
		if conf.Http.Listen != srv.Addr {
			// TODO: why doesn't this stop listening?
			err := srv.Shutdown(nil)
			if err != nil {
				panic(err)
			}

			srv = listenServer(conf.Http.Listen)
		}
	}
}
