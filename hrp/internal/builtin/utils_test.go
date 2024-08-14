package builtin

import (
	"testing"
)

func TestDownload(t *testing.T) {
	err := DownloadFile("/tmp/bytedance.ds.zip", "https://gtf-eapi-cn.bytedance.com/cn/mostRecent/bytedance.ds.zip")
	if err != nil {
		t.Fatal(err)
	}
}
