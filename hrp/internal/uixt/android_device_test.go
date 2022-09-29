package uixt

import (
	"fmt"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
	"testing"
)

func TestConvertPoints(t *testing.T) {
	data := "09-29 15:02:08.379 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434928378\tstartX=720.000000\tstartY=1462.000000\tendX=1296.000000\tendY=1462.000000\n09-29 15:02:09.433 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434929432\tstartX=720.000000\tstartY=1462.000000\tendX=1296.000000\tendY=1462.000000\n09-29 15:02:10.452 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434930452\tstartX=720.000000\tstartY=1462.000000\tendX=1296.000000\tendY=1462.000000\n09-29 15:02:11.451 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434931450\tstartX=720.000000\tstartY=1462.000000\tendX=1296.000000\tendY=1462.000000\n09-29 15:02:12.491 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434932489\tstartX=720.000000\tstartY=1462.000000\tendX=1296.000000\tendY=1462.000000\n09-29 15:02:16.028 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434936027\tstartX=720.000000\tstartY=1462.000000\tendX=144.000000\tendY=1462.000000\n09-29 15:02:21.424 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434941423\tstartX=720.000000\tstartY=1462.000000\tendX=144.000000\tendY=1462.000000\n09-29 15:02:27.923 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434947922\tstartX=720.000000\tstartY=1462.000000\tendX=144.000000\tendY=1462.000000\n09-29 15:02:33.628 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434953628\tstartX=720.000000\tstartY=1462.000000\tendX=144.000000\tendY=1462.000000\n09-29 15:02:39.347 I/iesqaMonitor( 9938): [tap]\ttimesec=1664434959347\tx=1259.5y=1868.5"
	eps := ConvertPoints(data)
	if len(eps) != 10 {
		t.Fatal()
	}
	jsons, _ := json.Marshal(eps)
	println(fmt.Sprintf("%v", string(jsons)))
}
