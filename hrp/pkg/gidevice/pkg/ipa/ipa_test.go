//go:build localtest

package ipa

import (
	"testing"
)

func TestInfo(t *testing.T) {
	name := "/Users/hero/Documents/Workspace/GitHub/taobao-iphone-device/tests/testdata/WebDriverAgentRunner.ipa"
	name = "/private/tmp/derivedDataPath/Build/Products/Release-iphoneos/WebDriverAgentRunner-Runner.ipa"

	info, err := Info(name)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range info {
		t.Logf("%-50s\t%v", k, v)
	}

	t.Log(info["CFBundleIdentifier"])
}
