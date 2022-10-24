//go:build localtest

package nskeyedarchiver

import (
	"fmt"
	"testing"
)

func TestNSArray_archive(t *testing.T) {
	objs := make([]interface{}, 0, 1)
	value := []interface{}{
		"a", 1,
		"b", "2",
		"c", false,
	}
	array := NewNSArray(value)
	objects := array.archive(objs)
	fmt.Println(objects)
}
