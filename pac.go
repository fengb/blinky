package main

import (
	"github.com/Jguer/go-alpm"
	"github.com/fsnotify/fsnotify"
	"github.com/serverhorror/rog-go/reverse"
	"log"
	"os"
	"strings"
	"time"
)

type Package struct {
	Name    string
	Version string
	Upgrade string
}

type Snapshot struct {
	Status   string
	LastSync time.Time
	Packages []Package
}

type Pac struct {
	snapshot *Snapshot
	conf     alpm.PacmanConfig
	dbpath   string
	logfile  string
	watcher  *fsnotify.Watcher
}

func NewPac(filename string) (*Pac, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf, err := alpm.ParseConfig(f)
	if err != nil {
		return nil, err
	}

	pac := Pac{conf: conf}

	pac.dbpath = conf.DBPath
	if pac.dbpath == "" {
		pac.dbpath = "/var/lib/pacman"
	}

	pac.logfile = conf.LogFile
	if pac.logfile == "" {
		pac.logfile = "/var/log/pacman.log"
	}

	pac.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = pac.watcher.Add(pac.dbpath)
	if err != nil {
		pac.watcher.Close()
		return nil, err
	}

	go pac.watchChanges()
	go pac.watchErrors()

	return &pac, nil
}

func (p *Pac) UpdateConf(conf *Conf) error {
	// TODO: reload pacman.conf / watcher
	return nil
}

func (p *Pac) GetSnapshot() (*Snapshot, error) {
	var err error
	if p.snapshot == nil {
		err = p.UpdateSnapshot()
	}

	return p.snapshot, err
}

func (p *Pac) UpdateSnapshot() error {
	log.Println("Updating pacman snapshot")

	snapshot := Snapshot{}

	err := Parallel(
		func() error {
			var err error
			snapshot.LastSync, err = p.FetchLastSync()
			return err
		},
		func() error {
			var err error
			snapshot.Packages, err = p.FetchPackages()
			snapshot.Status = "current"
			for _, pkg := range snapshot.Packages {
				if pkg.Upgrade != "" {
					snapshot.Status = "outdated"
					break
				}
			}
			return err
		},
	)

	if err != nil {
		return err
	}

	p.snapshot = &snapshot
	return nil
}

func (p *Pac) FetchLastSync() (time.Time, error) {
	f, err := os.Open(p.logfile)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	var line string
	scanner := reverse.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
		if strings.Contains(line, "synchronizing") {
			break
		}
	}

	err = scanner.Err()
	if err != nil {
		return time.Time{}, err
	}

	loc := time.Now().Location()
	return time.ParseInLocation("2006-01-02 15:04", line[1:17], loc)
}

func (p *Pac) FetchPackages() ([]Package, error) {
	handle, err := p.conf.CreateHandle()
	if err != nil {
		return nil, err
	}
	defer handle.Release()

	localDb, err := handle.LocalDb()
	if err != nil {
		return nil, err
	}

	syncDbs, err := handle.SyncDbs()
	if err != nil {
		return nil, err
	}

	packages := []Package{}
	localDb.PkgCache().ForEach(func(alpmPkg alpm.Package) error {
		pkg := Package{Name: alpmPkg.Name(), Version: alpmPkg.Version()}
		newAlpmPkg := alpmPkg.NewVersion(syncDbs)
		if newAlpmPkg != nil {
			pkg.Upgrade = newAlpmPkg.Version()
		}
		packages = append(packages, pkg)
		return nil
	})

	return packages, nil
}

func (p *Pac) watchChanges() {
	debouncedUpdate := NewDebounced(1000, func() {
		err := p.UpdateSnapshot()
		if err != nil {
			// How to handle?
			log.Println(err)
		}
	})

	for event := range p.watcher.Events {
		if strings.HasSuffix(event.Name, "db.lck") {
			if event.Op&fsnotify.Create == fsnotify.Create {
				if p.snapshot != nil {
					p.snapshot.Status = "updating"
				}
				debouncedUpdate.Cancel()
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				debouncedUpdate.Call()
			}
		}
	}
}

func (p *Pac) watchErrors() {
	for err := range p.watcher.Errors {
		log.Println(err)
	}
}
