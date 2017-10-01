package main

import (
	"fmt"
	"github.com/Jguer/go-alpm"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func alpmToy() {
	handle, err := alpm.Init("/", "/var/lib/pacman")
	if err != nil {
		panic(err)
	}

	localDb, err := handle.LocalDb()
	if err != nil {
		panic(err)
	}

	syncDbs, err := handle.SyncDbs()
	if err != nil {
		panic(err)
	}
	syncDbs.ForEach(func(db alpm.Db) error {
		fmt.Printf("%s\n", db.Name())
		return nil
	})

	for _, pkg := range localDb.PkgCache().Slice() {
		newPkg := pkg.NewVersion(syncDbs)
		if newPkg != nil {
			fmt.Printf("%s %s -> %s\n", pkg.Name(), pkg.Version(), newPkg.Version())
		} else {
			fmt.Printf("%s %s\n", pkg.Name(), pkg.Version())
		}
	}
}

func main() {
	alpmToy()
	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":8080", nil)
}
