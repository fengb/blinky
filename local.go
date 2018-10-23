package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/serverhorror/rog-go/reverse"
	"log"
	"os"
	"strings"
	"time"
)

type Local struct {
	Snapshot *Snapshot
	dbpath   string
	logfile  string
	watcher  *fsnotify.Watcher
}

func NewLocal() (*Local, error) {
	loc := Local{}

	if loc.dbpath == "" {
		loc.dbpath = "/var/lib/pacman"
	}

	if loc.logfile == "" {
		loc.logfile = "/var/log/pacman.log"
	}

	err := loc.updateSnapshot()
	if err != nil {
		return nil, err
	}
	if loc.Snapshot == nil {
		panic("wtf")
	}

	loc.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = loc.watcher.Add(loc.dbpath)
	if err != nil {
		loc.watcher.Close()
		return nil, err
	}

	go loc.watchChanges()
	go loc.watchErrors()

	return &loc, nil
}

func (l *Local) UpdateConf(conf *Conf) error {
	// TODO: reload pacman.conf / watcher
	return nil
}

func (l *Local) updateSnapshot() error {
	snapshot := Snapshot{}

	err := Parallel(
		func() (err error) {
			snapshot.LastSync, err = l.fetchLastSync()
			return err
		},
		func() (err error) {
			snapshot.Packages, err = l.fetchPackages()
			return err
		},
	)

	if err != nil {
		return err
	}

	l.Snapshot = &snapshot
	return nil
}

func (l *Local) fetchLastSync() (time.Time, error) {
	f, err := os.Open(l.logfile)
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

func (l *Local) fetchPackages() ([]Package, error) {
	stdout, stderr, err := CmdRun("pacman", "-Qu")
	if err != nil {
		log.Println(stderr)
		return make([]Package, 0), err
	}

	lines := strings.Split(stdout, "\n")
	packages := make([]Package, len(lines))
	for _, line := range lines {
		tokens := strings.Split(line, " ")
		pkg := Package{Name: tokens[0], Version: tokens[1], Upgrade: tokens[3]}
		packages = append(packages, pkg)
	}

	return packages, nil
}

func (l *Local) watchChanges() {
	debouncedUpdate := NewDebounced(1000, func() {
		err := l.updateSnapshot()
		if err != nil {
			// How to handle?
			log.Println(err)
		}
	})

	for event := range l.watcher.Events {
		if strings.HasSuffix(event.Name, "db.lck") {
			if event.Op&fsnotify.Create == fsnotify.Create {
				debouncedUpdate.Cancel()
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				debouncedUpdate.Call()
			}
		}
	}
}

func (l *Local) watchErrors() {
	for err := range l.watcher.Errors {
		log.Println(err)
	}
}
