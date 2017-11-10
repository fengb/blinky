package main

import (
	"github.com/Jguer/go-alpm"
	"github.com/fsnotify/fsnotify"
	"os"
)

type Package struct {
	alpm.Package
	Upgrade string
}

type Pac struct {
	snapshot []Package
	conf     alpm.PacmanConfig
	term     chan struct{}
}

func NewPac(filename string) (*Pac, error) {
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	conf, err := alpm.ParseConfig(reader)
	if err != nil {
		return nil, err
	}

	pac := Pac{conf: conf, term: make(chan struct{})}

	err = pac.startAutoUpdate(pac.term)
	if err != nil {
		pac.Close()
		return nil, err
	}
	return &pac, nil
}

func (p *Pac) Close() {
	if p.term == nil {
		return
	}

	close(p.term)
	p.term = nil
}

func (p *Pac) GetSnapshot() ([]Package, error) {
	if p.snapshot != nil {
		return p.snapshot, nil
	} else {
		return p.FetchSnapshot()
	}
}

func (p *Pac) FetchSnapshot() ([]Package, error) {
	handle, err := p.conf.CreateHandle()
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

	packages := make([]Package, 0)
	for _, alpmPkg := range localDb.PkgCache().Slice() {
		pkg := Package{Package: alpmPkg}
		newAlpmPkg := alpmPkg.NewVersion(syncDbs)
		if newAlpmPkg != nil {
			pkg.Upgrade = newAlpmPkg.Version()
		}
		packages = append(packages, pkg)
	}

	p.snapshot = packages
	return packages, nil
}

func (p *Pac) startAutoUpdate(term <-chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if p.conf.LogFile != "" {
		err = watcher.Add(p.conf.LogFile)
	} else {
		err = watcher.Add("/var/log/pacman.log")
	}
	if err != nil {
		watcher.Close()
		return err
	}

	debouncedFetch := NewDebounced(100, func() {
		_, err := p.FetchSnapshot()
		if err != nil {
			// How to handle?
		}
	})

	go func() {
		defer watcher.Close()

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					debouncedFetch.Call()
				}
			case err := <-watcher.Errors:
				// ???
				panic(err)
			case <-term:
				return
			}
		}
	}()

	return nil
}
