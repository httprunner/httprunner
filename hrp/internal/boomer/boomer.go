package boomer

import (
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Mode is the running mode of boomer, both standalone and distributed are supported.
type Mode int

const (
	// DistributedMasterMode requires being connected by each worker.
	DistributedMasterMode Mode = iota
	// DistributedWorkerMode requires connecting to a master.
	DistributedWorkerMode
	// StandaloneMode will run without a master.
	StandaloneMode
)

// A Boomer is used to run tasks.
type Boomer struct {
	mode Mode

	localRunner *localRunner

	cpuProfile         string
	cpuProfileDuration time.Duration

	memoryProfile         string
	memoryProfileDuration time.Duration

	disableKeepalive   bool
	disableCompression bool
}

// SetMode only accepts boomer.DistributedMasterMode、boomer.DistributedWorkerMode and boomer.StandaloneMode.
func (b *Boomer) SetMode(mode Mode) {
	switch mode {
	case DistributedMasterMode:
		b.mode = DistributedMasterMode
	case DistributedWorkerMode:
		b.mode = DistributedWorkerMode
	case StandaloneMode:
		b.mode = StandaloneMode
	default:
		log.Error().Err(errors.New("Invalid mode, ignored!"))
	}
}

// GetMode returns boomer operating mode
func (b *Boomer) GetMode() string {
	switch b.mode {
	case DistributedMasterMode:
		return "master"
	case DistributedWorkerMode:
		return "worker"
	case StandaloneMode:
		return "standalone"
	default:
		log.Error().Err(errors.New("Invalid mode, ignored!"))
		return ""
	}
}

// NewStandaloneBoomer returns a new Boomer, which can run without master.
func NewStandaloneBoomer(spawnCount int, spawnRate float64) *Boomer {
	return &Boomer{
		mode:        StandaloneMode,
		localRunner: newLocalRunner(spawnCount, spawnRate),
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

	if rateLimiter != nil {
		b.localRunner.rateLimitEnabled = true
		b.localRunner.rateLimiter = rateLimiter
	}
}

// SetDisableKeepAlive disable keep-alive for tcp
func (b *Boomer) SetDisableKeepAlive(disableKeepalive bool) {
	b.disableKeepalive = disableKeepalive
}

// SetDisableCompression disable compression to prevent the Transport from requesting compression with an "Accept-Encoding: gzip"
func (b *Boomer) SetDisableCompression(disableCompression bool) {
	b.disableCompression = disableCompression
}

func (b *Boomer) GetDisableKeepAlive() bool {
	return b.disableKeepalive
}

func (b *Boomer) GetDisableCompression() bool {
	return b.disableCompression
}

// SetLoopCount set loop count for test.
func (b *Boomer) SetLoopCount(loopCount int64) {
	// total loop count for testcase, it will be evenly distributed to each worker
	b.localRunner.loop = &Loop{loopCount: loopCount * int64(b.localRunner.spawnCount)}
}

// AddOutput accepts outputs which implements the boomer.Output interface.
func (b *Boomer) AddOutput(o Output) {
	b.localRunner.addOutput(o)
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

// EnableGracefulQuit catch SIGINT and SIGTERM signals to quit gracefully
func (b *Boomer) EnableGracefulQuit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		b.Quit()
	}()
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

	b.localRunner.setTasks(tasks)
	b.localRunner.start()
}

// RecordTransaction reports a transaction stat.
func (b *Boomer) RecordTransaction(name string, success bool, elapsedTime int64, contentSize int64) {
	b.localRunner.stats.transactionChan <- &transaction{
		name:        name,
		success:     success,
		elapsedTime: elapsedTime,
		contentSize: contentSize,
	}
}

// RecordSuccess reports a success.
func (b *Boomer) RecordSuccess(requestType, name string, responseTime int64, responseLength int64) {
	b.localRunner.stats.requestSuccessChan <- &requestSuccess{
		requestType:    requestType,
		name:           name,
		responseTime:   responseTime,
		responseLength: responseLength,
	}
}

// RecordFailure reports a failure.
func (b *Boomer) RecordFailure(requestType, name string, responseTime int64, exception string) {
	b.localRunner.stats.requestFailureChan <- &requestFailure{
		requestType:  requestType,
		name:         name,
		responseTime: responseTime,
		errMsg:       exception,
	}
}

// Quit will send a quit message to the master.
func (b *Boomer) Quit() {
	b.localRunner.stop()
}

func (b *Boomer) GetSpawnDoneChan() chan struct{} {
	return b.localRunner.spawnDone
}

func (b *Boomer) GetSpawnCount() int {
	return b.localRunner.spawnCount
}

func (b *Boomer) ResetStartTime() {
	b.localRunner.stats.total.resetStartTime()
}
