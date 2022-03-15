package boomer

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/hrp/internal/json"
)

// Output is primarily responsible for printing test results to different destinations
// such as consoles, files. You can write you own output and add to boomer.
// When running in standalone mode, the default output is ConsoleOutput, you can add more.
// When running in distribute mode, test results will be reported to master with or without
// an output.
// All the OnXXX function will be call in a separated goroutine, just in case some output will block.
// But it will wait for all outputs return to avoid data lost.
type Output interface {
	// OnStart will be call before the test starts.
	OnStart()

	// By default, each output receive stats data from runner every three seconds.
	// OnEvent is responsible for dealing with the data.
	OnEvent(data map[string]interface{})

	// OnStop will be called before the test ends.
	OnStop()
}

// ConsoleOutput is the default output for standalone mode.
type ConsoleOutput struct {
}

// NewConsoleOutput returns a ConsoleOutput.
func NewConsoleOutput() *ConsoleOutput {
	return &ConsoleOutput{}
}

func getMedianResponseTime(numRequests int64, responseTimes map[int64]int64) int64 {
	medianResponseTime := int64(0)
	if len(responseTimes) != 0 {
		pos := (numRequests - 1) / 2
		var sortedKeys []int64
		for k := range responseTimes {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Slice(sortedKeys, func(i, j int) bool {
			return sortedKeys[i] < sortedKeys[j]
		})
		for _, k := range sortedKeys {
			if pos < responseTimes[k] {
				medianResponseTime = k
				break
			}
			pos -= responseTimes[k]
		}
	}
	return medianResponseTime
}

func getAvgResponseTime(numRequests int64, totalResponseTime int64) (avgResponseTime float64) {
	avgResponseTime = float64(0)
	if numRequests != 0 {
		avgResponseTime = float64(totalResponseTime) / float64(numRequests)
	}
	return avgResponseTime
}

func getAvgContentLength(numRequests int64, totalContentLength int64) (avgContentLength int64) {
	avgContentLength = int64(0)
	if numRequests != 0 {
		avgContentLength = totalContentLength / numRequests
	}
	return avgContentLength
}

func getCurrentRps(numRequests int64, duration float64) (currentRps float64) {
	currentRps = float64(numRequests) / duration
	return currentRps
}

func getCurrentFailPerSec(numFailures int64, duration float64) (currentFailPerSec float64) {
	currentFailPerSec = float64(numFailures) / duration
	return currentFailPerSec
}

func getTotalFailRatio(totalRequests, totalFailures int64) (failRatio float64) {
	if totalRequests == 0 {
		return 0
	}
	return float64(totalFailures) / float64(totalRequests)
}

// OnStart of ConsoleOutput has nothing to do.
func (o *ConsoleOutput) OnStart() {

}

// OnStop of ConsoleOutput has nothing to do.
func (o *ConsoleOutput) OnStop() {

}

// OnEvent will print to the console.
func (o *ConsoleOutput) OnEvent(data map[string]interface{}) {
	output, err := convertData(data)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert data")
		return
	}

	var state string
	switch output.State {
	case 1:
		state = "initializing"
	case 2:
		state = "spawning"
	case 3:
		state = "running"
	case 4:
		state = "quitting"
	case 5:
		state = "stopped"
	}

	currentTime := time.Now()
	println(fmt.Sprintf("Current time: %s, Users: %d, State: %s, Total RPS: %.1f, Total Average Response Time: %.1fms, Total Fail Ratio: %.1f%%",
		currentTime.Format("2006/01/02 15:04:05"), output.UserCount, state, output.TotalRPS, output.TotalAvgResponseTime, output.TotalFailRatio*100))
	println(fmt.Sprintf("Accumulated Transactions: %d Passed, %d Failed",
		output.TransactionsPassed, output.TransactionsFailed))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "Name", "# requests", "# fails", "Median", "Average", "Min", "Max", "Content Size", "# reqs/sec", "# fails/sec"})

	for _, stat := range output.Stats {
		row := make([]string, 11)
		row[0] = stat.Method
		row[1] = stat.Name
		row[2] = strconv.FormatInt(stat.NumRequests, 10)
		row[3] = strconv.FormatInt(stat.NumFailures, 10)
		row[4] = strconv.FormatInt(stat.medianResponseTime, 10)
		row[5] = strconv.FormatFloat(stat.avgResponseTime, 'f', 2, 64)
		row[6] = strconv.FormatInt(stat.MinResponseTime, 10)
		row[7] = strconv.FormatInt(stat.MaxResponseTime, 10)
		row[8] = strconv.FormatInt(stat.avgContentLength, 10)
		row[9] = strconv.FormatFloat(stat.currentRps, 'f', 2, 64)
		row[10] = strconv.FormatFloat(stat.currentFailPerSec, 'f', 2, 64)
		table.Append(row)
	}
	table.Render()
	println()
}

