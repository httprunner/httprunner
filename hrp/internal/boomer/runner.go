package boomer

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
)

const (
	stateInit     = iota + 1 // initializing
	stateSpawning            // spawning
	stateRunning             // running
	stateQuitting            // quitting
	stateStopped             // stopped
)

const (
	reportStatsInterval = 3 * time.Second
)

type Loop struct {
	loopCount     int64 // more than 0
	acquiredCount int64 // count acquired of load testing
	finishedCount int64 // count finished of load testing
}

func (l *Loop) isFinished() bool {
	// return true when there are no remaining loop count to test
	return atomic.LoadInt64(&l.finishedCount) == l.loopCount
}

func (l *Loop) acquire() bool {
	// get one ticket when there are still remaining loop count to test
	// return true when getting ticket successfully
	if atomic.LoadInt64(&l.acquiredCount) < l.loopCount {
		atomic.AddInt64(&l.acquiredCount, 1)
		return true
	}
	return false
}

func (l *Loop) increaseFinishedCount() {
	atomic.AddInt64(&l.finishedCount, 1)
}

type runner struct {
	state int32

	tasks           []*Task
	totalTaskWeight int

	rateLimiter      RateLimiter
	rateLimitEnabled bool
	stats            *requestStats

	currentClientsNum int32 // current clients count
	spawnCount        int   // target clients to spawn
	spawnRate         float64
	loop              *Loop // specify loop count for testcase, count = loopCount * spawnCount
	spawnDone         chan struct{}

	outputs []Output
}

// safeRun runs fn and recovers from unexpected panics.
// it prevents panics from Task.Fn crashing boomer.
func (r *runner) safeRun(fn func()) {
	defer func() {
		// don't panic
		err := recover()
		if err != nil {
			stackTrace := debug.Stack()
			errMsg := fmt.Sprintf("%v", err)
			os.Stderr.Write([]byte(errMsg))
			os.Stderr.Write([]byte("\n"))
			os.Stderr.Write(stackTrace)
		}
	}()
	fn()
}

func (r *runner) addOutput(o Output) {
	r.outputs = append(r.outputs, o)
}

func (r *runner) outputOnStart() {
	size := len(r.outputs)
	if size == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(size)
	for _, output := range r.outputs {
		go func(o Output) {
			o.OnStart()
			wg.Done()
		}(output)
	}
	wg.Wait()
}

func (r *runner) outputOnEvent(data map[string]interface{}) {
	size := len(r.outputs)
	if size == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(size)
	for _, output := range r.outputs {
		go func(o Output) {
			o.OnEvent(data)
			wg.Done()
		}(output)
	}
	wg.Wait()
}

func (r *runner) outputOnStop() {
	size := len(r.outputs)
	if size == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(size)
	for _, output := range r.outputs {
		go func(o Output) {
			o.OnStop()
			wg.Done()
		}(output)
	}
	wg.Wait()
}

func (r *runner) reportStats() {
	data := r.stats.collectReportData()
	data["user_count"] = atomic.LoadInt32(&r.currentClientsNum)
	data["state"] = atomic.LoadInt32(&r.state)
	r.outputOnEvent(data)
}

func (r *runner) reportTestResult() {
	// convert stats in total
	var statsTotal interface{} = r.stats.total.serialize()
	entryTotalOutput, err := deserializeStatsEntry(statsTotal)
	if err != nil {
		return
	}
	duration := time.Duration(entryTotalOutput.LastRequestTimestamp-entryTotalOutput.StartTime) * time.Millisecond
	currentTime := time.Now()
	println(fmt.Sprint("=========================================== Statistics Summary =========================================="))
	println(fmt.Sprintf("Current time: %s, Users: %v, Duration: %v, Accumulated Transactions: %d Passed, %d Failed",
		currentTime.Format("2006/01/02 15:04:05"), atomic.LoadInt32(&r.currentClientsNum), duration, r.stats.transactionPassed, r.stats.transactionFailed))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "# requests", "# fails", "Median", "Average", "Min", "Max", "Content Size", "# reqs/sec", "# fails/sec"})
	row := make([]string, 10)
	row[0] = entryTotalOutput.Name
	row[1] = strconv.FormatInt(entryTotalOutput.NumRequests, 10)
	row[2] = strconv.FormatInt(entryTotalOutput.NumFailures, 10)
	row[3] = strconv.FormatInt(entryTotalOutput.medianResponseTime, 10)
	row[4] = strconv.FormatFloat(entryTotalOutput.avgResponseTime, 'f', 2, 64)
	row[5] = strconv.FormatInt(entryTotalOutput.MinResponseTime, 10)
	row[6] = strconv.FormatInt(entryTotalOutput.MaxResponseTime, 10)
	row[7] = strconv.FormatInt(entryTotalOutput.avgContentLength, 10)
	row[8] = strconv.FormatFloat(entryTotalOutput.currentRps, 'f', 2, 64)
	row[9] = strconv.FormatFloat(entryTotalOutput.currentFailPerSec, 'f', 2, 64)
	table.Append(row)
	table.Render()
	println()
}

