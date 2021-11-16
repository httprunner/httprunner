package hrp

import (
	"time"

	"github.com/myzhan/boomer"
)

func NewStandaloneBoomer(spawnCount int, spawnRate float64) *Boomer {
	b := &Boomer{
		Boomer: boomer.NewStandaloneBoomer(spawnCount, spawnRate),
		debug:  false,
	}
	b.AddOutput(boomer.NewConsoleOutput())
	return b
}

type Boomer struct {
	*boomer.Boomer
	debug bool
}

func (b *Boomer) SetDebug(debug bool) *Boomer {
	b.debug = debug
	return b
}

func (b *Boomer) Run(testcases ...ITestCase) {
	var taskSlice []*boomer.Task
	for _, iTestCase := range testcases {
		testcase, err := iTestCase.ToTestCase()
		if err != nil {
			panic(err)
		}
		task := b.convertBoomerTask(testcase)
		taskSlice = append(taskSlice, task)
	}
	b.Boomer.Run(taskSlice...)
}

func (b *Boomer) Quit() {
	b.Boomer.Quit()
}

func (b *Boomer) convertBoomerTask(testcase *TestCase) *boomer.Task {
	runner := NewRunner(nil).SetDebug(b.debug)
	return &boomer.Task{
		Name:   testcase.Config.Name,
		Weight: testcase.Config.Weight,
		Fn: func() {
			config := &testcase.Config
			for _, step := range testcase.TestSteps {
				var err error
				start := time.Now()
				stepData, err := runner.runStep(step, config)
				elapsed := time.Since(start).Nanoseconds() / int64(time.Millisecond)
				if err == nil {
					b.RecordSuccess(step.Type(), step.Name(), elapsed, stepData.ResponseLength)
				} else {
					b.RecordFailure(step.Type(), step.Name(), elapsed, err.Error())
				}
			}
		},
	}
}
