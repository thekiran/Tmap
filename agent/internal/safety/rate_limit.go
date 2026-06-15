package safety

import (
	"context"
	"time"
)

type RateLimiter struct {
	interval time.Duration
	tokens   chan struct{}
}

func NewRateLimiter(requestsPerSecond, burst int) *RateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 1
	}
	if burst <= 0 {
		burst = 1
	}
	rl := &RateLimiter{
		interval: time.Second / time.Duration(requestsPerSecond),
		tokens:   make(chan struct{}, burst),
	}
	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}
	go rl.refill()
	return rl
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	if rl == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.tokens:
		return nil
	}
}

func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()
	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
		}
	}
}
