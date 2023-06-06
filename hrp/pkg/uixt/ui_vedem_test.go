package uixt

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func checkUI(uiName string, source []byte) error {
	service, err := newVEDEMUIService()
	if err != nil {
		return err
	}
	uiResults, err := service.getUIResult(uiName, source)
	if err != nil {
		return err
	}
	fmt.Println(uiResults)
	return nil
}

func TestUIWithLocalFile(t *testing.T) {
	sourcePath := "/Users/bytedance/Desktop/lifeservice.png"
	file, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}

	if err := checkUI("dyhouse", file); err != nil {
		t.Fatal(err)
	}
}
