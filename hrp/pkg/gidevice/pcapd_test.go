//go:build localtest

package gidevice

import (
	"fmt"
	"testing"
	"time"
)

func TestPcapWithPID(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PcapStart(WithPcapPID(1234))
	if err != nil {
		t.Fatal(err)
	}
	timer := time.NewTimer(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			dev.PcapStop()
			return
		case d := <-data:
			fmt.Println(string(d))
		}
	}
}

func TestPcapWithProcName(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PcapStart(WithPcapProcName("Awe"))
	if err != nil {
		t.Fatal(err)
	}
	timer := time.NewTimer(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			dev.PcapStop()
			return
		case d := <-data:
			fmt.Println(string(d))
		}
	}
}

func TestPcapWithBundleID(t *testing.T) {
	setupLockdownSrv(t)

	data, err := dev.PcapStart(WithPcapBundleID("com.ss.iphone.ugc.Aweme"))
	if err != nil {
		t.Fatal(err)
	}
	timer := time.NewTimer(time.Duration(time.Second * 10))
	for {
		select {
		case <-timer.C:
			dev.PcapStop()
			return
		case d := <-data:
			fmt.Println(string(d))
		}
	}
}
