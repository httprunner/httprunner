package hrp

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type UIXTRunner struct {
	Ctx       context.Context
	Configs   *UIXTConfig
	Session   *SessionRunner
	DriverExt *uixt.XTDriver

	RestartCount int // app restart count
	RetryCount   int // retry count
}

type UIXTConfig struct {
	uixt.DriverCacheConfig // includes Platform, Serial, AIOptions

	// Runtime context
	Ctx    context.Context
	Cancel context.CancelFunc `json:"-"`

	// Test case configuration
	JSONCase ITestCase

	// Device specific options
	UIA2         bool // UIAutomator2（Android）
	LogOn        bool // 开启打点日志
	WDAPort      int  // iOS WebDriverAgent port
	WDAMjpegPort int  // iOS WebDriverAgent MJPEG port

	// Agent behavior configuration
	Timeout            int     // seconds
	AbortErrors        []error // abort errors
	MaxRestartAppCount int     // max app restart count
	MaxRetryCount      int     // max retry count

	// Backward compatibility fields - legacy API support
	OSType     string                // deprecated: use Platform from DriverCacheConfig
	Serial     string                // deprecated: use Serial from DriverCacheConfig
	LLMService option.LLMServiceType // deprecated: use AIOptions from DriverCacheConfig
}

const (
	DEFAULT_TIMEOUT               = 1200 // 20 minutes
	DEFAULT_MAX_RESTART_APP_COUNT = 3    // max app restart count
	DEFAULT_MAX_RETRY_COUNT       = 3    // max retry count
)

func NewUIXTRunner(configs *UIXTConfig) (runner *UIXTRunner, err error) {
	configs.addDefault()
	log.Info().Str("version", version.GetVersionInfo()).
		Interface("configs", configs).Msg("init UIXT runner")

	// init testcase config
	var config *TConfig
	var testSteps []IStep
	if configs.JSONCase != nil {
		// load testcase
		testCases, err := LoadTestCases(configs.JSONCase)
		if err != nil || len(testCases) == 0 {
			return nil, errors.Wrap(err, "load testcase failed")
		}
		testCase := testCases[0]
		config = testCase.Config.Get()
		testSteps = testCase.TestSteps
	} else {
		config = NewConfig("config agent")
	}
	config.SetAIOptions(configs.AIOptions...)

	switch configs.Platform {
	case "ios":
		port, err := configs.getWDALocalPort(configs.Serial)
		if err != nil {
			log.Error().Err(err).Msg("get ios agent WDA local port failed")
		} else {
			log.Info().Str("port", port).Msg("set WDA_LOCAL_PORT env")
			os.Setenv("WDA_LOCAL_PORT", port)
		}
		config.SetIOS(
			option.WithUDID(configs.Serial),
			option.WithWDAPort(configs.WDAPort),
			option.WithWDAMjpegPort(configs.WDAMjpegPort),
			option.WithWDALogOn(configs.LogOn),
		)
	case "harmony":
		config.SetHarmony(
			option.WithConnectKey(configs.Serial),
		)
	case "darwin":
		width, height := 1920, 1080
		osWidth := os.Getenv("OSWidth")
		osHeight := os.Getenv("OSHeight")
		if osHeight != "" && osWidth != "" {
			width, err = strconv.Atoi(osWidth)
			if err != nil {
				log.Warn().Msg("get OSWidth failed, use default value")
			}
			height, err = strconv.Atoi(osHeight)
			if err != nil {
				log.Warn().Msg("get OSHeight failed, use default value")
			}
		}
		log.Info().Int("width", width).Int("height", height).Msg("get darwin screen size")
		config.SetBrowser(
			option.WithBrowserLogOn(false),
			option.WithBrowserPageSize(width, height),
		)
	default:
		// default to android
		configs.Platform = "android"
		config.SetAndroid(
			option.WithSerialNumber(configs.Serial),
			option.WithUIA2(configs.UIA2),
			option.WithAdbLogOn(configs.LogOn),
		)
	}

	testcase := TestCase{
		Config:    config,
		TestSteps: testSteps,
	}

	// create runner with HTML report enabled for UIXT
	hrpRunner := NewRunner(nil).SetSaveTests(true).GenHTMLReport()
	caseRunner, err := NewCaseRunner(testcase, hrpRunner)
	if err != nil {
		return nil, errors.Wrap(err, "init case runner failed")
	}
	sessionRunner := caseRunner.NewSession()

	// Use configs directly as it inherits DriverCacheConfig
	driverCacheConfig := configs.DriverCacheConfig
	driverCacheConfig.AIOptions = config.AIOptions.Options()

	dExt, err := uixt.GetOrCreateXTDriver(driverCacheConfig)
	if err != nil {
		return nil, errors.Wrap(err, "get driver failed")
	}

	// check environment
	if err := CheckEnv(dExt); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(configs.Ctx)
	// create a channel to receive signals
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, syscall.SIGINT, syscall.SIGTERM)

	// cancel when interrupted
	go func() {
		<-interruptSignal
		log.Warn().Msg("interrupted in uixt runner")
		cancel()
	}()

	runner = &UIXTRunner{
		Ctx:       ctx,
		Configs:   configs,
		Session:   sessionRunner,
		DriverExt: dExt,
	}
	return runner, nil
}