type statsEntryOutput struct {
	statsEntry

	medianResponseTime int64   // median response time
	avgResponseTime    float64 // average response time, round float to 2 decimal places
	avgContentLength   int64   // average content size
	currentRps         float64 // # reqs/sec
	currentFailPerSec  float64 // # fails/sec
}

type dataOutput struct {
	UserCount            int32                             `json:"user_count"`
	State                int32                             `json:"state"`
	TotalStats           *statsEntryOutput                 `json:"stats_total"`
	TransactionsPassed   int64                             `json:"transactions_passed"`
	TransactionsFailed   int64                             `json:"transactions_failed"`
	TotalAvgResponseTime float64                           `json:"total_avg_response_time"`
	TotalRPS             float64                           `json:"total_rps"`
	TotalFailRatio       float64                           `json:"total_fail_ratio"`
	Stats                []*statsEntryOutput               `json:"stats"`
	Errors               map[string]map[string]interface{} `json:"errors"`
}

func convertData(data map[string]interface{}) (output *dataOutput, err error) {
	userCount, ok := data["user_count"].(int32)
	if !ok {
		return nil, fmt.Errorf("user_count is not int32")
	}
	state, ok := data["state"].(int32)
	if !ok {
		return nil, fmt.Errorf("state is not int32")
	}
	stats, ok := data["stats"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("stats is not []interface{}")
	}

	errors := data["errors"].(map[string]map[string]interface{})

	transactions, ok := data["transactions"].(map[string]int64)
	if !ok {
		return nil, fmt.Errorf("transactions is not map[string]int64")
	}
	transactionsPassed := transactions["passed"]
	transactionsFailed := transactions["failed"]

	// convert stats in total
	statsTotal, ok := data["stats_total"].(interface{})
	if !ok {
		return nil, fmt.Errorf("stats_total is not interface{}")
	}
	entryTotalOutput, err := deserializeStatsEntry(statsTotal)
	if err != nil {
		return nil, err
	}

	output = &dataOutput{
		UserCount:            userCount,
		State:                state,
		TotalStats:           entryTotalOutput,
		TransactionsPassed:   transactionsPassed,
		TransactionsFailed:   transactionsFailed,
		TotalAvgResponseTime: entryTotalOutput.avgResponseTime,
		TotalRPS:             entryTotalOutput.currentRps,
		TotalFailRatio:       getTotalFailRatio(entryTotalOutput.NumRequests, entryTotalOutput.NumFailures),
		Stats:                make([]*statsEntryOutput, 0, len(stats)),
		Errors:               errors,
	}

	// convert stats
	for _, stat := range stats {
		entryOutput, err := deserializeStatsEntry(stat)
		if err != nil {
			return nil, err
		}
		output.Stats = append(output.Stats, entryOutput)
	}
	// sort stats by type
	sort.Slice(output.Stats, func(i, j int) bool {
		return output.Stats[i].Method < output.Stats[j].Method
	})
	return
}