func (r *localRunner) spawnWorkers(spawnCount int, spawnRate float64, quit chan bool, spawnCompleteFunc func()) {
	log.Info().
		Int("spawnCount", spawnCount).
		Float64("spawnRate", spawnRate).
		Msg("Spawning workers")

	atomic.StoreInt32(&r.state, stateSpawning)
	for i := 1; i <= spawnCount; i++ {
		// spawn workers with rate limit
		sleepTime := time.Duration(1000000/r.spawnRate) * time.Microsecond
		time.Sleep(sleepTime)

		// loop count per worker
		var workerLoop *Loop
		if r.loop != nil {
			workerLoop = &Loop{loopCount: atomic.LoadInt64(&r.loop.loopCount) / int64(r.spawnCount)}
		}

		select {
		case <-quit:
			// quit spawning goroutine
			log.Info().Msg("Quitting spawning workers")
			return
		default:
			atomic.AddInt32(&r.currentClientsNum, 1)
			go func() {
				for {
					select {
					case <-quit:
						return
					default:
						if workerLoop != nil && !workerLoop.acquire() {
							return
						}
						if r.rateLimitEnabled {
							blocked := r.rateLimiter.Acquire()
							if !blocked {
								task := r.getTask()
								r.safeRun(task.Fn)
							}
						} else {
							task := r.getTask()
							r.safeRun(task.Fn)
						}
						if workerLoop != nil {
							// finished count of total
							r.loop.increaseFinishedCount()
							// finished count of single worker
							workerLoop.increaseFinishedCount()
							if r.loop.isFinished() {
								r.stop()
							}
						}
					}
				}
			}()
		}
	}

	close(r.spawnDone)
	if spawnCompleteFunc != nil {
		spawnCompleteFunc()
	}
	atomic.StoreInt32(&r.state, stateRunning)
}

// setTasks will set the runner's task list AND the total task weight
// which is used to get a random task later
func (r *runner) setTasks(t []*Task) {
	r.tasks = t

	weightSum := 0
	for _, task := range r.tasks {
		weightSum += task.Weight
	}
	r.totalTaskWeight = weightSum
}

func (r *runner) getTask() *Task {
	tasksCount := len(r.tasks)
	if tasksCount == 1 {
		// Fast path
		return r.tasks[0]
	}

	rs := rand.New(rand.NewSource(time.Now().UnixNano()))

	totalWeight := r.totalTaskWeight
	if totalWeight <= 0 {
		// If all the tasks have not weights defined, they have the same chance to run
		randNum := rs.Intn(tasksCount)
		return r.tasks[randNum]
	}

	randNum := rs.Intn(totalWeight)
	runningSum := 0
	for _, task := range r.tasks {
		runningSum += task.Weight
		if runningSum > randNum {
			return task
		}
	}

	return nil
}

type localRunner struct {
	runner

	// close this channel will stop all goroutines used in runner.
	stopChan chan bool
}

func newLocalRunner(spawnCount int, spawnRate float64) *localRunner {
	return &localRunner{
		runner: runner{
			state:      stateInit,
			spawnRate:  spawnRate,
			spawnCount: spawnCount,
			stats:      newRequestStats(),
			outputs:    make([]Output, 0),
			spawnDone:  make(chan struct{}),
		},
		stopChan: make(chan bool),
	}
}

func (r *localRunner) start() {
	// init state
	atomic.StoreInt32(&r.state, stateInit)
	atomic.StoreInt32(&r.currentClientsNum, 0)
	r.stats.clearAll()

	// start rate limiter
	if r.rateLimitEnabled {
		r.rateLimiter.Start()
	}

	// all running workers(goroutines) will select on this channel.
	// close this channel will stop all running workers.
	quitChan := make(chan bool)
	// when this channel is closed, all statistics are reported successfully
	reportedChan := make(chan bool)
	go r.spawnWorkers(r.spawnCount, r.spawnRate, quitChan, nil)

	// output setup
	r.outputOnStart()

	// start running
	go func() {
		var ticker = time.NewTicker(reportStatsInterval)
		for {
			select {
			// record stats
			case t := <-r.stats.transactionChan:
				r.stats.logTransaction(t.name, t.success, t.elapsedTime, t.contentSize)
			case m := <-r.stats.requestSuccessChan:
				r.stats.logRequest(m.requestType, m.name, m.responseTime, m.responseLength)
			case n := <-r.stats.requestFailureChan:
				r.stats.logRequest(n.requestType, n.name, n.responseTime, 0)
				r.stats.logError(n.requestType, n.name, n.errMsg)
			// report stats
			case <-ticker.C:
				r.reportStats()
				// close reportedChan and return if the last stats is reported successfully
				if atomic.LoadInt32(&r.state) == stateQuitting {
					close(reportedChan)
					return
				}
			}
		}
	}()

	// stop
	<-r.stopChan
	atomic.StoreInt32(&r.state, stateQuitting)

	// stop previous goroutines without blocking
	// those goroutines will exit when r.safeRun returns
	close(quitChan)

	// wait until all stats are reported successfully
	<-reportedChan

	// stop rate limiter
	if r.rateLimitEnabled {
		r.rateLimiter.Stop()
	}

	// report test result
	r.reportTestResult()

	// output teardown
	r.outputOnStop()

	atomic.StoreInt32(&r.state, stateStopped)
	return
}

func (r *localRunner) stop() {
	close(r.stopChan)
}
