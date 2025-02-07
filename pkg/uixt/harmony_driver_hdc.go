package uixt

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"time"

	"code.byted.org/iesqa/ghdc"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

type hdcDriver struct {
	*HarmonyDevice
	*DriverClient
	points   []ExportPoint
	uiDriver *ghdc.UIDriver
}

type PowerStatus string

const (
	POWER_STATUS_SUSPEND PowerStatus = "POWER_STATUS_SUSPEND"
	POWER_STATUS_OFF     PowerStatus = "POWER_STATUS_OFF"
	POWER_STATUS_ON      PowerStatus = "POWER_STATUS_ON"
)

func newHarmonyDriver(device *ghdc.Device) (driver *hdcDriver, err error) {
	driver = new(hdcDriver)
	driver.Device = device
	uiDriver, err := ghdc.NewUIDriver(*device)
	if err != nil {
		log.Error().Err(err).Msg("failed to new harmony ui driver")
		return nil, err
	}
	driver.uiDriver = uiDriver
	driver.NewSession(nil)
	return
}

func (hd *hdcDriver) NewSession(capabilities option.Capabilities) (SessionInfo, error) {
	hd.DriverClient.session.Reset()
	hd.Unlock()
	return SessionInfo{}, errDriverNotImplemented
}

func (hd *hdcDriver) DeleteSession() error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) GetSession() *DriverSession {
	return &hd.DriverClient.session
}

func (hd *hdcDriver) Status() (DeviceStatus, error) {
	return DeviceStatus{}, errDriverNotImplemented
}

func (hd *hdcDriver) DeviceInfo() (DeviceInfo, error) {
	return DeviceInfo{}, errDriverNotImplemented
}

func (hd *hdcDriver) Location() (Location, error) {
	return Location{}, errDriverNotImplemented
}

func (hd *hdcDriver) BatteryInfo() (BatteryInfo, error) {
	return BatteryInfo{}, errDriverNotImplemented
}

func (hd *hdcDriver) WindowSize() (size Size, err error) {
	display, err := hd.uiDriver.GetDisplaySize()
	if err != nil {
		log.Error().Err(err).Msg("failed to get window size")
		return Size{}, err
	}
	size.Width = display.Width
	size.Height = display.Height
	return size, err
}

func (hd *hdcDriver) Screen() (Screen, error) {
	return Screen{}, errDriverNotImplemented
}

func (hd *hdcDriver) Scale() (float64, error) {
	return 1, nil
}

func (hd *hdcDriver) GetTimestamp() (timestamp int64, err error) {
	return 0, errDriverNotImplemented
}

func (hd *hdcDriver) Homescreen() error {
	return hd.uiDriver.PressKey(ghdc.KEYCODE_HOME)
}

func (hd *hdcDriver) Unlock() (err error) {
	// Todo 检查是否锁屏 hdc shell hidumper -s RenderService -a screen
	screenInfo, err := hd.RunShellCommand("hidumper", "-s", "RenderService", "-a", "screen")
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`powerstatus=([\w_]+)`)
	match := re.FindStringSubmatch(screenInfo)
	log.Info().Msg("screen info: " + screenInfo)
	if len(match) <= 1 {
		return fmt.Errorf("failed to unlock; failed to find powerstatus")
	}
	if match[1] == string(POWER_STATUS_SUSPEND) || match[1] == string(POWER_STATUS_OFF) {
		err = hd.uiDriver.PressPowerKey()
		if err != nil {
			return err
		}
	}

	return hd.Swipe(500, 1500, 500, 500)
}

func (hd *hdcDriver) AppLaunch(packageName string) error {
	// Todo
	return errDriverNotImplemented
}

func (hd *hdcDriver) AppTerminate(packageName string) (bool, error) {
	_, err := hd.RunShellCommand("aa", "force-stop", packageName)
	if err != nil {
		log.Error().Err(err).Msg("failed to terminal app")
		return false, err
	}
	return true, nil
}

func (hd *hdcDriver) GetForegroundApp() (app AppInfo, err error) {
	// Todo
	return AppInfo{}, errDriverNotImplemented
}

func (hd *hdcDriver) AssertForegroundApp(packageName string, activityType ...string) error {
	// Todo
	return nil
}

func (hd *hdcDriver) StartCamera() error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) StopCamera() error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) Orientation() (orientation Orientation, err error) {
	return OrientationPortrait, nil
}

func (hd *hdcDriver) Tap(x, y float64, opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)

	if len(actionOptions.Offset) == 2 {
		x += float64(actionOptions.Offset[0])
		y += float64(actionOptions.Offset[1])
	}

	x += actionOptions.GetRandomOffset()
	y += actionOptions.GetRandomOffset()
	if actionOptions.Identifier != "" {
		startTime := int(time.Now().UnixMilli())
		hd.points = append(hd.points, ExportPoint{Start: startTime, End: startTime + 100, Ext: actionOptions.Identifier, RunTime: 100})
	}
	return hd.uiDriver.InjectGesture(ghdc.NewGesture().Start(ghdc.Point{X: int(x), Y: int(y)}).Pause(100))
}

