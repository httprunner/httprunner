package dial

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mehrdadrad/mylg/cli"
	"github.com/mehrdadrad/mylg/icmp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

var (
	regexIPAddr            = regexp.MustCompile(`([\d.]+)`)
	regexElapsedTime       = regexp.MustCompile(`(\d+\.\d+)`)
	regexTraceroutePass    = regexp.MustCompile(fmt.Sprintf(`(\d+)[\s*]+(\S+)\s+\(%s\)\s+%s\s+ms`, regexIPAddr, regexElapsedTime))
	regexTracerouteFailure = regexp.MustCompile(`(\d+)[\s*]+$`)
)

type TraceRouteOptions struct {
	MaxTTL    int
	Queries   int
	SaveTests bool
}

type TraceRouteResult struct {
	IP      string                 `json:"ip"`
	Details []TraceRouteResultNode `json:"details"`
	Suc     bool                   `json:"suc"`
	ErrMsg  string                 `json:"errMsg"`
}

type TraceRouteResultNode struct {
	Id   int    `json:"id"`
	Ip   string `json:"ip"`
	Time string `json:"time"`
}

type HopResp struct {
	Num     int     `json:"Id"`
	Hop     string  `json:"Hop"`
	Ip      string  `json:"Ip"`
	Elapsed float64 `json:"Elapsed"`
	Holder  string  `json:"Holder"`
	ASN     float64 `json:"ASN"`
	Last    bool    `json:"Last"`
}

func DoTraceRoute(traceRouteOptions *TraceRouteOptions, args []string) (err error) {
	if len(args) != 1 {
		return errors.New("there should be one argument")
	}
	var traceRouteResult TraceRouteResult
	defer func() {
		if traceRouteOptions.SaveTests {
			dir, _ := os.Getwd()
			traceRouteResultName := fmt.Sprintf("traceroute_result_%v.json", time.Now().Format("20060102150405"))
			traceRouteResultPath := filepath.Join(dir, traceRouteResultName)
			err = builtin.Dump2JSON(traceRouteResult, traceRouteResultPath)
			if err != nil {
				log.Error().Err(err).Msg("save traceroute result failed")
			}
		}
	}()

	traceRouteTarget := args[0]
	parsedURL, err := url.Parse(traceRouteTarget)
	if err == nil && parsedURL.Host != "" {
		log.Info().Msgf("parse input url %v and extract host %v", traceRouteTarget, parsedURL.Host)
		traceRouteTarget = strings.Split(parsedURL.Host, ":")[0]
	}

	cmd := exec.Command("traceroute", "-m", strconv.Itoa(traceRouteOptions.MaxTTL),
		"-q", strconv.Itoa(traceRouteOptions.Queries), traceRouteTarget)
	stdout, _ := cmd.StdoutPipe()

	startT := time.Now()
	defer func() {
		log.Info().Msgf("for target %s, traceroute costs %v", traceRouteTarget, time.Since(startT))
	}()

	log.Info().Msgf("start to traceroute %v", traceRouteTarget)
	err = cmd.Start()
	if err != nil {
		traceRouteResult.Suc = false
		traceRouteResult.ErrMsg = "execute traceroute failed"
		log.Error().Err(err).Msg("start command failed")
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		hopLine := scanner.Text()
		fmt.Println(hopLine)
		passLine := regexTraceroutePass.FindStringSubmatch(hopLine)
		if len(passLine) == 5 {
			hopID, _ := strconv.Atoi(passLine[1])
			traceRouteResult.Details = append(traceRouteResult.Details, TraceRouteResultNode{
				Id:   hopID,
				Ip:   passLine[3],
				Time: passLine[4],
			})
			traceRouteResult.Suc = true
			continue
		}
		failureLine := regexTracerouteFailure.FindStringSubmatch(hopLine)
		if len(failureLine) == 2 {
			hopID, _ := strconv.Atoi(failureLine[1])
			traceRouteResult.Details = append(traceRouteResult.Details, TraceRouteResultNode{
				Id: hopID,
			})
			continue
		}
	}
	hopCount := len(traceRouteResult.Details)
	traceRouteResult.IP = traceRouteResult.Details[hopCount-1].Ip
	err = cmd.Wait()
	if err != nil {
		traceRouteResult.Suc = false
		traceRouteResult.ErrMsg = "wait traceroute finish failed"
		log.Error().Err(err).Msg("wait command failed")
		return
	}
	return
}

