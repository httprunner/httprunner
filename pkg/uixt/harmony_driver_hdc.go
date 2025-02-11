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
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

func NewHDCDriver(device *HarmonyDevice) (*HDCDriver, error) {
	driver := &HDCDriver{
		Device: device,
	}
	driver.InitSession(nil)

	uiDriver, err := ghdc.NewUIDriver(*device.Device)
	if err != nil {
		log.Error().Err(err).Msg("failed to new harmony ui driver")
		return nil, err
	}
	driver.uiDriver = uiDriver

	return driver, nil
}

type HDCDriver struct {
	Device  *HarmonyDevice
	Session *Session

	points   []ExportPoint
	uiDriver *ghdc.UIDriver
}

func (hd *HDCDriver) InitSession(capabilities option.Capabilities) error {
	hd.Session = &Session{}
	hd.Session.Reset()
	hd.Unlock()
	return nil
}

func (hd *HDCDriver) DeleteSession() error {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) GetSession() *Session {
	return hd.Session
}

func (hd *HDCDriver) Status() (types.DeviceStatus, error) {
	return types.DeviceStatus{}, types.ErrDriverNotImplemented
}

func (hd *HDCDriver) GetDevice() IDevice {
	return hd.Device
}

func (hd *HDCDriver) DeviceInfo() (types.DeviceInfo, error) {
	return types.DeviceInfo{}, types.ErrDriverNotImplemented
}

func (hd *HDCDriver) BatteryInfo() (types.BatteryInfo, error) {
	return types.BatteryInfo{}, types.ErrDriverNotImplemented
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

func (hd *HDCDriver) Scale() (float64, error) {
	return 1, nil
}

func (hd *HDCDriver) Homescreen() error {
	return hd.uiDriver.PressKey(ghdc.KEYCODE_HOME)
}

type PowerStatus string

const (
	POWER_STATUS_SUSPEND PowerStatus = "POWER_STATUS_SUSPEND"
	POWER_STATUS_OFF     PowerStatus = "POWER_STATUS_OFF"
	POWER_STATUS_ON      PowerStatus = "POWER_STATUS_ON"
)

func (hd *HDCDriver) Unlock() (err error) {
	// Todo 检查是否锁屏 hdc shell hidumper -s RenderService -a screen
	screenInfo, err := hd.Device.RunShellCommand("hidumper", "-s", "RenderService", "-a", "screen")
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
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) AppTerminate(packageName string) (bool, error) {
	_, err := hd.Device.RunShellCommand("aa", "force-stop", packageName)
	if err != nil {
		log.Error().Err(err).Msg("failed to terminal app")
		return false, err
	}
	return true, nil
}

func (hd *HDCDriver) GetForegroundApp() (app types.AppInfo, err error) {
	// Todo
	return types.AppInfo{}, types.ErrDriverNotImplemented
}

func (hd *HDCDriver) AssertForegroundApp(packageName string, activityType ...string) error {
	// Todo
	return nil
}

func (hd *HDCDriver) Orientation() (orientation types.Orientation, err error) {
	return types.OrientationPortrait, nil
}

func (hd *HDCDriver) TapXY(x, y float64, opts ...option.ActionOption) error {
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

func (hd *HDCDriver) DoubleTapXY(x, y float64, opts ...option.ActionOption) error {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) (err error) {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	return types.ErrDriverNotImplemented
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

func (hd *HDCDriver) SetIme(ime string) error {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) SendKeys(text string, opts ...option.ActionOption) error {
	return hd.uiDriver.InputText(text)
}

func (hd *HDCDriver) Input(text string, opts ...option.ActionOption) error {
	return hd.uiDriver.InputText(text)
}

func (hd *HDCDriver) AppClear(packageName string) error {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) PressButton(devBtn types.DeviceButton) error {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) PressBack(opts ...option.ActionOption) error {
	return hd.uiDriver.PressBack()
}

func (hd *HDCDriver) Backspace(count int, opts ...option.ActionOption) (err error) {
	return nil
}

func (hd *HDCDriver) PressKeyCode(keyCode KeyCode) (err error) {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) PressHarmonyKeyCode(keyCode ghdc.KeyCode) (err error) {
	return hd.uiDriver.PressKey(keyCode)
}

func (hd *HDCDriver) ScreenShot() (*bytes.Buffer, error) {
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

func (hd *HDCDriver) TapByText(text string, opts ...option.ActionOption) error {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) TapByTexts(actions ...TapTextAction) error {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) StartCaptureLog(identifier ...string) (err error) {
	return types.ErrDriverNotImplemented
}

func (hd *HDCDriver) StopCaptureLog() (result interface{}, err error) {
	// defer clear(hd.points)
	return hd.points, nil
}

func (hd *HDCDriver) ScreenRecord(duration time.Duration) (videoPath string, err error) {
	return "", nil
}

func (hd *HDCDriver) Setup() error {
	return nil
}

func (hd *HDCDriver) TearDown() error {
	return nil
}

func (hd *HDCDriver) Rotation() (rotation types.Rotation, err error) {
	err = types.ErrDriverNotImplemented
	return
}

func (hd *HDCDriver) SetRotation(rotation types.Rotation) (err error) {
	err = types.ErrDriverNotImplemented
	return
}
