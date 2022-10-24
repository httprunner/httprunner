//go:build localtest

package gidevice

import (
	"testing"
)

var houseArrestSrv HouseArrest

func setupHouseArrestSrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	if houseArrestSrv, err = lockdownSrv.HouseArrestService(); err != nil {
		t.Fatal(err)
	}
}

func Test_houseArrest_Documents(t *testing.T) {
	setupHouseArrestSrv(t)

	bundleID = "com.apple.iMovie"
	appAfc, err := houseArrestSrv.Documents(bundleID)
	if err != nil {
		t.Fatal(err)
	}

	names, err := appAfc.ReadDir("Documents")
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range names {
		t.Log(name)
	}
}

func Test_houseArrest_Container(t *testing.T) {
	setupHouseArrestSrv(t)

	bundleID = "com.apple.iMovie"
	appAfc, err := houseArrestSrv.Documents(bundleID)
	if err != nil {
		t.Fatal(err)
	}

	names, err := appAfc.ReadDir("Documents")
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range names {
		t.Log(name)
	}
}
