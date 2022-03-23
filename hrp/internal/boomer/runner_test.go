package boomer

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type HitOutput struct {
	onStart bool
	onEvent bool
	onStop  bool
}

func (o *HitOutput) OnStart() {
	o.onStart = true
}

func (o *HitOutput) OnEvent(data map[string]interface{}) {
	o.onEvent = true
}

func (o *HitOutput) OnStop() {
	o.onStop = true
}

func TestSafeRun(t *testing.T) {
	runner := &runner{}
	runner.safeRun(func() {
		panic("Runner will catch this panic")
	})
}

func TestOutputOnStart(t *testing.T) {
	hitOutput := &HitOutput{}
	hitOutput2 := &HitOutput{}
	runner := &runner{}
	runner.addOutput(hitOutput)
	runner.addOutput(hitOutput2)
	runner.outputOnStart()
	if !hitOutput.onStart {
		t.Error("hitOutput's OnStart has not been called")
	}
	if !hitOutput2.onStart {
		t.Error("hitOutput2's OnStart has not been called")
	}
}

func TestOutputOnEvent(t *testing.T) {
	hitOutput := &HitOutput{}
	hitOutput2 := &HitOutput{}
	runner := &runner{}
	runner.addOutput(hitOutput)
	runner.addOutput(hitOutput2)
	runner.outputOnEvent(nil)
	if !hitOutput.onEvent {
		t.Error("hitOutput's OnEvent has not been called")
	}
	if !hitOutput2.onEvent {
		t.Error("hitOutput2's OnEvent has not been called")
	}
}

func TestOutputOnStop(t *testing.T) {
	hitOutput := &HitOutput{}
	hitOutput2 := &HitOutput{}
	runner := &runner{}
	runner.addOutput(hitOutput)
	runner.addOutput(hitOutput2)
	runner.outputOnStop()
	if !hitOutput.onStop {
		t.Error("hitOutput's OnStop has not been called")
	}
	if !hitOutput2.onStop {
		t.Error("hitOutput2's OnStop has not been called")
	}
}

func TestLocalRunner(t *testing.T) {
	taskA := &Task{
		Weight: 10,
		Fn: func() {
			time.Sleep(time.Second)
		},
		Name: "TaskA",
	}
	tasks := []*Task{taskA}
	runner := newLocalRunner(2, 2)
	runner.setTasks(tasks)
	go runner.start()
	time.Sleep(4 * time.Second)
	runner.stop()
}

func TestLoopCount(t *testing.T) {
	taskA := &Task{
		Weight: 10,
		Fn: func() {
			time.Sleep(time.Second)
		},
		Name: "TaskA",
	}
	tasks := []*Task{taskA}
	runner := newLocalRunner(2, 2)
	runner.loop = &Loop{loopCount: 4}
	runner.setTasks(tasks)
	go runner.start()
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	<-ticker.C
	if !assert.Equal(t, runner.loop.loopCount, atomic.LoadInt64(&runner.loop.finishedCount)) {
		t.Fail()
	}
}
