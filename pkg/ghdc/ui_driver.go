package ghdc

import (
	"embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed agent.so
var agentSO embed.FS

type UitestRequest struct {
	Module    string      `json:"module,omitempty"`
	Method    string      `json:"method,omitempty"`
	Params    interface{} `json:"params,omitempty"`
	RequestId string      `json:"request_id,omitempty"`
}

type UitestResponse struct {
	Result    interface{}      `json:"result,omitempty"`
	Exception *UitestException `json:"exception,omitempty"`
}

type UitestKitResponse struct {
	Result    interface{} `json:"result,omitempty"`
	Exception string      `json:"exception,omitempty"`
}

type UitestException struct {
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

type UIDriver struct {
	Device
	uTp  *uitestTransport
	uKTp *uitestKitTransport
}

type Dimension struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

func NewUIDriver(device Device) (d *UIDriver, err error) {
	d = new(UIDriver)
	d.Device = device
	err = d.prepareDevice()
	if err != nil {
		err = fmt.Errorf("[uitest] failed to prepare device \n%v", err)
		return
	}
	uTp, err := device.createUitestTransport()
	if err != nil {
		err = fmt.Errorf("[uitest] failed to create uitest transport \n%v", err)
		return
	}
	uKTp, err := device.createUitestKitTransport()
	if err != nil {
		err = fmt.Errorf("[uitest] failed to create uitest kit transport \n%v", err)
		return
	}
	d.uTp = &uTp
	d.uKTp = &uKTp
	return
}

func (d *UIDriver) Close() {
	if d.uTp != nil {
		d.uTp.Close()
	}
	if d.uKTp != nil {
		d.uKTp.Close()
	}
}

func (d *UIDriver) Reconnect() error {
	d.Close()
	uTp, err := d.createUitestTransport()
	if err != nil {
		err = fmt.Errorf("[uitest] failed to create uitest transport \n%v", err)
		return err
	}
	uKTp, err := d.createUitestKitTransport()
	if err != nil {
		err = fmt.Errorf("[uitest] failed to create uitest kit transport \n%v", err)
		return err
	}
	d.uTp = &uTp
	d.uKTp = &uKTp
	return nil
}

func (d *UIDriver) prepareDevice() error {
	uitestPid, err := d.Device.RunShellCommand("pidof uitest")
	if err != nil {
		return err
	}
	uitestPid = strings.TrimSpace(uitestPid)

	isLowerVersion, err := d.needUpdateLib()
	if err != nil {
		return err
	}

	if uitestPid != "" && !isLowerVersion {
		return nil
	}

	_, err = d.Device.RunShellCommand("param set persist.ace.testmode.enabled 1")
	if err != nil {
		return err
	}

	if isLowerVersion {
		if uitestPid != "" {
			_, err = d.Device.RunShellCommand("kill -9 " + uitestPid)
			if err != nil {
				return err
			}
			uitestPid = ""
		}

		err = d.updateLib()
		if err != nil {
			return err
		}
	}
	if uitestPid == "" {
		_, err = d.Device.RunShellCommand("uitest start-daemon singleness")
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *UIDriver) isServerRunning() bool {
	res, err := d.Device.RunShellCommand("top -H -n 1 -p $(pidof uitest)")
	if err != nil {
		return false
	}
	if strings.Contains(res, "rpc-") {
		return true
	}
	return false
}

func (d *UIDriver) updateLib() error {
	tmpDir := os.TempDir()
	soFileName := filepath.Join(tmpDir, "agent.so")
	soRaw, err := agentSO.ReadFile("agent.so")
	if err != nil {
		return err
	}
	err = os.WriteFile(soFileName, soRaw, os.ModePerm)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return err
	}

	_, err = d.Device.RunShellCommand("rm /data/tmp/local/agent.so")
	if err != nil {
		return err
	}
	err = d.Device.PushFile(soFileName, "/data/local/tmp/agent.so")
	if err != nil {
		return err
	}
	return nil
}

func (d *UIDriver) needUpdateLib() (res bool, err error) {
	deviceVersionStr, err := d.Device.RunShellCommand("cat data/local/tmp/agent.so |grep -a UITEST_AGENT_LIBRARY")
	if err != nil {
		return false, err
	}
	soRaw, err := agentSO.ReadFile("agent.so")
	if err != nil {
		return false, err
	}
	// 定义要搜索的字符串
	searchString := "UITEST_AGENT_LIBRARY"

	// 将二进制内容转换为字符串
	content := string(soRaw)

	// 按行分割内容
	lines := strings.Split(content, "\n")

	// 搜索包含特定字符串的行
	for _, line := range lines {
		if strings.Contains(line, searchString) {
			update := false
			deviceVersion, err := getVersion(deviceVersionStr)
			if err != nil {
				update = true
			}
			lowestVersion, err := getVersion(line)
			if err != nil {
				return false, err
			}
			if update || lowestVersion[0] > deviceVersion[0] || lowestVersion[1] > deviceVersion[1] || lowestVersion[2] > deviceVersion[2] {
				return true, nil
			}
			return false, err
		}
	}
	return false, err
}

func (d *UIDriver) supportDevice() error {
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(filepath.Join(rootDir, "minUiTestVersion.txt"))
	if err != nil {
		return err
	}
	lowestVersion := string(raw)
	uitestVersion, err := d.Device.RunShellCommand("uitest --version")
	if lowestVersion > uitestVersion {
		return fmt.Errorf("not supprt uitest lowest version %s, device version %s", lowestVersion, uitestVersion)
	}
	return nil
}

func (d *UIDriver) createDriver() (driver string, err error) {
	params := map[string]interface{}{
		"api":          "Driver.create",
		"this":         nil,
		"args":         []string{},
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		err = fmt.Errorf("[uitest] failed to create driver")
		return
	}
	if res.Exception != nil {
		err = fmt.Errorf("[uitest] failed to create driver msg: %s", res.Exception.Message)
		return
	}
	driver = res.Result.(string)
	return
}

func (d *UIDriver) createFocused() (onName string, err error) {
	params := map[string]interface{}{
		"api":          "On.focused",
		"this":         "On#seed",
		"args":         []bool{true},
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		err = fmt.Errorf("[uitest] failed to create focused")
		return
	}
	if res.Exception != nil {
		err = fmt.Errorf("[uitest] failed to create focused msg: %s", res.Exception.Message)
		return
	}
	onName = res.Result.(string)
	return
}

func (d *UIDriver) findComponent(driverName, onName string) (componentName string, err error) {
	params := map[string]interface{}{
		"api":          "Driver.waitForComponent",
		"this":         driverName,
		"args":         []any{onName, 5000},
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		err = fmt.Errorf("[uitest] failed to create focused")
		return
	}
	if res.Exception != nil {
		err = fmt.Errorf("[uitest] failed to create focused msg: %s", res.Exception.Message)
		return
	}
	componentName = res.Result.(string)
	return
}

func (d *UIDriver) createPointMatrix(pointerMatrix *PointerMatrix) (pointMatrixName string, err error) {
	fingers, steps := pointerMatrix.fingerIndexStats()
	params := map[string]interface{}{
		"api":          "PointerMatrix.create",
		"this":         nil,
		"args":         []int{fingers, steps},
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		err = fmt.Errorf("[uitest] failed to create PointerMatrix")
		return
	}
	if res.Exception != nil {
		err = fmt.Errorf("[uitest] failed to create PointerMatrix msg: %s", res.Exception.Message)
		return
	}
	pointMatrixName = res.Result.(string)
	return
}

func (d *UIDriver) releaseObj(obj []string) error {
	params := map[string]interface{}{
		"api":  "BackendObjectsCleaner",
		"this": nil,
		"args": obj,
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		err = fmt.Errorf("[uitest] failed to release driver")
		return err
	}
	if res.Exception != nil {
		err = fmt.Errorf("[uitest] failed to release driver msg: %s", res.Exception.Message)
		return err
	}
	return nil
}

func (d *UIDriver) Touch(x, y int) error {
	driverName, err := d.createDriver()
	if err != nil {
		return err
	}

	defer func() {
		_ = d.releaseObj([]string{driverName})
	}()

	params := map[string]interface{}{
		"api":          "Driver.click",
		"this":         driverName,
		"args":         []int{x, y},
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		return err
	}
	if res.Exception != nil {
		return fmt.Errorf("[uitest] failed to touch (%d, %d): %s", x, y, res.Exception.Message)
	}
	return nil
}

func (d *UIDriver) Drag(fromX, fromY, toX, toY int, duration float64) error {
	driverName, err := d.createDriver()
	if err != nil {
		return err
	}

	defer func() {
		_ = d.releaseObj([]string{driverName})
	}()

	distance := math.Sqrt(math.Pow(float64(fromX-toX), 2) + math.Pow(float64(fromX-toX), 2))
	speed := int(distance / duration)

	params := map[string]interface{}{
		"api":          "Driver.drag",
		"this":         driverName,
		"args":         []int{fromX, fromY, toX, toY, speed},
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		return err
	}
	if res.Exception != nil {
		return fmt.Errorf("[uitest] failed to Drag from (%d, %d) to (%d, %d): %s", fromX, fromY, toX, toY, res.Exception.Message)
	}
	return nil
}

func (d *UIDriver) PressKey(key KeyCode) error {
	driverName, err := d.createDriver()
	if err != nil {
		return err
	}

	defer func() {
		_ = d.releaseObj([]string{driverName})
	}()

	params := map[string]interface{}{
		"api":          "Driver.triggerKey",
		"this":         driverName,
		"args":         []KeyCode{key},
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		return err
	}
	if res.Exception != nil {
		return fmt.Errorf("[uitest] failed to Press Key code:%d: %s", key, res.Exception.Message)
	}
	return nil
}

func (d *UIDriver) PressKeys(keys []KeyCode) error {
	driverName, err := d.createDriver()
	if err != nil {
		return err
	}

	defer func() {
		_ = d.releaseObj([]string{driverName})
	}()

	params := map[string]interface{}{
		"api":          "Driver.triggerCombineKeys",
		"this":         driverName,
		"args":         keys,
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		return err
	}
	if res.Exception != nil {
		return fmt.Errorf("[uitest] failed to Press Key code:%v: %s", keys, res.Exception.Message)
	}
	return nil
}

func (d *UIDriver) InjectGesture(gesture *Gesture, speedArg ...int) error {
	return d.InjectMultiGesture([]*Gesture{gesture}, speedArg...)
}

func (d *UIDriver) InjectMultiGesture(gestures []*Gesture, speedArg ...int) error {
	speed := 2000
	var releaseObj []string
	defer func() {
		if len(releaseObj) > 0 {
			_ = d.releaseObj(releaseObj)
		}
	}()
	if len(speedArg) > 0 && speedArg[0] > 0 {
		speed = speedArg[0]
	}
	driverName, err := d.createDriver()
	if err != nil {
		return err
	}
	releaseObj = append(releaseObj, driverName)

	pointerMatrix := d.gestureToPointMatrix(gestures)

	pointerMatrixName, err := d.createPointMatrix(pointerMatrix)
	if err != nil {
		return err
	}
	releaseObj = append(releaseObj, pointerMatrixName)

	for step, point := range pointerMatrix.points {
		err = d.setPoint(pointerMatrixName, point.index, step, point.point)
		if err != nil {
			return err
		}
	}

	err = d.injectMultiPointerAction(driverName, pointerMatrixName, speed)
	if err != nil {
		return err
	}

	return nil
}

func (d *UIDriver) InputText(text string) error {
	params := map[string]interface{}{
		"api":  "Driver.inputText",
		"this": nil,
		"args": []any{map[string]interface{}{
			"x": 0,
			"y": 0,
		}, text},
		"message_type": "hypium",
	}

	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		return err
	}
	if res.Exception != nil {
		return fmt.Errorf("[uitest] failed to input text %s: %s", text, res.Exception.Message)
	}
	return nil
}

func (d *UIDriver) InputTextOnFocused(text string) error {
	driverName, err := d.createDriver()
	if err != nil {
		return err
	}
	defer func() {
		_ = d.releaseObj([]string{driverName})
	}()

	onName, err := d.createFocused()
	if err != nil {
		return err
	}
	defer func() {
		_ = d.releaseObj([]string{onName})
	}()
	componentName, err := d.findComponent(driverName, onName)
	if err != nil {
		return err
	}
	defer func() {
		_ = d.releaseObj([]string{componentName})
	}()
	params := map[string]interface{}{
		"api":          "Component.inputText",
		"this":         componentName,
		"args":         []string{text},
		"message_type": "hypium",
	}

	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		return err
	}
	if res.Exception != nil {
		return fmt.Errorf("[uitest] failed to input text %s: %s", text, res.Exception.Message)
	}
	return nil
}

