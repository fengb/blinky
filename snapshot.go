package main

import (
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
		return false
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