// DoTraceRouteSDK with golang SDK, which needs root privilege
func DoTraceRouteSDK(traceRouteOptions *TraceRouteOptions, args []string) (err error) {
	if len(args) != 1 {
		return errors.New("there should be one argument")
	}

	var traceRouteResult TraceRouteResult
	defer func() {
		if traceRouteOptions.SaveTests {
			dir, _ := os.Getwd()
			traceRouteResultName := fmt.Sprintf("traceroute_result_%v.json", time.Now().Format("20060102150405"))
			traceRouteResultPath := filepath.Join(dir, traceRouteResultName)
			err = builtin.Dump2JSON(traceRouteResult, traceRouteResultPath)
			if err != nil {
				log.Error().Err(err).Msg("save traceroute result failed")
			}
		}
	}()

	traceRouteTarget := args[0]
	parsedURL, err := url.Parse(traceRouteTarget)
	if err == nil && parsedURL.Host != "" {
		log.Info().Msgf("parse input url %v and extract host %v", traceRouteTarget, parsedURL.Host)
		traceRouteTarget = strings.Split(parsedURL.Host, ":")[0]
	}

	cfg, err := cli.ReadDefaultConfig()
	if err != nil {
		log.Error().Err(err).Msgf("fail to read default config")
		traceRouteResult.Suc = false
		traceRouteResult.ErrMsg = err.Error()
		return
	}

	traceRouter, err := icmp.NewTrace(traceRouteTarget, cfg)
	if err != nil {
		log.Error().Err(err).Msgf("fail to new traceRouter for %s", traceRouteTarget)
		traceRouteResult.Suc = false
		traceRouteResult.ErrMsg = err.Error()
		return
	}

	startT := time.Now()
	defer func() {
		log.Info().Msgf("for target %s, traceroute costs %v", traceRouteTarget, time.Since(startT))
	}()

	log.Info().Msgf("start to trace route of %v", traceRouteTarget)
	hopRespChan, err := traceRouter.MRun()
	if err != nil {
		log.Error().Err(err).Msgf("fail to trace route of %v", traceRouteTarget)
		traceRouteResult.Suc = false
		traceRouteResult.ErrMsg = err.Error()
	}
	count := 0
	t := time.NewTicker(2 * time.Minute)
	for {
		select {
		case <-t.C:
			log.Error().Err(err).Msgf("fail to do traceroute for %s because timeout", traceRouteTarget)
			traceRouteResult.Suc = false
			traceRouteResult.ErrMsg = "timeout"
			return
		case resp := <-hopRespChan:
			respJSON := resp.Marshal()
			fmt.Printf("traceroute hop: %v\n", respJSON)
			var hopResp HopResp
			err = json.Unmarshal([]byte(respJSON), &hopResp)
			if err != nil {
				log.Error().Err(err).Msgf("fail to do traceroute for %s because of hop response %v unmarshal error", traceRouteTarget, respJSON)
				traceRouteResult.Suc = false
				traceRouteResult.ErrMsg = "hop response unmarshal error"
			}
			traceRouteResult.Details = append(traceRouteResult.Details, TraceRouteResultNode{
				Id:   hopResp.Num,
				Ip:   hopResp.Ip,
				Time: fmt.Sprintf("%.2f", hopResp.Elapsed),
			})
			traceRouteResult.Suc = true
			traceRouteResult.ErrMsg = ""
			if hopResp.Last {
				traceRouteResult.IP = hopResp.Ip
				log.Info().Msgf("for target %s, traceroute completed", traceRouteTarget)
				return
			}
			count += 1
			if count > 30 {
				log.Info().Msgf("for target %s, traceroute hop counts reach limit", traceRouteTarget)
				return
			}
		}
	}
}
