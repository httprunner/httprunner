package uixt

import (
	"strconv"
	"strings"
	"testing"

	"github.com/electricbubble/gwda"
)

func TestDriverExt_GesturePassword(t *testing.T) {
	split := strings.Split("6304258", "")
	password := make([]int, len(split))
	for i := range split {
		password[i], _ = strconv.Atoi(split[i])
	}

	driver, err := gwda.NewUSBDriver(nil)
	checkErr(t, err)

	driverExt, err := Extend(driver, 0.95)
	checkErr(t, err)

	pathSearch := "/Users/hero/Documents/temp/2020-05/opencv/IMG_5.png"

	err = driverExt.GesturePassword(pathSearch, password...)
	checkErr(t, err)
}
