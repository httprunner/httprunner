package gidevice

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var ErrAfcStatNotExist = errors.New("afc stat: no such file or directory")

var _ Afc = (*afc)(nil)

func newAfc(client *libimobiledevice.AfcClient) *afc {
	return &afc{client: client}
}

type afc struct {
	client *libimobiledevice.AfcClient
}

func (c *afc) DiskInfo() (info *AfcDiskInfo, err error) {
	if err = c.client.Send(libimobiledevice.AfcOperationGetDeviceInfo, nil, nil); err != nil {
		return nil, fmt.Errorf("afc send 'DiskInfo': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return nil, fmt.Errorf("afc receive 'DiskInfo': %w", err)
	}

	m := respMsg.Map()
	info = &AfcDiskInfo{
		Model: m["Model"],
	}
	if info.TotalBytes, err = strconv.ParseUint(m["FSTotalBytes"], 10, 64); err != nil {
		return nil, fmt.Errorf("afc 'DiskInfo': %w", err)
	}
	if info.FreeBytes, err = strconv.ParseUint(m["FSFreeBytes"], 10, 64); err != nil {
		return nil, fmt.Errorf("afc 'DiskInfo': %w", err)
	}
	if info.BlockSize, err = strconv.ParseUint(m["FSBlockSize"], 10, 64); err != nil {
		return nil, fmt.Errorf("afc 'DiskInfo': %w", err)
	}

	return
}

func (c *afc) ReadDir(dirname string) (names []string, err error) {
	if err = c.client.Send(libimobiledevice.AfcOperationReadDir, toCString(dirname), nil); err != nil {
		return nil, fmt.Errorf("afc send 'ReadDir': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return nil, fmt.Errorf("afc receive 'ReadDir': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return nil, fmt.Errorf("afc 'ReadDir': %w", err)
	}

	names = respMsg.Strings()
	return
}

func (c *afc) Stat(filename string) (info *AfcFileInfo, err error) {
	if err = c.client.Send(libimobiledevice.AfcOperationGetFileInfo, toCString(filename), nil); err != nil {
		return nil, fmt.Errorf("afc send 'Stat': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return nil, fmt.Errorf("afc receive 'Stat': %w", err)
	}

	m := respMsg.Map()

	if len(m) == 0 {
		return nil, ErrAfcStatNotExist
	}

	info = &AfcFileInfo{
		source: m,
		name:   path.Base(filename),
		ifmt:   m["st_ifmt"],
	}
	if info.creationTime, err = strconv.ParseUint(m["st_birthtime"], 10, 64); err != nil {
		return nil, fmt.Errorf("afc 'Stat': %w", err)
	}
	if info.blocks, err = strconv.ParseUint(m["st_blocks"], 10, 64); err != nil {
		return nil, fmt.Errorf("afc 'Stat': %w", err)
	}
	if info.modTime, err = strconv.ParseUint(m["st_mtime"], 10, 64); err != nil {
		return nil, fmt.Errorf("afc 'Stat': %w", err)
	}
	if info.nlink, err = strconv.ParseUint(m["st_nlink"], 10, 64); err != nil {
		return nil, fmt.Errorf("afc 'Stat': %w", err)
	}
	if info.size, err = strconv.ParseUint(m["st_size"], 10, 64); err != nil {
		return nil, fmt.Errorf("afc 'Stat': %w", err)
	}

	return
}

func (c *afc) Open(filename string, mode AfcFileMode) (file *AfcFile, err error) {
	buf := new(bytes.Buffer)
	if err = binary.Write(buf, binary.LittleEndian, uint64(mode)); err != nil {
		return nil, fmt.Errorf("afc send 'Open': %w", err)
	}
	buf.Write(toCString(filename))

	if err = c.client.Send(libimobiledevice.AfcOperationFileOpen, buf.Bytes(), nil); err != nil {
		return nil, fmt.Errorf("afc send 'Open': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return nil, fmt.Errorf("afc receive 'Open': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return nil, fmt.Errorf("afc 'Open': %w", err)
	}

	if respMsg.Operation != libimobiledevice.AfcOperationFileOpenResult {
		return nil, fmt.Errorf("afc operation mistake 'Open': '%d'", respMsg.Operation)
	}

	file = &AfcFile{
		client: c.client,
		fd:     respMsg.Uint64(),
	}
	return
}

func (c *afc) Remove(filePath string) (err error) {
	if err = c.client.Send(libimobiledevice.AfcOperationRemovePath, toCString(filePath), nil); err != nil {
		return fmt.Errorf("afc send 'Remove': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return fmt.Errorf("afc receive 'Remove': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return fmt.Errorf("afc 'Remove': %w", err)
	}

	return
}

func (c *afc) Rename(oldPath string, newPath string) (err error) {
	if err = c.client.Send(libimobiledevice.AfcOperationRenamePath, toCString(oldPath, newPath), nil); err != nil {
		return fmt.Errorf("afc send 'Rename': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return fmt.Errorf("afc receive 'Rename': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return fmt.Errorf("afc 'Rename': %w", err)
	}

	return
}

func (c *afc) Mkdir(path string) (err error) {
	if err = c.client.Send(libimobiledevice.AfcOperationMakeDir, toCString(path), nil); err != nil {
		return fmt.Errorf("afc send 'Mkdir': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return fmt.Errorf("afc receive 'Mkdir': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return fmt.Errorf("afc 'Mkdir': %w", err)
	}

	return
}

func (c *afc) Link(oldName string, newName string, linkType AfcLinkType) (err error) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint64(linkType))
	buf.Write(toCString(oldName, newName))

	if err = c.client.Send(libimobiledevice.AfcOperationMakeLink, buf.Bytes(), nil); err != nil {
		return fmt.Errorf("afc send 'Link': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return fmt.Errorf("afc receive 'Link': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return fmt.Errorf("afc 'Link': %w", err)
	}

	return
}

func (c *afc) Truncate(filePath string, size int64) (err error) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint64(size))
	buf.Write(toCString(filePath))

	if err = c.client.Send(libimobiledevice.AfcOperationTruncateFile, buf.Bytes(), nil); err != nil {
		return fmt.Errorf("afc send 'Truncate': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return fmt.Errorf("afc receive 'Truncate': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return fmt.Errorf("afc 'Truncate': %w", err)
	}

	return
}

func (c *afc) SetFileModTime(filePath string, modTime time.Time) (err error) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint64(modTime.Unix()))
	buf.Write(toCString(filePath))

	if err = c.client.Send(libimobiledevice.AfcOperationSetFileModTime, buf.Bytes(), nil); err != nil {
		return fmt.Errorf("afc send 'SetFileModTime': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return fmt.Errorf("afc receive 'SetFileModTime': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return fmt.Errorf("afc 'SetFileModTime': %w", err)
	}

	return
}

func (c *afc) Hash(filePath string) ([]byte, error) {
	var err error
	if err = c.client.Send(libimobiledevice.AfcOperationGetFileHash, toCString(filePath), nil); err != nil {
		return nil, fmt.Errorf("afc send 'Hash': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return nil, fmt.Errorf("afc receive 'Hash': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return nil, fmt.Errorf("afc 'Hash': %w", err)
	}

	return respMsg.Payload, nil
}

func (c *afc) HashWithRange(filePath string, start, end uint64) ([]byte, error) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, start)
	_ = binary.Write(buf, binary.LittleEndian, end)
	buf.Write(toCString(filePath))

	var err error
	if err = c.client.Send(libimobiledevice.AfcOperationGetFileHashRange, buf.Bytes(), nil); err != nil {
		return nil, fmt.Errorf("afc send 'HashWithRange': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return nil, fmt.Errorf("afc receive 'HashWithRange': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return nil, fmt.Errorf("afc 'HashWithRange': %w", err)
	}

	return respMsg.Payload, nil
}

// RemoveAll since iOS6+
func (c *afc) RemoveAll(path string) (err error) {
	if err = c.client.Send(libimobiledevice.AfcOperationRemovePathAndContents, toCString(path), nil); err != nil {
		return fmt.Errorf("afc send 'RemoveAll': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = c.client.Receive(); err != nil {
		return fmt.Errorf("afc receive 'RemoveAll': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return fmt.Errorf("afc 'RemoveAll': %w", err)
	}

	return
}

func (c *afc) WriteFile(filename string, data []byte, perm AfcFileMode) (err error) {
	var file *AfcFile
	if file, err = c.Open(filename, perm); err != nil {
		return err
	}
	defer func() {
		err = file.Close()
	}()

	if _, err = file.Write(data); err != nil {
		return err
	}
	return
}

func toCString(s ...string) []byte {
	buf := new(bytes.Buffer)
	for _, v := range s {
		buf.WriteString(v)
		buf.WriteByte(0)
	}
	return buf.Bytes()
}

type AfcDiskInfo struct {
	Model      string
	TotalBytes uint64
	FreeBytes  uint64
	BlockSize  uint64
}

type AfcFileInfo struct {
	name string

	creationTime uint64
	blocks       uint64
	ifmt         string
	modTime      uint64
	nlink        uint64
	size         uint64

	source map[string]string
}

func (f *AfcFileInfo) Name() string {
	return f.name
}

func (f *AfcFileInfo) Size() int64 {
	return int64(f.size)
}

// func (f *AfcFileInfo) Mode() os.FileMode {
// 	return os.ModeType
// }

func (f *AfcFileInfo) ModTime() time.Time {
	return time.Unix(0, int64(f.modTime))
}

func (f *AfcFileInfo) IsDir() bool {
	return f.ifmt == "S_IFDIR"
}

// func (f *AfcFileInfo) Sys() interface{} {
// 	return f.source
// }

func (f *AfcFileInfo) CreationTime() time.Time {
	return time.Unix(0, int64(f.creationTime))
}

// func (f *AfcFileInfo) Blocks() uint64 {
// 	return f.blocks
// }

// func (f *AfcFileInfo) Format() string {
// 	return f.ifmt
// }

// func (f *AfcFileInfo) Link() uint64 {
// 	return f.nlink
// }

// func (f *AfcFileInfo) PhysicalSize(info *AfcDiskInfo) int64 {
// 	return int64(f.blocks * (info.BlockSize / 8))
// }

type AfcFileMode uint32

const (
	AfcFileModeRdOnly   AfcFileMode = 0x00000001
	AfcFileModeRw       AfcFileMode = 0x00000002
	AfcFileModeWrOnly   AfcFileMode = 0x00000003
	AfcFileModeWr       AfcFileMode = 0x00000004
	AfcFileModeAppend   AfcFileMode = 0x00000005
	AfcFileModeRdAppend AfcFileMode = 0x00000006
)

type AfcLockType int

const (
	AfcLockTypeSharedLock    AfcLockType = 1 | 4
	AfcLockTypeExclusiveLock AfcLockType = 2 | 4
	AfcLockTypeUnlock        AfcLockType = 8 | 4
)

type AfcFile struct {
	client *libimobiledevice.AfcClient
	fd     uint64
	reader *bytes.Reader
}

func (f *AfcFile) op(o ...uint64) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, f.fd)

	if len(o) == 0 {
		return buf.Bytes()
	}

	for _, v := range o {
		_ = binary.Write(buf, binary.LittleEndian, v)
	}

	return buf.Bytes()
}

func (f *AfcFile) Lock(lockType AfcLockType) (err error) {
	if err = f.client.Send(libimobiledevice.AfcOperationFileRefLock, f.op(uint64(lockType)), nil); err != nil {
		return fmt.Errorf("afc file send 'Lock': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = f.client.Receive(); err != nil {
		return fmt.Errorf("afc file receive 'Lock': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return fmt.Errorf("afc file 'Lock': %w", err)
	}
	return
}

func (f *AfcFile) Unlock() (err error) {
	return f.Lock(AfcLockTypeUnlock)
}

func (f *AfcFile) Read(b []byte) (n int, err error) {
	if err = f.client.Send(libimobiledevice.AfcOperationFileRead, f.op(uint64(len(b))), nil); err != nil {
		return -1, fmt.Errorf("afc file send 'Read': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = f.client.Receive(); err != nil {
		return -1, fmt.Errorf("afc file receive 'Read': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return -1, fmt.Errorf("afc file 'Read': %w", err)
	}

	if respMsg.Payload == nil {
		return 0, io.EOF
	}

	if f.reader == nil {
		f.reader = bytes.NewReader(respMsg.Payload)
	} else {
		f.reader.Reset(respMsg.Payload)
	}

	return f.reader.Read(b)
}

func (f *AfcFile) Write(b []byte) (n int, err error) {
	if err = f.client.Send(libimobiledevice.AfcOperationFileWrite, f.op(), b); err != nil {
		return -1, fmt.Errorf("afc file send 'Write': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = f.client.Receive(); err != nil {
		return -1, fmt.Errorf("afc file receive 'Write': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return -1, fmt.Errorf("afc file 'Write': %w", err)
	}

	n = len(b)
	return
}

func (f *AfcFile) Tell() (n uint64, err error) {
	if err = f.client.Send(libimobiledevice.AfcOperationFileTell, f.op(), nil); err != nil {
		return 0, fmt.Errorf("afc file 'Tell': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = f.client.Receive(); err != nil {
		return 0, fmt.Errorf("afc file receive 'Tell': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return 0, fmt.Errorf("afc file 'Tell': %w", err)
	}

	if respMsg.Operation != libimobiledevice.AfcOperationFileTellResult {
		return 0, fmt.Errorf("afc operation mistake 'Tell': '%d'", respMsg.Operation)
	}

	n = respMsg.Uint64()
	return
}

func (f *AfcFile) Seek(offset int64, whence int) (ret int64, err error) {
	if err = f.client.Send(libimobiledevice.AfcOperationFileSeek, f.op(uint64(whence), uint64(offset)), nil); err != nil {
		return -1, fmt.Errorf("afc file 'Seek': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = f.client.Receive(); err != nil {
		return -1, fmt.Errorf("afc file receive 'Seek': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return -1, fmt.Errorf("afc file 'Seek': %w", err)
	}

	var tell uint64
	if tell, err = f.Tell(); err != nil {
		return -1, err
	}

	ret = int64(tell)
	return
}

func (f *AfcFile) Truncate(size int64) (err error) {
	if err = f.client.Send(libimobiledevice.AfcOperationFileSetSize, f.op(uint64(size)), nil); err != nil {
		return fmt.Errorf("afc file 'Truncate': %w", err)
	}
	var respMsg *libimobiledevice.AfcMessage
	if respMsg, err = f.client.Receive(); err != nil {
		return fmt.Errorf("afc file receive 'Truncate': %w", err)
	}
	if err = respMsg.Err(); err != nil {
		return fmt.Errorf("afc file 'Truncate': %w", err)
	}

	return
}

func (f *AfcFile) Close() (err error) {
	if err = f.client.Send(libimobiledevice.AfcOperationFileClose, f.op(), nil); err != nil {
		return fmt.Errorf("afc file 'Close': %w", err)
	}
	if _, err = f.client.Receive(); err != nil {
		return fmt.Errorf("afc file receive 'Close': %w", err)
	}

	return
}

type AfcLinkType int

const (
	AfcLinkTypeHardLink AfcLinkType = 1
	AfcLinkTypeSymLink  AfcLinkType = 2
)
