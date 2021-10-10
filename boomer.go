package httpboomer

import (
	"time"

	"github.com/myzhan/boomer"
)

var (
	defaultMasterHost = "127.0.0.1"
	defaultMasterPort = 5557
)

// run load test with default configs
func Boom(testcases ...ITestCase) {
	NewBoomer(defaultMasterHost, defaultMasterPort).Run(testcases...)
}

func NewBoomer(masterHost string, masterPort int) *Boomer {
	return &Boomer{
		Boomer: boomer.NewBoomer(masterHost, masterPort),
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
	for _, testcase := range testcases {
		tcStruct, err := testcase.ToStruct()
		if err != nil {
			panic(err)
		}
		task := b.convertBoomerTask(tcStruct)
		taskSlice = append(taskSlice, task)
	}
	boomer.Run(taskSlice...)
}

func (b *Boomer) convertBoomerTask(testcase *TestCase) *boomer.Task {
	runner := NewRunner().SetDebug(b.debug)
	return &boomer.Task{
		Name:   testcase.Config.Name,
		Weight: testcase.Config.Weight,
		Fn: func() {
			config := &testcase.Config
			for _, step := range testcase.TestSteps {
				start := time.Now()
				_, err := runner.runStep(step, config)
				elapsed := time.Since(start).Nanoseconds() / int64(time.Millisecond)

				if err == nil {
					boomer.RecordSuccess(step.Type(), step.Name(), elapsed, int64(0))
				} else {
					boomer.RecordFailure(step.Type(), step.Name(), elapsed, err.Error())
				}
			}
		},
	}
}
