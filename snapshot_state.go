package main

import (
	"bytes"
	"log"
	"net"
	"sort"
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

type NetworkLink struct {
	Ip          net.IP
	Hostnames   []string
	LastContact time.Time
	Snapshot    *Snapshot
	raw         []byte
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

type SnapshotState struct {
	local         *Snapshot
	subs          []chan *Snapshot
	networkLookup map[string]NetworkLink
}

func NewSnapshotState() *SnapshotState {
	return &SnapshotState{networkLookup: make(map[string]NetworkLink)}
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
	for _, sub := range s.subs {
		sub <- snapshot
	}
}

func (s *SnapshotState) SubLocal() <-chan *Snapshot {
	sub := make(chan *Snapshot)
	s.subs = append(s.subs, sub)
	return sub
}

func (s *SnapshotState) UpdateNetworkLink(
	addr *net.UDPAddr,
	raw []byte,
	makeSnapshot func(raw []byte) (*Snapshot, error),
) error {
	ipString := addr.IP.String()
	cache, ok := s.networkLookup[ipString]
	if ok && bytes.Equal(raw, cache.raw) {
		cache.LastContact = time.Now()
		s.networkLookup[ipString] = cache
		return nil
	}

	snapshot, err := makeSnapshot(raw)
	if err != nil {
		return err
	}

	cache = NetworkLink{raw: raw, Ip: addr.IP, LastContact: time.Now(), Snapshot: snapshot}
	cache.Hostnames, _ = net.LookupAddr(ipString)

	s.networkLookup[ipString] = cache
	log.Println("Update received from", ipString)

	return nil
}

func (s *SnapshotState) Network() []NetworkLink {
	keys := []string{}
	for key := range s.networkLookup {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	slice := make([]NetworkLink, len(keys))
	for i, key := range keys {
		slice[i] = s.networkLookup[key]
	}
	return slice
}
