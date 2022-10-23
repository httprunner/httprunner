package gadb

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"time"
)

var ErrConnBroken = errors.New("socket connection broken")

var DefaultAdbReadTimeout time.Duration = 60

type transport struct {
	sock        net.Conn
	readTimeout time.Duration
}

func newTransport(address string, readTimeout ...time.Duration) (tp transport, err error) {
	if len(readTimeout) == 0 {
		readTimeout = []time.Duration{DefaultAdbReadTimeout}
	}
	tp.readTimeout = readTimeout[0]
	if tp.sock, err = net.Dial("tcp", address); err != nil {
		err = fmt.Errorf("adb transport: %w", err)
	}
	return
}

func (t transport) Send(command string) (err error) {
	msg := fmt.Sprintf("%04x%s", len(command), command)
	debugLog(fmt.Sprintf("--> %s", command))
	return _send(t.sock, []byte(msg))
}

func (t transport) VerifyResponse() (err error) {
	var status string
	if status, err = t.ReadStringN(4); err != nil {
		return err
	}
	if status == "OKAY" {
		debugLog(fmt.Sprintf("<-- %s", status))
		return nil
	}

	var sError string
	if sError, err = t.UnpackString(); err != nil {
		return err
	}
	err = fmt.Errorf("command failed: %s", sError)
	debugLog(fmt.Sprintf("<-- %s %s", status, sError))
	return
}

func (t transport) ReadStringAll() (s string, err error) {
	var raw []byte
	raw, err = t.ReadBytesAll()
	return string(raw), err
}

func (t transport) ReadBytesAll() (raw []byte, err error) {
	raw, err = ioutil.ReadAll(t.sock)
	debugLog(fmt.Sprintf("\r%s", raw))
	return
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

	raw, err = t.ReadBytesN(int(size))
	debugLog(fmt.Sprintf("\r%s", raw))
	return
}

func (t transport) ReadStringN(size int) (s string, err error) {
	var raw []byte
	if raw, err = t.ReadBytesN(size); err != nil {
		return "", err
	}
	return string(raw), nil
}

func (t transport) ReadBytesN(size int) (raw []byte, err error) {
	_ = t.sock.SetReadDeadline(time.Now().Add(time.Second * t.readTimeout))
	return _readN(t.sock, size)
}

func (t transport) Close() (err error) {
	if t.sock == nil {
		return nil
	}
	return t.sock.Close()
}

func (t transport) CreateSyncTransport() (sTp syncTransport, err error) {
	if err = t.Send("sync:"); err != nil {
		return syncTransport{}, err
	}
	if err = t.VerifyResponse(); err != nil {
		return syncTransport{}, err
	}
	sTp = newSyncTransport(t.sock, t.readTimeout)
	return
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
