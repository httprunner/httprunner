package libimobiledevice

import (
	"bytes"
	"log"
)

type Packet interface {
	Pack() ([]byte, error)
	Unpack(buffer *bytes.Buffer) (Packet, error)
	Unmarshal(v interface{}) error

	String() string
}

var debugFlag = false

// SetDebug sets debug mode
func SetDebug(debug bool) {
	debugFlag = debug
}

func debugLog(msg string) {
	if !debugFlag {
		return
	}
	log.Printf("[%s-debug] %s\n", ProgramName, msg)
}
