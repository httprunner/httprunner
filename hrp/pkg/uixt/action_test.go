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

func TestSleepRandom(t *testing.T) {
	startTime := time.Now()
	params := []interface{}{1}
	err := sleepRandom(params)
	checkErr(t, err)
	dur := time.Since(startTime).Seconds()
	if dur < 0.9 || dur > 1.1 {
		t.Fatal("sleepRandom failed")
	}

	startTime = time.Now()
	params = []interface{}{1, 2}
	err = sleepRandom(params)
	checkErr(t, err)
	dur = time.Since(startTime).Seconds()
	if dur < 1 || dur > 2 {
		t.Fatal("sleepRandom failed")
	}
}
