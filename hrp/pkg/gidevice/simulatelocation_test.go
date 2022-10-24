//go:build localtest

package gidevice

import "testing"

var simulateLocationSrv SimulateLocation

func setupSimulateLocationSrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	if simulateLocationSrv, err = lockdownSrv.SimulateLocationService(); err != nil {
		t.Fatal(err)
	}
}

func Test_simulateLocation_Update(t *testing.T) {
	setupSimulateLocationSrv(t)

	// https://api.map.baidu.com/lbsapi/getpoint/index.html
	// if err := dev.SimulateLocationUpdate(116.024067, 40.362639, CoordinateSystemBD09); err != nil {
	if err := simulateLocationSrv.Update(116.024067, 40.362639, CoordinateSystemBD09); err != nil {
		t.Fatal(err)
	}

	// https://developer.amap.com/tools/picker
	// https://lbs.qq.com/tool/getpoint/index.html
	// if err := simulateLocationSrv.Update(120.116979,30.252876, CoordinateSystemGCJ02); err != nil {
	// 	t.Fatal(err)
	// }

	if err := simulateLocationSrv.Update(121.499763, 31.239580); err != nil {
		// if err := simulateLocationSrv.Update(121.499763, 31.239580, CoordinateSystemWGS84); err != nil {
		t.Fatal(err)
	}
}

func Test_simulateLocation_Recover(t *testing.T) {
	setupSimulateLocationSrv(t)

	// if err := dev.SimulateLocationRecover(); err != nil {
	if err := simulateLocationSrv.Recover(); err != nil {
		t.Fatal(err)
	}
}
