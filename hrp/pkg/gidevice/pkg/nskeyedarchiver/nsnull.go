package nskeyedarchiver

import (
	"howett.net/plist"
)

type NSNull struct{}

func NewNSNull() *NSNull {
	return &NSNull{}
}

func (ns *NSNull) archive(objects []interface{}) []interface{} {
	info := map[string]interface{}{}

	objects = append(objects, info)

	info["$class"] = plist.UID(len(objects))

	cls := map[string]interface{}{
		"$classname": "NSNull",
		"$classes":   []interface{}{"NSNull", "NSObject"},
	}
	objects = append(objects, cls)

	return objects
}
