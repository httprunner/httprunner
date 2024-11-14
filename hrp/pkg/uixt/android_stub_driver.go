package uixt

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

type stubAndroidDriver struct {
	socket  net.Conn
	seq     int
	timeout time.Duration
	adbDriver
}

const StubSocketName = "com.bytest.device"

// newStubAndroidDriver
// 创建stub Driver address为forward后的端口格式127.0.0.1:${port}
func newStubAndroidDriver(address string, urlPrefix string, readTimeout ...time.Duration) (*stubAndroidDriver, error) {
	timeout := 10 * time.Second
	if len(readTimeout) > 0 {
		timeout = readTimeout[0]
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("failed to connect %s", address))
		return nil, err
	}

	driver := &stubAndroidDriver{
		socket:  conn,
		timeout: timeout,
	}

	if driver.urlPrefix, err = url.Parse(urlPrefix); err != nil {
		return nil, err
	}

	driver.NewSession(nil)
	return driver, nil
}

func (sad *stubAndroidDriver) httpGET(pathElem ...string) (rawResp rawResponse, err error) {
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
	sad.client = convertToHTTPClient(conn)
	return sad.httpRequest(http.MethodGet, sad.concatURL(nil, pathElem...), nil)
}

func (sad *stubAndroidDriver) httpPOST(data interface{}, pathElem ...string) (rawResp rawResponse, err error) {
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
	sad.client = convertToHTTPClient(conn)

	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return sad.httpRequest(http.MethodPost, sad.concatURL(nil, pathElem...), bsJSON)
}

func (sad *stubAndroidDriver) NewSession(capabilities Capabilities) (SessionInfo, error) {
	sad.Driver.session.Reset()
	return SessionInfo{}, errDriverNotImplemented
}

func (sad *stubAndroidDriver) sendCommand(packageName string, cmdType string, params map[string]interface{}, readTimeout ...time.Duration) (interface{}, error) {
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

	res, err := sad.adbClient.RunStubCommand(append(data, '\n'), packageName)
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

func (sad *stubAndroidDriver) DeleteSession() error {
	return sad.close()
}

func (sad *stubAndroidDriver) close() error {
	if sad.socket != nil {
		return sad.socket.Close()
	}
	return nil
}

func (sad *stubAndroidDriver) Status() (DeviceStatus, error) {
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

func (sad *stubAndroidDriver) Source(srcOpt ...SourceOption) (source string, err error) {
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

func (sad *stubAndroidDriver) LoginNoneUIBak(packageName, phoneNumber, captcha string) error {
	_, err := sad.adbClient.RunShellCommand(
		"am", "broadcast",
		"-a", fmt.Sprintf("%s.util.crony.action_login", packageName),
		"-e", "phone", phoneNumber,
		"-e", "code", captcha)
	if err != nil {
		return err
	}
	time.Sleep(10 * time.Second)
	login, err := sad.isLogin(packageName)
	if err != nil || !login {
		log.Err(err).Msg("failed to login")
		return fmt.Errorf("failed to login")
	}
	return err
}

func (sad *stubAndroidDriver) LoginNoneUI(packageName, phoneNumber, captcha string) error {
	params := map[string]interface{}{
		"phone": phoneNumber,
		"code":  captcha,
	}
	resp, err := sad.httpPOST(params, "/host", "/login", "account")
	if err != nil {
		return err
	}
	res, err := resp.valueConvertToJsonObject()
	if err != nil {
		return err
	}
	log.Info().Msgf("%v", res)
	if res["isSuccess"] != true {
		err = fmt.Errorf("failed to login %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return err
	}
	time.Sleep(10 * time.Second)
	login, err := sad.isLogin(packageName)
	if err != nil {
		return err
	}
	if !login {
		return fmt.Errorf("failed to login")
	}
	return nil
}

func (sad *stubAndroidDriver) LogoutNoneUI(packageName string) error {
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
		err = fmt.Errorf("failed to logout %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return err
	}
	log.Info().Interface("resp", resp).Msg("logout success")
	return nil
}

func (sad *stubAndroidDriver) LoginNoneUIDynamic(packageName, phoneNumber string, captcha string) error {
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

func (sad *stubAndroidDriver) isLogin(packageName string) (login bool, err error) {
	resp, err := sad.httpGET("/host", "/login", "/check")
	if err != nil {
		return false, err
	}
	res, err := resp.valueConvertToJsonObject()
	if err != nil {
		return false, err
	}
	log.Info().Msgf("%v", res)
	if res["isSuccess"] != true {
		err = fmt.Errorf("failed to check login %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return false, err
	}
	log.Info().Interface("resp", resp).Msg("check login success")
	return true, nil
}
