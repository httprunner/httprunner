package nskeyedarchiver

import "howett.net/plist"

type NSArray struct {
	internal []interface{}
}

func NewNSArray(value []interface{}) *NSArray {
	return &NSArray{
		internal: value,
	}
}

func (ns *NSArray) archive(objects []interface{}) []interface{} {
	objs := make([]interface{}, 0, len(ns.internal))

	info := map[string]interface{}{}
	objects = append(objects, info)

	for _, v := range ns.internal {
		var uid plist.UID
		objects, uid = archive(objects, v)
		objs = append(objs, uid)
	}

	info["NS.objects"] = objs
	info["$class"] = plist.UID(len(objects))

	cls := map[string]interface{}{
		"$classname": "NSArray",
		"$classes":   []interface{}{"NSArray", "NSObject"},
	}
	objects = append(objects, cls)

	return objects
}
