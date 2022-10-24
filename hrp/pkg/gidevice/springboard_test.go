//go:build localtest

package gidevice

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"testing"
)

var springBoardSrv SpringBoard

func setupSpringBoardSrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	if springBoardSrv, err = lockdownSrv.SpringBoardService(); err != nil {
		t.Fatal(err)
	}
}

func Test_springBoard_GetIcon(t *testing.T) {
	setupSpringBoardSrv(t)
	raw, _ := springBoardSrv.GetIconPNGData("com.tencent.xin")
	img, format, err := image.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create("./abc." + format)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	switch format {
	case "png":
		err = png.Encode(file, img)
	case "jpeg":
		err = jpeg.Encode(file, img, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func Test_springBoard_GetOrient(t *testing.T) {
	setupSpringBoardSrv(t)
	fmt.Println(springBoardSrv.GetInterfaceOrientation())
}