func deserializeStatsEntry(stat interface{}) (entryOutput *statsEntryOutput, err error) {
	statBytes, err := json.Marshal(stat)
	if err != nil {
		return nil, err
	}
	entry := statsEntry{}
	if err = json.Unmarshal(statBytes, &entry); err != nil {
		return nil, err
	}

	var duration float64
	if entry.Name == "Total" {
		duration = float64(entry.LastRequestTimestamp - entry.StartTime)
		// fix: avoid divide by zero
		if duration < 1 {
			duration = 1
		}
	} else {
		duration = float64(reportStatsInterval / time.Second)
	}

	numRequests := entry.NumRequests
	entryOutput = &statsEntryOutput{
		statsEntry:         entry,
		medianResponseTime: getMedianResponseTime(numRequests, entry.ResponseTimes),
		avgResponseTime:    getAvgResponseTime(numRequests, entry.TotalResponseTime),
		avgContentLength:   getAvgContentLength(numRequests, entry.TotalContentLength),
		currentRps:         getCurrentRps(numRequests, duration),
		currentFailPerSec:  getCurrentFailPerSec(entry.NumFailures, duration),
	}
	return
}

// gauge vectors for requests
var (
	gaugeNumRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "num_requests",
			Help: "The number of requests",
		},
		[]string{"method", "name"},
	)
	gaugeNumFailures = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "num_failures",
			Help: "The number of failures",
		},
		[]string{"method", "name"},
	)
	gaugeMedianResponseTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "median_response_time",
			Help: "The median response time",
		},
		[]string{"method", "name"},
	)
	gaugeAverageResponseTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "average_response_time",
			Help: "The average response time",
		},
		[]string{"method", "name"},
	)
	gaugeMinResponseTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "min_response_time",
			Help: "The min response time",
		},
		[]string{"method", "name"},
	)
	gaugeMaxResponseTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "max_response_time",
			Help: "The max response time",
		},
		[]string{"method", "name"},
	)
	gaugeAverageContentLength = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "average_content_length",
			Help: "The average content length",
		},
		[]string{"method", "name"},
	)
	gaugeCurrentRPS = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "current_rps",
			Help: "The current requests per second",
		},
		[]string{"method", "name"},
	)
	gaugeCurrentFailPerSec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "current_fail_per_sec",
			Help: "The current failure number per second",
		},
		[]string{"method", "name"},
	)
)

// counter for total
var (
	counterErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors",
			Help: "The errors of load testing",
		},
		[]string{"method", "name", "error"},
	)
)

// summary for total
var (
	summaryResponseTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "response_time",
			Help: "The summary of response time",
			Objectives: map[float64]float64{
				0.5:  0.01,
				0.9:  0.01,
				0.95: 0.005,
			},
			AgeBuckets: 1,
			MaxAge:     100000 * time.Second,
		},
		[]string{"method", "name"},
	)
)

// gauges for total
var (
	gaugeUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "users",
			Help: "The current number of users",
		},
	)
	gaugeState = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "state",
			Help: "The current runner state, 1=initializing, 2=spawning, 3=running, 4=quitting, 5=stopped",
		},
	)
	gaugeTotalAverageResponseTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_average_response_time",
			Help: "The average response time in total milliseconds",
		},
	)
	gaugeTotalRPS = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_rps",
			Help: "The requests per second in total",
		},
	)
	gaugeTotalFailRatio = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "fail_ratio",
			Help: "The ratio of request failures in total",
		},
	)
	gaugeTransactionsPassed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "transactions_passed",
			Help: "The accumulated number of passed transactions",
		},
	)
	gaugeTransactionsFailed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "transactions_failed",
			Help: "The accumulated number of failed transactions",
		},
	)
)

