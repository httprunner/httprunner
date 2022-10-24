//go:build localtest

package nskeyedarchiver

import (
	"fmt"
	"testing"
)

func TestNSURL_archive(t *testing.T) {
	objs := make([]interface{}, 0, 1)
	nsurl := NewNSURL("/tmp")
	objects := nsurl.archive(objs)
	fmt.Println(objects)
}
