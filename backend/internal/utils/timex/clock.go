package timex

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().In(MSK) }

func System() Clock { return SystemClock{} }

type FrozenClock struct {
	mu sync.RWMutex
	t  time.Time
}

func NewFrozen(t time.Time) *FrozenClock {
	return &FrozenClock{t: t.In(MSK)}
}

func (f *FrozenClock) Now() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.t
}

func (f *FrozenClock) Set(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.t = t.In(MSK)
}

func (f *FrozenClock) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.t = f.t.Add(d)
}
