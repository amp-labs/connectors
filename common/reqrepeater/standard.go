package reqrepeater

import (
	"errors"
	"time"
)

var ErrRetry = errors.New("try again later")

// Strategy is configuration of how a connector handles faulty requests
// connector will create Retry based on this config every time it attempts to make API call.
type Strategy interface {
	Start() Retry
}

// Retry records how many request fails occurred
// once some limit is reached it will fire to indicate surrender, only then we acknowledge API call failure.
type Retry interface {
	Completed() bool
}

// UniformRetryStrategy tries to call API with equally distributed intervals until retry limit is reached.
type UniformRetryStrategy struct {
	RetryLimit int
	Interval   time.Duration
}

func (r UniformRetryStrategy) Start() Retry {
	return &UniformRetry{
		RetriesLeft: r.RetryLimit,
		Interval:    r.Interval,
	}
}

// UniformRetry is a concrete instance that keeps track of remaining retries of UniformRetryStrategy.
type UniformRetry struct {
	RetriesLeft int
	Interval    time.Duration
}

func (r *UniformRetry) Completed() bool {
	if r.RetriesLeft == 0 {
		return true
	}

	r.RetriesLeft -= 1
	time.Sleep(r.Interval)

	return false
}

// NullStrategy it doesn't even try.
type NullStrategy struct{}

func (NullStrategy) Start() Retry {
	return &NullRetry{}
}

type NullRetry struct{}

func (NullRetry) Completed() bool {
	return true
}
