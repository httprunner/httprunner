package driver_ext

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

type StubAndroidDriver struct {
	*uixt.ADBDriver

	seq                 int
	timeout             time.Duration
	douyinUrlPrefix     string
	douyinLiteUrlPrefix string
}

const (
	StubSocketName        = "com.bytest.device"
	AndroidDouyinPort     = 32316
	AndroidDouyinLitePort = 32792
)

type AppLoginInfo struct {
	Did     string `json:"did,omitempty" yaml:"did,omitempty"`
	Uid     string `json:"uid,omitempty" yaml:"uid,omitempty"`
	IsLogin bool   `json:"is_login,omitempty" yaml:"is_login,omitempty"`
}

func NewStubAndroidDriver(dev *uixt.AndroidDevice) (*StubAndroidDriver, error) {
	adbDriver, err := uixt.NewADBDriver(dev)
	if err != nil {
		return nil, err
	}
	driver := &StubAndroidDriver{
		timeout:   10 * time.Second,
		ADBDriver: adbDriver,
	}

	// setup driver
	if err = driver.Setup(); err != nil {
		return nil, err
	}

	return driver, nil
}

func (sad *StubAndroidDriver) GetDriver() uixt.IDriver {
	return sad.ADBDriver
}

func (sad *StubAndroidDriver) Setup() error {
	socketLocalPort, err := sad.Device.Forward(StubSocketName)
	if err != nil {
		return errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%s failed: %v",
				socketLocalPort, StubSocketName, err))
	}

	douyinLocalPort, err := sad.Device.Forward(AndroidDouyinPort)
	if err != nil {
		return errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%d failed: %v",
				douyinLocalPort, AndroidDouyinPort, err))
	}
	sad.douyinUrlPrefix = fmt.Sprintf("http://127.0.0.1:%d", douyinLocalPort)

	douyinLiteLocalPort, err := sad.Device.Forward(AndroidDouyinLitePort)
	if err != nil {
		return errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%d failed: %v",
				douyinLiteLocalPort, AndroidDouyinLitePort, err))
	}
	sad.douyinLiteUrlPrefix = fmt.Sprintf("http://127.0.0.1:%d", douyinLiteLocalPort)

	return nil
}

func (sad *StubAndroidDriver) sendCommand(packageName string, cmdType string, params map[string]interface{}) (
	interface{}, error) {
	sad.seq++
	packet := map[string]interface{}{
		"Seq": sad.seq,
		"Cmd": cmdType,
		"v":   "",
	}
	for key, value := range params {
		if key == "Cmd" || key == "Seq" {
			return "", errors.New("params cannot be Cmd or Seq")
		}
		packet[key] = value
	}
	data, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}

	res, err := sad.Device.RunStubCommand(append(data, '\n'), packageName)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(res), &resultMap); err != nil {
		return nil, err
	}
	if resultMap["Error"] != nil {
		return nil, fmt.Errorf("failed to call stub command: %s", resultMap["Error"].(string))
	}

	return resultMap["Result"], nil
}

func (sad *StubAndroidDriver) AppLaunch(packageName string) (err error) {
	_ = sad.EnableDevtool(packageName, true)
	err = sad.ADBDriver.AppLaunch(packageName)
	if err != nil {
		return err
	}
	return nil
}

func (sad *StubAndroidDriver) Status() (types.DeviceStatus, error) {
	app, err := sad.ForegroundInfo()
	if err != nil {
		return types.DeviceStatus{}, err
	}
	res, err := sad.sendCommand(app.PackageName, "Hello", nil)
	if err != nil {
		return types.DeviceStatus{}, err
	}
	log.Info().Msg(fmt.Sprintf("ping stub result :%v", res))
	return types.DeviceStatus{}, nil
}

func (sad *StubAndroidDriver) Source(srcOpt ...option.SourceOption) (source string, err error) {
	app, err := sad.ForegroundInfo()
	if err != nil {
		return "", err
	}
	params := map[string]interface{}{
		"ClassName": "com.bytedance.byteinsight.MockOperator",
		"Method":    "getLayout",
		"RetType":   "",
		"Args":      []string{},
	}
	res, err := sad.sendCommand(app.PackageName, "CallStaticMethod", params)
	if err != nil {
		if app.PackageName == "com.ss.android.ugc.aweme" {
			log.Error().Err(err).Msg("failed to get source")
		}
		return "", nil
	}
	if res.(string) == "{}" {
		res, err = sad.sendCommand(app.PackageName, "CallStaticMethod", params)
		if err != nil {
			if app.PackageName == "com.ss.android.ugc.aweme" {
				log.Error().Err(err).Msg("failed to get source")
			}
			return "", nil
		}
	}
	return res.(string), nil
}

