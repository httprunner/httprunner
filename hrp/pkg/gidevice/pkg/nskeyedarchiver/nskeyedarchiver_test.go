//go:build localtest

package nskeyedarchiver

import (
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestMarshal(t *testing.T) {
	// value := map[string]interface{}{
	// 	"a": 1,
	// 	"b": "2",
	// 	"c": true,
	// }

	// value := []interface{}{
	// 	"a", 1,
	// 	"b", "2",
	// 	"c", false,
	// }

	// value := NewNSUUID(uuid.NewV4().Bytes())

	// value := NewNSURL("/tmp")

	value := NewXCTestConfiguration(NewNSUUID(uuid.NewV4().Bytes()), NewNSURL("/tmp"), "", "")

	raw, err := Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range raw {
		fmt.Printf("%x", v)
	}
	fmt.Println()
	// fmt.Println(raw)
}
