package boomer

import (
	"math"
	"testing"
)

func TestGetMedianResponseTime(t *testing.T) {
	numRequests := int64(10)
	responseTimes := map[int64]int64{
		100: 1,
		200: 3,
		300: 6,
	}

	medianResponseTime := getMedianResponseTime(numRequests, responseTimes)
	if medianResponseTime != 300 {
		t.Error("medianResponseTime should be 300")
	}

	responseTimes = map[int64]int64{}

	medianResponseTime = getMedianResponseTime(numRequests, responseTimes)
	if medianResponseTime != 0 {
		t.Error("medianResponseTime should be 0")
	}
}

func TestGetAvgResponseTime(t *testing.T) {
	numRequests := int64(3)
	totalResponseTime := int64(100)

	avgResponseTime := getAvgResponseTime(numRequests, totalResponseTime)
	if math.Dim(float64(33.33), avgResponseTime) > 0.01 {
		t.Error("avgResponseTime should be close to 33.33")
	}

	avgResponseTime = getAvgResponseTime(int64(0), totalResponseTime)
	if avgResponseTime != float64(0) {
		t.Error("avgResponseTime should be close to 0")
	}
}

func TestGetAvgContentLength(t *testing.T) {
	numRequests := int64(3)
	totalContentLength := int64(100)

	avgContentLength := getAvgContentLength(numRequests, totalContentLength)
	if avgContentLength != 33 {
		t.Error("avgContentLength should be 33")
	}

	avgContentLength = getAvgContentLength(int64(0), totalContentLength)
	if avgContentLength != 0 {
		t.Error("avgContentLength should be 0")
	}
}

func TestGetCurrentRps(t *testing.T) {
	numRequests := int64(10)
	numReqsPerSecond := map[int64]int64{}

	currentRps := getCurrentRps(numRequests, numReqsPerSecond)
	if currentRps != 0 {
		t.Error("currentRps should be 0")
	}

	numReqsPerSecond[1] = 2
	numReqsPerSecond[2] = 3
	numReqsPerSecond[3] = 2
	numReqsPerSecond[4] = 3

	currentRps = getCurrentRps(numRequests, numReqsPerSecond)
	if currentRps != 2 {
		t.Error("currentRps should be 2")
	}
}

func TestConsoleOutput(t *testing.T) {
	o := NewConsoleOutput()
	o.OnStart()

	data := map[string]interface{}{}
	stat := map[string]interface{}{}
	data["stats"] = []interface{}{stat}

	stat["name"] = "http"
	stat["method"] = "post"
	stat["num_requests"] = int64(100)
	stat["num_failures"] = int64(10)
	stat["response_times"] = map[int64]int64{
		10:  1,
		100: 99,
	}
	stat["total_response_time"] = int64(9910)
	stat["min_response_time"] = int64(10)
	stat["max_response_time"] = int64(100)
	stat["total_content_length"] = int64(100000)
	stat["num_reqs_per_sec"] = map[int64]int64{
		1: 20,
		2: 40,
		3: 40,
	}

	o.OnEvent(data)

	o.OnStop()
}
