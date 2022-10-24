package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type AfcMessage struct {
	Operation uint64
	Data      []byte
	Payload   []byte
}

func (m *AfcMessage) Map() map[string]string {
	ret := make(map[string]string)
	ss := m.Strings()
	if ss != nil {
		for i := 0; i < len(ss); i += 2 {
			ret[ss[i]] = ss[i+1]
		}
	}
	return ret
}

func (m *AfcMessage) Strings() []string {
	if m.Operation == AfcOperationData {
		bs := bytes.Split(m.Payload, []byte{0})
		ss := make([]string, len(bs)-1)
		for i := 0; i < len(ss); i++ {
			ss[i] = string(bs[i])
		}
		return ss
	}
	return nil
}

func (m *AfcMessage) Uint64() uint64 {
	return binary.LittleEndian.Uint64(m.Data)
}

func (m *AfcMessage) Err() error {
	if m.Operation == AfcOperationStatus {
		status := m.Uint64()
		if status != AfcErrSuccess {
			return toError(status)
		}
	}
	return nil
}

func toError(status uint64) error {
	switch status {
	case AfcErrUnknownError:
		return errors.New("UnknownError")
	case AfcErrOperationHeaderInvalid:
		return errors.New("OperationHeaderInvalid")
	case AfcErrNoResources:
		return errors.New("NoResources")
	case AfcErrReadError:
		return errors.New("ReadError")
	case AfcErrWriteError:
		return errors.New("WriteError")
	case AfcErrUnknownPacketType:
		return errors.New("UnknownPacketType")
	case AfcErrInvalidArgument:
		return errors.New("InvalidArgument")
	case AfcErrObjectNotFound:
		return errors.New("ObjectNotFound")
	case AfcErrObjectIsDir:
		return errors.New("ObjectIsDir")
	case AfcErrPermDenied:
		return errors.New("PermDenied")
	case AfcErrServiceNotConnected:
		return errors.New("ServiceNotConnected")
	case AfcErrOperationTimeout:
		return errors.New("OperationTimeout")
	case AfcErrTooMuchData:
		return errors.New("TooMuchData")
	case AfcErrEndOfData:
		return errors.New("EndOfData")
	case AfcErrOperationNotSupported:
		return errors.New("OperationNotSupported")
	case AfcErrObjectExists:
		return errors.New("ObjectExists")
	case AfcErrObjectBusy:
		return errors.New("ObjectBusy")
	case AfcErrNoSpaceLeft:
		return errors.New("NoSpaceLeft")
	case AfcErrOperationWouldBlock:
		return errors.New("OperationWouldBlock")
	case AfcErrIoError:
		return errors.New("IoError")
	case AfcErrOperationInterrupted:
		return errors.New("OperationInterrupted")
	case AfcErrOperationInProgress:
		return errors.New("OperationInProgress")
	case AfcErrInternalError:
		return errors.New("InternalError")
	case AfcErrMuxError:
		return errors.New("MuxError")
	case AfcErrNoMemory:
		return errors.New("NoMemory")
	case AfcErrNotEnoughData:
		return errors.New("NotEnoughData")
	case AfcErrDirNotEmpty:
		return errors.New("DirNotEmpty")
	}
	return nil
}

