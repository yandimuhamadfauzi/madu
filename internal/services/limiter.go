package services

import (
	"sync"
	"time"
)

type entry struct {
	fails       int
	windowStart time.Time
	lockedUntil time.Time
}

// FailLimiter menghitung kegagalan per kunci (IP atau username) dan mengunci sementara.
type FailLimiter struct {
	mu     sync.Mutex
	items  map[string]*entry
	max    int
	window time.Duration
	lock   time.Duration
}

func NewFailLimiter(max int, window, lock time.Duration) *FailLimiter {
	l := &FailLimiter{items: map[string]*entry{}, max: max, window: window, lock: lock}
	go l.cleanup()
	return l
}

// Blocked mengembalikan sisa waktu kunci, atau 0 jika tidak terkunci.
func (l *FailLimiter) Blocked(key string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e, ok := l.items[key]; ok {
		if d := time.Until(e.lockedUntil); d > 0 {
			return d
		}
	}
	return 0
}

func (l *FailLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	e, ok := l.items[key]
	if !ok || now.Sub(e.windowStart) > l.window {
		e = &entry{windowStart: now}
		l.items[key] = e
	}
	e.fails++
	if e.fails >= l.max {
		e.lockedUntil = now.Add(l.lock)
		e.fails = 0
		e.windowStart = now
	}
}

func (l *FailLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.items, key)
}

// cleanup membuang data lama agar memori tidak terus bertambah.
func (l *FailLimiter) cleanup() {
	for range time.Tick(10 * time.Minute) {
		l.mu.Lock()
		now := time.Now()
		for k, e := range l.items {
			if now.Sub(e.windowStart) > l.window && now.After(e.lockedUntil) {
				delete(l.items, k)
			}
		}
		l.mu.Unlock()
	}
}
