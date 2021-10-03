package builtin

import (
	"math"
	"math/rand"
	"time"
)

var Functions = map[string]interface{}{
	"sleep":             sleep,           // call with one argument
	"gen_random_string": genRandomString, // call with one argument
	"max":               math.Max,        // call with two arguments
}

func sleep(nSecs int) {
	time.Sleep(time.Duration(nSecs) * time.Second)
}

func genRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	const lettersLen = len(letters)

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(lettersLen)]
	}
	return string(b)
}