func (configs *UIXTConfig) addDefault() {
	// Handle backward compatibility - sync legacy fields to embedded DriverCacheConfig
	if configs.OSType != "" && configs.Platform == "" {
		configs.Platform = configs.OSType
	}
	if configs.Serial != "" && configs.DriverCacheConfig.Serial == "" {
		configs.DriverCacheConfig.Serial = configs.Serial
	}
	if configs.LLMService != "" && len(configs.AIOptions) == 0 {
		configs.AIOptions = []option.AIServiceOption{
			option.WithLLMService(configs.LLMService),
		}
	}

	if configs.Ctx == nil {
		configs.Ctx = context.Background()
	}
	if configs.Timeout == 0 {
		configs.Timeout = DEFAULT_TIMEOUT
	}
	if configs.MaxRestartAppCount == 0 {
		configs.MaxRestartAppCount = DEFAULT_MAX_RESTART_APP_COUNT
	}
	if configs.MaxRetryCount == 0 {
		configs.MaxRetryCount = DEFAULT_MAX_RETRY_COUNT
	}
	if len(configs.AbortErrors) == 0 {
		configs.AbortErrors = []error{
			// risk control error, abort
			code.RiskControlAccountActivation,
			code.RiskControlSlideVerification,
			code.RiskControlLogout,
			// network error, abort
			code.NetworkError,
		}
	}
	if configs.WDAPort == 0 {
		configs.WDAPort = 8700
	}
	if configs.WDAMjpegPort == 0 {
		configs.WDAMjpegPort = 8800
	}
}

var client = &http.Client{
	Timeout: 10 * time.Minute,
}

func (configs *UIXTConfig) getWDALocalPort(udid string) (string, error) {
	payloadBytes, _ := json.Marshal(map[string]string{
		"device_id": udid,
	})
	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://localhost:%d/get_device_port", configs.WDAMjpegPort),
		bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", errors.Wrap(err, "create request failed")
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "request ios agent failed")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", errors.Wrap(err, "read ios agent response body failed")
	}

	var resp iosAgentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", errors.Wrap(err, "unmarshal ios agent response failed")
	}

	log.Info().Interface("resp", resp).Msg("get ios agent WDA local port")
	if resp.Code != 0 {
		return "", errors.New("ios agent response code != 0")
	}
	return resp.Port, nil
}

type iosAgentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Port    string `json:"port"`
}

func CheckEnv(driverExt *uixt.XTDriver) (err error) {
	log.Info().Msg("check runner environment")

	// 检查设备是否正常
	if err := CheckDevice(driverExt); err != nil {
		log.Error().Err(err).Str("screenshot", "").Msg("check device failed")
		return err
	}

	return nil
}

func CheckDevice(driverExt *uixt.XTDriver) error {
	// 检测截图功能是否正常
	bufSource, err := driverExt.ScreenShot()
	if err != nil {
		return errors.Wrap(err, "screenshot abnormal")
	}

	// 检测设备是否锁屏（截图是否全黑）
	img, _, err := image.Decode(bufSource)
	if err != nil {
		return errors.Wrap(err, "decode screenshot image failed")
	}

	if isImageBlack(img) {
		return errors.Wrap(code.DeviceConfigureError,
			"device screen is locked")
	}

	return nil
}

func isBlack(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r == 0 && g == 0 && b == 0
}

// 判断图片是否全黑
func isImageBlack(img image.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if !isBlack(img.At(x, y)) {
				return false
			}
		}
	}
	return true
}
