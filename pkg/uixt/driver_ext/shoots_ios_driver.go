package driver_ext

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

const (
	defaultBightInsightPort = 8000
	defaultDouyinServerPort = 32921
)

func NewShootsIOSDriver(device *uixt.IOSDevice) (driver *ShootsIOSDriver, err error) {
	localShootsPort, err := builtin.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}

	if err = device.Forward(localShootsPort, defaultBightInsightPort); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}

	localServerPort, err := builtin.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = device.Forward(localServerPort, defaultDouyinServerPort); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}

	capabilities := option.NewCapabilities()
	capabilities.WithDefaultAlertAction(option.AlertActionAccept)
	wdaDriver, err := device.NewHTTPDriver(capabilities)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init WDA driver for shoots IOS")
	}

	host := "localhost"
	timeout := 10 * time.Second
	driver = &ShootsIOSDriver{
		WDADriver: wdaDriver.(*uixt.WDADriver),
	}
	driver.bightInsightPrefix = fmt.Sprintf("http://%s:%d", host, localShootsPort)
	driver.serverPrefix = fmt.Sprintf("http://%s:%d", host, localServerPort)
	driver.timeout = timeout

	return driver, nil
}

type ShootsIOSDriver struct {
	*uixt.WDADriver

	bightInsightPrefix string
	serverPrefix       string
	timeout            time.Duration
}

func (s *ShootsIOSDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := s.Session.Request(http.MethodGet, fmt.Sprintf("%s/source?format=json&onlyWeb=false", s.bightInsightPrefix), []byte{})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (s *ShootsIOSDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
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
	resp, err := s.Session.Request(http.MethodPost, fmt.Sprintf("%s/host/login/account/", s.serverPrefix), bsJSON)
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

func (s *ShootsIOSDriver) LogoutNoneUI(packageName string) error {
	resp, err := s.Session.Request(http.MethodGet, fmt.Sprintf("%s/host/loginout/", s.serverPrefix), []byte{})
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

func (s *ShootsIOSDriver) TearDown() error {
	s.WDADriver.TearDown()
	return nil
}

func (s *ShootsIOSDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
	resp, err := s.Session.Request(http.MethodGet, fmt.Sprintf("%s/host/app/info/", s.serverPrefix), []byte{})
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
