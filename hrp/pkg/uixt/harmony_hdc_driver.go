package uixt

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"time"

	"code.byted.org/iesqa/ghdc"
	"github.com/rs/zerolog/log"
)

type hdcDriver struct {
	points []ExportPoint
	Driver
	device   *ghdc.Device
	uiDriver *ghdc.UIDriver
}

type PowerStatus string

const (
	POWER_STATUS_SUSPEND PowerStatus = "POWER_STATUS_SUSPEND"
	POWER_STATUS_ON      PowerStatus = "POWER_STATUS_ON"
)

func newHarmonyDriver(device *ghdc.Device) (driver *hdcDriver, err error) {
	driver = new(hdcDriver)
	driver.device = device
	uiDriver, err := ghdc.NewUIDriver(*device)
	if err != nil {
		log.Error().Err(err).Msg("failed to new harmony ui driver")
		return nil, err
	}
	driver.uiDriver = uiDriver
	driver.NewSession(nil)
	return
}

func (hd *hdcDriver) NewSession(capabilities Capabilities) (SessionInfo, error) {
	hd.Driver.session.Init()
	hd.Unlock()
	return SessionInfo{}, errDriverNotImplemented
}

func (hd *hdcDriver) DeleteSession() error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) GetSession() *DriverSession {
	return &hd.Driver.session
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
	screenInfo, err := hd.device.RunShellCommand("hidumper", "-s", "RenderService", "-a", "screen")
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`powerstatus=([\w_]+)`)
	match := re.FindStringSubmatch(screenInfo)
	if len(match) <= 1 {
		return fmt.Errorf("failed to unlock; failed to find powerstatus")
	}
	if match[1] == string(POWER_STATUS_SUSPEND) {
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
	_, err := hd.device.RunShellCommand("aa", "force-stop", packageName)
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
	return errDriverNotImplemented
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

func (hd *hdcDriver) Tap(x, y int, options ...ActionOption) error {
	return hd.TapFloat(float64(x), float64(y), options...)
}

func (hd *hdcDriver) TapFloat(x, y float64, options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)

	if len(actionOptions.Offset) == 2 {
		x += float64(actionOptions.Offset[0])
		y += float64(actionOptions.Offset[1])
	}

	x += actionOptions.getRandomOffset()
	y += actionOptions.getRandomOffset()
	if actionOptions.Identifier != "" {
		startTime := int(time.Now().UnixMilli())
		hd.points = append(hd.points, ExportPoint{Start: startTime, End: startTime + 100, Ext: actionOptions.Identifier, RunTime: 100})
	}
	return hd.uiDriver.InjectGesture(ghdc.NewGesture().Start(ghdc.Point{X: int(x), Y: int(y)}).Pause(100))
}

func (hd *hdcDriver) DoubleTap(x, y int, options ...ActionOption) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) DoubleTapFloat(x, y float64, options ...ActionOption) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) TouchAndHold(x, y int, second ...float64) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) TouchAndHoldFloat(x, y float64, second ...float64) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) Drag(fromX, fromY, toX, toY int, options ...ActionOption) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) DragFloat(fromX, fromY, toX, toY float64, options ...ActionOption) error {
	return errDriverNotImplemented
}

// Swipe works like Drag, but `pressForDuration` value is 0
func (hd *hdcDriver) Swipe(fromX, fromY, toX, toY int, options ...ActionOption) error {
	return hd.SwipeFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (hd *hdcDriver) SwipeFloat(fromX, fromY, toX, toY float64, options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)
	if len(actionOptions.Offset) == 4 {
		fromX += float64(actionOptions.Offset[0])
		fromY += float64(actionOptions.Offset[1])
		toX += float64(actionOptions.Offset[2])
		toY += float64(actionOptions.Offset[3])
	}
	fromX += actionOptions.getRandomOffset()
	fromY += actionOptions.getRandomOffset()
	toX += actionOptions.getRandomOffset()
	toY += actionOptions.getRandomOffset()

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

func (hd *hdcDriver) SendKeys(text string, options ...ActionOption) error {
	return hd.uiDriver.InputText(text)
}

func (hd *hdcDriver) Input(text string, options ...ActionOption) error {
	return hd.uiDriver.InputText(text)
}

func (hd *hdcDriver) Clear(packageName string) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) PressButton(devBtn DeviceButton) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) PressBack(options ...ActionOption) error {
	return hd.uiDriver.PressBack()
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

func (hd *hdcDriver) Source(srcOpt ...SourceOption) (string, error) {
	return "", nil
}

func (hd *hdcDriver) LoginNoneUI(packageName, phoneNumber string, captcha string) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) LogoutNoneUI(packageName string) error {
	return errDriverNotImplemented
}

func (hd *hdcDriver) TapByText(text string, options ...ActionOption) error {
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
