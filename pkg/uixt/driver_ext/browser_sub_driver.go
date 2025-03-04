package driver_ext

import (
	"encoding/json"
	"net/http"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/pkg/errors"
)

type StubBrowserDriver struct {
	*uixt.BrowserDriver

	sessionId string
}

func NewStubBrowserDriver(device *uixt.BrowserDevice) (driver *StubBrowserDriver, err error) {
	browserDriver, err := uixt.NewBrowserDriver(device)
	if err != nil {
		return nil, errors.Wrap(err, "create browser session failed")
	}
	driver = &StubBrowserDriver{
		BrowserDriver: browserDriver,
	}
	driver.sessionId = device.UUID()
	return driver, nil
}

func (wd *StubBrowserDriver) GetDriver() uixt.IDriver {
	return wd.BrowserDriver
}

// Source Return application elements tree
func (wd *StubBrowserDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := wd.BrowserDriver.HttpGet(http.MethodGet, wd.sessionId, "stub/source")
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(resp.Data)
	if err != nil {
		return "", err
	}

	return string(jsonData), err
}

func (wd *StubBrowserDriver) LoginNoneUI(packageName, phoneNumber, captcha, password string) (
	info AppLoginInfo, err error) {
	data := map[string]interface{}{
		"url":        packageName,
		"web_cookie": password,
	}
	resp, err := wd.HttpPOST(data, wd.sessionId, "stub/login")
	if err != nil {
		return info, err
	}
	respdata := resp.Data.(map[string]interface{})
	loginSuccss := AppLoginInfo{
		IsLogin: true,
		Uid:     respdata["webid"].(string),
		Did:     password,
	}
	return loginSuccss, err
}

func (wd *StubBrowserDriver) LogoutNoneUI(packageName string) error {
	return errors.New("not implemented")
}
