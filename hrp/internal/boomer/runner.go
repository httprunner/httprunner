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

	"github.com/go-errors/errors"

	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
)

const (
	StateInit     = iota + 1 // initializing
	StateSpawning            // spawning
	StateRunning             // running
	StateStopping            // stopping
	StateStopped             // stopped
	StateQuitting            // quitting
	StateMissing             // missing
)

const (
	reportStatsInterval = 3 * time.Second
	heartbeatInterval   = 1 * time.Second
	heartbeatLiveness   = 3 * time.Second
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

type Controller struct {
	mutex             sync.RWMutex
	once              sync.Once
	currentClientsNum int64 // current clients count
	spawnCount        int64 // target clients to spawn
	spawnRate         float64
	spawnDone         chan struct{}
	tasks             []*Task
}

func (c *Controller) setSpawn(spawnCount int64, spawnRate float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if spawnCount > 0 {
		atomic.StoreInt64(&c.spawnCount, spawnCount)
	}
	if spawnRate > 0 {
		c.spawnRate = spawnRate
	}
}

func (c *Controller) setSpawnCount(spawnCount int64) {
	if spawnCount > 0 {
		atomic.StoreInt64(&c.spawnCount, spawnCount)
	}
}

func (c *Controller) setSpawnRate(spawnRate float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if spawnRate > 0 {
		c.spawnRate = spawnRate
	}
}

func (c *Controller) getSpawnCount() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return atomic.LoadInt64(&c.spawnCount)
}

func (c *Controller) getSpawnRate() float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.spawnRate
}

func (c *Controller) getSpawnDone() chan struct{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.spawnDone
}

func (c *Controller) getCurrentClientsNum() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return atomic.LoadInt64(&c.currentClientsNum)
}

func (c *Controller) spawnCompete() {
	close(c.spawnDone)
}

func (c *Controller) isFinished() bool {
	// return true when workers acquired
	return atomic.LoadInt64(&c.currentClientsNum) == atomic.LoadInt64(&c.spawnCount)
}

func (c *Controller) acquire() bool {
	// get one ticket when there are still remaining spawn count to test
	// return true when getting ticket successfully
	if atomic.LoadInt64(&c.currentClientsNum) < atomic.LoadInt64(&c.spawnCount) {
		atomic.AddInt64(&c.currentClientsNum, 1)
		return true
	}
	return false
}

func (c *Controller) erase() bool {
	// return true if acquiredCount > spawnCount
	if atomic.LoadInt64(&c.currentClientsNum) > atomic.LoadInt64(&c.spawnCount) {
		atomic.AddInt64(&c.currentClientsNum, -1)
		return true
	}
	return false
}

func (c *Controller) increaseFinishedCount() {
	atomic.AddInt64(&c.currentClientsNum, -1)
}

func (c *Controller) reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.spawnCount = 0
	c.spawnRate = 0
	c.currentClientsNum = 0
	c.spawnDone = make(chan struct{})
	c.tasks = []*Task{}
	c.once = sync.Once{}
}

type runner struct {
	state int32

	tasks           []*Task
	totalTaskWeight int
	mutex           sync.RWMutex

	rateLimiter      RateLimiter
	rateLimitEnabled bool
	stats            *requestStats

	spawnCount int64 // target clients to spawn
	spawnRate  float64

	controller *Controller
	loop       *Loop // specify loop count for testcase, count = loopCount * spawnCount

	// rebalance spawn
	rebalance chan bool

	// all running workers(goroutines) will select on this channel.
	// close this channel will stop all running workers.
	stopChan chan bool

	// close this channel will stop all goroutines used in runner.
	closeChan chan bool

	outputs []Output

	once *sync.Once
}

func (r *runner) setSpawnRate(spawnRate float64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if spawnRate > 0 {
		r.spawnRate = spawnRate
	}
}

func (r *runner) getSpawnRate() float64 {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.spawnRate
}

func (r *runner) getSpawnCount() int64 {
	return atomic.LoadInt64(&r.spawnCount)
}

func (r *runner) setSpawnCount(spawnCount int64) {
	atomic.StoreInt64(&r.spawnCount, spawnCount)
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
	defer func() {
		r.outputs = make([]Output, 0)
	}()
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
	data["user_count"] = r.controller.getCurrentClientsNum()
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
		currentTime.Format("2006/01/02 15:04:05"), r.controller.getCurrentClientsNum(), duration, r.stats.transactionPassed, r.stats.transactionFailed))
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

func (r *runner) reset() {
	r.updateState(StateInit)
	r.controller.reset()
	r.stats.clearAll()
	r.rebalance = make(chan bool)
	r.stopChan = make(chan bool)
}

