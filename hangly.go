package main

import (
	"fmt"
	"github.com/Jguer/go-alpm"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func alpmToy() {
	reader, err := os.Open("/etc/pacman.conf")
	if err != nil {
		panic(err)
	}

	conf, err := alpm.ParseConfig(reader)
	if err != nil {
		panic(err)
	}

	handle, err := conf.CreateHandle()
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

	for _, pkg := range localDb.PkgCache().Slice() {
		newPkg := pkg.NewVersion(syncDbs)
		if newPkg != nil {
			fmt.Printf("%s %s -> %s\n", pkg.Name(), pkg.Version(), newPkg.Version())
		}
	}
}

func main() {
	alpmToy()
	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":8080", nil)
}
