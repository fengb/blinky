package main

import (
	"github.com/Jguer/go-alpm"
	"os"
)

type Outdated struct {
	Installed alpm.Package
	Latest    *alpm.Package
}

func PacOutdated() ([]Outdated, error) {
	reader, err := os.Open("/etc/pacman.conf")
	if err != nil {
		return nil, err
	}

	conf, err := alpm.ParseConfig(reader)
	if err != nil {
		return nil, err
	}

	handle, err := conf.CreateHandle()
	if err != nil {
		return nil, err
	}

	localDb, err := handle.LocalDb()
	if err != nil {
		return nil, err
	}

	syncDbs, err := handle.SyncDbs()
	if err != nil {
		return nil, err
	}

	outdated := make([]Outdated, 0)
	for _, pkg := range localDb.PkgCache().Slice() {
		newPkg := pkg.NewVersion(syncDbs)
		if newPkg != nil {
			outdated = append(outdated, Outdated{pkg, newPkg})
		}
	}

	return outdated, nil
}
