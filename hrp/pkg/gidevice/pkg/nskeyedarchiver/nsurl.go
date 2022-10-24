package nskeyedarchiver

import (
	"fmt"

	"howett.net/plist"
)

type NSURL struct {
	internal string
}

func NewNSURL(path string) *NSURL {
	return &NSURL{
		internal: path,
	}
}

func (ns *NSURL) archive(objects []interface{}) []interface{} {
	info := map[string]interface{}{}

	objects = append(objects, info)

	uid := plist.UID(0)
	info["NS.base"] = uid
	objects, uid = archive(objects, fmt.Sprintf("file://%s", ns.internal))
	info["NS.relative"] = uid

	info["$class"] = plist.UID(len(objects))

	cls := map[string]interface{}{
		"$classname": "NSURL",
		"$classes":   []interface{}{"NSURL", "NSObject"},
	}
	objects = append(objects, cls)

	return objects
}
