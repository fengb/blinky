package main

import (
	"time"
)

func Debounced(delay int64, fn func()) func() chan struct{} {
	var timer *time.Timer
	var drains []chan struct{}
	duration := time.Duration(delay) * time.Millisecond

	return func() chan struct{} {
		drain := make(chan struct{})
		drains = append(drains, drain)

		if timer == nil {
			timer = time.NewTimer(duration)
			go func() {
				<-timer.C
				timer = nil
				fn()
				for _, drain := range drains {
					select {
					case drain <- struct{}{}:
						close(drain)
					default:
					}
				}
				drains = nil
			}()
		} else {
			timer.Stop()
			timer.Reset(duration)
		}

		return drain
	}
}
