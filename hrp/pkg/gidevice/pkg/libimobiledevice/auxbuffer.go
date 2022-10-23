package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/nskeyedarchiver"
)

type AuxBuffer struct {
	buf *bytes.Buffer
}

func NewAuxBuffer() *AuxBuffer {
	return &AuxBuffer{
		buf: new(bytes.Buffer),
	}
}

func (m *AuxBuffer) AppendObject(obj interface{}) error {
	marshal, err := nskeyedarchiver.Marshal(obj)
	if err != nil {
		return err
	}
	m.AppendUInt32(10)
	m.AppendUInt32(2)
	m.AppendUInt32(uint32(len(marshal)))
	m.buf.Write(marshal)

	return nil
}

func (m *AuxBuffer) AppendInt64(v int64) {
	m.AppendUInt32(10)
	m.AppendUInt32(4)
	m.AppendUInt64(uint64(v))
}

func (m *AuxBuffer) AppendInt32(v int32) {
	m.AppendUInt32(10)
	m.AppendUInt32(3)
	m.AppendUInt32(uint32(v))
}

func (m *AuxBuffer) AppendUInt32(v uint32) {
	_ = binary.Write(m.buf, binary.LittleEndian, v)
}

func (m *AuxBuffer) AppendUInt64(v uint64) {
	_ = binary.Write(m.buf, binary.LittleEndian, v)
}

func (m *AuxBuffer) AppendBytes(b []byte) {
	m.buf.Write(b)
}

func (m *AuxBuffer) Len() int {
	return m.buf.Len()
}

func (m *AuxBuffer) Bytes() []byte {
	dup := m.buf.Bytes()
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b, 0x01f0)
	binary.LittleEndian.PutUint64(b[8:], uint64(m.Len()))
	return append(b, dup...)
}

func UnmarshalAuxBuffer(b []byte) ([]interface{}, error) {
	reader := bytes.NewReader(b)
	var magic, pkgLen uint64
	if err := binary.Read(reader, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &pkgLen); err != nil {
		return nil, err
	}

	// if magic != 0x1df0 {
	// 	TODO magic
	// 	return nil, errors.New("magic not equal 0x1df0")
	// }

	if pkgLen > uint64(len(b)-16) {
		return nil, errors.New("package length not enough")
	}

	var ret []interface{}

	for reader.Len() > 0 {
		var flag, typ uint32
		if err := binary.Read(reader, binary.LittleEndian, &flag); err != nil {
			return nil, err
		}
		if err := binary.Read(reader, binary.LittleEndian, &typ); err != nil {
			return nil, err
		}
		switch typ {
		case 2:
			var l uint32
			if err := binary.Read(reader, binary.LittleEndian, &l); err != nil {
				return nil, err
			}
			plistBuf := make([]byte, l)
			if _, err := reader.Read(plistBuf); err != nil {
				return nil, err
			}
			archiver := NewNSKeyedArchiver()
			d, err := archiver.Unmarshal(plistBuf)
			if err != nil {
				return nil, err
			}
			ret = append(ret, d)
		case 3, 5:
			var i int32
			if err := binary.Read(reader, binary.LittleEndian, &i); err != nil {
				return nil, err
			}
			ret = append(ret, i)
		case 4, 6:
			var i int64
			if err := binary.Read(reader, binary.LittleEndian, &i); err != nil {
				return nil, err
			}
			ret = append(ret, i)
		case 10:
			// TODO Dictionary key
			// fmt.Println("Dictionary key!")
			continue
		default:
			// fmt.Printf("unknown type %d\n", typ)
			break
		}
	}

	return ret, nil
}