func (d *UIDriver) TouchDown(x, y int) error {
	params := map[string]interface{}{
		"api":  "touchDown",
		"this": nil,
		"args": map[string]interface{}{
			"x": x,
			"y": y,
		},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "Gestures", DEFAULT, nil)
}

func (d *UIDriver) TouchDownAsync(x, y int) {
	go func() {
		err := d.TouchDown(x, y)
		if err != nil {
			debugLog(fmt.Sprintf("%v", err))
		}
	}()
}

func (d *UIDriver) TouchMove(x, y int) error {
	params := map[string]interface{}{
		"api":  "touchMove",
		"this": nil,
		"args": map[string]interface{}{
			"x": x,
			"y": y,
		},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "Gestures", DEFAULT, nil)
}

func (d *UIDriver) TouchMoveAsync(x, y int) {
	go func() {
		err := d.TouchMove(x, y)
		if err != nil {
			debugLog(fmt.Sprintf("%v", err))
		}
	}()
}

func (d *UIDriver) TouchUp(x, y int) error {
	params := map[string]interface{}{
		"api":  "touchUp",
		"this": nil,
		"args": map[string]interface{}{
			"x": x,
			"y": y,
		},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "Gestures", DEFAULT, nil)
}

func (d *UIDriver) TouchUpAsync(x, y int) {
	go func() {
		err := d.TouchUp(x, y)
		if err != nil {
			debugLog(fmt.Sprintf("%v", err))
		}
	}()
}

