package gidevice

import "github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"

var _ SimulateLocation = (*simulateLocation)(nil)

func newSimulateLocation(client *libimobiledevice.SimulateLocationClient) *simulateLocation {
	return &simulateLocation{
		client: client,
	}
}

type simulateLocation struct {
	client *libimobiledevice.SimulateLocationClient
}

func (s *simulateLocation) Update(longitude float64, latitude float64, coordinateSystem ...CoordinateSystem) (err error) {
	if len(coordinateSystem) == 0 {
		coordinateSystem = []CoordinateSystem{CoordinateSystemWGS84}
	}
	pkt := s.client.NewLocationPacket(longitude, latitude, coordinateSystem[0])
	return s.client.SendPacket(pkt)
}

func (s *simulateLocation) Recover() (err error) {
	return s.client.Recover()
}
