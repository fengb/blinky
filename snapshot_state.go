package main

import (
	"log"
	"time"
)

type Package struct {
	Name    string
	Version string
	Upgrade string
}

type Snapshot struct {
	LastSync time.Time
	Packages []Package
}

func (s *Snapshot) Status() string {
	for _, pkg := range s.Packages {
		if pkg.Upgrade != "" {
			return "outdated"
		}
	}
	return "current"
}

func (s *Snapshot) Equal(o *Snapshot) bool {
	if o == nil {
		return s == o
	}

	if s.LastSync != o.LastSync ||
		len(s.Packages) != len(o.Packages) {
		return false
	}

	for i, _ := range s.Packages {
		if s.Packages[i] != o.Packages[i] {
			return false
		}
	}

	return true
}

type SnapshotState struct {
	local  *Snapshot
	Remote map[string]*Snapshot
}

func NewSnapshotState() *SnapshotState {
	return &SnapshotState{Remote: make(map[string]*Snapshot)}
}

func (s *SnapshotState) Local() *Snapshot {
	return s.local
}

func (s *SnapshotState) UpdateLocal(snapshot *Snapshot) {
	if snapshot.Equal(s.local) {
		return
	}

	log.Println("Local snapshot changed")

	s.local = snapshot
}

func (s *SnapshotState) UpdateRemote(key string, snapshot *Snapshot) {
	log.Println("Remote snapshot changed —", key)
	s.Remote[key] = snapshot
}
