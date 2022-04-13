package hrp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"

	"github.com/httprunner/httprunner/hrp/internal/builtin"
	"github.com/httprunner/httprunner/hrp/internal/sdk"
)

// Run starts to run API test with default configs.
func Run(testcases ...ITestCase) error {
	t := &testing.T{}
	return NewRunner(t).SetRequestsLogOn().Run(testcases...)
}

// NewRunner constructs a new runner instance.
func NewRunner(t *testing.T) *HRPRunner {
	if t == nil {
		t = &testing.T{}
	}
	return &HRPRunner{
		t:             t,
		failfast:      true, // default to failfast
		genHTMLReport: false,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 30 * time.Second,
		},
		http2Client: &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 30 * time.Second,
		},
	}
}

type HRPRunner struct {
	t             *testing.T
	failfast      bool
	requestsLogOn bool
	pluginLogOn   bool
	saveTests     bool
	genHTMLReport bool
	httpClient    *http.Client
	http2Client   *http.Client
}

// SetClientTransport configures transport of http client for high concurrency load testing
func (r *HRPRunner) SetClientTransport(maxConns int, disableKeepAlive bool, disableCompression bool) *HRPRunner {
	log.Info().
		Int("maxConns", maxConns).
		Bool("disableKeepAlive", disableKeepAlive).
		Bool("disableCompression", disableCompression).
		Msg("[init] SetClientTransport")
	r.httpClient.Transport = &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DialContext:         (&net.Dialer{}).DialContext,
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: maxConns,
		DisableKeepAlives:   disableKeepAlive,
		DisableCompression:  disableCompression,
	}
	r.http2Client.Transport = &http2.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: disableCompression,
	}
	return r
}

// SetFailfast configures whether to stop running when one step fails.
func (r *HRPRunner) SetFailfast(failfast bool) *HRPRunner {
	log.Info().Bool("failfast", failfast).Msg("[init] SetFailfast")
	r.failfast = failfast
	return r
}

// SetRequestsLogOn turns on request & response details logging.
func (r *HRPRunner) SetRequestsLogOn() *HRPRunner {
	log.Info().Msg("[init] SetRequestsLogOn")
	r.requestsLogOn = true
	return r
}

// SetPluginLogOn turns on plugin logging.
func (r *HRPRunner) SetPluginLogOn() *HRPRunner {
	log.Info().Msg("[init] SetPluginLogOn")
	r.pluginLogOn = true
	return r
}

// SetProxyUrl configures the proxy URL, which is usually used to capture HTTP packets for debugging.
func (r *HRPRunner) SetProxyUrl(proxyUrl string) *HRPRunner {
	log.Info().Str("proxyUrl", proxyUrl).Msg("[init] SetProxyUrl")
	p, err := url.Parse(proxyUrl)
	if err != nil {
		log.Error().Err(err).Str("proxyUrl", proxyUrl).Msg("[init] invalid proxyUrl")
		return r
	}
	r.httpClient.Transport = &http.Transport{
		Proxy:           http.ProxyURL(p),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return r
}

// SetSaveTests configures whether to save summary of tests.
func (r *HRPRunner) SetSaveTests(saveTests bool) *HRPRunner {
	log.Info().Bool("saveTests", saveTests).Msg("[init] SetSaveTests")
	r.saveTests = saveTests
	return r
}

// GenHTMLReport configures whether to gen html report of api tests.
func (r *HRPRunner) GenHTMLReport() *HRPRunner {
	log.Info().Bool("genHTMLReport", true).Msg("[init] SetgenHTMLReport")
	r.genHTMLReport = true
	return r
}

// Run starts to execute one or multiple testcases.
func (r *HRPRunner) Run(testcases ...ITestCase) error {
	event := sdk.EventTracking{
		Category: "RunAPITests",
		Action:   "hrp run",
	}
	// report start event
	go sdk.SendEvent(event)
	// report execution timing event
	defer sdk.SendEvent(event.StartTiming("execution"))
	// record execution data to summary
	s := newOutSummary()

	// load all testcases
	testCases, err := loadTestCases(testcases...)
	if err != nil {
		return err
	}

	// run testcase one by one
	for _, testcase := range testCases {
		// each testcase has its own session runner
		sessionRunner, err := r.NewSessionRunner(testcase)
		if err != nil {
			log.Error().Err(err).Msg("[Run] init session runner failed")
			return err
		}
		defer sessionRunner.parser.plugin.Quit()

		cfg := testcase.Config
		// parse config parameters
		err = initParameterIterator(cfg, "runner")
		if err != nil {
			log.Error().Interface("parameters", cfg.Parameters).Err(err).Msg("parse config parameters failed")
			return err
		}
		// 在runner模式下，指定整体策略，cfg.ParametersSetting.Iterators仅包含一个CartesianProduct的迭代器
		for it := cfg.ParametersSetting.Iterators[0]; it.HasNext(); {
			var parameterVariables map[string]interface{}
			// iterate through all parameter iterators and update case variables
			for _, it := range cfg.ParametersSetting.Iterators {
				if it.HasNext() {
					parameterVariables = it.Next()
				}
			}
			sessionRunner.parseConfig(parameterVariables)
			if err = sessionRunner.Start(); err != nil {
				log.Error().Err(err).Msg("[Run] run testcase failed")
				return err
			}
			caseSummary := sessionRunner.GetSummary()
			s.appendCaseSummary(caseSummary)
		}
	}
	s.Time.Duration = time.Since(s.Time.StartAt).Seconds()

	// save summary
	if r.saveTests {
		dir, _ := filepath.Split(summaryPath)
		err := builtin.EnsureFolderExists(dir)
		if err != nil {
			return err
		}
		err = builtin.Dump2JSON(s, fmt.Sprintf(summaryPath, s.Time.StartAt.Unix()))
		if err != nil {
			return err
		}
	}

	// generate HTML report
	if r.genHTMLReport {
		err := s.genHTMLReport()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *HRPRunner) NewSessionRunner(testcase *TestCase) (*SessionRunner, error) {
	sessionRunner := &SessionRunner{
		testCase:  testcase,
		hrpRunner: r,
		parser:    newParser(),
		summary:   newSummary(),
	}

	// init parser plugin
	plugin, err := initPlugin(testcase.Config.Path, r.pluginLogOn)
	if err != nil {
		return nil, errors.Wrap(err, "init plugin failed")
	}
	sessionRunner.parser.plugin = plugin

	// parse testcase config
	if err := sessionRunner.parseConfig(nil); err != nil {
		return nil, errors.Wrap(err, "parse testcase config failed")
	}

	return sessionRunner, nil
}
