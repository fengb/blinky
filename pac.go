package main

import (
	"bufio"
	"github.com/Jguer/go-alpm"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"strings"
	"time"
)

type Package struct {
	alpm.Package
	Upgrade string
}

type Snapshot struct {
	LastSync time.Time
	Packages []Package
}

type Pac struct {
	snapshot *Snapshot
	conf     alpm.PacmanConfig
	logfile  string
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

	logfile := conf.LogFile
	if logfile == "" {
		logfile = "/var/log/pacman.log"
	}

	pac := Pac{conf: conf, logfile: logfile, term: make(chan struct{})}

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

func (p *Pac) GetSnapshot() (*Snapshot, error) {
	var err error
	if p.snapshot == nil {
		err = p.UpdateSnapshot()
	}

	return p.snapshot, err
}

func (p *Pac) UpdateSnapshot() error {
	var snapshot Snapshot

	var syncError error
	done := make(chan struct{}, 2)

	go func() {
		snapshot.LastSync, syncError = p.FetchLastSync()
		done <- struct{}{}
	}()

	var packageError error
	go func() {
		snapshot.Packages, packageError = p.FetchPackages()
		done <- struct{}{}
	}()

	<-done
	<-done

	if syncError != nil {
		return syncError
	}

	if packageError != nil {
		return packageError
	}

	p.snapshot = &snapshot
	return nil
}

func (p *Pac) FetchLastSync() (time.Time, error) {
	var lastSync time.Time

	f, err := os.Open(p.logfile)
	if err != nil {
		return lastSync, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "synchronizing") {
			lastSync, err = time.Parse("2006-01-02 15:04", line[1:17])
			if err != nil {
				return lastSync, err
			}
		}
	}

	err = scanner.Err()
	if err != nil {
		return lastSync, err
	}

	return lastSync, nil
}

func (p *Pac) FetchPackages() ([]Package, error) {
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

	return packages, nil
}

func (p *Pac) startAutoUpdate(term <-chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = watcher.Add(p.logfile)
	if err != nil {
		watcher.Close()
		return err
	}

	debouncedUpdate := NewDebounced(100, func() {
		err := p.UpdateSnapshot()
		if err != nil {
			// How to handle?
			log.Println(err)
		}
	})

	go func() {
		defer watcher.Close()

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					debouncedUpdate.Call()
				}
			case err := <-watcher.Errors:
				// ???
				log.Println(err)
			case <-term:
				return
			}
		}
	}()

	return nil
}
