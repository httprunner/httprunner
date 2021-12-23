package boomer

import (
	"math"
	"time"

	"github.com/rs/zerolog/log"
)

// A Boomer is used to run tasks.
type Boomer struct {
	rateLimiter RateLimiter

	localRunner *localRunner
	spawnCount  int
	spawnRate   float64

	cpuProfile         string
	cpuProfileDuration time.Duration

	memoryProfile         string
	memoryProfileDuration time.Duration

	outputs []Output
}

// NewStandaloneBoomer returns a new Boomer, which can run without master.
func NewStandaloneBoomer(spawnCount int, spawnRate float64) *Boomer {
	return &Boomer{
		spawnCount: spawnCount,
		spawnRate:  spawnRate,
	}
}

// SetRateLimiter creates rate limiter with the given limit and burst.
func (b *Boomer) SetRateLimiter(maxRPS int64, requestIncreaseRate string) {
	var rateLimiter RateLimiter
	var err error
	if requestIncreaseRate != "-1" {
		if maxRPS <= 0 {
			maxRPS = math.MaxInt64
		}
		log.Warn().Int64("maxRPS", maxRPS).Str("increaseRate", requestIncreaseRate).Msg("set ramp up rate limiter")
		rateLimiter, err = NewRampUpRateLimiter(maxRPS, requestIncreaseRate, time.Second)
	} else {
		if maxRPS > 0 {
			log.Warn().Int64("maxRPS", maxRPS).Msg("set stable rate limiter")
			rateLimiter = NewStableRateLimiter(maxRPS, time.Second)
		}
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to create rate limiter")
		return
	}
	b.rateLimiter = rateLimiter
}

// AddOutput accepts outputs which implements the boomer.Output interface.
func (b *Boomer) AddOutput(o Output) {
	b.outputs = append(b.outputs, o)
}

// EnableCPUProfile will start cpu profiling after run.
func (b *Boomer) EnableCPUProfile(cpuProfile string, duration time.Duration) {
	b.cpuProfile = cpuProfile
	b.cpuProfileDuration = duration
}

// EnableMemoryProfile will start memory profiling after run.
func (b *Boomer) EnableMemoryProfile(memoryProfile string, duration time.Duration) {
	b.memoryProfile = memoryProfile
	b.memoryProfileDuration = duration
}

// Run accepts a slice of Task and connects to the locust master.
func (b *Boomer) Run(tasks ...*Task) {
	if b.cpuProfile != "" {
		err := startCPUProfile(b.cpuProfile, b.cpuProfileDuration)
		if err != nil {
			log.Error().Err(err).Msg("failed to start cpu profiling")
		}
	}
	if b.memoryProfile != "" {
		err := startMemoryProfile(b.memoryProfile, b.memoryProfileDuration)
		if err != nil {
			log.Error().Err(err).Msg("failed to start memory profiling")
		}
	}

	b.localRunner = newLocalRunner(tasks, b.rateLimiter, b.spawnCount, b.spawnRate)
	for _, o := range b.outputs {
		b.localRunner.addOutput(o)
	}
	b.localRunner.start()
}

// RecordTransaction reports a transaction stat.
func (b *Boomer) RecordTransaction(name string, success bool, elapsedTime int64, contentSize int64) {
	if b.localRunner == nil {
		log.Warn().Msg("boomer not initialized")
		return
	}
	b.localRunner.stats.transactionChan <- &transaction{
		name:        name,
		success:     success,
		elapsedTime: elapsedTime,
		contentSize: contentSize,
	}
}

// RecordSuccess reports a success.
func (b *Boomer) RecordSuccess(requestType, name string, responseTime int64, responseLength int64) {
	if b.localRunner == nil {
		log.Warn().Msg("boomer not initialized")
		return
	}
	b.localRunner.stats.requestSuccessChan <- &requestSuccess{
		requestType:    requestType,
		name:           name,
		responseTime:   responseTime,
		responseLength: responseLength,
	}
}

// RecordFailure reports a failure.
func (b *Boomer) RecordFailure(requestType, name string, responseTime int64, exception string) {
	if b.localRunner == nil {
		log.Warn().Msg("boomer not initialized")
		return
	}
	b.localRunner.stats.requestFailureChan <- &requestFailure{
		requestType:  requestType,
		name:         name,
		responseTime: responseTime,
		errMsg:       exception,
	}
}

// Quit will send a quit message to the master.
func (b *Boomer) Quit() {
	if b.localRunner == nil {
		log.Warn().Msg("boomer not initialized")
		return
	}
	b.localRunner.stop()
}
