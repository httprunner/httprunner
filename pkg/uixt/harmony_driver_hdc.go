package uixt

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"time"

	"code.byted.org/iesqa/ghdc"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

type HDCDriver struct {
	*HarmonyDevice
	*Session
	IDriver
	*DriverExt
	points   []ExportPoint
	uiDriver *ghdc.UIDriver
}

type PowerStatus string

const (
	POWER_STATUS_SUSPEND PowerStatus = "POWER_STATUS_SUSPEND"
	POWER_STATUS_OFF     PowerStatus = "POWER_STATUS_OFF"
	POWER_STATUS_ON      PowerStatus = "POWER_STATUS_ON"
)

func NewHDCDriver(device *HarmonyDevice) (driver *HDCDriver, err error) {
	driver = new(HDCDriver)
	driver.HarmonyDevice = device
	uiDriver, err := ghdc.NewUIDriver(*device.Device)
	if err != nil {
		log.Error().Err(err).Msg("failed to new harmony ui driver")
		return nil, err
	}
	driver.uiDriver = uiDriver
	driver.NewSession(nil)
	return
}

func (hd *HDCDriver) NewSession(capabilities option.Capabilities) (Session, error) {
	hd.Reset()
	hd.Unlock()
	return Session{}, errDriverNotImplemented
}

func (hd *HDCDriver) DeleteSession() error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) GetSession() *Session {
	return hd.Session
}

func (hd *HDCDriver) Status() (types.DeviceStatus, error) {
	return types.DeviceStatus{}, errDriverNotImplemented
}

func (hd *HDCDriver) GetDevice() IDevice {
	return hd.HarmonyDevice
}

func (hd *HDCDriver) DeviceInfo() (types.DeviceInfo, error) {
	return types.DeviceInfo{}, errDriverNotImplemented
}

func (hd *HDCDriver) Location() (types.Location, error) {
	return types.Location{}, errDriverNotImplemented
}

func (hd *HDCDriver) BatteryInfo() (types.BatteryInfo, error) {
	return types.BatteryInfo{}, errDriverNotImplemented
}

func (hd *HDCDriver) WindowSize() (size types.Size, err error) {
	display, err := hd.uiDriver.GetDisplaySize()
	if err != nil {
		log.Error().Err(err).Msg("failed to get window size")
		return types.Size{}, err
	}
	size.Width = display.Width
	size.Height = display.Height
	return size, err
}

func (hd *HDCDriver) Screen() (ai.Screen, error) {
	return ai.Screen{}, errDriverNotImplemented
}

func (hd *HDCDriver) Scale() (float64, error) {
	return 1, nil
}

func (hd *HDCDriver) Homescreen() error {
	return hd.uiDriver.PressKey(ghdc.KEYCODE_HOME)
}

func (hd *HDCDriver) Unlock() (err error) {
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

func (hd *HDCDriver) AppLaunch(packageName string) error {
	// Todo
	return errDriverNotImplemented
}

func (hd *HDCDriver) AppTerminate(packageName string) (bool, error) {
	_, err := hd.RunShellCommand("aa", "force-stop", packageName)
	if err != nil {
		log.Error().Err(err).Msg("failed to terminal app")
		return false, err
	}
	return true, nil
}

func (hd *HDCDriver) GetForegroundApp() (app types.AppInfo, err error) {
	// Todo
	return types.AppInfo{}, errDriverNotImplemented
}

func (hd *HDCDriver) AssertForegroundApp(packageName string, activityType ...string) error {
	// Todo
	return nil
}

func (hd *HDCDriver) StartCamera() error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) StopCamera() error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) Orientation() (orientation types.Orientation, err error) {
	return types.OrientationPortrait, nil
}

func (hd *HDCDriver) Tap(x, y float64, opts ...option.ActionOption) error {
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

func (hd *HDCDriver) DoubleTap(x, y float64, opts ...option.ActionOption) error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) (err error) {
	return errDriverNotImplemented
}

func (hd *HDCDriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	return errDriverNotImplemented
}

// Swipe works like Drag, but `pressForDuration` value is 0
func (hd *HDCDriver) Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
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

func (hd *HDCDriver) SetPasteboard(contentType types.PasteboardType, content string) error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) GetPasteboard(contentType types.PasteboardType) (raw *bytes.Buffer, err error) {
	return nil, errDriverNotImplemented
}

func (hd *HDCDriver) SetIme(ime string) error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) SendKeys(text string, opts ...option.ActionOption) error {
	return hd.uiDriver.InputText(text)
}

func (hd *HDCDriver) Input(text string, opts ...option.ActionOption) error {
	return hd.uiDriver.InputText(text)
}

func (hd *HDCDriver) Clear(packageName string) error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) PressButton(devBtn types.DeviceButton) error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) PressBack(opts ...option.ActionOption) error {
	return hd.uiDriver.PressBack()
}

func (hd *HDCDriver) Backspace(count int, opts ...option.ActionOption) (err error) {
	return nil
}

func (hd *HDCDriver) PressKeyCode(keyCode KeyCode) (err error) {
	return errDriverNotImplemented
}

func (hd *HDCDriver) PressHarmonyKeyCode(keyCode ghdc.KeyCode) (err error) {
	return hd.uiDriver.PressKey(keyCode)
}

func (hd *HDCDriver) Screenshot() (*bytes.Buffer, error) {
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

func (hd *HDCDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	return "", nil
}

func (hd *HDCDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
	err = errDriverNotImplemented
	return
}

func (hd *HDCDriver) LogoutNoneUI(packageName string) error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) TapByText(text string, opts ...option.ActionOption) error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) TapByTexts(actions ...TapTextAction) error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) AccessibleSource() (string, error) {
	return "", errDriverNotImplemented
}

func (hd *HDCDriver) HealthCheck() error {
	return errDriverNotImplemented
}

func (hd *HDCDriver) GetAppiumSettings() (map[string]interface{}, error) {
	return nil, errDriverNotImplemented
}

func (hd *HDCDriver) SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	return nil, errDriverNotImplemented
}

func (hd *HDCDriver) IsHealthy() (bool, error) {
	return false, errDriverNotImplemented
}

func (hd *HDCDriver) StartCaptureLog(identifier ...string) (err error) {
	return errDriverNotImplemented
}

func (hd *HDCDriver) StopCaptureLog() (result interface{}, err error) {
	// defer clear(hd.points)
	return hd.points, nil
}

func (hd *HDCDriver) GetDriverResults() []*DriverRequests {
	return nil
}

func (hd *HDCDriver) RecordScreen(folderPath string, duration time.Duration) (videoPath string, err error) {
	return "", nil
}

func (hd *HDCDriver) Setup() error {
	return nil
}

func (hd *HDCDriver) TearDown() error {
	return nil
}

func (hd *HDCDriver) Rotation() (rotation types.Rotation, err error) {
	err = errDriverNotImplemented
	return
}

func (hd *HDCDriver) SetRotation(rotation types.Rotation) (err error) {
	err = errDriverNotImplemented
	return
}
