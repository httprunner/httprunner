package uixt

import (
	"fmt"
	"testing"

	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

func TestConvertPoints(t *testing.T) {
	data := "10-09 20:16:48.216 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317808206,\"ext\":\"输入\",\"from\":{\"x\":0.0,\"y\":0.0},\"operation\":\"Gtf-SendKeys\",\"run_time\":627,\"start\":1665317807579,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":0.0,\"y\":0.0}}\n10-09 20:18:22.899 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317902898,\"ext\":\"进入直播间\",\"from\":{\"x\":717.0,\"y\":2117.5},\"operation\":\"Gtf-Tap\",\"run_time\":121,\"start\":1665317902777,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":717.0,\"y\":2117.5}}\n10-09 20:18:32.063 I/iesqaMonitor(17845): {\"duration\":0,\"end\":1665317912062,\"ext\":\"第一次上划\",\"from\":{\"x\":1437.0,\"y\":2409.9},\"operation\":\"Gtf-Swipe\",\"run_time\":32,\"start\":1665317912030,\"start_first\":0,\"start_last\":0,\"to\":{\"x\":1437.0,\"y\":2409.9}}"
	eps := ConvertPoints(data)
	if len(eps) != 3 {
		t.Fatal()
	}
	jsons, _ := json.Marshal(eps)
	println(fmt.Sprintf("%v", string(jsons)))
}
