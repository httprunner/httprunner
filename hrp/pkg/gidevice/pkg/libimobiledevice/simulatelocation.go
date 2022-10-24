package libimobiledevice

import (
	"fmt"
	"math"
	"strings"
)

const SimulateLocationServiceName = "com.apple.dt.simulatelocation"

type CoordinateSystem string

const (
	CoordinateSystemWGS84 CoordinateSystem = "WGS84"
	CoordinateSystemBD09  CoordinateSystem = "BD09"
	CoordinateSystemGCJ02 CoordinateSystem = "GCJ02"
)

func NewSimulateLocationClient(innerConn InnerConn) *SimulateLocationClient {
	return &SimulateLocationClient{
		client: newServicePacketClient(innerConn),
	}
}

type SimulateLocationClient struct {
	client *servicePacketClient
}

func (c *SimulateLocationClient) NewLocationPacket(lon, lat float64, coordinateSystem CoordinateSystem) Packet {
	switch CoordinateSystem(strings.ToUpper(string(coordinateSystem))) {
	case CoordinateSystemGCJ02:
		lon, lat = gcj02ToWGS84(lon, lat)
	case CoordinateSystemBD09:
		lon, lat = bd09ToWGS84(lon, lat)
	case CoordinateSystemWGS84:
		_, _ = lon, lat
	default:
		_, _ = lon, lat
	}

	pkt := new(locationPacket)
	pkt.lon = lon
	pkt.lat = lat
	return pkt
}

func (c *SimulateLocationClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

// Recover try to revert back
func (c *SimulateLocationClient) Recover() error {
	data := []byte{0x00, 0x00, 0x00, 0x01}
	debugLog(fmt.Sprintf("--> %+v\n", data))
	return c.client.innerConn.Write(data)
}

const (
	xPi    = math.Pi * 3000.0 / 180.0
	offset = 0.00669342162296594323
	axis   = 6378245.0
)

func isOutOfChina(lon, lat float64) bool {
	return !(lon > 73.66 && lon < 135.05 && lat > 3.86 && lat < 53.55)
}

func delta(lon, lat float64) (float64, float64) {
	dLat := transformLat(lon-105.0, lat-35.0)
	dLon := transformLng(lon-105.0, lat-35.0)

	radLat := lat / 180.0 * math.Pi
	magic := math.Sin(radLat)
	magic = 1 - offset*magic*magic
	sqrtMagic := math.Sqrt(magic)

	dLat = (dLat * 180.0) / ((axis * (1 - offset)) / (magic * sqrtMagic) * math.Pi)
	dLon = (dLon * 180.0) / (axis / sqrtMagic * math.Cos(radLat) * math.Pi)

	mgLat := lat + dLat
	mgLon := lon + dLon

	return mgLon, mgLat
}

func transformLat(lon, lat float64) float64 {
	ret := -100.0 + 2.0*lon + 3.0*lat + 0.2*lat*lat + 0.1*lon*lat + 0.2*math.Sqrt(math.Abs(lon))
	ret += (20.0*math.Sin(6.0*lon*math.Pi) + 20.0*math.Sin(2.0*lon*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lat*math.Pi) + 40.0*math.Sin(lat/3.0*math.Pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(lat/12.0*math.Pi) + 320*math.Sin(lat*math.Pi/30.0)) * 2.0 / 3.0
	return ret
}

func transformLng(lon, lat float64) float64 {
	ret := 300.0 + lon + 2.0*lat + 0.1*lon*lon + 0.1*lon*lat + 0.1*math.Sqrt(math.Abs(lon))
	ret += (20.0*math.Sin(6.0*lon*math.Pi) + 20.0*math.Sin(2.0*lon*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lon*math.Pi) + 40.0*math.Sin(lon/3.0*math.Pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(lon/12.0*math.Pi) + 300.0*math.Sin(lon/30.0*math.Pi)) * 2.0 / 3.0
	return ret
}

func gcj02ToWGS84(lon, lat float64) (float64, float64) {
	if isOutOfChina(lon, lat) {
		return lon, lat
	}

	mgLon, mgLat := delta(lon, lat)

	return lon*2 - mgLon, lat*2 - mgLat
}

func bd09ToGCJ02(lon, lat float64) (float64, float64) {
	x := lon - 0.0065
	y := lat - 0.006

	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*xPi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*xPi)

	gLon := z * math.Cos(theta)
	gLat := z * math.Sin(theta)

	return gLon, gLat
}

func bd09ToWGS84(lon, lat float64) (float64, float64) {
	lon, lat = bd09ToGCJ02(lon, lat)
	return gcj02ToWGS84(lon, lat)
}
