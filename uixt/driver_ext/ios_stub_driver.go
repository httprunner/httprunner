package driver_ext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type StubIOSDriver struct {
	Device              *uixt.IOSDevice
	Session             *uixt.DriverSession
	WDADriver           *uixt.WDADriver
	timeout             time.Duration
	douyinUrlPrefix     string
	douyinLiteUrlPrefix string
}

const (
	IOSDouyinPort           = 32921
	IOSDouyinLitePort       = 33461
	defaultBightInsightPort = 8000
)

func NewStubIOSDriver(dev *uixt.IOSDevice) (*StubIOSDriver, error) {
	driver := &StubIOSDriver{
		Device:  dev,
		timeout: 10 * time.Second,
		Session: uixt.NewDriverSession(),
	}

	// setup driver
	if err := driver.Setup(); err != nil {
		return nil, err
	}

	return driver, nil
}

func (s *StubIOSDriver) SetupWda() (err error) {
	if s.WDADriver != nil {
		return nil
	}
	s.WDADriver, err = uixt.NewWDADriver(s.Device)
	return err
}

func (s *StubIOSDriver) GetDriver() uixt.IDriver {
	return s.WDADriver
}

func (s *StubIOSDriver) Setup() error {
	localPort, err := s.getLocalPort()
	if err != nil {
		return err
	}
	s.Session.SetBaseURL(fmt.Sprintf("http://127.0.0.1:%d", localPort))

	localDouyinPort, err := builtin.GetFreePort()
	if err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = s.Device.Forward(localDouyinPort, IOSDouyinPort); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}
	s.douyinUrlPrefix = fmt.Sprintf("http://127.0.0.1:%d", localDouyinPort)

	localDouyinLitePort, err := builtin.GetFreePort()
	if err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = s.Device.Forward(localDouyinLitePort, IOSDouyinLitePort); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}
	s.douyinLiteUrlPrefix = fmt.Sprintf("http://127.0.0.1:%d", localDouyinLitePort)
	return nil
}

func (s *StubIOSDriver) getLocalPort() (int, error) {
	localStubPort, err := builtin.GetFreePort()
	if err != nil {
		return 0, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = s.Device.Forward(localStubPort, defaultBightInsightPort); err != nil {
		return 0, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}
	return localStubPort, nil
}

func (s *StubIOSDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := s.Session.GET("/source?format=json&onlyWeb=false")
	if err != nil {
		log.Error().Err(err).Msg("get source err")
		return "", nil
	}
	return string(resp), nil
}

func (s *StubIOSDriver) OpenUrl(urlStr string, opts ...option.ActionOption) (err error) {
	targetUrl := fmt.Sprintf("/openURL?url=%s", url.QueryEscape(urlStr))
	_, err = s.Session.GET(targetUrl)
	if err != nil {
		log.Error().Err(err).Msg("get source err")
		return nil
	}
	return nil
}

func (s *StubIOSDriver) LoginNoneUI(packageName, phoneNumber, captcha, password string) (info AppLoginInfo, err error) {
	appInfo, err := s.ForegroundInfo()
	if err != nil {
		return info, err
	}
	if appInfo.BundleId == "com.ss.iphone.ugc.AwemeInhouse" || appInfo.BundleId == "com.ss.iphone.ugc.awemeinhouse.lite" {
		return s.LoginDouyin(appInfo.BundleId, phoneNumber, captcha, password)
	} else if appInfo.BundleId == "com.ss.iphone.InHouse.article.Video" {
		return s.LoginXigua(appInfo.BundleId, phoneNumber, captcha, password)
	} else {
		return info, fmt.Errorf("not support app")
	}
}

func (s *StubIOSDriver) LoginXigua(packageName, phoneNumber, captcha, password string) (info AppLoginInfo, err error) {
	loginSchema := ""
	if captcha != "" {
		loginSchema = fmt.Sprintf("snssdk32://local_channel_autologin?login_type=1&account=%s&smscode=%s", phoneNumber, captcha)
	} else if password != "" {
		loginSchema = fmt.Sprintf("snssdk32://local_channel_autologin?login_type=2&account=%s&password=%s", phoneNumber, password)
	} else {
		return info, fmt.Errorf("password and capcha is empty")
	}
	info.IsLogin = true
	return info, s.OpenUrl(loginSchema)
}

func (s *StubIOSDriver) LoginDouyin(packageName, phoneNumber, captcha, password string) (info AppLoginInfo, err error) {
	params := map[string]interface{}{
		"phone": phoneNumber,
	}
	if captcha != "" {
		params["captcha"] = captcha
	} else if password != "" {
		params["password"] = password
	} else {
		return info, fmt.Errorf("password and capcha is empty")
	}
	urlPrefix, err := s.getUrlPrefix(packageName)
	if err != nil {
		return info, err
	}
	fullUrl := urlPrefix + "/host/login/account/"
	resp, err := s.Session.POST(params, fullUrl)
	if err != nil {
		return info, err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return info, err
	}
	log.Info().Msgf("%v", res)
	// {'isSuccess': True, 'data': '登录成功', 'code': 0}
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to logout %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return info, err
	}
	time.Sleep(20 * time.Second)
	info, err = s.getLoginAppInfo(packageName)
	if err != nil || !info.IsLogin {
		return info, fmt.Errorf("falied to login %v", info)
	}
	return info, nil
}