func (r *runner) spawnWorkers(spawnCount int64, spawnRate float64, quit chan bool, spawnCompleteFunc func()) {
	r.updateState(StateSpawning)
	log.Info().
		Int64("spawnCount", spawnCount).
		Float64("spawnRate", spawnRate).
		Msg("Spawning workers")

	r.controller.setSpawn(spawnCount, spawnRate)

	for {
		select {
		case <-quit:
			// quit spawning goroutine
			log.Info().Msg("Quitting spawning workers")
			return
		default:
			if r.isStarted() && r.controller.acquire() {
				// spawn workers with rate limit
				sleepTime := time.Duration(1000000/r.controller.getSpawnRate()) * time.Microsecond
				time.Sleep(sleepTime)

				// loop count per worker
				var workerLoop *Loop
				if r.loop != nil {
					workerLoop = &Loop{loopCount: atomic.LoadInt64(&r.loop.loopCount) / r.controller.spawnCount}
				}
				go func() {
					for {
						select {
						case <-quit:
							atomic.AddInt64(&r.controller.currentClientsNum, -1)
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
									close(r.rebalance)
								}
							}
							if r.controller.erase() {
								return
							}
						}
					}
				}()
				continue
			}

			r.controller.once.Do(func() {
				// spawning compete
				r.controller.spawnCompete()
				if spawnCompleteFunc != nil {
					spawnCompleteFunc()
				}
				r.updateState(StateRunning)
			})

			<-r.rebalance
			if r.isStarted() {
				// rebalance spawn count
				r.controller.setSpawn(r.getSpawnCount(), r.getSpawnRate())
			}
		}
	}
}

// setTasks will set the runner's task list AND the total task weight
// which is used to get a random task later
func (r *runner) setTasks(t []*Task) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.tasks = t

	weightSum := 0
	for _, task := range r.tasks {
		weightSum += task.Weight
	}
	r.totalTaskWeight = weightSum
}

func (r *runner) getTask() *Task {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	tasksCount := len(r.tasks)
	if tasksCount == 0 {
		log.Error().Msg("no valid testcase found")
		os.Exit(1)
	} else if tasksCount == 1 {
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

func (r *runner) statsStart() {
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
			if !r.isStarted() {
				log.Info().Msg("Quitting statsStart")
				return
			}
		}
	}
}

func (r *runner) stop() {
	// stop previous goroutines without blocking
	// those goroutines will exit when r.safeRun returns
	close(r.stopChan)
	if r.rateLimitEnabled {
		r.rateLimiter.Stop()
	}
}

func (r *runner) getState() int32 {
	return atomic.LoadInt32(&r.state)
}

func (r *runner) updateState(state int32) {
	log.Debug().Int32("from", atomic.LoadInt32(&r.state)).Int32("to", state).Msg("update runner state")
	atomic.StoreInt32(&r.state, state)
}

func (r *runner) isStarted() bool {
	return r.getState() == StateRunning || r.getState() == StateSpawning
}

type localRunner struct {
	runner

	profile *Profile
}

func newLocalRunner(spawnCount int64, spawnRate float64) *localRunner {
	return &localRunner{
		runner: runner{
			state:      StateInit,
			stats:      newRequestStats(),
			spawnCount: spawnCount,
			spawnRate:  spawnRate,
			controller: &Controller{},
			outputs:    make([]Output, 0),
			closeChan:  make(chan bool),
			once:       &sync.Once{},
		},
	}
}

func (r *localRunner) start() {
	// init localRunner
	r.reset()

	// start rate limiter
	if r.rateLimitEnabled {
		r.rateLimiter.Start()
	}
	// output setup
	r.outputOnStart()

	go r.spawnWorkers(r.getSpawnCount(), r.getSpawnRate(), r.stopChan, nil)

	// start stats report
	r.statsStart()

	// stop
	<-r.stopChan
	r.updateState(StateStopped)

	// stop rate limiter
	if r.rateLimitEnabled {
		r.rateLimiter.Stop()
	}

	// report test result
	r.reportTestResult()

	// output teardown
	r.outputOnStop()

	r.updateState(StateQuitting)
	return
}

func (r *localRunner) stop() {
	if r.runner.isStarted() {
		r.runner.stop()
		close(r.rebalance)
	}
}

// workerRunner connects to the master, spawns goroutines and collects stats.
type workerRunner struct {
	runner

	nodeID     string
	masterHost string
	masterPort int
	client     *grpcClient

	profile *Profile

	tasksChan chan *profileMessage

	ignoreQuit bool
}

