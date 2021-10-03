package builtin

import "time"

var FunctionsMap = map[string]interface{}{
	"sleep": Sleep,
}

func Sleep(nSecs int) {
	time.Sleep(time.Duration(nSecs) * time.Second)
}
