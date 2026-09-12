package limiter

import (
	"slices"
	"sync"
	"time"
)

type Limiter struct {
	mu    sync.Mutex
	done  chan struct{}
	cache map[string][]time.Time
}

const (
	perMinute = 10
	perHour   = 50
)

func (l *Limiter) Eligible(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	requests := slices.DeleteFunc(l.cache[ip], func(r time.Time) bool {
		return time.Since(r) >= time.Hour
	})
	l.cache[ip] = requests

	if len(requests) >= perHour {
		return false
	}

	recent := 0
	for _, r := range requests {
		if now.Sub(r) < time.Minute {
			recent++
		}
	}
	if recent >= perMinute {
		return false
	}

	l.cache[ip] = append(requests, now)
	return true
}

func (l *Limiter) sweep() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-l.done:
			return
		case <-ticker.C:
			l.mu.Lock()
			for ip, requests := range l.cache {
				shouldDelete := len(requests) == 0 || time.Since(requests[len(requests)-1]) >= time.Hour
				if shouldDelete {
					delete(l.cache, ip)
				}
			}
			l.mu.Unlock()
		}
	}
}

func (l *Limiter) Stop() {
	close(l.done)
}

func New() *Limiter {
	l := &Limiter{cache: map[string][]time.Time{}, done: make(chan struct{})}
	go l.sweep()
	return l
}
