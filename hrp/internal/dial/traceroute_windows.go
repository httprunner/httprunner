//go:build windows

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

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

var (
	regexTracertPass    = regexp.MustCompile(`(\d+)[\s*<]+(\d+)\s+ms`)
	regexTracertFailure = regexp.MustCompile(`(\d+)[\s*]+Request timed out`)
)

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

	cmd := exec.Command("tracert", "-h", strconv.Itoa(traceRouteOptions.MaxTTL), traceRouteTarget)
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
		failureLine := regexTracertFailure.FindStringSubmatch(hopLine)
		if len(failureLine) == 2 {
			hopID, _ := strconv.Atoi(failureLine[1])
			traceRouteResult.Details = append(traceRouteResult.Details, TraceRouteResultNode{
				Id: hopID,
			})
			continue
		}
		passLine := regexTracertPass.FindStringSubmatch(hopLine)
		if len(passLine) == 3 {
			hopID, _ := strconv.Atoi(passLine[1])
			fields := strings.Fields(hopLine)
			hopIP := strings.Trim(fields[len(fields)-1], "[]")
			traceRouteResult.Details = append(traceRouteResult.Details, TraceRouteResultNode{
				Id:   hopID,
				Ip:   hopIP,
				Time: passLine[2],
			})
			traceRouteResult.Suc = true
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
