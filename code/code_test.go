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

func TestGetErrorByCode(t *testing.T) {
	code := 0
	err := GetErrorByCode(code)
	fmt.Println("[TestGetErrorByCode]:err:", err)
}
