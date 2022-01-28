package util

import "time"

const defaultInterval = 3 * time.Second

type Repeat struct {
	times    int
	interval time.Duration
}

func (self Repeat) Do(f func() error) error {
	return self.DoWithBreak(func() (error, bool) {
		err := f()
		return err, (err == nil)
	})
}

func (self Repeat) DoWithBreak(f func() (error, bool)) error {
	var err error
	for i := 0; i < self.times; i++ {
		e, stop := f()

		if e == nil || stop {
			return e
		}

		err = e

		// Don't sleep if it's the last try
		if i < self.times-1 {
			time.Sleep(self.interval)
		}
	}
	return err
}

func Times(times int) *Repeat {
	return &Repeat{
		times:    times,
		interval: defaultInterval,
	}
}

func (self *Repeat) Interval(interval time.Duration) *Repeat {
	self.interval = interval
	return self
}

type Duration struct {
	limit    time.Duration
	interval time.Duration
}

func (self Duration) Do(f func() error) error {
	var err error
	endTime := time.Now().Add(self.limit)
	for time.Now().Before(endTime) {
		err = f()
		if err == nil {
			return nil
		}

		if time.Now().Add(self.interval).Before(endTime) {
			time.Sleep(self.interval)
		} else {
			break
		}
	}
	return err
}

func For(t time.Duration) *Duration {
	return &Duration{
		limit:    t,
		interval: defaultInterval,
	}
}

func (self *Duration) Interval(interval time.Duration) *Duration {
	self.interval = interval
	return self
}
