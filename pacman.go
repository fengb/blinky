package main

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/serverhorror/rog-go/reverse"
	"log"
	"os"
	"strings"
	"time"
)

type Pacman struct {
	conf          *Conf
	snapshotState *SnapshotState

	refresher *DailyTicker
	watcher   *fsnotify.Watcher
}

func NewPacman(conf *Conf, snapshotState *SnapshotState) (*Pacman, error) {
	pac := Pacman{conf: conf, snapshotState: snapshotState, refresher: NewDailyTicker(nil)}
	err := pac.UpdateConf(conf)
	if err != nil {
		return nil, err
	}

	go func() {
		for _ = range pac.refresher.C {
			log.Println("Running pacman --noconfirm -Syuwq")
			_, _, err := CmdRun("pacman", "--noconfirm", "-Syuwq")
			if err != nil {
				log.Println(err)
			}
		}
	}()

	return &pac, nil
}

func (p *Pacman) UpdateConf(conf *Conf) error {
	p.conf = conf
	return Parallel(
		p.updateWatcher,
		p.updateRefresh,
	)
}

func (p *Pacman) updateWatcher() error {
	err := p.updateSnapshot()
	if err != nil {
		return err
	}

	p.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = p.watcher.Add(p.conf.Pacman.Conf.DBPath)
	if err != nil {
		p.watcher.Close()
		return err
	}

	go p.watchChanges()
	go p.watchErrors()

	return nil
}

func (p *Pacman) updateRefresh() error {
	if p.conf.Pacman.Refresh != nil {
		// Ideally use an exit status but all errors are 1
		_, stderr, _ := CmdRun("pacman", "-S")
		allowsPacmanSync := !strings.Contains(stderr, "root")

		if !allowsPacmanSync {
			return errors.New("Cannot run pacman -S")
		}

		p.refresher.Reset(p.conf.Pacman.Refresh)
		log.Println("Next refresh:", p.refresher.NextRun())
	} else {
		p.refresher.Stop()
	}

	return nil
}

func (p *Pacman) updateSnapshot() error {
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

func (p *Pacman) fetchLastSync() (time.Time, error) {
	f, err := os.Open(p.conf.Pacman.Conf.LogFile)
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

func (p *Pacman) fetchPackages() ([]Package, error) {
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

func (p *Pacman) watchChanges() {
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

func (p *Pacman) watchErrors() {
	for err := range p.watcher.Errors {
		log.Println(err)
	}
}
