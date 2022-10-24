//go:build localtest

package nskeyedarchiver

import (
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestNSUUID_archive(t *testing.T) {
	objs := make([]interface{}, 0, 1)
	nsuuid := NewNSUUID(uuid.NewV4().Bytes())
	objects := nsuuid.archive(objs)
	fmt.Println(objects)
}
