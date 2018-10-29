package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/serverhorror/rog-go/reverse"
	"log"
	"os"
	"strings"
	"time"
)

var dbpath = "/var/lib/pacman"
var logfile = "/var/log/pacman.log"

type LocalPacman struct {
	C chan *Snapshot

	watcher *fsnotify.Watcher
}

func NewLocalPacman() (*LocalPacman, error) {
	pac := LocalPacman{C: make(chan *Snapshot)}
	var err error

	pac.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = pac.watcher.Add(dbpath)
	if err != nil {
		pac.watcher.Close()
		return nil, err
	}

	go pac.watchChanges()
	go pac.watchErrors()

	return &pac, nil
}

func (p *LocalPacman) FetchSnapshot() (*Snapshot, error) {
	snapshot := Snapshot{}

	err := Parallel(
		func() (err error) {
			snapshot.LastSync, err = p.fetchLastSync()
			return err
		},
		func() (err error) {
			snapshot.Packages, err = p.fetchPackages()
			return err
		},
	)

	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (p *LocalPacman) fetchLastSync() (time.Time, error) {
	f, err := os.Open(logfile)
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

func (p *LocalPacman) fetchPackages() ([]Package, error) {
	stdout, _, err := CmdRun("pacman", "-Qu")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(stdout, "\n")
	packages := make([]Package, 0, len(lines))
	for _, line := range lines {
		tokens := strings.Fields(line)
		pkg := Package{Name: tokens[0], Version: tokens[1], Upgrade: tokens[3]}
		packages = append(packages, pkg)
	}

	return packages, nil
}

func (p *LocalPacman) watchChanges() {
	debouncedUpdate := NewDebounced(1000, func() {
		snapshot, err := p.FetchSnapshot()
		if err != nil {
			// How to handle?
			log.Println(err)
		} else {
			p.C <- snapshot
		}
	})

	for event := range p.watcher.Events {
		if strings.HasSuffix(event.Name, "db.lck") {
			if event.Op&fsnotify.Create == fsnotify.Create {
				debouncedUpdate.Cancel()
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				debouncedUpdate.Call()
			}
		}
	}
}

func (p *LocalPacman) watchErrors() {
	for err := range p.watcher.Errors {
		log.Println(err)
	}
}
