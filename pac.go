package main

import (
	"errors"
	"github.com/Jguer/go-alpm"
	"github.com/fsnotify/fsnotify"
	"os"
)

type Package struct {
	alpm.Package
	Upgrade string
}

type Pac struct {
	Packages []Package
	conf     alpm.PacmanConfig
	watch    chan []Package
	unwatch  chan struct{}
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

	pac := Pac{conf: conf, unwatch: make(chan struct{})}
	return &pac, nil
}

func (p *Pac) GetPackages() ([]Package, error) {
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

	p.Packages = packages
	return packages, nil
}

func (p *Pac) Watch() (chan []Package, error) {
	if p.watch != nil {
		return nil, errors.New("already watching")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if p.conf.LogFile != "" {
		err = watcher.Add(p.conf.LogFile)
	} else {
		err = watcher.Add("/var/log/pacman.log")
	}
	if err != nil {
		return nil, err
	}

	p.watch = make(chan []Package)

	debouncedPackages := NewDebounced(1000, func() {
		pkgs, err := p.GetPackages()
		if err != nil {
			// How to handle?
		}
		p.watch <- pkgs
	})

	go func() {
		defer func() {
			watcher.Close()
			close(p.watch)
			p.watch = nil
		}()

		debouncedPackages.CallImmediate()

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					debouncedPackages.Call()
				}
			case <-watcher.Errors:
				// File disappeared? Job completed I suppose...
				return
			case <-p.unwatch:
				return
			}
		}
	}()

	return p.watch, nil
}

func (p *Pac) Unwatch() error {
	if p.watch == nil {
		return errors.New("not watching")
	}

	p.unwatch <- struct{}{}
	return nil
}
