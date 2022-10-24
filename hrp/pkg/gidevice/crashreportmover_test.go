//go:build localtest

package gidevice

import (
	"fmt"
	"os"
	"testing"
)

var crashReportMoverSrv CrashReportMover

func setupCrashReportMoverSrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	if crashReportMoverSrv, err = lockdownSrv.CrashReportMoverService(); err != nil {
		t.Fatal(err)
	}
}

func Test_crashReportMover_Move(t *testing.T) {
	setupCrashReportMoverSrv(t)

	SetDebug(true)
	userHomeDir, _ := os.UserHomeDir()
	// err := crashReportMoverSrv.Move(userHomeDir + "/Documents/temp/2021-04/out_gidevice")
	// err := crashReportMoverSrv.Move(userHomeDir+"/Documents/temp/2021-04/out_gidevice",
	err := crashReportMoverSrv.Move(userHomeDir+"/Documents/temp/2021-04/out_gidevice_extract",
		WithKeepCrashReport(true),
		WithExtractRawCrashReport(true),
		WithWhenMoveIsDone(func(filename string) {
			fmt.Println("Copy:", filename)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
}