func (hd *hdcDriver) DoubleTap(x, y float64, opts ...option.ActionOption) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) (err error) {
	return errDriverNotImplemented
}

func (hd *hdcDriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	return errDriverNotImplemented
}

// Swipe works like Drag, but `pressForDuration` value is 0
func (hd *hdcDriver) Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)
	if len(actionOptions.Offset) == 4 {
		fromX += float64(actionOptions.Offset[0])
		fromY += float64(actionOptions.Offset[1])
		toX += float64(actionOptions.Offset[2])
		toY += float64(actionOptions.Offset[3])
	}
	fromX += actionOptions.GetRandomOffset()
	fromY += actionOptions.GetRandomOffset()
	toX += actionOptions.GetRandomOffset()
	toY += actionOptions.GetRandomOffset()

	duration := 200
	if actionOptions.PressDuration > 0 {
		duration = int(actionOptions.PressDuration * 1000)
	}
	if actionOptions.Identifier != "" {
		startTime := int(time.Now().UnixMilli())
		hd.points = append(hd.points, ExportPoint{Start: startTime, End: startTime + 100, Ext: actionOptions.Identifier, RunTime: 100})
	}
	return hd.uiDriver.InjectGesture(ghdc.NewGesture().Start(ghdc.Point{X: int(fromX), Y: int(fromY)}).MoveTo(ghdc.Point{X: int(toX), Y: int(toY)}, duration))
}

func (hd *hdcDriver) SetPasteboard(contentType PasteboardType, content string) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error) {
	return nil, errDriverNotImplemented
}

func (hd *hdcDriver) SetIme(ime string) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) SendKeys(text string, opts ...option.ActionOption) error {
	return hd.uiDriver.InputText(text)
}

func (hd *hdcDriver) Input(text string, opts ...option.ActionOption) error {
	return hd.uiDriver.InputText(text)
}

func (hd *hdcDriver) Clear(packageName string) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) PressButton(devBtn DeviceButton) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) PressBack(opts ...option.ActionOption) error {
	return hd.uiDriver.PressBack()
}

func (hd *hdcDriver) Backspace(count int, opts ...option.ActionOption) (err error) {
	return nil
}

func (hd *hdcDriver) PressKeyCode(keyCode KeyCode) (err error) {
	return errDriverNotImplemented
}

func (hd *hdcDriver) PressHarmonyKeyCode(keyCode ghdc.KeyCode) (err error) {
	return hd.uiDriver.PressKey(keyCode)
}

func (hd *hdcDriver) Screenshot() (*bytes.Buffer, error) {
	tempDir := os.TempDir()
	screenshotPath := fmt.Sprintf("%s/screenshot_%d.png", tempDir, time.Now().Unix())
	err := hd.uiDriver.Screenshot(screenshotPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to screenshot")
		return nil, err
	}
	defer func() {
		_ = os.Remove(screenshotPath)
	}()

	raw, err := os.ReadFile(screenshotPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to screenshot")
		return nil, err
	}
	return bytes.NewBuffer(raw), nil
}

func (hd *hdcDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	return "", nil
}

func (hd *hdcDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
	err = errDriverNotImplemented
	return
}

func (hd *hdcDriver) LogoutNoneUI(packageName string) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) TapByText(text string, opts ...option.ActionOption) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) TapByTexts(actions ...TapTextAction) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) AccessibleSource() (string, error) {
	return "", errDriverNotImplemented
}

func (hd *hdcDriver) HealthCheck() error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) GetAppiumSettings() (map[string]interface{}, error) {
	return nil, errDriverNotImplemented
}

func (hd *hdcDriver) SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	return nil, errDriverNotImplemented
}

func (hd *hdcDriver) IsHealthy() (bool, error) {
	return false, errDriverNotImplemented
}

func (hd *hdcDriver) StartCaptureLog(identifier ...string) (err error) {
	return errDriverNotImplemented
}

func (hd *hdcDriver) StopCaptureLog() (result interface{}, err error) {
	// defer clear(hd.points)
	return hd.points, nil
}

func (hd *hdcDriver) GetDriverResults() []*DriverResult {
	return nil
}

func (hd *hdcDriver) RecordScreen(folderPath string, duration time.Duration) (videoPath string, err error) {
	return "", nil
}

func (hd *hdcDriver) TearDown() error {
	return nil
}

func (hd *hdcDriver) Rotation() (rotation Rotation, err error) {
	err = errDriverNotImplemented
	return
}

func (hd *hdcDriver) SetRotation(rotation Rotation) (err error) {
	err = errDriverNotImplemented
	return
}
