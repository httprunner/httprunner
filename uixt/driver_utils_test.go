package uixt

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/stretchr/testify/assert"
)

func TestGetSimulationDuration(t *testing.T) {
	params := []float64{1.23}
	duration := getSimulationDuration(params)
	if duration != 1230 {
		t.Fatal("getSimulationDuration failed")
	}

	params = []float64{1, 2}
	duration = getSimulationDuration(params)
	if duration < 1000 || duration > 2000 {
		t.Fatal("getSimulationDuration failed")
	}

	params = []float64{1, 5, 0.7, 5, 10, 0.3}
	duration = getSimulationDuration(params)
	if duration < 1000 || duration > 10000 {
		t.Fatal("getSimulationDuration failed")
	}
}

func TestSleepStrict(t *testing.T) {
	ctx := context.Background()
	startTime := time.Now()
	sleepStrict(ctx, startTime, 1230)
	dur := time.Since(startTime).Milliseconds()
	t.Log(dur)
	if dur < 1230 || dur > 1300 {
		t.Fatalf("sleepRandom failed, dur: %d", dur)
	}
}

func TestUtils_GetFreePort(t *testing.T) {
	freePort, err := builtin.GetFreePort()
	assert.Nil(t, err)
	assert.Greater(t, freePort, 10000)
	t.Log(freePort)
}

func TestUtils_ConvertPoints(t *testing.T) {
	data := "10-09 20:16:48.216 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317808206,\"ext\":\"输入\",\"from\":{\"x\":0.0,\"y\":0.0},\"operation\":\"Gtf-SendKeys\",\"run_time\":627,\"start\":1665317807579,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":0.0,\"y\":0.0}}\n10-09 20:18:22.899 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317902898,\"ext\":\"进入直播间\",\"from\":{\"x\":717.0,\"y\":2117.5},\"operation\":\"Gtf-Tap\",\"run_time\":121,\"start\":1665317902777,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":717.0,\"y\":2117.5}}\n10-09 20:18:32.063 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317912062,\"ext\":\"第一次上划\",\"from\":{\"x\":1437.0,\"y\":2409.9},\"operation\":\"Gtf-Swipe\",\"run_time\":32,\"start\":1665317912030,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":1437.0,\"y\":2409.9}}"
	eps := ConvertPoints(strings.Split(data, "\n"))
	assert.Equal(t, 3, len(eps))
}
