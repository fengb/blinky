package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/serverhorror/rog-go/reverse"
	"log"
	"os"
	"strings"
	"time"
)

type Pac struct {
	snapshotState *SnapshotState
	dbpath        string
	logfile       string
	watcher       *fsnotify.Watcher
}

func NewPac(filename string, snapshotState *SnapshotState) (*Pac, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pac := Pac{snapshotState: snapshotState}

	if pac.dbpath == "" {
		pac.dbpath = "/var/lib/pacman"
	}

	if pac.logfile == "" {
		pac.logfile = "/var/log/pacman.log"
	}

	err = pac.updateSnapshot()
	if err != nil {
		return nil, err
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

func (p *Pac) updateSnapshot() error {
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
		return err
	}

	p.snapshotState.UpdateLocal(&snapshot)
	return nil
}

func (p *Pac) fetchLastSync() (time.Time, error) {
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

func (p *Pac) fetchPackages() ([]Package, error) {
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

func (p *Pac) watchChanges() {
	debouncedUpdate := NewDebounced(1000, func() {
		err := p.updateSnapshot()
		if err != nil {
			// How to handle?
			log.Println(err)
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

func (p *Pac) watchErrors() {
	for err := range p.watcher.Errors {
		log.Println(err)
	}
}
