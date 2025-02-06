//go:build localtest

package uixt

import (
	"regexp"
	"testing"
)

func TestCheckPopup(t *testing.T) {
	setupAndroidAdbDriver(t)
	popup, err := driverExt.CheckPopup()
	if err != nil {
		t.Logf("check popup failed, err: %v", err)
	} else if popup == nil {
		t.Log("no popup found")
	} else {
		t.Logf("found popup: %v", popup)
	}
}

func TestClosePopup(t *testing.T) {
	setupAndroidAdbDriver(t)

	if err := driverExt.ClosePopupsHandler(); err != nil {
		t.Fatal(err)
	}
}

func matchPopup(text string) bool {
	for _, popup := range popups {
		if regexp.MustCompile(popup[1]).MatchString(text) {
			return true
		}
	}
	return false
}

func TestMatchRegex(t *testing.T) {
	testData := []string{
		"以后再说", "我知道了", "同意", "拒绝", "稍后",
		"始终允许", "继续使用", "仅在使用中允许",
	}
	for _, text := range testData {
		if !matchPopup(text) {
			t.Fatal(text)
		}
	}
}
