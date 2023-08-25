package uixt

import (
	"testing"
	"time"
)

func checkErr(t *testing.T, err error, msg ...string) {
	if err != nil {
		if len(msg) == 0 {
			t.Fatal(err)
		} else {
			t.Fatal(msg, err)
		}
	}
}

func TestGetSimulationDuration(t *testing.T) {
	params := []interface{}{1.23}
	duration := getSimulationDuration(params)
	if duration != 1230 {
		t.Fatal("getSimulationDuration failed")
	}

	params = []interface{}{1, 2}
	duration = getSimulationDuration(params)
	if duration < 1000 || duration > 2000 {
		t.Fatal("getSimulationDuration failed")
	}

	params = []interface{}{1, 5, 0.7, 5, 10, 0.3}
	duration = getSimulationDuration(params)
	if duration < 1000 || duration > 10000 {
		t.Fatal("getSimulationDuration failed")
	}
}

func TestSleepStrict(t *testing.T) {
	startTime := time.Now()
	sleepStrict(startTime, 1230)
	dur := time.Since(startTime).Milliseconds()
	t.Log(dur)
	if dur < 1230 || dur > 1300 {
		t.Fatalf("sleepRandom failed, dur: %d", dur)
	}
}
