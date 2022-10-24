package gadb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type syncTransport struct {
	sock        net.Conn
	readTimeout time.Duration
}

func newSyncTransport(sock net.Conn, readTimeout time.Duration) syncTransport {
	return syncTransport{sock: sock, readTimeout: readTimeout}
}

func (sync syncTransport) Send(command, data string) (err error) {
	if len(command) != 4 {
		return errors.New("sync commands must have length 4")
	}
	msg := bytes.NewBufferString(command)
	if err = binary.Write(msg, binary.LittleEndian, int32(len(data))); err != nil {
		return fmt.Errorf("sync transport write: %w", err)
	}
	msg.WriteString(data)

	debugLog(fmt.Sprintf("--> %s", msg.String()))
	return _send(sync.sock, msg.Bytes())
}

func (sync syncTransport) SendStream(reader io.Reader) (err error) {
	syncMaxChunkSize := 64 * 1024
	for err == nil {
		tmp := make([]byte, syncMaxChunkSize)
		var n int
		n, err = reader.Read(tmp)
		if err == io.EOF {
			err = nil
			break
		}
		if err == nil {
			err = sync.sendChunk(tmp[:n])
		}
	}

	return
}

func (sync syncTransport) SendStatus(statusCode string, n uint32) (err error) {
	msg := bytes.NewBufferString(statusCode)
	if err = binary.Write(msg, binary.LittleEndian, n); err != nil {
		return fmt.Errorf("sync transport write: %w", err)
	}
	debugLog(fmt.Sprintf("--> %s", msg.String()))
	return _send(sync.sock, msg.Bytes())
}

func (sync syncTransport) sendChunk(buffer []byte) (err error) {
	msg := bytes.NewBufferString("DATA")
	if err = binary.Write(msg, binary.LittleEndian, int32(len(buffer))); err != nil {
		return fmt.Errorf("sync transport write: %w", err)
	}
	debugLog(fmt.Sprintf("--> %s ......", msg.String()))
	msg.Write(buffer)
	return _send(sync.sock, msg.Bytes())
}

func (sync syncTransport) VerifyStatus() (err error) {
	var status string
	if status, err = sync.ReadStringN(4); err != nil {
		return err
	}

	log := bytes.NewBufferString(fmt.Sprintf("<-- %s", status))
	defer func() {
		debugLog(log.String())
	}()

	var tmpUint32 uint32
	if tmpUint32, err = sync.ReadUint32(); err != nil {
		return fmt.Errorf("sync transport read (status): %w", err)
	}
	log.WriteString(fmt.Sprintf(" %d\t", tmpUint32))

	var msg string
	if msg, err = sync.ReadStringN(int(tmpUint32)); err != nil {
		return err
	}
	log.WriteString(msg)

	if status == "FAIL" {
		err = fmt.Errorf("sync verify status (fail): %s", msg)
		return
	}

	if status != "OKAY" {
		err = fmt.Errorf("sync verify status: Unknown error: %s", msg)
		return
	}

	return
}

var syncReadChunkDone = errors.New("sync read chunk done")

func (sync syncTransport) WriteStream(dest io.Writer) (err error) {
	var chunk []byte
	save := func() error {
		if chunk, err = sync.readChunk(); err != nil && err != syncReadChunkDone {
			return fmt.Errorf("sync read chunk: %w", err)
		}
		if err == syncReadChunkDone {
			return err
		}
		if err = _send(dest, chunk); err != nil {
			return fmt.Errorf("sync write stream: %w", err)
		}
		return nil
	}

	for err == nil {
		err = save()
	}

	if err == syncReadChunkDone {
		err = nil
	}
	return
}

func (sync syncTransport) readChunk() (chunk []byte, err error) {
	var status string
	if status, err = sync.ReadStringN(4); err != nil {
		return nil, err
	}

	log := bytes.NewBufferString("")
	defer func() { debugLog(log.String()) }()

	var tmpUint32 uint32
	if tmpUint32, err = sync.ReadUint32(); err != nil {
		return nil, fmt.Errorf("read chunk (length): %w", err)
	}

	if status == "FAIL" {
		log.WriteString(fmt.Sprintf("<-- %s\t%d\t", status, tmpUint32))
		var sError string
		if sError, err = sync.ReadStringN(int(tmpUint32)); err != nil {
			return nil, fmt.Errorf("read chunk (error message): %w", err)
		}
		err = fmt.Errorf("status (fail): %s", sError)
		log.WriteString(sError)
		return
	}

	switch status {
	case "DONE":
		log.WriteString(fmt.Sprintf("<-- %s", status))
		err = syncReadChunkDone
		return
	case "DATA":
		log.WriteString(fmt.Sprintf("<-- %s\t%d\t", status, tmpUint32))
		if chunk, err = sync.ReadBytesN(int(tmpUint32)); err != nil {
			return nil, err
		}
	default:
		log.WriteString(fmt.Sprintf("<-- %s\t%d\t", status, tmpUint32))
		err = errors.New("unknown error")
	}

	log.WriteString("......")

	return

}

func (sync syncTransport) ReadDirectoryEntry() (entry DeviceFileInfo, err error) {
	var status string
	if status, err = sync.ReadStringN(4); err != nil {
		return DeviceFileInfo{}, err
	}

	log := bytes.NewBufferString(fmt.Sprintf("<-- %s", status))
	defer func() {
		debugLog(log.String())
	}()

	if status == "DONE" {
		return
	}

	log = bytes.NewBufferString(fmt.Sprintf("<-- %s\t", status))

	if err = binary.Read(sync.sock, binary.LittleEndian, &entry.Mode); err != nil {
		return DeviceFileInfo{}, fmt.Errorf("sync transport read (mode): %w", err)
	}
	log.WriteString(entry.Mode.String() + "\t")

	if entry.Size, err = sync.ReadUint32(); err != nil {
		return DeviceFileInfo{}, fmt.Errorf("sync transport read (size): %w", err)
	}
	log.WriteString(fmt.Sprintf("%10d", entry.Size) + "\t")

	var tmpUint32 uint32
	if tmpUint32, err = sync.ReadUint32(); err != nil {
		return DeviceFileInfo{}, fmt.Errorf("sync transport read (time): %w", err)
	}
	entry.LastModified = time.Unix(int64(tmpUint32), 0)
	log.WriteString(entry.LastModified.String() + "\t")

	if tmpUint32, err = sync.ReadUint32(); err != nil {
		return DeviceFileInfo{}, fmt.Errorf("sync transport read (file name length): %w", err)
	}
	log.WriteString(fmt.Sprintf("%d\t", tmpUint32))

	if entry.Name, err = sync.ReadStringN(int(tmpUint32)); err != nil {
		return DeviceFileInfo{}, fmt.Errorf("sync transport read (file name): %w", err)
	}
	log.WriteString(entry.Name + "\t")

	return
}

func (sync syncTransport) ReadUint32() (n uint32, err error) {
	err = binary.Read(sync.sock, binary.LittleEndian, &n)
	return
}

func (sync syncTransport) ReadStringN(size int) (s string, err error) {
	var raw []byte
	if raw, err = sync.ReadBytesN(size); err != nil {
		return "", err
	}
	return string(raw), nil
}

func (sync syncTransport) ReadBytesN(size int) (raw []byte, err error) {
	_ = sync.sock.SetReadDeadline(time.Now().Add(time.Second * sync.readTimeout))
	return _readN(sync.sock, size)
}

func (sync syncTransport) Close() (err error) {
	if sync.sock == nil {
		return nil
	}
	return sync.sock.Close()
}
