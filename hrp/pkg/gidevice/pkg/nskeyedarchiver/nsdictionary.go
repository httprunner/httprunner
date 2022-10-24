package nskeyedarchiver

import (
	"howett.net/plist"
)

type NSDictionary struct {
	internal map[string]interface{}
}

func NewNSDictionary(value map[string]interface{}) *NSDictionary {
	return &NSDictionary{
		internal: value,
	}
}

func (ns *NSDictionary) archive(objects []interface{}) []interface{} {
	keys := make([]interface{}, 0, len(ns.internal))
	objs := make([]interface{}, 0, len(ns.internal))

	info := map[string]interface{}{}
	objects = append(objects, info)

	for k, v := range ns.internal {
		uid := plist.UID(len(objects))
		keys = append(keys, uid)
		objects = append(objects, k)

		objects, uid = archive(objects, v)
		objs = append(objs, uid)
	}

	info["NS.keys"] = keys
	info["NS.objects"] = objs
	info["$class"] = plist.UID(len(objects))

	cls := map[string]interface{}{
		"$classname": "NSDictionary",
		"$classes":   []interface{}{"NSDictionary", "NSObject"},
	}
	objects = append(objects, cls)

	return objects
}
