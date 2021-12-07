package hrp

import (
	"time"

	"github.com/myzhan/boomer"

	"github.com/httprunner/hrp/internal/ga"
)

func NewStandaloneBoomer(spawnCount int, spawnRate float64) *Boomer {
	b := &Boomer{
		Boomer: boomer.NewStandaloneBoomer(spawnCount, spawnRate),
		debug:  false,
	}
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
	event := ga.EventTracking{
		Category: "RunLoadTests",
		Action:   "hrp boom",
	}
	// report start event
	go ga.SendEvent(event)
	// report execution timing event
	defer ga.SendEvent(event.StartTiming("execution"))

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
	return &boomer.Task{
		Name:   testcase.Config.Name,
		Weight: testcase.Config.Weight,
		Fn: func() {
			runner := NewRunner(nil).SetDebug(b.debug)
			config := testcase.Config
			for _, step := range testcase.TestSteps {
				var err error
				start := time.Now()
				stepData, err := runner.runStep(step, config)
				elapsed := time.Since(start).Nanoseconds() / int64(time.Millisecond)
				if err == nil {
					b.RecordSuccess(step.getType(), step.name(), elapsed, stepData.responseLength)
				} else {
					b.RecordFailure(step.getType(), step.name(), elapsed, err.Error())
				}
			}
		},
	}
}
