package boomer

// Task is like the "Locust object" in locust, the python version.
// When boomer receives a start message from master, it will spawn several goroutines to run Task.Fn.
// But users can keep some information in the python version, they can't do the same things in boomer.
// Because Task.Fn is a pure function.
type Task struct {
	// The weight is used to distribute goroutines over multiple tasks.
	Weight int
	// Fn is called by the goroutines allocated to this task, in a loop.
	Fn   func()
	Name string
}
