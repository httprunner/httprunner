//go:build localtest

package nskeyedarchiver

import (
	"fmt"
	"testing"
)

func TestNSDictionary_archive(t *testing.T) {
	objs := make([]interface{}, 0, 1)
	value := map[string]interface{}{
		"a": 1,
		"b": "2",
		"c": true,
	}
	dict := NewNSDictionary(value)
	objects := dict.archive(objs)
	fmt.Println(objects)
}
