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
	startTime1 := time.Now()
	params := []interface{}{1}
	err := sleepRandom(startTime1, params)
	checkErr(t, err)
	dur := time.Since(startTime1).Seconds()
	t.Log(dur)
	if dur < 1 || dur > 1.1 {
		t.Fatal("sleepRandom failed")
	}

	params = []interface{}{0, 2}
	err = sleepRandom(startTime1, params)
	checkErr(t, err)
	dur = time.Since(startTime1).Seconds()
	t.Log(dur)
	if dur < 1 || dur > 2 {
		t.Fatal("sleepRandom failed")
	}

	startTime2 := time.Now()
	params = []interface{}{1, 2}
	err = sleepRandom(startTime2, params)
	checkErr(t, err)
	dur = time.Since(startTime2).Seconds()
	t.Log(dur)
	if dur < 1 || dur > 2 {
		t.Fatal("sleepRandom failed")
	}
}
