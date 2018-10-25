package main

type Local struct {
	Snapshot *Snapshot
}

func NewLocal() (*Local, error) {
	worker, err := NewLocalPacman()
	if err != nil {
		return nil, err
	}

	snapshot, err := worker.FetchSnapshot()
	if err != nil {
		return nil, err
	}

	loc := Local{Snapshot: snapshot}
	go func() {
		for snapshot := range worker.C {
			loc.Snapshot = snapshot
		}
	}()

	return &loc, err
}

func (l *Local) UpdateConf(conf *Conf) error {
	// TODO: reload worker
	return nil
}
