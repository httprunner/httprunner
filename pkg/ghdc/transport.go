package ghdc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

var ErrConnBroken = errors.New("socket connection broken")

var DefaultReadTimeout time.Duration = 60 * time.Second

var DATA_UNIT_LENGTH = 4

type transport struct {
	sock        net.Conn
	readTimeout time.Duration
	connectKey  string
}

func newTransport(address string, alive bool, connectKey string, readTimeout ...time.Duration) (tp transport, err error) {
	if len(readTimeout) == 0 {
		readTimeout = []time.Duration{DefaultReadTimeout}
	}
	tp.readTimeout = readTimeout[0]
	tp.connectKey = connectKey
	if tp.sock, err = net.Dial("tcp", address); err != nil {
		err = fmt.Errorf("ghdc transport: %w", err)
		return
	}
	if err = _handleShake(tp.sock, connectKey); err != nil {
		return
	}
	if alive {
		if err = tp.setAlive(); err != nil {
			return
		}
	}
	return
}

func _handleShake(sock net.Conn, connectKey string) (err error) {
	data, err := _readN(sock, 48)
	if err != nil {
		return err
	}
	if !(data[4] == 79 && data[5] == 72 && data[6] == 79 && data[7] == 83 && data[9] == 72 && data[10] == 68 && data[11] == 67) {
		return fmt.Errorf("handle shake error")
	}
	bannerStr := []byte("OHOS HDC\x00\x00\x00\x00")
	connectKeyBytes256 := [256]byte{}
	copy(connectKeyBytes256[:], connectKey)
	size := 12 + 256
	buffer := new(bytes.Buffer)
	if err = binary.Write(buffer, binary.BigEndian, uint32(size)); err != nil {
		return fmt.Errorf("transport write: %w", err)
	}
	if err = binary.Write(buffer, binary.BigEndian, bannerStr); err != nil {
		return fmt.Errorf("transport write: %w", err)
	}
	if err = binary.Write(buffer, binary.BigEndian, connectKeyBytes256); err != nil {
		return fmt.Errorf("transport write: %w", err)
	}
	return _send(sock, buffer.Bytes())
}

func (t transport) setAlive() (err error) {
	return _sendCommand(t.sock, "alive")
}

func (t transport) SendCommand(command string) (err error) {
	return _sendCommand(t.sock, command)
}

func _sendCommand(writer io.Writer, command string) (err error) {
	command += "\x00"
	buf := new(bytes.Buffer)
	if err = binary.Write(buf, binary.BigEndian, uint32(len(command))); err != nil {
		return err
	}
	if err = binary.Write(buf, binary.BigEndian, []byte(command)); err != nil {
		return err
	}
	debugLog(fmt.Sprintf("--> %s", command))
	return _send(writer, buf.Bytes())
}

func (t transport) SendBytes(data []byte) (err error) {
	length := uint32(len(data))

	newData := make([]byte, 4+len(data))

	binary.BigEndian.PutUint32(newData[:4], length)

	copy(newData[4:], data)
	return _send(t.sock, newData)
}

func (t transport) ReadStringAll() (s string, err error) {
	var raw []byte
	raw, err = _readAll(t.sock)
	return string(raw), err
}

func (t transport) ReadAll() (raw []byte, err error) {
	return _readAll(t.sock)
}

func _readAll(reader io.Reader) (raw []byte, err error) {
	buffer := new(bytes.Buffer)
	for true {
		lengthBuf := make([]byte, 4)
		_, err := io.ReadFull(reader, lengthBuf)
		if err != nil {
			if err == io.EOF {
				return buffer.Bytes(), nil
			} else if errors.Is(err, io.ErrUnexpectedEOF) {
				err = fmt.Errorf("reached unexpected EOF, read partial data: %s %v", string(buffer.Bytes()), err)
				return nil, err
			} else {
				return nil, err
			}
		}
		length := binary.BigEndian.Uint32(lengthBuf)

		data, err := _readN(reader, int(length))
		if err != nil {
			return nil, err
		}
		buffer.Write(data)

	}
	return buffer.Bytes(), nil
}

func (t transport) UnpackString() (s string, err error) {
	var raw []byte
	raw, err = t.UnpackBytes()
	return string(raw), err
}

func (t transport) UnpackBytes() (raw []byte, err error) {
	var length string
	if length, err = t.ReadStringN(4); err != nil {
		return nil, err
	}
	var size int64
	if size, err = strconv.ParseInt(length, 16, 64); err != nil {
		return nil, err
	}

	raw, err = t.RehdytesN(int(size))
	debugLog(fmt.Sprintf("\r%s", raw))
	return
}

func (t transport) ReadStringN(size int) (s string, err error) {
	var raw []byte
	if raw, err = t.RehdytesN(size); err != nil {
		return "", err
	}
	return string(raw), nil
}

func (t transport) RehdytesN(size int) (raw []byte, err error) {
	_ = t.sock.SetReadDeadline(time.Now().Add(t.readTimeout))
	return _readN(t.sock, size)
}

func _readResponse(reader io.Reader) error {
	raw, err := _readN(reader, DATA_UNIT_LENGTH)
	if err != nil {
		return fmt.Errorf("failed to read response length: %w", err)
	}
	if !bytes.Equal(raw[0:4], []byte("OKAY")) {
		fmt.Printf("failed to push file: %s\n", string(raw))
	} else {
		return nil
	}
	raw, err = _readAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response data: %w", err)
	}
	return fmt.Errorf("read response error %s", string(raw))
}

func (t transport) Close() (err error) {
	if t.sock == nil {
		return nil
	}
	return t.sock.Close()
}

func _send(writer io.Writer, msg []byte) (err error) {
	for totalSent := 0; totalSent < len(msg); {
		var sent int
		if sent, err = writer.Write(msg[totalSent:]); err != nil {
			return err
		}
		if sent == 0 {
			return ErrConnBroken
		}
		totalSent += sent
	}
	return
}

func _read(reader io.Reader) (data []byte, err error) {
	buf := make([]byte, 4*1024*1024)
	n, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func _readN(reader io.Reader, size int) (raw []byte, err error) {
	raw = make([]byte, 0, size)
	for len(raw) < size {
		buf := make([]byte, size-len(raw))
		var n int
		if n, err = io.ReadFull(reader, buf); err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, ErrConnBroken
		}
		raw = append(raw, buf...)
	}
	return
}
