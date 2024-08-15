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

type ShootsAndroidDriver struct {
	socket  net.Conn
	seq     int
	timeout time.Duration
	adbDriver
}

const ShootsSocketName = "com.bytest.device"

// newShootsAndroidDriver
// 创建shoots Driver address为forward后的端口格式127.0.0.1:${port}
func newShootsAndroidDriver(address string, urlPrefix string, readTimeout ...time.Duration) (*ShootsAndroidDriver, error) {
	timeout := 10 * time.Second
	if len(readTimeout) > 0 {
		timeout = readTimeout[0]
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("failed to connect %s", address))
		return nil, err
	}

	driver := &ShootsAndroidDriver{
		socket:  conn,
		timeout: timeout,
	}

	if driver.urlPrefix, err = url.Parse(urlPrefix); err != nil {
		return nil, err
	}

	return driver, nil
}

func (sad *ShootsAndroidDriver) httpGET(pathElem ...string) (rawResp rawResponse, err error) {
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

func (sad *ShootsAndroidDriver) httpPOST(data interface{}, pathElem ...string) (rawResp rawResponse, err error) {
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

func (sad *ShootsAndroidDriver) NewSession(capabilities Capabilities) (SessionInfo, error) {
	return SessionInfo{}, errDriverNotImplemented
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

	res, err := sad.adbClient.RunShootsCommand(append(data, '\n'), packageName)
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

func (sad *ShootsAndroidDriver) Status() (DeviceStatus, error) {
	app, err := sad.GetForegroundApp()
	if err != nil {
		return DeviceStatus{}, err
	}
	res, err := sad.sendCommand(app.PackageName, "Hello", nil)
	if err != nil {
		return DeviceStatus{}, err
	}
	log.Info().Msg(fmt.Sprintf("pint shoots result :%v", res))
	return DeviceStatus{}, nil
}

func (sad *ShootsAndroidDriver) Source(srcOpt ...SourceOption) (source string, err error) {
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

func (sad *ShootsAndroidDriver) LoginNoneUIBak(packageName, phoneNumber, captcha string) error {
	_, err := sad.adbClient.RunShellCommand("am", "broadcast", "-a", fmt.Sprintf("%s.util.crony.action_login", packageName), "-e", "phone", phoneNumber, "-e", "code", captcha)
	time.Sleep(10 * time.Second)
	login, err := sad.isLogin(packageName)
	if err != nil || !login {
		log.Err(err).Msg("failed to login")
		return fmt.Errorf("failed to login")
	}
	return err
}

func (sad *ShootsAndroidDriver) LoginNoneUI(packageName, phoneNumber, captcha string) error {
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
		err = fmt.Errorf("falied to logout %s", res["data"])
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

func (sad *ShootsAndroidDriver) LogoutNoneUI(packageName string) error {
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

func (sad *ShootsAndroidDriver) isLogin(packageName string) (login bool, err error) {
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
		err = fmt.Errorf("falied to get is login %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return false, err
	}
	fmt.Printf("%v", resp)
	if err != nil {
		return false, err
	}
	return true, nil
}
