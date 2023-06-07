package uixt

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func checkUI(uiTypes []string, source []byte) error {
	service, err := newVEDEMUIService()
	if err != nil {
		return err
	}
	uiResults, err := service.FindUI(uiTypes, source)
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

	if err := checkUI([]string{"dyhouse", "shoppingbag"}, file); err != nil {
		t.Fatal(err)
	}
}