func (sad *StubAndroidDriver) LoginNoneUI(packageName, phoneNumber, captcha, password string) (
	info AppLoginInfo, err error) {
	app, err := sad.ForegroundInfo()
	if err != nil {
		return info, err
	}
	// app.PackageName in ["com.ss.android.ugc.aweme", "com.ss.android.ugc.aweme.lite"]
	if app.PackageName == "com.ss.android.ugc.aweme" || app.PackageName == "com.ss.android.ugc.aweme.lite" {
		return sad.LoginDouyin(app.PackageName, phoneNumber, captcha, password)
	} else if app.PackageName == "com.ss.android.article.video" {
		return sad.LoginXigua(app.PackageName, phoneNumber, captcha, password)
	} else {
		return info, fmt.Errorf("not support app %s", app.PackageName)
	}
}

func (sad *StubAndroidDriver) LoginXigua(packageName, phoneNumber, captcha, password string) (
	info AppLoginInfo, err error) {
	loginSchema := ""
	if captcha != "" {
		loginSchema = fmt.Sprintf("snssdk32://local_channel_autologin?login_type=1&account=%s&smscode=%s",
			phoneNumber, captcha)
	} else if password != "" {
		loginSchema = fmt.Sprintf("snssdk32://local_channel_autologin?login_type=2&account=%s&password=%s",
			phoneNumber, password)
	} else {
		return info, fmt.Errorf("password and capcha is empty")
	}
	info.IsLogin = true
	return info, sad.OpenUrl(loginSchema)
}

func (sad *StubAndroidDriver) LoginDouyin(packageName, phoneNumber, captcha, password string) (
	info AppLoginInfo, err error) {

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

	info, err = sad.getLoginAppInfo(packageName)
	if err != nil {
		log.Err(err).Msg("failed to get login info")
		return info, err
	}
	if info.Did == "" {
		_ = sad.Home()
		_ = sad.AppLaunch(packageName)
		time.Sleep(20 * time.Second)
	}
	if info.IsLogin {
		_ = sad.LogoutNoneUI(packageName)
	}

	urlPrefix, err := sad.getUrlPrefix(packageName)
	if err != nil {
		return info, err
	}
	fullUrl := urlPrefix + "/host/login/account/"
	resp, err := sad.Session.POST(params, fullUrl)
	if err != nil {
		return info, err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return info, err
	}
	log.Info().Msgf("%v", res)
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to login %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return info, err
	}
	time.Sleep(20 * time.Second)
	info, err = sad.getLoginAppInfo(packageName)
	if err != nil || !info.IsLogin {
		return info, fmt.Errorf("falied to login %v", info)
	}
	return info, nil
}

func (sad *StubAndroidDriver) LogoutNoneUI(packageName string) error {
	urlPrefix, err := sad.getUrlPrefix(packageName)
	if err != nil {
		return err
	}
	fullUrl := urlPrefix + "/host/logout"
	resp, err := sad.Session.GET(fullUrl)
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
	return nil
}

func (sad *StubAndroidDriver) EnableDevtool(packageName string, enable bool) error {
	urlPrefix, err := sad.getUrlPrefix(packageName)
	if err != nil {
		return err
	}
	fullUrl := urlPrefix + "/host/devtool/enable"
	params := map[string]interface{}{
		"enable": enable,
	}
	resp, err := sad.Session.POST(params, fullUrl)
	if err != nil {
		return err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return err
	}
	log.Info().Msgf("%v", res)
	return nil
}

func (sad *StubAndroidDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
	urlPrefix, err := sad.getUrlPrefix(packageName)
	if err != nil {
		return info, err
	}
	fullUrl := urlPrefix + "/host/app/info"
	resp, err := sad.Session.GET(fullUrl)
	if err != nil {
		return info, err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return info, err
	}
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to get app info %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return info, err
	}

	err = json.Unmarshal([]byte(res["data"].(string)), &info)
	if err != nil {
		err = fmt.Errorf("falied to parse app info %s", res["data"])
		return
	}
	return info, nil
}

func (sad *StubAndroidDriver) getUrlPrefix(packageName string) (urlPrefix string, err error) {
	if packageName == "com.ss.android.ugc.aweme" {
		urlPrefix = sad.douyinUrlPrefix
	} else if packageName == "com.ss.android.ugc.aweme.lite" {
		urlPrefix = sad.douyinLiteUrlPrefix
	} else {
		return "", fmt.Errorf("not support app %s", packageName)
	}
	return urlPrefix, nil
}
