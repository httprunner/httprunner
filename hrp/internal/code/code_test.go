package code

import (
	"fmt"
	"testing"
)

func TestGetErrorCode(t *testing.T) {
	err := LoadYAMLError
	code := GetErrorCode(err)
	fmt.Println(code)
}
