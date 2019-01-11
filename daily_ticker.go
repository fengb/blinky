package main

import "time"

type DailyTicker struct {
	C      <-chan time.Time
	target Clock
	timer  *time.Timer
}

func NewDailyTicker(target Clock) *DailyTicker {
	c := make(chan time.Time)
	daily := DailyTicker{c, target, time.NewTimer(1 * time.Hour)}
	daily.Reset(target)

	go func() {
		for t := range daily.timer.C {
			c <- t
			daily.Reset(daily.target)
		}
	}()

	return &daily
}

func (d *DailyTicker) NextRun() time.Time {
	now := time.Now()
	targetTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		d.target.Hour(),
		d.target.Minute(),
		d.target.Second(),
		0,
		now.Location(),
	)
	if now.Before(targetTime) {
		return targetTime
	} else {
		return targetTime.Add(24 * time.Hour)
	}
}

func (d *DailyTicker) Reset(target Clock) {
	d.timer.Stop()
	d.target = target
	d.timer.Reset(time.Until(d.NextRun()))
}

func (d *DailyTicker) Stop() {
	d.timer.Stop()
}
