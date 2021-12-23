package boomer

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// RateLimiter is used to put limits on task executions.
type RateLimiter interface {
	// Start is used to enable the rate limiter.
	// It can be implemented as a noop if not needed.
	Start()

	// Acquire() is called before executing a task.Fn function.
	// If Acquire() returns true, the task.Fn function will be executed.
	// If Acquire() returns false, the task.Fn function won't be executed this time, but Acquire() will be called very soon.
	// It works like:
	// for {
	//      blocked := rateLimiter.Acquire()
	//      if !blocked {
	//	        task.Fn()
	//      }
	// }
	// Acquire() should block the caller until execution is allowed.
	Acquire() bool

	// Stop is used to disable the rate limiter.
	// It can be implemented as a noop if not needed.
	Stop()
}

// A StableRateLimiter uses the token bucket algorithm.
// the bucket is refilled according to the refill period, no burst is allowed.
type StableRateLimiter struct {
	threshold        int64
	currentThreshold int64
	refillPeriod     time.Duration
	broadcastChannel chan bool
	quitChannel      chan bool
}

// NewStableRateLimiter returns a StableRateLimiter.
func NewStableRateLimiter(threshold int64, refillPeriod time.Duration) (rateLimiter *StableRateLimiter) {
	rateLimiter = &StableRateLimiter{
		threshold:        threshold,
		currentThreshold: threshold,
		refillPeriod:     refillPeriod,
		broadcastChannel: make(chan bool),
	}
	return rateLimiter
}

// Start to refill the bucket periodically.
func (limiter *StableRateLimiter) Start() {
	limiter.quitChannel = make(chan bool)
	quitChannel := limiter.quitChannel
	go func() {
		for {
			select {
			case <-quitChannel:
				return
			default:
				atomic.StoreInt64(&limiter.currentThreshold, limiter.threshold)
				time.Sleep(limiter.refillPeriod)
				close(limiter.broadcastChannel)
				limiter.broadcastChannel = make(chan bool)
			}
		}
	}()
}

// Acquire a token from the bucket, returns true if the bucket is exhausted.
func (limiter *StableRateLimiter) Acquire() (blocked bool) {
	permit := atomic.AddInt64(&limiter.currentThreshold, -1)
	if permit < 0 {
		blocked = true
		// block until the bucket is refilled
		<-limiter.broadcastChannel
	} else {
		blocked = false
	}
	return blocked
}

// Stop the rate limiter.
func (limiter *StableRateLimiter) Stop() {
	close(limiter.quitChannel)
}

// ErrParsingRampUpRate is the error returned if the format of rampUpRate is invalid.
var ErrParsingRampUpRate = errors.New("ratelimiter: invalid format of rampUpRate, try \"1\" or \"1/1s\"")

// A RampUpRateLimiter uses the token bucket algorithm.
// the threshold is updated according to the warm up rate.
// the bucket is refilled according to the refill period, no burst is allowed.
type RampUpRateLimiter struct {
	maxThreshold     int64
	nextThreshold    int64
	currentThreshold int64
	refillPeriod     time.Duration
	rampUpRate       string
	rampUpStep       int64
	rampUpPeroid     time.Duration
	broadcastChannel chan bool
	rampUpChannel    chan bool
	quitChannel      chan bool
}

// NewRampUpRateLimiter returns a RampUpRateLimiter.
// Valid formats of rampUpRate are "1", "1/1s".
func NewRampUpRateLimiter(maxThreshold int64, rampUpRate string, refillPeriod time.Duration) (rateLimiter *RampUpRateLimiter, err error) {
	rateLimiter = &RampUpRateLimiter{
		maxThreshold:     maxThreshold,
		nextThreshold:    0,
		currentThreshold: 0,
		rampUpRate:       rampUpRate,
		refillPeriod:     refillPeriod,
		broadcastChannel: make(chan bool),
	}
	rateLimiter.rampUpStep, rateLimiter.rampUpPeroid, err = rateLimiter.parseRampUpRate(rateLimiter.rampUpRate)
	if err != nil {
		return nil, err
	}
	return rateLimiter, nil
}

func (limiter *RampUpRateLimiter) parseRampUpRate(rampUpRate string) (rampUpStep int64, rampUpPeroid time.Duration, err error) {
	if strings.Contains(rampUpRate, "/") {
		tmp := strings.Split(rampUpRate, "/")
		if len(tmp) != 2 {
			return rampUpStep, rampUpPeroid, ErrParsingRampUpRate
		}
		rampUpStep, err := strconv.ParseInt(tmp[0], 10, 64)
		if err != nil {
			return rampUpStep, rampUpPeroid, ErrParsingRampUpRate
		}
		rampUpPeroid, err := time.ParseDuration(tmp[1])
		if err != nil {
			return rampUpStep, rampUpPeroid, ErrParsingRampUpRate
		}
		return rampUpStep, rampUpPeroid, nil
	}

	rampUpStep, err = strconv.ParseInt(rampUpRate, 10, 64)
	if err != nil {
		return rampUpStep, rampUpPeroid, ErrParsingRampUpRate
	}
	rampUpPeroid = time.Second
	return rampUpStep, rampUpPeroid, nil
}

// Start to refill the bucket periodically.
func (limiter *RampUpRateLimiter) Start() {
	limiter.quitChannel = make(chan bool)
	quitChannel := limiter.quitChannel
	// bucket updater
	go func() {
		for {
			select {
			case <-quitChannel:
				return
			default:
				atomic.StoreInt64(&limiter.currentThreshold, limiter.nextThreshold)
				time.Sleep(limiter.refillPeriod)
				close(limiter.broadcastChannel)
				limiter.broadcastChannel = make(chan bool)
			}
		}
	}()
	// threshold updater
	go func() {
		for {
			select {
			case <-quitChannel:
				return
			default:
				nextValue := limiter.nextThreshold + limiter.rampUpStep
				if nextValue < 0 {
					// int64 overflow
					nextValue = int64(math.MaxInt64)
				}
				if nextValue > limiter.maxThreshold {
					nextValue = limiter.maxThreshold
				}
				atomic.StoreInt64(&limiter.nextThreshold, nextValue)
				time.Sleep(limiter.rampUpPeroid)
			}
		}
	}()
}

// Acquire a token from the bucket, returns true if the bucket is exhausted.
func (limiter *RampUpRateLimiter) Acquire() (blocked bool) {
	permit := atomic.AddInt64(&limiter.currentThreshold, -1)
	if permit < 0 {
		blocked = true
		// block until the bucket is refilled
		<-limiter.broadcastChannel
	} else {
		blocked = false
	}
	return blocked
}

// Stop the rate limiter.
func (limiter *RampUpRateLimiter) Stop() {
	limiter.nextThreshold = 0
	close(limiter.quitChannel)
}
