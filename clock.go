package throttler

import (
	"time"
)

type clock interface {
	Now() time.Time
}

type defaultClock struct {
}

func (*defaultClock) Now() time.Time {
	return time.Now()
}

var (
	defaultClockInstance = &defaultClock{}
)
