package boomer

import (
	"testing"
)

func TestLogRequest(t *testing.T) {
	newStats := newRequestStats()
	newStats.logRequest("http", "success", 2, 30)
	newStats.logRequest("http", "success", 3, 40)
	newStats.logRequest("http", "success", 2, 40)
	newStats.logRequest("http", "success", 1, 20)
	entry := newStats.get("success", "http")

	if entry.NumRequests != 4 {
		t.Error("numRequests is wrong, expected: 4, got:", entry.NumRequests)
	}
	if entry.MinResponseTime != 1 {
		t.Error("minResponseTime is wrong, expected: 1, got:", entry.MinResponseTime)
	}
	if entry.MaxResponseTime != 3 {
		t.Error("maxResponseTime is wrong, expected: 3, got:", entry.MaxResponseTime)
	}
	if entry.TotalResponseTime != 8 {
		t.Error("totalResponseTime is wrong, expected: 8, got:", entry.TotalResponseTime)
	}
	if entry.TotalContentLength != 130 {
		t.Error("totalContentLength is wrong, expected: 130, got:", entry.TotalContentLength)
	}

	// check newStats.total
	if newStats.total.NumRequests != 4 {
		t.Error("newStats.total.numRequests is wrong, expected: 4, got:", newStats.total.NumRequests)
	}
	if newStats.total.MinResponseTime != 1 {
		t.Error("newStats.total.minResponseTime is wrong, expected: 1, got:", newStats.total.MinResponseTime)
	}
	if newStats.total.MaxResponseTime != 3 {
		t.Error("newStats.total.maxResponseTime is wrong, expected: 3, got:", newStats.total.MaxResponseTime)
	}
	if newStats.total.TotalResponseTime != 8 {
		t.Error("newStats.total.totalResponseTime is wrong, expected: 8, got:", newStats.total.TotalResponseTime)
	}
	if newStats.total.TotalContentLength != 130 {
		t.Error("newStats.total.totalContentLength is wrong, expected: 130, got:", newStats.total.TotalContentLength)
	}
}

func BenchmarkLogRequest(b *testing.B) {
	newStats := newRequestStats()
	for i := 0; i < b.N; i++ {
		newStats.logRequest("http", "success", 2, 30)
	}
}

func TestRoundedResponseTime(t *testing.T) {
	newStats := newRequestStats()
	newStats.logRequest("http", "success", 147, 1)
	newStats.logRequest("http", "success", 3432, 1)
	newStats.logRequest("http", "success", 58760, 1)
	entry := newStats.get("success", "http")
	responseTimes := entry.ResponseTimes

	if len(responseTimes) != 3 {
		t.Error("len(responseTimes) is wrong, expected: 3, got:", len(responseTimes))
	}

	if val, ok := responseTimes[150]; !ok || val != 1 {
		t.Error("Rounded response time should be", 150)
	}

	if val, ok := responseTimes[3400]; !ok || val != 1 {
		t.Error("Rounded response time should be", 3400)
	}

	if val, ok := responseTimes[59000]; !ok || val != 1 {
		t.Error("Rounded response time should be", 59000)
	}
}

func TestLogError(t *testing.T) {
	newStats := newRequestStats()
	newStats.logError("http", "failure", "500 error")
	newStats.logError("http", "failure", "400 error")
	newStats.logError("http", "failure", "400 error")
	entry := newStats.get("failure", "http")

	if entry.NumFailures != 3 {
		t.Error("numFailures is wrong, expected: 3, got:", entry.NumFailures)
	}

	if newStats.total.NumFailures != 3 {
		t.Error("newStats.total.numFailures is wrong, expected: 3, got:", newStats.total.NumFailures)
	}

	// md5("httpfailure500 error") = 547c38e4e4742c1c581f9e2809ba4f55
	err500 := newStats.errors["547c38e4e4742c1c581f9e2809ba4f55"]
	if err500.errMsg != "500 error" {
		t.Error("Error message is wrong, expected: 500 error, got:", err500.errMsg)
	}
	if err500.occurrences != 1 {
		t.Error("Error occurrences is wrong, expected: 1, got:", err500.occurrences)
	}

	// md5("httpfailure400 error") = f391c310401ad8e10e929f2ee1a614e4
	err400 := newStats.errors["f391c310401ad8e10e929f2ee1a614e4"]
	if err400.errMsg != "400 error" {
		t.Error("Error message is wrong, expected: 400 error, got:", err400.errMsg)
	}
	if err400.occurrences != 2 {
		t.Error("Error occurrences is wrong, expected: 2, got:", err400.occurrences)
	}
}