// NewPrometheusPusherOutput returns a PrometheusPusherOutput.
func NewPrometheusPusherOutput(gatewayURL, jobName string) *PrometheusPusherOutput {
	nodeUUID, _ := uuid.NewUUID()
	return &PrometheusPusherOutput{
		pusher: push.New(gatewayURL, jobName).Grouping("instance", nodeUUID.String()),
	}
}

// PrometheusPusherOutput pushes boomer stats to Prometheus Pushgateway.
type PrometheusPusherOutput struct {
	pusher *push.Pusher // Prometheus Pushgateway Pusher
}

// OnStart will register all prometheus metric collectors
func (o *PrometheusPusherOutput) OnStart() {
	log.Info().Msg("register prometheus metric collectors")
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		// gauge vectors for requests
		gaugeNumRequests,
		gaugeNumFailures,
		gaugeMedianResponseTime,
		gaugeAverageResponseTime,
		gaugeMinResponseTime,
		gaugeMaxResponseTime,
		gaugeAverageContentLength,
		gaugeCurrentRPS,
		gaugeCurrentFailPerSec,
		// counter for total
		counterErrors,
		// summary for total
		summaryResponseTime,
		// gauges for total
		gaugeUsers,
		gaugeState,
		gaugeTotalAverageResponseTime,
		gaugeTotalRPS,
		gaugeTotalFailRatio,
		gaugeTransactionsPassed,
		gaugeTransactionsFailed,
	)
	o.pusher = o.pusher.Gatherer(registry)
}

// OnStop of PrometheusPusherOutput has nothing to do.
func (o *PrometheusPusherOutput) OnStop() {

}

// OnEvent will push metric to Prometheus Pushgataway
func (o *PrometheusPusherOutput) OnEvent(data map[string]interface{}) {
	output, err := convertData(data)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert data")
		return
	}

	// user count
	gaugeUsers.Set(float64(output.UserCount))

	// runner state
	gaugeState.Set(float64(output.State))

	// avg response time in total
	gaugeTotalAverageResponseTime.Set(output.TotalAvgResponseTime)

	// rps in total
	gaugeTotalRPS.Set(output.TotalRPS)

	// failure ratio in total
	gaugeTotalFailRatio.Set(output.TotalFailRatio)

	// accumulated number of transactions
	gaugeTransactionsPassed.Set(float64(output.TransactionsPassed))
	gaugeTransactionsFailed.Set(float64(output.TransactionsFailed))

	for _, stat := range output.Stats {
		method := stat.Method
		name := stat.Name
		gaugeNumRequests.WithLabelValues(method, name).Set(float64(stat.NumRequests))
		gaugeNumFailures.WithLabelValues(method, name).Set(float64(stat.NumFailures))
		gaugeMedianResponseTime.WithLabelValues(method, name).Set(float64(stat.medianResponseTime))
		gaugeAverageResponseTime.WithLabelValues(method, name).Set(float64(stat.avgResponseTime))
		gaugeMinResponseTime.WithLabelValues(method, name).Set(float64(stat.MinResponseTime))
		gaugeMaxResponseTime.WithLabelValues(method, name).Set(float64(stat.MaxResponseTime))
		gaugeAverageContentLength.WithLabelValues(method, name).Set(float64(stat.avgContentLength))
		gaugeCurrentRPS.WithLabelValues(method, name).Set(stat.currentRps)
		gaugeCurrentFailPerSec.WithLabelValues(method, name).Set(stat.currentFailPerSec)
		for responseTime, count := range stat.ResponseTimes {
			var i int64
			for i = 0; i < count; i++ {
				summaryResponseTime.WithLabelValues(method, name).Observe(float64(responseTime))
			}
		}
	}

	// errors
	for _, requestError := range output.Errors {
		counterErrors.WithLabelValues(
			requestError["method"].(string),
			requestError["name"].(string),
			requestError["error"].(string),
		).Add(float64(requestError["occurrences"].(int64)))
	}

	if err := o.pusher.Push(); err != nil {
		log.Error().Err(err).Msg("push to Pushgateway failed")
	}
}