func newWorkerRunner(masterHost string, masterPort int) (r *workerRunner) {
	r = &workerRunner{
		runner: runner{
			stats:      newRequestStats(),
			outputs:    make([]Output, 0),
			controller: &Controller{},
			closeChan:  make(chan bool),
			once:       &sync.Once{},
		},
		masterHost: masterHost,
		masterPort: masterPort,
		nodeID:     getNodeID(),
		tasksChan:  make(chan *profileMessage, 10),
	}
	return r
}

func (r *workerRunner) spawnComplete() {
	data := make(map[string]int64)
	data["count"] = r.controller.getSpawnCount()
	r.client.sendChannel() <- newGenericMessage("spawning_complete", data, r.nodeID)
}

func (r *workerRunner) onSpawnMessage(msg *genericMessage) {
	r.client.sendChannel() <- newGenericMessage("spawning", nil, r.nodeID)
	if msg.Profile == nil {
		log.Error().Msg("miss profile")
	}
	profile := BytesToProfile(msg.Profile)
	r.setSpawnCount(profile.SpawnCount)
	r.setSpawnRate(profile.SpawnRate)

	if msg.Tasks == nil && len(r.tasks) == 0 {
		log.Error().Msg("miss tasks")
	}
	r.tasksChan <- &profileMessage{
		Profile: profile,
		Tasks:   msg.Tasks,
	}
	log.Info().Msg("on spawn message successful")
}

func (r *workerRunner) onRebalanceMessage(msg *genericMessage) {
	if msg.Profile == nil {
		log.Error().Msg("miss profile")
	}
	profile := BytesToProfile(msg.Profile)
	r.setSpawnCount(profile.SpawnCount)
	r.setSpawnRate(profile.SpawnRate)

	r.tasksChan <- &profileMessage{
		Profile: profile,
	}
	log.Info().Msg("on rebalance message successful")
}

// Runner acts as a state machine.
func (r *workerRunner) onMessage(msg *genericMessage) {
	switch r.getState() {
	case StateInit:
		switch msg.Type {
		case "spawn":
			r.onSpawnMessage(msg)
		case "quit":
			r.close()
		}
	case StateSpawning:
		fallthrough
	case StateRunning:
		switch msg.Type {
		case "spawn":
			r.onSpawnMessage(msg)
		case "rebalance":
			r.onRebalanceMessage(msg)
		case "stop":
			r.stop()
			log.Info().Msg("Recv stop message from master, all the goroutines are stopped")
			r.client.sendChannel() <- newGenericMessage("client_stopped", nil, r.nodeID)
		case "quit":
			r.close()
			log.Info().Msg("Recv quit message from master, all the goroutines are stopped")
		}
	case StateStopped:
		switch msg.Type {
		case "spawn":
			r.onSpawnMessage(msg)
			go r.start()
		case "quit":
			r.close()
		}
	}
}

func (r *workerRunner) onQuiting() {
	if r.getState() != StateQuitting {
		r.client.sendChannel() <- newQuitMessage(r.nodeID)
	}
	r.updateState(StateQuitting)
}

func (r *workerRunner) startListener() {
	for {
		select {
		case msg := <-r.client.recvChannel():
			r.onMessage(msg)
		case <-r.closeChan:
			return
		}
	}
}

// run worker service
func (r *workerRunner) run() {
	r.updateState(StateInit)
	r.client = newClient(r.masterHost, r.masterPort, r.nodeID)

	err := r.client.connect()
	if err != nil {
		log.Printf("Failed to connect to master(%s:%d) with error %v\n", r.masterHost, r.masterPort, err)
	}

	// listen to master
	go r.startListener()

	// register worker information to master
	r.client.sendChannel() <- newGenericMessage("register", nil, r.nodeID)
	// tell master, I'm ready
	log.Info().Msg("send client ready signal")
	r.client.sendChannel() <- newClientReadyMessageToMaster(r.nodeID)

	// heartbeat
	// See: https://github.com/locustio/locust/commit/a8c0d7d8c588f3980303358298870f2ea394ab93
	go func() {
		var ticker = time.NewTicker(heartbeatInterval)
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&r.client.failCount) > 2 {
					r.updateState(StateMissing)
				}
				if r.getState() == StateMissing {
					if r.client.reConnect() == nil {
						r.updateState(StateInit)
					}
				}
				CPUUsage := GetCurrentCPUUsage()
				data := map[string]int64{
					"state":             int64(r.getState()),
					"current_cpu_usage": int64(CPUUsage),
					"spawn_count":       r.controller.getCurrentClientsNum(),
				}
				r.client.sendChannel() <- newGenericMessage("heartbeat", data, r.nodeID)
			case <-r.closeChan:
				return
			}
		}
	}()
	<-r.closeChan
}

