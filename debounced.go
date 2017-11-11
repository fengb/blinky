package main

import "time"

type Debounced struct {
	delay  time.Duration
	fn     func()
	timer  *time.Timer
	drains []chan struct{}
}

func NewDebounced(delay int64, fn func()) Debounced {
	return Debounced{
		delay: time.Duration(delay) * time.Millisecond,
		fn:    fn,
	}
}

func (d Debounced) IsActive() bool {
	return d.timer != nil
}

func (d *Debounced) CallImmediate() {
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}

	d.fn()
	for _, drain := range d.drains {
		close(drain)
	}
	d.drains = nil
}

func (d *Debounced) Call() {
	if d.IsActive() {
		d.timer.Stop()
		d.timer.Reset(d.delay)
		return
	}

	d.timer = time.NewTimer(d.delay)
	go func() {
		<-d.timer.C
		d.CallImmediate()
	}()
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