func (d *UIDriver) PressRecentApp() error {
	params := map[string]interface{}{
		"api":          "pressRecentApp",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "Gestures", DEFAULT, nil)
}

func (d *UIDriver) PressBack() error {
	params := map[string]interface{}{
		"api":          "pressBack",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "Gestures", DEFAULT, nil)
}

func (d *UIDriver) PressPowerKey() error {
	params := map[string]interface{}{
		"api":          "pressPowerKey",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "CtrlCmd", DEFAULT, nil)
}

func (d *UIDriver) GetDisplayRotation() (direction int, err error) {
	params := map[string]interface{}{
		"api":          "getDisplayRotation",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	res, err := d.sendUitestKit(params, "CtrlCmd", DEFAULT, nil)
	if err != nil {
		return
	}
	if res.Result == false {
		err = fmt.Errorf("[uitest] failed to exec method getDisplayRotation msg: %s", res.Exception)
		return
	}
	direction = (int)(res.Result.(float64))
	return direction, err
}

func (d *UIDriver) UpVolume() error {
	params := map[string]interface{}{
		"api":          "upVolume",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "CtrlCmd", DEFAULT, nil)
}

func (d *UIDriver) DownVolume() error {
	params := map[string]interface{}{
		"api":          "downVolume",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "CtrlCmd", DEFAULT, nil)
}

func (d *UIDriver) RotationDisplay(direction int) error {
	params := map[string]interface{}{
		"api":  "rotationDisplay",
		"this": nil,
		"args": map[string]interface{}{
			"direction": direction,
		},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "CtrlCmd", DEFAULT, nil)
}

func (d *UIDriver) GetDisplaySize() (display Dimension, err error) {
	params := map[string]interface{}{
		"api":          "getDisplaySize",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	res, err := d.sendUitestKit(params, "CtrlCmd", DEFAULT, nil)
	if err != nil {
		return
	}
	if res.Result == false {
		err = fmt.Errorf("[uitest] failed to exec method getDisplaySize msg: %s", res.Exception)
		return
	}
	raw, err := json.Marshal(res.Result)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw, &display)
	return
}

func (d *UIDriver) StartCaptureScreen(callback UitestKitCallback) error {
	params := map[string]interface{}{
		"api":          "startCaptureScreen",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "Captures", SCREEN_CAPTURE, callback)
}

func (d *UIDriver) StopCaptureScreen() error {
	params := map[string]interface{}{
		"api":          "stopCaptureScreen",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "Captures", DEFAULT, nil)
}

func (d *UIDriver) StartCaptureUiAction(callback UitestKitCallback) error {
	params := map[string]interface{}{
		"api":          "startCaptureUiAction",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "Captures", UI_ACTION_CAPTURE, callback)
}

func (d *UIDriver) StopCaptureUiAction() error {
	params := map[string]interface{}{
		"api":          "stopCaptureUiAction",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	return d.sendUitestKitNoResult(params, "Captures", DEFAULT, nil)
}

func (d *UIDriver) CaptureLayout() (layout interface{}, err error) {
	params := map[string]interface{}{
		"api":          "captureLayout",
		"this":         nil,
		"args":         map[string]interface{}{},
		"message_type": "hypium",
	}
	res, err := d.sendUitestKit(params, "Captures", DEFAULT, nil)
	if err != nil {
		return
	}
	if res.Result == false {
		err = fmt.Errorf("[uitest] failed to exec method captureLayout msg: %s", res.Exception)
		return
	}
	return res.Result, err
}

func (d *UIDriver) sendUitestKitNoResult(params map[string]interface{}, method string, reqType ReqTypeEnum, callback UitestKitCallback) error {
	res, err := d.sendUitestKit(params, method, reqType, callback)
	if err != nil {
		return err
	}
	if res.Result == false {
		err = fmt.Errorf("[uitest] failed to exec method %s params %v msg: %s", method, params, res.Exception)
		return err
	}

	return nil
}

func (d *UIDriver) sendUitestReq(req UitestRequest) (res UitestResponse, err error) {
	res, err = d.uTp.SendReq(req)
	if err != nil {
		fmt.Printf("[uitest] failed to send req first. try reconnect \n%v \n", err)
		if err = d.Reconnect(); err != nil {
			return
		}
		res, err = d.uTp.SendReq(req)
	}
	return
}

func (d *UIDriver) sendUitestKit(params map[string]interface{}, method string, reqType ReqTypeEnum, callback UitestKitCallback) (response UitestKitResponse, err error) {
	request := newHypiumRequest(params, method)
	requestStr, err := request.ToString()
	if err != nil {
		err = fmt.Errorf("[uitest] failed to create req while exec method %s %v", method, err)
		return
	}
	sessionId := hashCode(fmt.Sprintf("%s%d", requestStr, time.Now().Unix()))
	if sessionId <= (1 << 24) {
		sessionId += 1 << 24
	}
	err = d.uKTp.registerCallback(reqType, sessionId, nil)
	if err != nil {
		fmt.Printf("[uitest] failed to register callback try reconnect %s %v", method, err)
		if err = d.Reconnect(); err != nil {
			return
		}
		if err = d.uKTp.registerCallback(reqType, sessionId, nil); err != nil {
			return
		}
	}
	res, err := d.uKTp.sendMessage(reqType, sessionId, requestStr)
	if err != nil {
		err = fmt.Errorf("[uitest] failed to send message while exec method %s sessionId: %d %v", method, sessionId, err)
		return
	}
	err = d.uKTp.registerCallback(reqType, sessionId, callback)
	if err != nil {
		err = fmt.Errorf("[uitest] failed to register callback while exec method %s %v", method, err)
		return
	}
	return res, nil
}

func (d *UIDriver) gestureToPointMatrix(gestures []*Gesture) *PointerMatrix {
	pointerMatrix := &PointerMatrix{}
	for fingerIndex, gestures := range gestures {
		var curPoint Point
		for _, gestureStep := range gestures.steps {
			if gestureStep.GestureType == "start" {
				pointerMatrix.setPoint(gestureStep.Point, fingerIndex, gestureStep.Duration)
				curPoint = gestureStep.Point
			}
			if gestureStep.GestureType == "move" {
				toPoint := gestureStep.Point
				offsetX := toPoint.X - curPoint.X
				offsetY := toPoint.Y - curPoint.Y
				steps := gestureStep.calculateSteps()
				for i := 0; i < steps-1; i++ {
					curPoint = Point{X: curPoint.X + (offsetX / steps), Y: curPoint.Y + (offsetY / steps)}
					pointerMatrix.setPoint(curPoint, fingerIndex, EVENT_INJECTION_DELAY_MS)
				}
				curPoint = toPoint
				if steps == 1 {
					pointerMatrix.setPoint(curPoint, fingerIndex, gestureStep.Duration%EVENT_INJECTION_DELAY_MS)
				} else {
					pointerMatrix.setPoint(curPoint, fingerIndex, EVENT_INJECTION_DELAY_MS+(gestureStep.Duration%EVENT_INJECTION_DELAY_MS))
				}
			}
			if gestureStep.GestureType == "pause" {
				steps := gestureStep.calculateSteps()
				for i := 0; i < steps-1; i++ {
					pointerMatrix.setPoint(curPoint, fingerIndex, EVENT_INJECTION_DELAY_MS)
				}
				pointerMatrix.setPoint(curPoint, fingerIndex, EVENT_INJECTION_DELAY_MS+(gestureStep.Duration%EVENT_INJECTION_DELAY_MS))
			}
		}
	}
	return pointerMatrix
}

func (d *UIDriver) setPoint(pointerMatrixName string, fingerIndex int, step int, point Point) error {
	params := map[string]interface{}{
		"api":          "PointerMatrix.setPoint",
		"this":         pointerMatrixName,
		"args":         []any{fingerIndex, step, point},
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		return err
	}
	if res.Exception != nil {
		return fmt.Errorf("[uitest] failed to setPoint from: %s", res.Exception.Message)
	}
	return nil
}

func (d *UIDriver) injectMultiPointerAction(driverName, pointerMatrixName string, speed int) error {
	params := map[string]interface{}{
		"api":          "Driver.injectMultiPointerAction",
		"this":         driverName,
		"args":         []any{pointerMatrixName, speed},
		"message_type": "hypium",
	}
	res, err := d.sendUitestReq(newHypiumRequest(params, "callHypiumApi"))
	if err != nil {
		return err
	}
	if res.Exception != nil {
		return fmt.Errorf("[uitest] failed to injectMultiPointerAction from: %s", res.Exception.Message)
	}
	return nil
}

func hashCode(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func (r UitestRequest) ToString() (result string, err error) {
	data, err := json.Marshal(r)
	if err != nil {
		err = fmt.Errorf("error: \n%v", err)
		return
	}
	return string(data), nil
}

func getVersion(str string) (version []string, err error) {
	index := strings.Index(str, "@")
	if index == -1 {
		err = fmt.Errorf("invalid version str")
		return
	}
	version = strings.Split(str[index+1:], ".")
	return
}
