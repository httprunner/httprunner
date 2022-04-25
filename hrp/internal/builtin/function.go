package builtin

import (
	"crypto/md5"
	"encoding/hex"
	"math"
	"math/rand"
	"os"
	"time"
)

var Functions = map[string]interface{}{
	"get_timestamp":     getTimestamp,    // call without arguments
	"sleep":             sleep,           // call with one argument
	"gen_random_string": genRandomString, // call with one argument
	"max":               math.Max,        // call with two arguments
	"md5":               MD5,             // call with one argument
	"parameterize":      loadFromCSV,
	"P":                 loadFromCSV,
	"environ":           os.Getenv,
	"ENV":               os.Getenv,
	"load_ws_message":   loadMessage,
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func sleep(nSecs int) {
	time.Sleep(time.Duration(nSecs) * time.Second)
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func genRandomString(n int) string {
	lettersLen := len(letters)
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(lettersLen)]
	}
	return string(b)
}

func MD5(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}
