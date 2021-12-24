package boomer

import (
	"math"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewStandaloneBoomer(t *testing.T) {
	b := NewStandaloneBoomer(100, 10)

	if b.localRunner.spawnCount != 100 {
		t.Error("spawnCount should be 100")
	}

	if b.localRunner.spawnRate != 10 {
		t.Error("spawnRate should be 10")
	}
}

func TestSetRateLimiter(t *testing.T) {
	b := NewStandaloneBoomer(100, 10)
	b.SetRateLimiter(10, "10/1s")

	if b.localRunner.rateLimiter == nil {
		t.Error("b.rateLimiter should not be nil")
	}
}

func TestAddOutput(t *testing.T) {
	b := NewStandaloneBoomer(100, 10)
	b.AddOutput(NewConsoleOutput())
	b.AddOutput(NewConsoleOutput())

	if len(b.localRunner.outputs) != 2 {
		t.Error("length of outputs should be 2")
	}
}

func TestEnableCPUProfile(t *testing.T) {
	b := NewStandaloneBoomer(100, 10)
	b.EnableCPUProfile("cpu.prof", time.Second)

	if b.cpuProfile != "cpu.prof" {
		t.Error("cpuProfile should be cpu.prof")
	}

	if b.cpuProfileDuration != time.Second {
		t.Error("cpuProfileDuration should 1 second")
	}
}

func TestEnableMemoryProfile(t *testing.T) {
	b := NewStandaloneBoomer(100, 10)
	b.EnableMemoryProfile("mem.prof", time.Second)

	if b.memoryProfile != "mem.prof" {
		t.Error("memoryProfile should be mem.prof")
	}

	if b.memoryProfileDuration != time.Second {
		t.Error("memoryProfileDuration should 1 second")
	}
}

func TestStandaloneRun(t *testing.T) {
	b := NewStandaloneBoomer(10, 10)
	b.EnableCPUProfile("cpu.pprof", 2*time.Second)
	b.EnableMemoryProfile("mem.pprof", 2*time.Second)

	count := int64(0)
	taskA := &Task{
		Name: "increaseCount",
		Fn: func() {
			atomic.AddInt64(&count, 1)
			runtime.Goexit()
		},
	}
	go b.Run(taskA)

	time.Sleep(5 * time.Second)

	b.Quit()

	if atomic.LoadInt64(&count) != 10 {
		t.Error("count is", count, "expected: 10")
	}

	if _, err := os.Stat("cpu.pprof"); os.IsNotExist(err) {
		t.Error("File cpu.pprof is not generated")
	} else {
		os.Remove("cpu.pprof")
	}

	if _, err := os.Stat("mem.pprof"); os.IsNotExist(err) {
		t.Error("File mem.pprof is not generated")
	} else {
		os.Remove("mem.pprof")
	}
}

func TestCreateRatelimiter(t *testing.T) {
	b := NewStandaloneBoomer(10, 10)
	b.SetRateLimiter(100, "-1")

	if stableRateLimiter, ok := b.localRunner.rateLimiter.(*StableRateLimiter); !ok {
		t.Error("Expected stableRateLimiter")
	} else {
		if stableRateLimiter.threshold != 100 {
			t.Error("threshold should be equals to math.MaxInt64, was", stableRateLimiter.threshold)
		}
	}

	b.SetRateLimiter(0, "1")
	if rampUpRateLimiter, ok := b.localRunner.rateLimiter.(*RampUpRateLimiter); !ok {
		t.Error("Expected rampUpRateLimiter")
	} else {
		if rampUpRateLimiter.maxThreshold != math.MaxInt64 {
			t.Error("maxThreshold should be equals to math.MaxInt64, was", rampUpRateLimiter.maxThreshold)
		}
		if rampUpRateLimiter.rampUpRate != "1" {
			t.Error("rampUpRate should be equals to \"1\", was", rampUpRateLimiter.rampUpRate)
		}
	}

	b.SetRateLimiter(10, "2/2s")
	if rampUpRateLimiter, ok := b.localRunner.rateLimiter.(*RampUpRateLimiter); !ok {
		t.Error("Expected rampUpRateLimiter")
	} else {
		if rampUpRateLimiter.maxThreshold != 10 {
			t.Error("maxThreshold should be equals to 10, was", rampUpRateLimiter.maxThreshold)
		}
		if rampUpRateLimiter.rampUpRate != "2/2s" {
			t.Error("rampUpRate should be equals to \"2/2s\", was", rampUpRateLimiter.rampUpRate)
		}
		if rampUpRateLimiter.rampUpStep != 2 {
			t.Error("rampUpStep should be equals to 2, was", rampUpRateLimiter.rampUpStep)
		}
		if rampUpRateLimiter.rampUpPeroid != 2*time.Second {
			t.Error("rampUpPeroid should be equals to 2 seconds, was", rampUpRateLimiter.rampUpPeroid)
		}
	}
}
