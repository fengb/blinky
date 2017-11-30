package main

import "time"

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

func (s *Snapshot) PackageStatus() string {
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

	if s.Status != o.Status ||
		s.LastSync != o.LastSync ||
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
