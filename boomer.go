package hrp

import (
	"time"

	"github.com/debugtalk/boomer"
)

func NewStandaloneBoomer(spawnCount int, spawnRate float64) *Boomer {
	return &Boomer{
		Boomer: boomer.NewStandaloneBoomer(spawnCount, spawnRate),
		debug:  false,
	}
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

func (b *Boomer) convertBoomerTask(testcase *TestCase) *boomer.Task {
	runner := NewRunner(nil).SetDebug(b.debug)
	return &boomer.Task{
		Name:   testcase.Config.Name,
		Weight: testcase.Config.Weight,
		Fn: func() {
			config := &testcase.Config
			extractedVariables := make(map[string]interface{})
			for _, step := range testcase.TestSteps {
				var err error
				start := time.Now()
				stepData, err := runner.runStep(step, config, extractedVariables)
				elapsed := time.Since(start).Nanoseconds() / int64(time.Millisecond)
				if err == nil {
					boomer.RecordSuccess(step.Type(), step.Name(), elapsed, stepData.ResponseLength)
				} else {
					boomer.RecordFailure(step.Type(), step.Name(), elapsed, err.Error())
				}
			}
		},
	}
}
