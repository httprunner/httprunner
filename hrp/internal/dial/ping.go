package dial

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
)

type PingOptions struct {
	Count     int
	Timeout   time.Duration
	Interval  time.Duration
	SaveTests bool
}

type PingResult struct {
	Suc                bool   `json:"suc"`
	ErrMsg             string `json:"errMsg"`
	Ip                 string `json:"ip"`
	AvgCost            int    `json:"avgCost"`
	MaxCost            int    `json:"maxCost"`
	MinCost            int    `json:"minCost"`
	Lost               int    `json:"lost"`
	PingCount          int    `json:"pingCount"`
	PacketSize         int    `json:"packetSize"`
	ReceivePacketCount int    `json:"receivePacketCount"`
	SendPacketCount    int    `json:"sendPacketCount"`
	SuccessCount       int    `json:"successCount"`
	DebugLog           string `json:"debugLog"`
}

func DoPing(pingOptions *PingOptions, args []string) (err error) {
	if len(args) != 1 {
		return errors.New("there should be one argument")
	}

	var pingResult PingResult
	defer func() {
		if pingOptions.SaveTests {
			pingResultName := fmt.Sprintf("ping_result_%v.json", env.StartTimeStr)
			pingResultPath := filepath.Join(env.RootDir, pingResultName)
			err = builtin.Dump2JSON(pingResult, pingResultPath)
			if err != nil {
				log.Error().Err(err).Msg("save ping result failed")
			}
		}
	}()

	pingTarget := args[0]

	parsedURL, err := url.Parse(pingTarget)
	if err == nil && parsedURL.Host != "" {
		log.Info().Msgf("parse input url %v and extract host %v", pingTarget, parsedURL.Host)
		pingTarget = strings.Split(parsedURL.Host, ":")[0]
	}

	log.Info().Msgf("ping host %v", pingTarget)
	pinger, err := ping.NewPinger(pingTarget)
	if err != nil {
		log.Error().Err(err).Msgf("fail to get pinger for %s", pingTarget)
		pingResult.Suc = false
		pingResult.ErrMsg = err.Error()
		pingResult.DebugLog = err.Error()
		return
	}
	pinger.Count = pingOptions.Count
	pinger.Timeout = pingOptions.Timeout
	pinger.Interval = pingOptions.Interval

	pinger.OnRecv = func(pkt *ping.Packet) {
		pingResult.DebugLog += fmt.Sprintf("%d bytes from %s: icmp_seq=%d time=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		pingResult.DebugLog += fmt.Sprintf("\n--- %s ping statistics ---\n", stats.Addr)
		pingResult.DebugLog += fmt.Sprintf("%d packets transmitted, %d packets received, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		pingResult.DebugLog += fmt.Sprintf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}
	pingResult.DebugLog += fmt.Sprintf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())

	err = pinger.Run() // blocks until finished
	if err != nil {
		log.Error().Err(err).Msgf("fail to run ping for %s", parsedURL)
		pingResult.Suc = false
		pingResult.ErrMsg = err.Error()
		pingResult.DebugLog = err.Error()
		return
	}
	fmt.Print(pingResult.DebugLog)
	stats := pinger.Statistics() // get send/receive/rtt stats
	pingResult.Ip = pinger.IPAddr().String()
	pingResult.AvgCost = int(stats.AvgRtt / time.Millisecond)
	pingResult.MaxCost = int(stats.MaxRtt / time.Millisecond)
	pingResult.MinCost = int(stats.MinRtt / time.Millisecond)
	pingResult.Lost = int(stats.PacketLoss)
	pingResult.PingCount = pingOptions.Count
	pingResult.PacketSize = pinger.Size
	pingResult.ReceivePacketCount = stats.PacketsRecv
	pingResult.SendPacketCount = stats.PacketsSent
	pingResult.SuccessCount = stats.PacketsRecv
	pingResult.Suc = true
	pingResult.ErrMsg = ""
	return
}
