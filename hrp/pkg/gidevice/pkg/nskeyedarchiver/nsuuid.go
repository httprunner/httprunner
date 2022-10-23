package nskeyedarchiver

import (
	"encoding/hex"

	"howett.net/plist"
)

type NSUUID struct {
	internal []byte
}

func NewNSUUID(uuid []byte) *NSUUID {
	return &NSUUID{
		internal: uuid,
	}
}

func (ns *NSUUID) archive(objects []interface{}) []interface{} {
	info := map[string]interface{}{
		"NS.uuidbytes": ns.internal,
	}

	objects = append(objects, info)

	info["$class"] = plist.UID(len(objects))

	cls := map[string]interface{}{
		"$classname": "NSUUID",
		"$classes":   []interface{}{"NSUUID", "NSObject"},
	}
	objects = append(objects, cls)

	return objects
}

func (ns *NSUUID) String() string {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], ns.internal[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], ns.internal[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], ns.internal[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], ns.internal[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], ns.internal[10:])

	return string(buf)
}
