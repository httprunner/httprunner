package gadb

import (
	"net"
)

func DisableTimeWait(conn *net.TCPConn) error {
	return conn.SetLinger(0)
}
