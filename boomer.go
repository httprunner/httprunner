package httpboomer

import (
	"time"

	"github.com/myzhan/boomer"
)

func HttpBoomer() *Boomer {
	return &Boomer{}
}

type Boomer struct {
}

func (b *Boomer) Run(testcases ...*TestCase) {
	var taskSlice []*boomer.Task
	for _, testcase := range testcases {
		task := convertBoomerTask(testcase)
		taskSlice = append(taskSlice, task)
	}
	boomer.Run(taskSlice...)
}

func convertBoomerTask(testcase *TestCase) *boomer.Task {
	return &boomer.Task{
		Name:   testcase.Config.Name,
		Weight: testcase.Config.Weight,
		Fn: func() {
			for _, step := range testcase.TestSteps {
				start := time.Now()
				err := step.Run()
				elapsed := time.Since(start).Nanoseconds() / int64(time.Millisecond)

				tStep := step.ToStruct()
				if err == nil {
					boomer.RecordSuccess(string(tStep.Request.Method), tStep.Name, elapsed, int64(0))
				} else {
					boomer.RecordFailure(string(tStep.Request.Method), tStep.Name, elapsed, err.Error())
				}
			}
		},
	}
}