// start load test
func (r *workerRunner) start() {
	r.reset()

	// start rate limiter
	if r.rateLimitEnabled {
		r.rateLimiter.Start()
	}

	r.once.Do(r.outputOnStart)

	go r.spawnWorkers(r.getSpawnCount(), r.getSpawnRate(), r.stopChan, r.spawnComplete)

	// start stats report
	r.statsStart()

	r.reportTestResult()
	r.outputOnStop()
}

func (r *workerRunner) stop() {
	if r.isStarted() {
		r.runner.stop()
		close(r.rebalance)
		r.updateState(StateStopped)
	}
}

func (r *workerRunner) close() {
	r.stop()
	if r.ignoreQuit {
		return
	}
	// waiting report finished
	time.Sleep(1 * time.Second)
	close(r.closeChan)
	var ticker = time.NewTicker(1 * time.Second)
	if r.client != nil {
		// waitting for quit message is sent to master
		select {
		case <-r.client.disconnectedChannel():
			break
		case <-ticker.C:
			log.Warn().Msg("Timeout waiting for sending quit message to master, boomer will quit any way.")
			r.onQuiting()
		}
		r.client.close()
	}
}

// masterRunner controls worker to spawn goroutines and collect stats.
type masterRunner struct {
	runner

	masterBindHost string
	masterBindPort int
	server         *grpcServer

	autoStart            bool
	expectWorkers        int
	expectWorkersMaxWait int

	profile *Profile

	parseTestCasesChan chan bool
	testCaseBytes      chan []byte
	// set profile to worker
	profileBytes chan []byte
}

func newMasterRunner(masterBindHost string, masterBindPort int) *masterRunner {
	return &masterRunner{
		runner: runner{
			state:     StateInit,
			closeChan: make(chan bool),
		},
		masterBindHost:     masterBindHost,
		masterBindPort:     masterBindPort,
		server:             newServer(masterBindHost, masterBindPort),
		parseTestCasesChan: make(chan bool),
		testCaseBytes:      make(chan []byte),
	}
}

func (r *masterRunner) setExpectWorkers(expectWorkers int, expectWorkersMaxWait int) {
	r.expectWorkers = expectWorkers
	r.expectWorkersMaxWait = expectWorkersMaxWait
}

func (r *masterRunner) heartbeatWorker() {
	log.Info().Msg("heartbeatWorker, listen and record heartbeat from worker")
	var ticker = time.NewTicker(heartbeatInterval)
	for {
		select {
		case <-r.closeChan:
			return
		case <-ticker.C:
			r.server.clients.Range(func(key, value interface{}) bool {
				workerInfo, ok := value.(*WorkerNode)
				if !ok {
					log.Error().Msg("failed to get worker information")
				}
				if atomic.LoadInt32(&workerInfo.Heartbeat) <= 0 && workerInfo.getState() != StateMissing {
					workerInfo.setState(StateMissing)
					if r.getState() == StateRunning {
						// all running workers missed, stopping runner
						if r.server.getClientsLength() <= 0 {
							r.updateState(StateStopped)
						}
					}
				} else {
					atomic.AddInt32(&workerInfo.Heartbeat, -1)
				}
				return true
			})
		}
	}
}

func (r *masterRunner) clientListener() {
	log.Info().Msg("clientListener, start to deal message from worker")
	for {
		select {
		case <-r.closeChan:
			return
		case msg := <-r.server.recvChannel():
			worker, ok := r.server.getClients().Load(msg.NodeID)
			if !ok {
				continue
			}
			workerInfo, ok := worker.(*WorkerNode)
			if !ok {
				continue
			}
			switch msg.Type {
			case typeClientReady:
				if workerInfo.getState() == StateInit {
					break
				}
				workerInfo.setState(StateInit)
				if r.getState() == StateRunning {
					println(fmt.Sprintf("worker(%s) joined, ready to rebalance the load of each worker", workerInfo.ID))
					err := r.rebalance()
					if err != nil {
						log.Error().Err(err).Msg("failed to rebalance")
					}
				}
			case typeClientStopped:
				workerInfo.setState(StateStopped)
				if r.server.getWorkersLengthByState(StateStopped)+r.server.getWorkersLengthByState(StateInit) == r.server.getClientsLength() {
					r.updateState(StateStopped)
				}
			case typeHeartbeat:
				if workerInfo.getState() != int32(msg.Data["state"]) {
					workerInfo.setState(int32(msg.Data["state"]))
				}
				workerInfo.updateHeartbeat(3)
				if workerInfo.getCPUUsage() != float64(msg.Data["current_cpu_usage"]) {
					workerInfo.updateCPUUsage(float64(msg.Data["current_cpu_usage"]))
				}
				if workerInfo.getSpawnCount() != msg.Data["spawn_count"] {
					workerInfo.updateSpawnCount(msg.Data["spawn_count"])
				}
			case typeSpawning:
				workerInfo.setState(StateSpawning)
			case typeSpawningComplete:
				workerInfo.setState(StateRunning)
				if r.server.getWorkersLengthByState(StateRunning) == r.server.getClientsLength() {
					println(fmt.Sprintf("all(%v) workers spawn done, setting state as running", r.server.getClientsLength()))
					r.updateState(StateRunning)
				}
			case typeQuit:
				if workerInfo.getState() == StateQuitting {
					break
				}
				workerInfo.setState(StateQuitting)
				if r.isStarted() {
					if r.server.getClientsLength() > 0 {
						println(fmt.Sprintf("worker(%s) quited, ready to rebalance the load of each worker", workerInfo.ID))
						err := r.rebalance()
						if err != nil {
							log.Error().Err(err).Msg("failed to rebalance")
						}
					}
				}
			case typeException:
				// Todo
			default:
			}
		}
	}
}

