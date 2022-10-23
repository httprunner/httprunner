package nskeyedarchiver

import (
	"reflect"

	"howett.net/plist"
)

func Marshal(obj interface{}) (raw []byte, err error) {
	objects := []interface{}{"$null"}
	objects, _ = archive(objects, obj)
	archiver := map[string]interface{}{
		"$version":  100000,
		"$archiver": "NSKeyedArchiver",
		"$top":      map[string]interface{}{"root": plist.UID(1)},
		"$objects":  objects,
	}
	// if len(format) == 0 {
	// 	format = []int{plist.BinaryFormat}
	// }
	// return plist.Marshal(archiver, format[0])
	return plist.Marshal(archiver, plist.BinaryFormat)
}

func archive(_objects []interface{}, _value interface{}) (objects []interface{}, uid plist.UID) {
	val := reflect.ValueOf(_value)
	typ := val.Type()

	switch typ.Kind() {
	case reflect.String,
		reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr:
		uid = plist.UID(len(_objects))
		objects = append(_objects, _value)
		return
	case reflect.Map:
		uid = plist.UID(len(_objects))
		vv := make(map[string]interface{})
		keys := val.MapKeys()
		for _, k := range keys {
			vv[k.String()] = val.MapIndex(k).Interface()
		}
		objects = NewNSDictionary(vv).archive(_objects)
		return
	case reflect.Slice, reflect.Array:
		uid = plist.UID(len(_objects))
		vv := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			vv[i] = val.Index(i).Interface()
		}
		objects = NewNSArray(vv).archive(_objects)
		return
	case reflect.Struct, reflect.Ptr:
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			val = val.Elem()
		}

		switch typ.Name() {
		case "NSUUID":
			uid = plist.UID(len(_objects))
			objects = NewNSUUID(val.Field(0).Bytes()).archive(_objects)
			return
		case "NSURL":
			uid = plist.UID(len(_objects))
			objects = NewNSURL(val.Field(0).String()).archive(_objects)
			return
		case "XCTestConfiguration":
			uid = plist.UID(len(_objects))
			objects = newXCTestConfiguration(_value).archive(_objects)
			return
		}
	}

	return
}

// TODO unarchive
