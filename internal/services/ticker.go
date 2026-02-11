package services

import "time"

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type timeTicker struct {
	t *time.Ticker
}

func (tt *timeTicker) C() <-chan time.Time { return tt.t.C }
func (tt *timeTicker) Stop()               { tt.t.Stop() }

func NewTicker(d time.Duration) Ticker {
	return &timeTicker{t: time.NewTicker(d)}
}