func (s *StubIOSDriver) LogoutNoneUI(packageName string) error {
	urlPrefix, err := s.getUrlPrefix(packageName)
	if err != nil {
		return err
	}
	fullUrl := urlPrefix + "/host/loginout/"
	resp, err := s.Session.GET(fullUrl)
	if err != nil {
		return err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return err
	}
	log.Info().Msgf("%v", res)
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to logout %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return err
	}
	time.Sleep(10 * time.Second)
	return nil
}

func (s *StubIOSDriver) EnableDevtool(packageName string, enable bool) (err error) {
	urlPrefix, err := s.getUrlPrefix(packageName)
	if err != nil {
		return err
	}
	fullUrl := urlPrefix + "/host/devtool/enable"

	params := map[string]interface{}{
		"enable": enable,
	}
	resp, err := s.Session.POST(params, fullUrl)
	if err != nil {
		return err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return err
	}
	log.Info().Msgf("%v", res)
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to enable devtool %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return err
	}
	return nil
}

func (s *StubIOSDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
	urlPrefix, err := s.getUrlPrefix(packageName)
	if err != nil {
		return info, err
	}
	fullUrl := urlPrefix + "/host/app/info/"

	resp, err := s.Session.GET(fullUrl)
	if err != nil {
		return info, err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return info, err
	}
	log.Info().Msgf("%v", res)
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to get is login %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return info, err
	}
	err = json.Unmarshal([]byte(res["data"].(string)), &info)
	if err != nil {
		return info, err
	}
	return info, nil
}

func (s *StubIOSDriver) getUrlPrefix(packageName string) (urlPrefix string, err error) {
	if packageName == "com.ss.iphone.ugc.AwemeInhouse" {
		urlPrefix = s.douyinUrlPrefix
	} else if packageName == "com.ss.iphone.ugc.awemeinhouse.lite" {
		urlPrefix = s.douyinLiteUrlPrefix
	} else {
		return "", fmt.Errorf("not support app %s", packageName)
	}
	return urlPrefix, nil
}

func (s *StubIOSDriver) ScreenShot(opts ...option.ActionOption) (*bytes.Buffer, error) {
	if err := s.SetupWda(); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.ScreenShot(opts...)
}

func (s *StubIOSDriver) AppLaunch(packageName string) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	err := s.WDADriver.AppLaunch(packageName)
	if err != nil {
		return err
	}
	_ = s.EnableDevtool(packageName, true)
	return nil
}

func (s *StubIOSDriver) GetDevice() uixt.IDevice {
	return s.Device
}

func (s *StubIOSDriver) TearDown() error {
	return nil
}

