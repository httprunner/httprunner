//go:build localtest

package gidevice

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

var afcSrv Afc

func setupAfcSrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	if afcSrv, err = lockdownSrv.AfcService(); err != nil {
		t.Fatal(err)
	}
}

func Test_afc_DiskInfo(t *testing.T) {
	setupAfcSrv(t)

	info, err := afcSrv.DiskInfo()
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("%10s: %s\n", "Model", info.Model)
	log.Printf("%10s: %d\n", "BlockSize", info.BlockSize/8)
	log.Printf("%10s: %s\n", "FreeSpace", byteCountDecimal(int64(info.FreeBytes)))
	log.Printf("%10s: %s\n", "UsedSpace", byteCountDecimal(int64(info.TotalBytes-info.FreeBytes)))
	log.Printf("%10s: %s\n", "TotalSpace", byteCountDecimal(int64(info.TotalBytes)))
}

func byteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func Test_afc_ReadDir(t *testing.T) {
	setupAfcSrv(t)

	names, err := afcSrv.ReadDir("Downloads")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		t.Log(name)
	}
}

func Test_afc_Stat(t *testing.T) {
	setupAfcSrv(t)

	fileInfo, err := afcSrv.Stat("Downloads/downloads.28.sqlitedb")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fileInfo.Name())
	t.Log(fileInfo.IsDir())
	t.Log(fileInfo.CreationTime())
	t.Log(fileInfo.ModTime())
	t.Log(fileInfo.Size())
	t.Log(byteCountDecimal(fileInfo.Size()))
}

func Test_afc_Open(t *testing.T) {
	setupAfcSrv(t)

	afcFile, err := afcSrv.Open("DCIM/105APPLE/IMG_5977.JPEG", AfcFileModeRdOnly)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = afcFile.Close()
	}()

	userHomeDir, _ := os.UserHomeDir()
	file, err := os.Create(userHomeDir + "/Desktop/tmp.jpeg")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = io.Copy(file, afcFile); err != nil {
		t.Fatal(err)
	}
}
