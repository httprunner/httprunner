package httpboomer

func RunTestCase(name string) *TestCase {
	return &TestCase{
		Config: TConfig{
			Name: name,
		},
	}
}

func (tc *TestCase) WithVariables(variables Variables) *TestCase {
	tc.Config.Variables = variables
	return tc
}

func (tc *TestCase) ToStruct() *TStep {
	return &TStep{
		TestCase: tc,
	}
}

func (tc *TestCase) Run() error {
	for _, step := range tc.TestSteps {
		if err := step.Run(); err != nil {
			return err
		}
	}
	return nil
}