// session
func (s *StubIOSDriver) InitSession(capabilities option.Capabilities) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.InitSession(capabilities)
}

func (s *StubIOSDriver) GetSession() *uixt.DriverSession {
	if err := s.SetupWda(); err != nil {
		_ = errors.Wrap(code.DeviceHTTPDriverError, err.Error())
		return nil
	}
	return s.WDADriver.GetSession()
}

func (s *StubIOSDriver) DeleteSession() error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.DeleteSession()
}

// device info and status
func (s *StubIOSDriver) Status() (types.DeviceStatus, error) {
	if err := s.SetupWda(); err != nil {
		return types.DeviceStatus{}, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Status()
}

func (s *StubIOSDriver) DeviceInfo() (types.DeviceInfo, error) {
	if err := s.SetupWda(); err != nil {
		return types.DeviceInfo{}, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.DeviceInfo()
}

func (s *StubIOSDriver) BatteryInfo() (types.BatteryInfo, error) {
	if err := s.SetupWda(); err != nil {
		return types.BatteryInfo{}, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.BatteryInfo()
}

func (s *StubIOSDriver) ForegroundInfo() (types.AppInfo, error) {
	if err := s.SetupWda(); err != nil {
		return types.AppInfo{}, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.ForegroundInfo()
}

func (s *StubIOSDriver) WindowSize() (types.Size, error) {
	if err := s.SetupWda(); err != nil {
		return types.Size{}, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.WindowSize()
}

func (s *StubIOSDriver) ScreenRecord(duration time.Duration) (videoPath string, err error) {
	if err := s.SetupWda(); err != nil {
		return "", errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.ScreenRecord(duration)
}

func (s *StubIOSDriver) Orientation() (types.Orientation, error) {
	if err := s.SetupWda(); err != nil {
		return "", errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Orientation()
}

func (s *StubIOSDriver) Rotation() (types.Rotation, error) {
	if err := s.SetupWda(); err != nil {
		return types.Rotation{}, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Rotation()
}

func (s *StubIOSDriver) SetRotation(rotation types.Rotation) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.SetRotation(rotation)
}

func (s *StubIOSDriver) SetIme(ime string) error {
	return types.ErrDriverNotImplemented
}

func (s *StubIOSDriver) Home() error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Home()
}

func (s *StubIOSDriver) Unlock() error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Unlock()
}

func (s *StubIOSDriver) Back() error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Back()
}

func (s *StubIOSDriver) TapXY(x, y float64, opts ...option.ActionOption) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.TapXY(x, y, opts...)
}

func (s *StubIOSDriver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.TapAbsXY(x, y, opts...)
}

func (s *StubIOSDriver) DoubleTap(x, y float64, opts ...option.ActionOption) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.DoubleTap(x, y, opts...)
}

func (s *StubIOSDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.TouchAndHold(x, y, opts...)
}

func (s *StubIOSDriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Drag(fromX, fromY, toX, toY, opts...)
}

func (s *StubIOSDriver) Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Swipe(fromX, fromY, toX, toY, opts...)
}

func (s *StubIOSDriver) Input(text string, opts ...option.ActionOption) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Input(text, opts...)
}

func (s *StubIOSDriver) Backspace(count int, opts ...option.ActionOption) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.Backspace(count, opts...)
}

func (s *StubIOSDriver) AppTerminate(packageName string) (bool, error) {
	if err := s.SetupWda(); err != nil {
		return false, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.AppTerminate(packageName)
}

func (s *StubIOSDriver) AppClear(packageName string) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.AppClear(packageName)
}

// image related
func (s *StubIOSDriver) PushImage(localPath string) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.PushImage(localPath)
}

func (s *StubIOSDriver) ClearImages() error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.ClearImages()
}

// triggers the log capture and returns the log entries
func (s *StubIOSDriver) StartCaptureLog(identifier ...string) error {
	if err := s.SetupWda(); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.StartCaptureLog(identifier...)
}

func (s *StubIOSDriver) StopCaptureLog() (interface{}, error) {
	if err := s.SetupWda(); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	return s.WDADriver.StopCaptureLog()
}
