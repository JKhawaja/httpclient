package httpclient

import (
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
)

// RetryPolicy ...
type RetryPolicy interface {
	Backoffs() []time.Duration
	Retry(*http.Request) bool
}

// NoRetryPolicy ...
type NoRetryPolicy struct{}

// NewNoRetryPolicy ...
func NewNoRetryPolicy() RetryPolicy {
	return &NoRetryPolicy{}
}

// Backoffs ...
func (n *NoRetryPolicy) Backoffs() []time.Duration {
	return []time.Duration{}
}

// Retry ...
func (n *NoRetryPolicy) Retry(req *http.Request) bool {
	return false
}

// SingleRetryPolicy ...
type SingleRetryPolicy struct {
	Interval time.Duration
}

// NewSingleRetryPolicy ...
func NewSingleRetryPolicy(interval time.Duration) RetryPolicy {
	return &SingleRetryPolicy{
		Interval: interval,
	}
}

// Backoffs ...
func (s *SingleRetryPolicy) Backoffs() []time.Duration {
	return []time.Duration{s.Interval}
}

// Retry ...
func (s *SingleRetryPolicy) Retry(req *http.Request) bool {
	// do not retry POST and PATCH methods (as they are not indempotent)
	if req.Method == "POST" || req.Method == "PATCH" {
		return false
	}
	return true
}

// ConstantRetryPolicy ...
type ConstantRetryPolicy struct {
	Interval time.Duration
	Max      int
}

// NewConstantRetryPolicy ...
func NewConstantRetryPolicy(interval time.Duration, maxRetries int) RetryPolicy {
	return &ConstantRetryPolicy{
		Interval: interval,
		Max:      maxRetries,
	}
}

// Backoffs ...
func (c *ConstantRetryPolicy) Backoffs() []time.Duration {
	durations := make([]time.Duration, c.Max)
	for i := range durations {
		durations[i] = c.Interval
	}

	return durations
}

// Retry ...
func (c *ConstantRetryPolicy) Retry(req *http.Request) bool {
	// do not retry POST and PATCH methods (as they are not indempotent)
	if req.Method == "POST" || req.Method == "PATCH" {
		return false
	}
	return true
}

// ExponentialRetryPolicy ...
type ExponentialRetryPolicy struct {
	config backoff.ExponentialBackOff
}

// BackoffConfig ...
type BackoffConfig struct {
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
	MaxElapsedTime      time.Duration
}

// DefaultBackoffConfig ...
var DefaultBackoffConfig = BackoffConfig{
	InitialInterval:     100 * time.Millisecond,
	MaxInterval:         1500 * time.Millisecond,
	MaxElapsedTime:      10 * time.Second,
	RandomizationFactor: backoff.DefaultRandomizationFactor,
	Multiplier:          backoff.DefaultMultiplier,
}

// NewExponentialRetryPolicy ...
func NewExponentialRetryPolicy(config BackoffConfig) RetryPolicy {
	if config.RandomizationFactor == 0 {
		config.RandomizationFactor = backoff.DefaultRandomizationFactor
	}

	if config.Multiplier == 0 {
		config.Multiplier = backoff.DefaultMultiplier
	}

	if config.InitialInterval == 0 {
		config.InitialInterval = DefaultBackoffConfig.InitialInterval
	}

	if config.MaxInterval == 0 {
		config.MaxInterval = DefaultBackoffConfig.MaxInterval
	}

	if config.MaxElapsedTime == 0 {
		config.MaxElapsedTime = DefaultBackoffConfig.MaxElapsedTime
	}

	newConfig := backoff.ExponentialBackOff{
		InitialInterval:     config.InitialInterval,
		Multiplier:          config.Multiplier,
		RandomizationFactor: config.RandomizationFactor,
		MaxInterval:         config.MaxInterval,
		MaxElapsedTime:      config.MaxElapsedTime,
	}

	return &ExponentialRetryPolicy{
		config: newConfig,
	}
}

// Backoffs ...
func (e *ExponentialRetryPolicy) Backoffs() []time.Duration {
	var durations []time.Duration

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = e.config.InitialInterval
	b.MaxInterval = e.config.MaxInterval
	b.MaxElapsedTime = e.config.MaxElapsedTime
	b.RandomizationFactor = e.config.RandomizationFactor
	b.Multiplier = e.config.Multiplier

	total := 1 * time.Millisecond
	for {
		d := b.NextBackOff()
		switch d {
		case backoff.Stop:
			return durations
		default:
			if total >= b.MaxElapsedTime {
				return durations
			} else {
				durations = append(durations, d)
				total += d
			}
		}
	}
}

// Retry ...
func (e *ExponentialRetryPolicy) Retry(req *http.Request) bool {
	// do not retry POST and PATCH methods (as they are not indempotent)
	if req.Method == "POST" || req.Method == "PATCH" {
		return false
	}
	return true
}
