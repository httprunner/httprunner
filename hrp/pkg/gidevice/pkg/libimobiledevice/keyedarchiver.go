package libimobiledevice

import (
	"reflect"
	"strconv"
	"time"

	"howett.net/plist"
)

const nsNull = "$null"

func newKeyedArchiver() *KeyedArchiver {
	return &KeyedArchiver{
		Archiver: "NSKeyedArchiver",
		Version:  100000,
	}
}

type KeyedArchiver struct {
	Archiver string        `plist:"$archiver"`
	Objects  []interface{} `plist:"$objects"`
	Top      ArchiverRoot  `plist:"$top"`
	Version  int           `plist:"$version"`
}

func (ka *KeyedArchiver) UID() plist.UID {
	return plist.UID(len(ka.Objects))
}

type ArchiverRoot struct {
	Root plist.UID `plist:"root"`
}

type ArchiverClasses struct {
	Classes   []string `plist:"$classes"`
	ClassName string   `plist:"$classname"`
}

var (
	NSMutableDictionaryClass = &ArchiverClasses{
		Classes:   []string{"NSMutableDictionary", "NSDictionary", "NSObject"},
		ClassName: "NSMutableDictionary",
	}
	NSDictionaryClass = &ArchiverClasses{
		Classes:   []string{"NSDictionary", "NSObject"},
		ClassName: "NSDictionary",
	}
	NSMutableArrayClass = &ArchiverClasses{
		Classes:   []string{"NSMutableArray", "NSArray", "NSObject"},
		ClassName: "NSMutableArray",
	}
	NSArrayClass = &ArchiverClasses{
		Classes:   []string{"NSArray", "NSObject"},
		ClassName: "NSArray",
	}
	NSMutableDataClass = &ArchiverClasses{
		Classes:   []string{"NSMutableArray", "NSArray", "NSObject"},
		ClassName: "NSMutableArray",
	}
	NSDataClass = &ArchiverClasses{
		Classes:   []string{"NSData", "NSObject"},
		ClassName: "NSData",
	}
	NSDateClass = &ArchiverClasses{
		Classes:   []string{"NSDate", "NSObject"},
		ClassName: "NSDate",
	}
	NSErrorClass = &ArchiverClasses{
		Classes:   []string{"NSError", "NSObject"},
		ClassName: "NSError",
	}
)

type NSObject struct {
	Class plist.UID `plist:"$class"`
}

type NSArray struct {
	NSObject
	Values []plist.UID `plist:"NS.objects"`
}

type NSDictionary struct {
	NSArray
	Keys []plist.UID `plist:"NS.keys"`
}

type NSData struct {
	NSObject
	Data []byte `plist:"NS.data"`
}

type NSError struct {
	NSCode     int
	NSDomain   string
	NSUserInfo interface{}
}

type NSKeyedArchiver struct {
	objRefVal []interface{}
	objRef    map[interface{}]plist.UID
}

func NewNSKeyedArchiver() *NSKeyedArchiver {
	return &NSKeyedArchiver{
		objRef: make(map[interface{}]plist.UID),
	}
}

func (ka *NSKeyedArchiver) id(v interface{}) plist.UID {
	var ref plist.UID
	if id, ok := ka.objRef[v]; !ok {
		ref = plist.UID(len(ka.objRef))
		ka.objRefVal = append(ka.objRefVal, v)
		ka.objRef[v] = ref
	} else {
		ref = id
	}
	return ref
}

func (ka *NSKeyedArchiver) flushToStruct(root *KeyedArchiver) {
	for i := 0; i < len(ka.objRefVal); i++ {
		val := ka.objRefVal[i]
		vt := reflect.ValueOf(val)
		if vt.Kind() == reflect.Ptr {
			val = vt.Elem().Interface()
		}
		root.Objects = append(root.Objects, val)
	}
}

func (ka *NSKeyedArchiver) clear() {
	ka.objRef = make(map[interface{}]plist.UID)
	ka.objRefVal = []interface{}{}
}

