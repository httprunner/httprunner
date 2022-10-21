package demo

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/pkg/httpstat"
)

func TestMain(t *testing.T) {
	var httpStat httpstat.Stat

	req, _ := http.NewRequest("GET", "https://httprunner.com", nil)
	ctx := httpstat.WithHTTPStat(req, &httpStat)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	// get stat
	httpStat.Finish()
	result := httpStat.Durations()
	fmt.Println(result)

	// print stat
	httpStat.Print()
}
