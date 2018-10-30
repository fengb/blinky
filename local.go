package main

type Local struct {
	Snapshot *Snapshot

	update chan *Snapshot
}

func NewLocal(conf *Conf) (*Local, error) {
	update := make(chan *Snapshot)
	worker, err := NewWorkerPacman(update)
	if err != nil {
		close(update)
		return nil, err
	}

	snapshot, err := worker.FetchSnapshot()
	if err != nil {
		close(update)
		return nil, err
	}

	loc := Local{Snapshot: snapshot, update: update}
	go func() {
		for snapshot := range update {
			loc.Snapshot = snapshot
		}
	}()

	return &loc, err
}

func (l *Local) UpdateConf(conf *Conf) error {
	// TODO: reload worker
	return nil
}
