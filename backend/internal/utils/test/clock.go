package test

import (
	"time"

	"github.com/els/backend/internal/utils/timex"
)

var FixedTime = time.Date(2026, time.April, 15, 12, 34, 56, 0, timex.MSK)

func FrozenClock() *timex.FrozenClock {
	return timex.NewFrozen(FixedTime)
}

func FrozenAt(t time.Time) *timex.FrozenClock {
	return timex.NewFrozen(t)
}
