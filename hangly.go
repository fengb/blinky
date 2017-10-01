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

func alpmToy() error {
	reader, err := os.Open("/etc/pacman.conf")
	if err != nil {
		return err
	}

	conf, err := alpm.ParseConfig(reader)
	if err != nil {
		return err
	}

	handle, err := conf.CreateHandle()
	if err != nil {
		return err
	}

	localDb, err := handle.LocalDb()
	if err != nil {
		return err
	}

	syncDbs, err := handle.SyncDbs()
	if err != nil {
		return err
	}

	for _, pkg := range localDb.PkgCache().Slice() {
		newPkg := pkg.NewVersion(syncDbs)
		if newPkg != nil {
			fmt.Printf("%s %s -> %s\n", pkg.Name(), pkg.Version(), newPkg.Version())
		}
	}

	return nil
}

func main() {
	err := alpmToy()
	if err != nil {
		panic(err)
	}
	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":8080", nil)
}
