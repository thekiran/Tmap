package ippool

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// ProbeFunc runs one reachability probe.
type ProbeFunc func(ctx context.Context, ip string) ProbeOutcome

// Scheduler runs probes through a bounded worker pool with a global rate limit
// and jitter, so the network is never flooded and bursts are de-synchronized.
// All work stops promptly on context cancellation.
type Scheduler struct {
	Concurrency int
	RatePerSec  float64
	Jitter      float64 // 0..1 fraction of the base interval
	rng         *rand.Rand
	mu          sync.Mutex
}

func NewScheduler(cfg Config) *Scheduler {
	return &Scheduler{
		Concurrency: cfg.MaxConcurrency,
		RatePerSec:  cfg.MaxProbesPerSec,
		Jitter:      cfg.JitterFraction,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Run pulls IPs from `tasks`, probes them with bounded concurrency at the
// configured rate, and writes each outcome to `results`. It returns when tasks
// is closed and drained or ctx is cancelled. results is NOT closed (the caller
// owns it, since multiple Run calls may share a results channel).
func (s *Scheduler) Run(ctx context.Context, tasks <-chan string, probe ProbeFunc, results chan<- ProbeOutcome) {
	conc := s.Concurrency
	if conc <= 0 {
		conc = 8
	}
	rate := s.RatePerSec
	if rate <= 0 {
		rate = 8
	}
	base := time.Duration(float64(time.Second) / rate)

	work := make(chan string)
	var wg sync.WaitGroup
	wg.Add(conc)
	for i := 0; i < conc; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case ip, ok := <-work:
					if !ok {
						return
					}
					out := probe(ctx, ip)
					select {
					case results <- out:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// Dispatcher: rate-limits the rate at which new probes START.
	go func() {
		defer close(work)
		for {
			select {
			case <-ctx.Done():
				return
			case ip, ok := <-tasks:
				if !ok {
					return
				}
				if !s.sleepWithJitter(ctx, base) {
					return
				}
				select {
				case work <- ip:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	wg.Wait()
}

// sleepWithJitter waits base ± (Jitter*base), returning false if ctx is cancelled.
func (s *Scheduler) sleepWithJitter(ctx context.Context, base time.Duration) bool {
	d := base
	if s.Jitter > 0 {
		s.mu.Lock()
		delta := (s.rng.Float64()*2 - 1) * s.Jitter // -j..+j
		s.mu.Unlock()
		d = time.Duration(float64(base) * (1 + delta))
	}
	if d <= 0 {
		return ctx.Err() == nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