func BenchmarkLogError(b *testing.B) {
	newStats := newRequestStats()
	for i := 0; i < b.N; i++ {
		// LogError use md5 to calculate hash keys, it may slow down the only goroutine,
		// which consumes both requestSuccessChannel and requestFailureChannel.
		newStats.logError("http", "failure", "500 error")
	}
}

func TestClearAll(t *testing.T) {
	newStats := newRequestStats()
	newStats.logRequest("http", "success", 1, 20)
	newStats.clearAll()

	if newStats.total.NumRequests != 0 {
		t.Error("After clearAll(), newStats.total.numRequests is wrong, expected: 0, got:", newStats.total.NumRequests)
	}
}

func TestClearAllByChannel(t *testing.T) {
	newStats := newRequestStats()
	newStats.logRequest("http", "success", 1, 20)
	newStats.clearAll()

	if newStats.total.NumRequests != 0 {
		t.Error("After clearAll(), newStats.total.numRequests is wrong, expected: 0, got:", newStats.total.NumRequests)
	}
}

func TestSerializeStats(t *testing.T) {
	newStats := newRequestStats()
	newStats.logRequest("http", "success", 1, 20)

	serialized := newStats.serializeStats()
	if len(serialized) != 1 {
		t.Error("The length of serialized results is wrong, expected: 1, got:", len(serialized))
		return
	}

	first := serialized[0]
	entry, err := deserializeStatsEntry(first)
	if err != nil {
		t.Fatal()
	}

	if entry.Name != "success" {
		t.Error("The name is wrong, expected:", "success", "got:", entry.Name)
	}
	if entry.Method != "http" {
		t.Error("The method is wrong, expected:", "http", "got:", entry.Method)
	}
	if entry.NumRequests != int64(1) {
		t.Error("The num_requests is wrong, expected:", 1, "got:", entry.NumRequests)
	}
	if entry.NumFailures != int64(0) {
		t.Error("The num_failures is wrong, expected:", 0, "got:", entry.NumFailures)
	}
}

func TestSerializeErrors(t *testing.T) {
	newStats := newRequestStats()
	newStats.logError("http", "failure", "500 error")
	newStats.logError("http", "failure", "400 error")
	newStats.logError("http", "failure", "400 error")
	serialized := newStats.serializeErrors()

	if len(serialized) != 2 {
		t.Error("The length of serialized results is wrong, expected: 2, got:", len(serialized))
		return
	}

	for key, value := range serialized {
		if key == "f391c310401ad8e10e929f2ee1a614e4" {
			err := value["error"].(string)
			if err != "400 error" {
				t.Error("expected: 400 error, got:", err)
			}
			occurrences := value["occurrences"].(int64)
			if occurrences != int64(2) {
				t.Error("expected: 2, got:", occurrences)
			}
		}
	}
}

func TestCollectReportData(t *testing.T) {
	newStats := newRequestStats()
	newStats.logRequest("http", "success", 2, 30)
	newStats.logError("http", "failure", "500 error")
	result := newStats.collectReportData()

	if _, ok := result["stats"]; !ok {
		t.Error("Key stats not found")
	}
	if _, ok := result["stats_total"]; !ok {
		t.Error("Key stats not found")
	}
	if _, ok := result["errors"]; !ok {
		t.Error("Key stats not found")
	}
}
