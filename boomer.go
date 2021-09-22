package httpboomer

import (
	"time"

	"github.com/myzhan/boomer"
)

var defaultBoomer = NewBoomer()

func Run(testcases ...*TestCase) {
	defaultBoomer.Run(testcases...)
}

func NewBoomer() *Boomer {
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

				if err == nil {
					boomer.RecordSuccess(step.Type(), step.Name(), elapsed, int64(0))
				} else {
					boomer.RecordFailure(step.Type(), step.Name(), elapsed, err.Error())
				}
			}
		},
	}
}
