package driver_ext

import (
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

const (
	shootsSocketName = "com.bytest.device"
	douyinServerPort = 32316
	forwardToPrefix  = "forward-to-"
)

func NewShootsAndroidDriver(device *uixt.AndroidDevice) (driver *ShootsAndroidDriver, err error) {
	socketLocalPort, err := device.Forward(shootsSocketName)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%s failed: %v",
				socketLocalPort, shootsSocketName, err))
	}
	address := fmt.Sprintf("127.0.0.1:%d", socketLocalPort)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("failed to connect %s", address))
		return nil, err
	}
	driver = &ShootsAndroidDriver{
		socket:  conn,
		timeout: 10 * time.Second,
	}

	driver.InitSession(nil)
	serverLocalPort, err := device.Forward(douyinServerPort)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%d failed: %v",
				serverLocalPort, douyinServerPort, err))
	}
	rawURL := fmt.Sprintf("http://forward-to-%d:%d",
		serverLocalPort, douyinServerPort)
	driver.Session.Init(rawURL)

	driver.Device = device.Device
	driver.Logcat = device.Logcat
	return driver, nil
}

type ShootsAndroidDriver struct {
	*uixt.ADBDriver
	socket  net.Conn
	seq     int
	timeout time.Duration
}

func (sad *ShootsAndroidDriver) sendCommand(packageName string, cmdType string, params map[string]interface{}, readTimeout ...time.Duration) (interface{}, error) {
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
		return nil, fmt.Errorf("failed to call shoots command: %s", resultMap["Error"].(string))
	}

	return resultMap["Result"], nil
}

func (sad *ShootsAndroidDriver) DeleteSession() error {
	return sad.close()
}

func (sad *ShootsAndroidDriver) close() error {
	if sad.socket != nil {
		return sad.socket.Close()
	}
	return nil
}

func (sad *ShootsAndroidDriver) Status() (types.DeviceStatus, error) {
	app, err := sad.GetForegroundApp()
	if err != nil {
		return types.DeviceStatus{}, err
	}
	res, err := sad.sendCommand(app.PackageName, "Hello", nil)
	if err != nil {
		return types.DeviceStatus{}, err
	}
	log.Info().Msg(fmt.Sprintf("ping shoots result :%v", res))
	return types.DeviceStatus{}, nil
}

func (sad *ShootsAndroidDriver) Source(srcOpt ...option.SourceOption) (source string, err error) {
	app, err := sad.GetForegroundApp()
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
		return "", err
	}
	return res.(string), nil
}

func (sad *ShootsAndroidDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
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
	resp, err := sad.Session.POST(params, "/host", "/login", "account")
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

func (sad *ShootsAndroidDriver) LogoutNoneUI(packageName string) error {
	resp, err := sad.Session.GET("/host", "/logout")
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
	fmt.Printf("%v", resp)
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	return nil
}

func (sad *ShootsAndroidDriver) LoginNoneUIDynamic(packageName, phoneNumber string, captcha string) error {
	params := map[string]interface{}{
		"ClassName": "qe.python.test.LoginUtil",
		"Method":    "loginSync",
		"RetType":   "",
		"Args":      []string{phoneNumber, captcha},
	}
	res, err := sad.sendCommand(packageName, "CallStaticMethod", params)
	if err != nil {
		return err
	}
	log.Info().Msg(res.(string))
	return nil
}

func (sad *ShootsAndroidDriver) SetHDTStatus(status bool) error {
	_, err := sad.Device.RunShellCommand("settings", "put", "global", "feedbacker_sso_bypass_token", "default_sso_bypass_token")
	if err != nil {
		log.Warn().Msg(fmt.Sprintf("failed to disable sso, error: %v", err))
	}
	params := map[string]interface{}{
		"ClassName": "com.bytedance.ies.stark.framework.HybridDevTool",
		"Method":    "setEnabled",
		"RetType":   "",
		"Args":      []bool{status},
	}
	res, err := sad.sendCommand("com.ss.android.ugc.aweme", "CallStaticMethod", params)
	if err != nil {
		return fmt.Errorf("failed to set hds status %v, error: %v", status, err)
	}
	log.Info().Msg(fmt.Sprintf("set hdt status result: %s", res))
	return nil
}

func (sad *ShootsAndroidDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
	resp, err := sad.Session.GET("/host", "/app", "/info")
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

type AppLoginInfo struct {
	Did     string `json:"did,omitempty" yaml:"did,omitempty"`
	Uid     string `json:"uid,omitempty" yaml:"uid,omitempty"`
	IsLogin bool   `json:"is_login,omitempty" yaml:"is_login,omitempty"`
}
