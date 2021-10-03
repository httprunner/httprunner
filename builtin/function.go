package builtin

import "time"

var Functions = map[string]interface{}{
	"sleep": Sleep,
}

func Sleep(nSecs int) {
	time.Sleep(time.Duration(nSecs) * time.Second)
}
