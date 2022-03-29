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
	stateStopping            // stopping
	stateStopped             // stopped
	stateQuitting            // quitting
	stateMissing             // missing
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

	// when this channel is closed, all statistics are reported successfully
	reportedChan chan bool

	// all running workers(goroutines) will select on this channel.
	// close this channel will stop all running workers.
	stopChan chan bool

	// close this channel will stop all goroutines used in runner.
	closeChan chan bool

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
	duration := time.Duration(entryTotalOutput.LastRequestTimestamp-entryTotalOutput.StartTime) * time.Second
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

func (r *runner) startSpawning(spawnCount int, spawnRate float64, spawnCompleteFunc func()) {
	r.stats.clearAll()
	r.stopChan = make(chan bool)
	r.spawnDone = make(chan struct{})
	r.reportedChan = make(chan bool)

	r.spawnCount = spawnCount
	r.spawnRate = spawnRate

	atomic.StoreInt32(&r.currentClientsNum, 0)

	go r.spawnWorkers(spawnCount, spawnRate, r.stopChan, spawnCompleteFunc)
}

func (r *runner) spawnWorkers(spawnCount int, spawnRate float64, quit chan bool, spawnCompleteFunc func()) {
	log.Info().
		Int("spawnCount", spawnCount).
		Float64("spawnRate", spawnRate).
		Msg("Spawning workers")

	r.updateState(stateSpawning)
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
	r.updateState(stateRunning)
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
			// close reportedChan and return if the last stats is reported successfully
			if atomic.LoadInt32(&r.state) != stateSpawning && atomic.LoadInt32(&r.state) != stateRunning {
				close(r.reportedChan)
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

type localRunner struct {
	runner
}

func newLocalRunner(spawnCount int, spawnRate float64) *localRunner {
	return &localRunner{
		runner: runner{
			state:        stateInit,
			spawnRate:    spawnRate,
			spawnCount:   spawnCount,
			stats:        newRequestStats(),
			outputs:      make([]Output, 0),
			spawnDone:    make(chan struct{}),
			reportedChan: make(chan bool),
			stopChan:     make(chan bool),
			closeChan:    make(chan bool),
		},
	}
}

func (r *localRunner) start() {
	// init state
	r.updateState(stateInit)
	atomic.StoreInt32(&r.currentClientsNum, 0)
	r.stats.clearAll()

	// start rate limiter
	if r.rateLimitEnabled {
		r.rateLimiter.Start()
	}

	go r.spawnWorkers(r.spawnCount, r.spawnRate, r.stopChan, nil)

	// output setup
	r.outputOnStart()

	// start stats report
	go r.runner.statsStart()

	// stop
	<-r.stopChan
	r.updateState(stateStopped)

	// wait until all stats are reported successfully
	<-r.reportedChan

	// stop rate limiter
	if r.rateLimitEnabled {
		r.rateLimiter.Stop()
	}

	// report test result
	r.reportTestResult()

	// output teardown
	r.outputOnStop()

	r.updateState(stateQuitting)
	return
}

func (r *localRunner) stop() {
	close(r.stopChan)
}

// workerRunner connects to the master, spawns goroutines and collects stats.
type workerRunner struct {
	runner

	nodeID     string
	masterHost string
	masterPort int
	client     *grpcClient

	// this channel will start worker for spawning.
	spawnStartChan chan bool
}

func newWorkerRunner(masterHost string, masterPort int) (r *workerRunner) {
	r = &workerRunner{
		runner: runner{
			stats:        newRequestStats(),
			spawnDone:    make(chan struct{}),
			stopChan:     make(chan bool),
			reportedChan: make(chan bool),
			closeChan:    make(chan bool),
		},
		masterHost:     masterHost,
		masterPort:     masterPort,
		nodeID:         getNodeID(),
		spawnStartChan: make(chan bool),
	}
	return r
}

func (r *workerRunner) spawnComplete() {
	data := make(map[string]int64)
	data["count"] = int64(r.spawnCount)
	r.client.sendChannel() <- newGenericMessage("spawning_complete", data, r.nodeID)
	r.updateState(stateRunning)
}

func (r *workerRunner) onSpawnMessage(msg *genericMessage) {
	r.client.sendChannel() <- newGenericMessage("spawning", nil, r.nodeID)
	spawnCount, ok := msg.Data["spawn_count"]
	if ok {
		r.spawnCount = int(spawnCount)
	}
	spawnRate, ok := msg.Data["spawn_rate"]
	if ok {
		r.spawnRate = float64(spawnRate)
	}

	// start rate limiter
	if r.rateLimitEnabled {
		r.rateLimiter.Start()
	}

	r.spawnStartChan <- true
}

// Runner acts as a state machine.
func (r *workerRunner) onMessage(msg *genericMessage) {
	switch r.getState() {
	case stateInit:
		switch msg.Type {
		case "spawn":
			r.onSpawnMessage(msg)
		case "quit":
			r.close()
		}
	case stateSpawning:
		fallthrough
	case stateRunning:
		switch msg.Type {
		case "spawn":
			r.stop()
			r.onSpawnMessage(msg)
		case "stop":
			r.stop()
			log.Info().Msg("Recv stop message from master, all the goroutines are stopped")
			r.client.sendChannel() <- newGenericMessage("client_stopped", nil, r.nodeID)
		case "quit":
			r.close()
			log.Info().Msg("Recv quit message from master, all the goroutines are stopped")
		}
	case stateStopped:
		switch msg.Type {
		case "spawn":
			r.onSpawnMessage(msg)
		case "quit":
			r.close()
		}
	}
}

func (r *workerRunner) onQuiting() {
	if r.getState() != stateQuitting {
		r.client.sendChannel() <- newGenericMessage("quit", nil, r.nodeID)
	}
	r.updateState(stateQuitting)
}

func (r *workerRunner) startListener() {
	go func() {
		for {
			select {
			case msg := <-r.client.recvChannel():
				r.onMessage(msg)
			case <-r.closeChan:
				return
			}
		}
	}()
}

func (r *workerRunner) run() {
	r.updateState(stateInit)
	r.client = newClient(r.masterHost, r.masterPort, r.nodeID)

	err := r.client.connect()
	if err != nil {
		log.Printf("Failed to connect to master(%s:%d) with error %v\n", r.masterHost, r.masterPort, err)
		return
	}

	// listen to master
	r.startListener()

	go r.client.recv()
	go r.client.send()

	// register worker info to master
	r.client.sendChannel() <- newGenericMessage("register", nil, r.nodeID)
	// tell master, I'm ready
	log.Info().Msg(fmt.Sprint("send client ready signal"))
	r.client.sendChannel() <- newClientReadyMessage(r.nodeID)

	// heartbeat
	// See: https://github.com/locustio/locust/commit/a8c0d7d8c588f3980303358298870f2ea394ab93
	go func() {
		var ticker = time.NewTicker(heartbeatInterval)
		for {
			select {
			case <-ticker.C:
				CPUUsage := GetCurrentCPUUsage()
				data := map[string]int64{
					"state":             int64(r.getState()),
					"current_cpu_usage": int64(CPUUsage),
				}
				r.client.sendChannel() <- newGenericMessage("heartbeat", data, r.nodeID)
			case <-r.closeChan:
				return
			}
		}
	}()
	<-r.closeChan
}

func (r *workerRunner) start() {
	r.outputOnStart()
	for {
		select {
		case <-r.closeChan:
			return
		case <-r.spawnStartChan:
			r.stats.clearAll()
			r.startSpawning(r.spawnCount, r.spawnRate, r.spawnComplete)

			// start stats report
			go r.runner.statsStart()

			// stop
			<-r.stopChan

			<-r.reportedChan

			r.reportTestResult()
			r.outputOnStop()
			r.updateState(stateStopped)
		}
	}
}

func (r *workerRunner) stop() {
	if r.getState() == stateRunning || r.getState() == stateSpawning {
		close(r.stopChan)
		r.updateState(stateStopped)
	}
}

func (r *workerRunner) close() {
	var ticker = time.NewTicker(3 * time.Second)
	r.stop()
	if r.client != nil {
		r.onQuiting()
		// wait for quit message is sent to master
		select {
		case <-r.client.disconnectedChannel():
			break
		case <-ticker.C:
			log.Warn().Msg("Timeout waiting for sending quit message to master, boomer will quit any way.")
			break
		}
		r.client.close()
	}
	close(r.closeChan)
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

	startChan chan bool

	mutex sync.Mutex
}

func newMasterRunner(masterBindHost string, masterBindPort int) *masterRunner {
	return &masterRunner{
		runner: runner{
			state:        stateInit,
			spawnDone:    make(chan struct{}),
			stopChan:     make(chan bool),
			closeChan:    make(chan bool),
			reportedChan: make(chan bool),
		},
		masterBindHost: masterBindHost,
		masterBindPort: masterBindPort,
		server:         newServer(masterBindHost, masterBindPort),
		startChan:      make(chan bool),
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
				workerInfo, ok := value.(*workerNode)
				if !ok {
					log.Error().Msg("failed to get worker information")
				}
				if atomic.LoadInt32(&workerInfo.heartbeat) <= 0 && workerInfo.getState() != stateMissing {
					workerInfo.setState(stateMissing)
					if r.getState() == stateRunning {
						// all running workers missed, stopping runner
						if r.server.getClientsLength() <= 0 {
							r.stop()
						}
					}
				} else {
					atomic.AddInt32(&workerInfo.heartbeat, -1)
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
			workerInfo, ok := worker.(*workerNode)
			if !ok {
				continue
			}
			switch msg.Type {
			case typeClientReady:
				if workerInfo.getState() == stateInit {
					break
				}
				workerInfo.setState(stateInit)
				if r.getState() == stateRunning {
					println(fmt.Sprintf("worker(%s) joined, ready to rebalance the load of each worker", workerInfo.id))
					r.rebalance()
				}
			case typeClientStopped:
				workerInfo.setState(stateStopped)
				if len(r.server.getWorkersByState(stateStopped))+len(r.server.getWorkersByState(stateInit)) == r.server.getClientsLength() {
					r.updateState(stateStopped)
				}
			case typeHeartbeat:
				if workerInfo.getState() != int32(msg.Data["state"]) {
					workerInfo.setState(int32(msg.Data["state"]))
				}
				atomic.StoreInt32(&workerInfo.heartbeat, 3)
				workerInfo.cpuUsage = float64(msg.Data["current_cpu_usage"])
			case typeSpawning:
				workerInfo.setState(stateSpawning)
			case typeSpawningComplete:
				workerInfo.setState(stateRunning)
				if len(r.server.getWorkersByState(stateRunning)) == r.server.getClientsLength() {
					println(fmt.Sprintf("all(%v) workers spawn done, setting state as running", r.server.getClientsLength()))
					r.updateState(stateRunning)
				}
			case typeQuit:
				if workerInfo.getState() == stateQuitting {
					break
				}
				workerInfo.setState(stateQuitting)
				// disconnect
				r.server.disconnectedChannel() <- workerInfo.id
				if r.getState() == stateRunning || r.getState() == stateSpawning {
					if r.server.getClientsLength() == 0 {
						println(fmt.Sprintf("worker(%s) quited, current worker count: 0", workerInfo.id))
					} else {
						println(fmt.Sprintf("worker(%s) quited, ready to rebalance the load of each worker", workerInfo.id))
						r.rebalance()
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
	r.updateState(stateInit)

	// start grpc server
	err := r.server.start()
	if err != nil {
		log.Error().Err(err).Msg("failed to start grpc server")
		return
	}

	go r.server.recv()
	go r.server.send()

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
						r.startChan <- true
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

	for {
		select {
		case <-r.closeChan:
			return
		case <-r.startChan:
			r.stopChan = make(chan bool)
			numWorkers := r.server.getClientsLength()
			if numWorkers == 0 {
				return
			}
			workerSpawnRate := r.spawnRate / float64(numWorkers)
			workerSpawnCount := r.spawnCount / numWorkers

			if r.getState() != stateSpawning && r.getState() != stateRunning {
				log.Info().Msg("send spawn data to worker")
				r.updateState(stateSpawning)
				r.server.sendChannel() <- &genericMessage{
					Type: "spawn",
					Data: map[string]int64{
						"spawn_count": int64(workerSpawnCount),
						"spawn_rate":  int64(workerSpawnRate),
					},
				}
			}
			<-r.stopChan
			r.updateState(stateStopped)
		default:
		}
	}
}

func (r *masterRunner) rebalance() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.stop()
	if len(r.startChan) == 0 {
		r.startChan <- true
	}
}

func (r *masterRunner) stop() {
	if r.getState() == stateRunning || r.getState() == stateSpawning {
		r.updateState(stateStopping)
		r.server.sendChannel() <- &genericMessage{Type: "stop", Data: map[string]int64{}}
		close(r.stopChan)
		r.updateState(stateStopped)
	}
}

func (r *masterRunner) onQuiting() {
	if r.getState() != stateQuitting {
		r.server.sendChannel() <- &genericMessage{
			Type: "quit",
			Data: map[string]int64{},
		}
	}
	r.updateState(stateQuitting)
}

func (r *masterRunner) close() {
	var ticker = time.NewTicker(1 * time.Second)
	r.stop()
	r.onQuiting()
	// wait to notify all workers to quit
	<-ticker.C
	close(r.closeChan)
	r.server.close()
}
