//go:build localtest

package nskeyedarchiver

import (
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestXCTestConfiguration_archive(t *testing.T) {
	objs := make([]interface{}, 0, 1)
	xcTestConfiguration := NewXCTestConfiguration(NewNSUUID(uuid.NewV4().Bytes()), NewNSURL("/tmp"), "", "")
	objects := xcTestConfiguration.archive(objs)
	fmt.Println(objects)
}
