package middleware

import (
	"sync"
	"time"
)

type Limiter struct {
	tokens map[string]float64
	last   map[string]time.Time
	mu     sync.Mutex
	rate   float64
	cap    float64
	stop   chan struct{}
}

func NewLimiter(rate float64, cap float64) *Limiter {
	l := &Limiter{
		tokens: make(map[string]float64),
		last:   make(map[string]time.Time),
		rate:   rate,
		cap:    cap,
		stop:   make(chan struct{}),
	}
	go l.cleanupLoop()
	return l
}

func (l *Limiter) Close() {
	close(l.stop)
}

func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			for key, lastTime := range l.last {
				if now.Sub(lastTime) > 1*time.Hour {
					delete(l.tokens, key)
					delete(l.last, key)
				}
			}
			l.mu.Unlock()
		case <-l.stop:
			return
		}
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	lastTime, exists := l.last[key]
	if !exists {
		l.tokens[key] = l.cap
		l.last[key] = now
		lastTime = now
	}

	elapsed := now.Sub(lastTime).Seconds()
	l.last[key] = now

	tokens := l.tokens[key] + elapsed*l.rate
	if tokens > l.cap {
		tokens = l.cap
	}

	if tokens >= 1.0 {
		l.tokens[key] = tokens - 1.0
		return true
	}

	l.tokens[key] = tokens
	return false
}
