package driver_ext

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type StubIOSDriver struct {
	*uixt.WDADriver

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
	wdaDriver, err := uixt.NewWDADriver(dev)
	if err != nil {
		return nil, err
	}
	driver := &StubIOSDriver{
		WDADriver: wdaDriver,
		timeout:   10 * time.Second,
	}

	// setup driver
	if err := driver.Setup(); err != nil {
		return nil, err
	}

	// register driver session reset handler
	driver.Session.RegisterResetHandler(driver.Setup)

	return driver, nil
}

func (s *StubIOSDriver) Setup() error {
	localPort, err := s.getLocalPort()
	if err != nil {
		return err
	}
	err = s.Session.SetupPortForward(localPort)
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

func (s *StubIOSDriver) OpenUrl(urlStr string, options ...option.ActionOption) (err error) {
	targetUrl := fmt.Sprintf("/openURL?url=%s", url.QueryEscape(urlStr))
	_, err = s.Session.GET(targetUrl)
	if err != nil {
		log.Error().Err(err).Msg("get source err")
		return nil
	}
	return nil
}

func (s *StubIOSDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
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

func (s *StubIOSDriver) LoginXigua(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
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

func (s *StubIOSDriver) LoginDouyin(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
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
	bsJSON, err := json.Marshal(params)
	if err != nil {
		return info, err
	}

	urlPrefix, err := s.getUrlPrefix(packageName)
	if err != nil {
		return info, err
	}
	fullUrl := urlPrefix + "/host/login/account/" + urlPrefix
	resp, err := s.Session.POST(bsJSON, fullUrl)
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
	bsJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}
	resp, err := s.Session.POST(bsJSON, fullUrl)
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
