package limiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	bucket   chan struct{}
	rate     float64
	capacity int
	mu       sync.Mutex
}

func New(qps float64, capacity int) *RateLimiter {
	limiter := &RateLimiter{
		bucket:   make(chan struct{}, capacity),
		rate:     qps,
		capacity: capacity,
	}

	for i := 0; i < capacity; i++ {
		limiter.bucket <- struct{}{}
	}

	go limiter.fill()

	return limiter
}

func (l *RateLimiter) fill() {
	ticker := time.NewTicker(time.Duration(float64(time.Second) / l.rate))
	defer ticker.Stop()

	for range ticker.C {
		select {
		case l.bucket <- struct{}{}:
		default:
		}
	}
}

func (l *RateLimiter) Allow() bool {
	select {
	case <-l.bucket:
		return true
	default:
		return false
	}
}

func (l *RateLimiter) Wait() {
	<-l.bucket
}
