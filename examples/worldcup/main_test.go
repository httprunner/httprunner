//go:build localtest

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertTimeToSeconds(t *testing.T) {
	testData := []struct {
		timeStr string
		seconds int
	}{
		{"00:00", 0},
		{"00:01", 1},
		{"01:00", 60},
		{"01:01", 61},
		{"00:01:02", 62},
		{"01:02:03", 3723},
	}

	for _, td := range testData {
		seconds, err := convertTimeToSeconds(td.timeStr)
		assert.Nil(t, err)
		assert.Equal(t, td.seconds, seconds)
	}
}

func TestMainIOS(t *testing.T) {
	device := initIOSDevice(uuid)
	bundleID := "com.ss.iphone.ugc.Aweme"
	wc := NewWorldCupLive(device, "", bundleID, 30, 10)
	wc.EnterLive(bundleID)
	wc.Start()
	wc.DumpResult()
}

func TestMainAndroid(t *testing.T) {
	device := initAndroidDevice(uuid)
	bundleID := "com.ss.android.ugc.aweme"
	wc := NewWorldCupLive(device, "", bundleID, 30, 10)
	wc.EnterLive(bundleID)
	wc.Start()
	wc.DumpResult()
}
