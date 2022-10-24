package gadb

import "log"

var debugFlag = false

// SetDebug set debug mode
func SetDebug(debug bool) {
	debugFlag = debug
}

func debugLog(msg string) {
	if !debugFlag {
		return
	}
	log.Println("[DEBUG] [gadb] " + msg)
}
