package uixt

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

const (
	StubSocketName   = "com.bytest.device"
	DouyinServerPort = 32316
)

func NewStubDriver(device *AndroidDevice) (driver *StubAndroidDriver, err error) {
	socketLocalPort, err := device.Forward(StubSocketName)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%s failed: %v",
				socketLocalPort, StubSocketName, err))
	}

	serverLocalPort, err := device.Forward(DouyinServerPort)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%d failed: %v",
				serverLocalPort, DouyinServerPort, err))
	}

	address := fmt.Sprintf("127.0.0.1:%d", socketLocalPort)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("failed to connect %s", address))
		return nil, err
	}

	driver = &StubAndroidDriver{
		socket:  conn,
		timeout: 10 * time.Second,
	}

	rawURL := fmt.Sprintf("http://forward-to-%d:%d",
		serverLocalPort, DouyinServerPort)
	if driver.urlPrefix, err = url.Parse(rawURL); err != nil {
		return nil, err
	}

	driver.NewSession(nil)
	driver.Device = device.Device
	driver.Logcat = device.Logcat
	return driver, nil
}

type StubAndroidDriver struct {
	*ADBDriver
	socket  net.Conn
	seq     int
	timeout time.Duration
}

type AppLoginInfo struct {
	Did     string `json:"did,omitempty" yaml:"did,omitempty"`
	Uid     string `json:"uid,omitempty" yaml:"uid,omitempty"`
	IsLogin bool   `json:"is_login,omitempty" yaml:"is_login,omitempty"`
}

func (sad *StubAndroidDriver) httpGET(pathElem ...string) (rawResp rawResponse, err error) {
	var localPort int
	{
		tmpURL, _ := url.Parse(sad.urlPrefix.String())
		hostname := tmpURL.Hostname()
		if strings.HasPrefix(hostname, forwardToPrefix) {
			localPort, _ = strconv.Atoi(strings.TrimPrefix(hostname, forwardToPrefix))
		}
	}

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		return nil, fmt.Errorf("adb forward: %w", err)
	}
	sad.Client = convertToHTTPClient(conn)
	return sad.Request(http.MethodGet, sad.concatURL(nil, pathElem...), nil)
}

func (sad *StubAndroidDriver) httpPOST(data interface{}, pathElem ...string) (rawResp rawResponse, err error) {
	var localPort int
	{
		tmpURL, _ := url.Parse(sad.urlPrefix.String())
		hostname := tmpURL.Hostname()
		if strings.HasPrefix(hostname, forwardToPrefix) {
			localPort, _ = strconv.Atoi(strings.TrimPrefix(hostname, forwardToPrefix))
		}
	}

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		return nil, fmt.Errorf("adb forward: %w", err)
	}
	sad.Client = convertToHTTPClient(conn)

	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return sad.Request(http.MethodPost, sad.concatURL(nil, pathElem...), bsJSON)
}

func (sad *StubAndroidDriver) NewSession(capabilities option.Capabilities) (SessionInfo, error) {
	sad.Reset()
	return SessionInfo{}, errDriverNotImplemented
}

func (sad *StubAndroidDriver) sendCommand(packageName string, cmdType string, params map[string]interface{}, readTimeout ...time.Duration) (interface{}, error) {
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

func (sad *StubAndroidDriver) DeleteSession() error {
	return sad.close()
}

func (sad *StubAndroidDriver) close() error {
	if sad.socket != nil {
		return sad.socket.Close()
	}
	return nil
}

func (sad *StubAndroidDriver) Status() (DeviceStatus, error) {
	app, err := sad.GetForegroundApp()
	if err != nil {
		return DeviceStatus{}, err
	}
	res, err := sad.sendCommand(app.PackageName, "Hello", nil)
	if err != nil {
		return DeviceStatus{}, err
	}
	log.Info().Msg(fmt.Sprintf("ping stub result :%v", res))
	return DeviceStatus{}, nil
}

func (sad *StubAndroidDriver) Source(srcOpt ...option.SourceOption) (source string, err error) {
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

func (sad *StubAndroidDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
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
	resp, err := sad.httpPOST(params, "/host", "/login", "account")
	if err != nil {
		return info, err
	}
	res, err := resp.valueConvertToJsonObject()
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
	resp, err := sad.httpGET("/host", "/logout")
	if err != nil {
		return err
	}
	res, err := resp.valueConvertToJsonObject()
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

func (sad *StubAndroidDriver) LoginNoneUIDynamic(packageName, phoneNumber string, captcha string) error {
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

func (sad *StubAndroidDriver) SetHDTStatus(status bool) error {
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

func (sad *StubAndroidDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
	resp, err := sad.httpGET("/host", "/app", "/info")
	if err != nil {
		return info, err
	}
	res, err := resp.valueConvertToJsonObject()
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
