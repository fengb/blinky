package main

import "time"

type Debounced struct {
	delay    time.Duration
	fn       func()
	timer    *time.Timer
	isActive bool
	drains   []chan struct{}
}

func NewDebounced(delay int64, fn func()) *Debounced {
	d := Debounced{
		delay: time.Duration(delay) * time.Millisecond,
		fn:    fn,
		timer: time.NewTimer(1 * time.Hour),
	}
	d.timer.Stop()

	go func() {
		for {
			<-d.timer.C
			d.CallImmediate()
		}
	}()

	return &d
}

func (d *Debounced) IsActive() bool {
	return d.isActive
}

func (d *Debounced) Cancel() {
	d.timer.Stop()
	d.isActive = false
}

func (d *Debounced) CallImmediate() {
	d.timer.Stop()
	d.isActive = false
	d.fn()

	for _, drain := range d.drains {
		close(drain)
	}
	d.drains = nil
}

func (d *Debounced) Call() {
	d.isActive = true
	d.timer.Stop()
	d.timer.Reset(d.delay)
	return
}

func (d *Debounced) Drain() <-chan struct{} {
	drain := make(chan struct{})

	if d.IsActive() {
		d.drains = append(d.drains, drain)
	} else {
		close(drain)
	}

	return drain
}
