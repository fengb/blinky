package main

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

		err = conf.Templates.Index.Execute(w, snapshot)
		if err != nil {
			// ???
		}
	})

	listenString := fmt.Sprintf("%s:%d", conf.Web.Host, conf.Web.Port)
	fmt.Println("Listening on", listenString)
	http.ListenAndServe(listenString, nil)
	return nil
}