type XCTestConfiguration struct {
	Contents map[string]interface{}
}

func (ka *NSKeyedArchiver) Marshal(obj interface{}) ([]byte, error) {
	val := reflect.ValueOf(obj)
	typ := val.Type()

	root := newKeyedArchiver()

	var tmpTop plist.UID

	ka.id(nsNull)

	switch typ.Kind() {
	case reflect.Map:
		m := &NSDictionary{}
		m.Class = ka.id(NSDictionaryClass)
		keys := val.MapKeys()
		for _, v := range keys {
			m.Keys = append(m.Keys, ka.id(v.Interface()))
			m.Values = append(m.Values, ka.id(val.MapIndex(v).Interface()))
		}
		tmpTop = ka.id(m)
	case reflect.Slice, reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			d := &NSData{}
			d.Class = ka.id(NSDataClass)
			var w []byte
			for i := 0; i < val.Len(); i++ {
				w = append(w, uint8(val.Index(i).Uint()))
			}
			d.Data = w
		}
		a := &NSArray{}
		a.Class = ka.id(NSArrayClass)
		for i := 0; i < val.Len(); i++ {
			a.Values = append(a.Values, ka.id(val.Index(i).Interface()))
		}
		tmpTop = ka.id(a)
	case reflect.String:
		tmpTop = ka.id(obj)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		tmpTop = ka.id(obj)
	}

	root.Top.Root = tmpTop

	ka.flushToStruct(root)

	ka.clear()

	return plist.Marshal(root, plist.BinaryFormat)
}

func (ka *NSKeyedArchiver) convertValue(v interface{}) interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		className := ka.objRefVal[m["$class"].(plist.UID)].(map[string]interface{})["$classname"]

		switch className {
		case NSMutableDictionaryClass.Classes[0], NSDictionaryClass.Classes[0]:
			ret := make(map[string]interface{})
			keys := m["NS.keys"].([]interface{})
			values := m["NS.objects"].([]interface{})

			for i := 0; i < len(keys); i++ {
				var keyValue string
				key := ka.objRefVal[keys[i].(plist.UID)]
				switch key.(type) {
				case uint64:
					keyValue = strconv.Itoa(int(key.(uint64)))
					break
				default:
					keyValue = key.(string)
				}
				val := ka.convertValue(ka.objRefVal[values[i].(plist.UID)])
				ret[keyValue] = val
			}
			return ret
		case NSMutableArrayClass.Classes[0], NSArrayClass.Classes[0]:
			ret := make([]interface{}, 0)
			values := m["NS.objects"].([]interface{})
			for i := 0; i < len(values); i++ {
				ret = append(ret, ka.convertValue(values[i]))
			}
			return ret
		case NSMutableDataClass.Classes[0], NSDataClass.Classes[0]:
			return m["NS.data"].([]byte)
		case NSDateClass.Classes[0]:
			return time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC).
				Add(time.Duration(m["NS.time"].(float64)) * time.Second)
		case NSErrorClass.Classes[0]:
			err := &NSError{}
			err.NSCode = int(m["NSCode"].(uint64))
			err.NSDomain = ka.objRefVal[m["NSDomain"].(plist.UID)].(string)
			err.NSUserInfo = ka.convertValue(ka.objRefVal[m["NSUserInfo"].(plist.UID)])
			return *err
		}
	} else if uid, ok := v.(plist.UID); ok {
		return ka.convertValue(ka.objRefVal[uid])
	}
	return v
}

func (ka *NSKeyedArchiver) Unmarshal(b []byte) (interface{}, error) {
	archiver := new(KeyedArchiver)

	if _, err := plist.Unmarshal(b, archiver); err != nil {
		return nil, err
	}

	for _, v := range archiver.Objects {
		ka.objRefVal = append(ka.objRefVal, v)
	}

	ret := ka.convertValue(ka.objRefVal[archiver.Top.Root])

	ka.clear()

	return ret, nil
}