const (
	AfcOperationInvalid              = 0x00000000 /* Invalid */
	AfcOperationStatus               = 0x00000001 /* Status */
	AfcOperationData                 = 0x00000002 /* Data */
	AfcOperationReadDir              = 0x00000003 /* ReadDir */
	AfcOperationReadFile             = 0x00000004 /* ReadFile */
	AfcOperationWriteFile            = 0x00000005 /* WriteFile */
	AfcOperationWritePart            = 0x00000006 /* WritePart */
	AfcOperationTruncateFile         = 0x00000007 /* TruncateFile */
	AfcOperationRemovePath           = 0x00000008 /* RemovePath */
	AfcOperationMakeDir              = 0x00000009 /* MakeDir */
	AfcOperationGetFileInfo          = 0x0000000A /* GetFileInfo */
	AfcOperationGetDeviceInfo        = 0x0000000B /* GetDeviceInfo */
	AfcOperationWriteFileAtomic      = 0x0000000C /* WriteFileAtomic (tmp file+rename) */
	AfcOperationFileOpen             = 0x0000000D /* FileRefOpen */
	AfcOperationFileOpenResult       = 0x0000000E /* FileRefOpenResult */
	AfcOperationFileRead             = 0x0000000F /* FileRefRead */
	AfcOperationFileWrite            = 0x00000010 /* FileRefWrite */
	AfcOperationFileSeek             = 0x00000011 /* FileRefSeek */
	AfcOperationFileTell             = 0x00000012 /* FileRefTell */
	AfcOperationFileTellResult       = 0x00000013 /* FileRefTellResult */
	AfcOperationFileClose            = 0x00000014 /* FileRefClose */
	AfcOperationFileSetSize          = 0x00000015 /* FileRefSetFileSize (ftruncate) */
	AfcOperationGetConnectionInfo    = 0x00000016 /* GetConnectionInfo */
	AfcOperationSetConnectionOptions = 0x00000017 /* SetConnectionOptions */
	AfcOperationRenamePath           = 0x00000018 /* RenamePath */
	AfcOperationSetFSBlockSize       = 0x00000019 /* SetFSBlockSize (0x800000) */
	AfcOperationSetSocketBlockSize   = 0x0000001A /* SetSocketBlockSize (0x800000) */
	AfcOperationFileRefLock          = 0x0000001B /* FileRefLock */
	AfcOperationMakeLink             = 0x0000001C /* MakeLink */
	AfcOperationGetFileHash          = 0x0000001D /* GetFileHash */
	AfcOperationSetFileModTime       = 0x0000001E /* SetModTime */
	AfcOperationGetFileHashRange     = 0x0000001F /* GetFileHashWithRange */
	/* iOS 6+ */
	AfcOperationFileSetImmutableHint             = 0x00000020 /* FileRefSetImmutableHint */
	AfcOperationGetSizeOfPathContents            = 0x00000021 /* GetSizeOfPathContents */
	AfcOperationRemovePathAndContents            = 0x00000022 /* RemovePathAndContents */
	AfcOperationDirectoryEnumeratorRefOpen       = 0x00000023 /* DirectoryEnumeratorRefOpen */
	AfcOperationDirectoryEnumeratorRefOpenResult = 0x00000024 /* DirectoryEnumeratorRefOpenResult */
	AfcOperationDirectoryEnumeratorRefRead       = 0x00000025 /* DirectoryEnumeratorRefRead */
	AfcOperationDirectoryEnumeratorRefClose      = 0x00000026 /* DirectoryEnumeratorRefClose */
	/* iOS 7+ */
	AfcOperationFileRefReadWithOffset  = 0x00000027 /* FileRefReadWithOffset */
	AfcOperationFileRefWriteWithOffset = 0x00000028 /* FileRefWriteWithOffset */
)

const (
	AfcErrSuccess                = 0
	AfcErrUnknownError           = 1
	AfcErrOperationHeaderInvalid = 2
	AfcErrNoResources            = 3
	AfcErrReadError              = 4
	AfcErrWriteError             = 5
	AfcErrUnknownPacketType      = 6
	AfcErrInvalidArgument        = 7
	AfcErrObjectNotFound         = 8
	AfcErrObjectIsDir            = 9
	AfcErrPermDenied             = 10
	AfcErrServiceNotConnected    = 11
	AfcErrOperationTimeout       = 12
	AfcErrTooMuchData            = 13
	AfcErrEndOfData              = 14
	AfcErrOperationNotSupported  = 15
	AfcErrObjectExists           = 16
	AfcErrObjectBusy             = 17
	AfcErrNoSpaceLeft            = 18
	AfcErrOperationWouldBlock    = 19
	AfcErrIoError                = 20
	AfcErrOperationInterrupted   = 21
	AfcErrOperationInProgress    = 22
	AfcErrInternalError          = 23
	AfcErrMuxError               = 30
	AfcErrNoMemory               = 31
	AfcErrNotEnoughData          = 32
	AfcErrDirNotEmpty            = 33
)
