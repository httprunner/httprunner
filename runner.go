package httpboomer

func HttpRunner() *Runner {
	return &Runner{}
}

type Runner struct {
}

func (r *Runner) Run(testcases ...*TestCase) error {
	for _, testcase := range testcases {
		if err := r.runCase(testcase); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runCase(testcase *TestCase) error {
	for _, step := range testcase.TestSteps {
		if err := r.runStep(step); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runStep(req IStep) error {
	return req.Run()
}

func (r *Runner) GetSummary() *TestCaseSummary {
	return &TestCaseSummary{}
}