func (r *masterRunner) run() {
	r.updateState(StateInit)

	// start grpc server
	err := r.server.start()
	if err != nil {
		log.Error().Err(err).Msg("failed to start grpc server")
		return
	}

	// listen and deal message from worker
	go r.clientListener()
	// listen and record heartbeat from worker
	go r.heartbeatWorker()

	if r.autoStart {
		log.Info().Msg("auto start, waiting expected workers joined")
		var ticker = time.NewTicker(1 * time.Second)
		var tickerMaxWait = time.NewTicker(time.Duration(r.expectWorkersMaxWait) * time.Second)
	FOR:
		for {
			select {
			case <-r.closeChan:
				return
			case <-ticker.C:
				c := r.server.getClientsLength()
				log.Info().Msg(fmt.Sprintf("expected worker number: %v, current worker count: %v", r.expectWorkers, c))
				if c >= r.expectWorkers {
					go func() {
						err = r.start()
						if err != nil {
							log.Error().Err(err).Msg("failed to run")
							os.Exit(1)
						}
					}()
					break FOR
				}
			case <-tickerMaxWait.C:
				log.Warn().Msg("reached max wait time, quiting")
				r.onQuiting()
				os.Exit(1)
			}
		}
	}
	<-r.closeChan
}

func (r *masterRunner) start() error {
	numWorkers := r.server.getClientsLength()
	if numWorkers == 0 {
		return errors.New("current workers: 0")
	}

	log.Info().Msg("send spawn data to worker")
	r.updateState(StateSpawning)
	// fetching testcase
	testcase, err := r.fetchTestCase()
	if err != nil {
		return err
	}
	profile := r.profile.dispatch(int64(numWorkers))

	r.server.sendChannel() <- newMessageToWorker("spawn", ProfileToBytes(profile), nil, testcase)
	println("send spawn data to worker successful")
	log.Info().Msg("send spawn data to worker successful")
	return nil
}

func (r *masterRunner) rebalance() error {
	numWorkers := r.server.getClientsLength()
	if numWorkers == 0 {
		return errors.New("current workers: 0")
	}
	profile := r.profile.dispatch(int64(numWorkers))

	r.server.sendChannel() <- newMessageToWorker("rebalance", ProfileToBytes(profile), nil, nil)
	println("send rebalance data to worker successful")
	return nil
}

func (r *masterRunner) fetchTestCase() ([]byte, error) {
	ticker := time.NewTicker(30 * time.Second)
	if len(r.testCaseBytes) > 0 {
		<-r.testCaseBytes
	}
	r.parseTestCasesChan <- true
	select {
	case <-ticker.C:
		return nil, errors.New("parse testcases timeout")
	case tcb := <-r.testCaseBytes:
		return tcb, nil
	}
}

func (r *masterRunner) stop() error {
	if r.isStarted() {
		r.updateState(StateStopping)
		r.server.sendChannel() <- &genericMessage{Type: "stop", Data: map[string]int64{}}
		r.updateState(StateStopped)
		return nil
	} else {
		return errors.New("already stopped")
	}
}

func (r *masterRunner) onQuiting() {
	if r.getState() != StateQuitting {
		r.server.sendChannel() <- &genericMessage{
			Type: "quit",
		}
	}
	r.updateState(StateQuitting)
}

func (r *masterRunner) close() {
	r.onQuiting()
	r.server.wg.Wait()
	close(r.closeChan)
	r.server.close()
}
