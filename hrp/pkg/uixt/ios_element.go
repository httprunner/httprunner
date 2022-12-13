package uixt

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

// All elements returned by search endpoints have assigned element_id.
// Given element_id you can query properties like:
// enabled, rect, size, location, text, displayed, accessible, name
type wdaElement struct {
	parent *wdaDriver
	id     string // element_id
}

func (we wdaElement) Click() (err error) {
	// [[FBRoute POST:@"/element/:uuid/click"] respondWithTarget:self action:@selector(handleClick:)]
	_, err = we.parent.httpPOST(nil, "/session", we.parent.sessionId, "/element", we.id, "/click")
	return
}

func (we wdaElement) SendKeys(text string, options ...DataOption) (err error) {
	// [[FBRoute POST:@"/element/:uuid/value"] respondWithTarget:self action:@selector(handleSetValue:)]
	data := map[string]interface{}{
		"value": strings.Split(text, ""),
	}
	// new data options in post data for extra uiautomator configurations
	newData := NewData(data, options...)

	_, err = we.parent.httpPOST(newData, "/session", we.parent.sessionId, "/element", we.id, "/value")
	return
}

func (we wdaElement) Clear() (err error) {
	// [[FBRoute POST:@"/element/:uuid/clear"] respondWithTarget:self action:@selector(handleClear:)]
	_, err = we.parent.httpPOST(nil, "/session", we.parent.sessionId, "/element", we.id, "/clear")
	return
}

func (we wdaElement) Tap(x, y int) error {
	return we.TapFloat(float64(x), float64(y))
}

func (we wdaElement) TapFloat(x, y float64) (err error) {
	// [[FBRoute POST:@"/wda/tap/:uuid"] respondWithTarget:self action:@selector(handleTap:)]
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/tap/", we.id)
	return
}

func (we wdaElement) DoubleTap() (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/doubleTap"] respondWithTarget:self action:@selector(handleDoubleTap:)]
	_, err = we.parent.httpPOST(nil, "/session", we.parent.sessionId, "/wda/element", we.id, "/doubleTap")
	return
}

func (we wdaElement) TouchAndHold(second ...float64) (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/touchAndHold"] respondWithTarget:self action:@selector(handleTouchAndHold:)]
	data := make(map[string]interface{})
	if len(second) == 0 || second[0] <= 0 {
		second = []float64{1.0}
	}
	data["duration"] = second[0]
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/element", we.id, "/touchAndHold")
	return
}

func (we wdaElement) TwoFingerTap() (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/twoFingerTap"] respondWithTarget:self action:@selector(handleTwoFingerTap:)]
	_, err = we.parent.httpPOST(nil, "/session", we.parent.sessionId, "/wda/element", we.id, "/twoFingerTap")
	return
}

func (we wdaElement) TapWithNumberOfTaps(numberOfTaps, numberOfTouches int) (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/tapWithNumberOfTaps"] respondWithTarget:self action:@selector(handleTapWithNumberOfTaps:)]
	if numberOfTouches <= 0 {
		return errors.New("'numberOfTouches' must be greater than zero")
	}
	if numberOfTouches > 5 {
		return errors.New("'numberOfTouches' cannot be greater than 5")
	}
	if numberOfTaps <= 0 {
		return errors.New("'numberOfTaps' must be greater than zero")
	}
	if numberOfTaps > 10 {
		return errors.New("'numberOfTaps' cannot be greater than 10")
	}
	data := map[string]interface{}{
		"numberOfTaps":    numberOfTaps,
		"numberOfTouches": numberOfTouches,
	}
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/element", we.id, "/tapWithNumberOfTaps")
	return
}

func (we wdaElement) ForceTouch(pressure float64, second ...float64) (err error) {
	return we.ForceTouchFloat(-1, -1, pressure, second...)
}

func (we wdaElement) ForceTouchFloat(x, y, pressure float64, second ...float64) (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/forceTouch"] respondWithTarget:self action:@selector(handleForceTouch:)]
	data := make(map[string]interface{})
	if x != -1 && y != -1 {
		data["x"] = x
		data["y"] = y
	}
	if len(second) == 0 || second[0] <= 0 {
		second = []float64{1.0}
	}
	data["pressure"] = pressure
	data["duration"] = second[0]
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/element", we.id, "/forceTouch")
	return
}

func (we wdaElement) Drag(fromX, fromY, toX, toY int, pressForDuration ...float64) error {
	return we.DragFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), pressForDuration...)
}

func (we wdaElement) DragFloat(fromX, fromY, toX, toY float64, pressForDuration ...float64) (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/dragfromtoforduration"] respondWithTarget:self action:@selector(handleDrag:)]
	data := map[string]interface{}{
		"fromX": fromX,
		"fromY": fromY,
		"toX":   toX,
		"toY":   toY,
	}
	if len(pressForDuration) == 0 || pressForDuration[0] < 0 {
		pressForDuration = []float64{1.0}
	}
	data["duration"] = pressForDuration[0]
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/element", we.id, "/dragfromtoforduration")
	return
}

func (we wdaElement) Swipe(fromX, fromY, toX, toY int) error {
	return we.SwipeFloat(float64(fromX), float64(fromY), float64(toX), float64(toY))
}

func (we wdaElement) SwipeFloat(fromX, fromY, toX, toY float64) error {
	return we.DragFloat(fromX, fromY, toX, toY, 0)
}

func (we wdaElement) SwipeDirection(direction Direction, velocity ...float64) (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/swipe"] respondWithTarget:self action:@selector(handleSwipe:)]
	data := map[string]interface{}{"direction": direction}
	if len(velocity) != 0 && velocity[0] > 0 {
		data["velocity"] = velocity[0]
	}
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/element", we.id, "/swipe")
	return
}

func (we wdaElement) Pinch(scale, velocity float64) (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/pinch"] respondWithTarget:self action:@selector(handlePinch:)]
	if scale <= 0 {
		return errors.New("'scale' must be greater than zero")
	}
	if scale == 1 {
		return errors.New("'scale' must be greater or less than 1")
	}
	if scale < 1 && velocity > 0 {
		return errors.New("'velocity' must be less than zero when 'scale' is less than 1")
	}
	if scale > 1 && velocity <= 0 {
		return errors.New("'velocity' must be greater than zero when 'scale' is greater than 1")
	}
	data := map[string]interface{}{
		"scale":    scale,
		"velocity": velocity,
	}
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/element", we.id, "/pinch")
	return
}

func (we wdaElement) PinchToZoomOutByW3CAction(scale ...float64) (err error) {
	if len(scale) == 0 {
		scale = []float64{1.0}
	} else if scale[0] > 23 {
		scale[0] = 23
	}
	var size Size
	if size, err = we.Size(); err != nil {
		return err
	}
	r := scale[0] * 2 / 100.0
	offsetX, offsetY := float64(size.Width)*r, float64(size.Height)*r

	actions := NewW3CActions().SwipeFloat(0-offsetX, 0-offsetY, 0, 0, we).SwipeFloat(offsetX, offsetY, 0, 0, we)
	return we.parent.PerformW3CActions(actions)
}

func (we wdaElement) Rotate(rotation float64, velocity ...float64) (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/rotate"] respondWithTarget:self action:@selector(handleRotate:)]
	if rotation > math.Pi*2 || rotation < math.Pi*-2 {
		return errors.New("'rotation' must not be more than 2π or less than -2π")
	}
	if len(velocity) == 0 || velocity[0] == 0 {
		velocity = []float64{rotation}
	}
	if rotation > 0 && velocity[0] < 0 || rotation < 0 && velocity[0] > 0 {
		return errors.New("'rotation' and 'velocity' must have the same sign")
	}
	data := map[string]interface{}{
		"rotation": rotation,
		"velocity": velocity[0],
	}
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/element", we.id, "/rotate")
	return
}

func (we wdaElement) PickerWheelSelect(order PickerWheelOrder, offset ...int) (err error) {
	// [[FBRoute POST:@"/wda/pickerwheel/:uuid/select"] respondWithTarget:self action:@selector(handleWheelSelect:)]
	if len(offset) == 0 {
		offset = []int{2}
	} else if offset[0] <= 0 || offset[0] > 5 {
		return fmt.Errorf("'offset' value is expected to be in range (0, 5]. '%d' was given instead", offset[0])
	}
	data := map[string]interface{}{
		"order":  order,
		"offset": float64(offset[0]) * 0.1,
	}
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/pickerwheel", we.id, "/select")
	return
}

func (we wdaElement) scroll(data interface{}) (err error) {
	// [[FBRoute POST:@"/wda/element/:uuid/scroll"] respondWithTarget:self action:@selector(handleScroll:)]
	_, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/wda/element", we.id, "/scroll")
	return
}

func (we wdaElement) ScrollElementByName(name string) error {
	data := map[string]interface{}{"name": name}
	return we.scroll(data)
}

func (we wdaElement) ScrollElementByPredicate(predicate string) error {
	data := map[string]interface{}{"predicateString": predicate}
	return we.scroll(data)
}

func (we wdaElement) ScrollToVisible() error {
	data := map[string]interface{}{"toVisible": true}
	return we.scroll(data)
}

func (we wdaElement) ScrollDirection(direction Direction, distance ...float64) error {
	if len(distance) == 0 || distance[0] <= 0 {
		distance = []float64{0.5}
	}
	data := map[string]interface{}{
		"direction": direction,
		"distance":  distance[0],
	}
	return we.scroll(data)
}

func (we wdaElement) FindElement(by BySelector) (element WebElement, err error) {
	// [[FBRoute POST:@"/element/:uuid/element"] respondWithTarget:self action:@selector(handleFindSubElement:)]
	using, value := by.getUsingAndValue()
	data := map[string]interface{}{
		"using": using,
		"value": value,
	}
	var rawResp rawResponse
	if rawResp, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/element", we.id, "/element"); err != nil {
		return nil, err
	}
	var elementID string
	if elementID, err = rawResp.valueConvertToElementID(); err != nil {
		if errors.Is(err, errNoSuchElement) {
			return nil, fmt.Errorf("%w: unable to find an element using '%s', value '%s'", err, using, value)
		}
		return nil, err
	}
	element = &wdaElement{parent: we.parent, id: elementID}
	return
}

func (we wdaElement) FindElements(by BySelector) (elements []WebElement, err error) {
	// [[FBRoute POST:@"/element/:uuid/elements"] respondWithTarget:self action:@selector(handleFindSubElements:)]
	using, value := by.getUsingAndValue()
	data := map[string]interface{}{
		"using": using,
		"value": value,
	}
	var rawResp rawResponse
	if rawResp, err = we.parent.httpPOST(data, "/session", we.parent.sessionId, "/element", we.id, "/elements"); err != nil {
		return nil, err
	}
	var elementIDs []string
	if elementIDs, err = rawResp.valueConvertToElementIDs(); err != nil {
		if errors.Is(err, errNoSuchElement) {
			return nil, fmt.Errorf("%w: unable to find an element using '%s', value '%s'", err, using, value)
		}
		return nil, err
	}
	elements = make([]WebElement, len(elementIDs))
	for i := range elementIDs {
		elements[i] = &wdaElement{parent: we.parent, id: elementIDs[i]}
	}
	return
}

func (we wdaElement) FindVisibleCells() (elements []WebElement, err error) {
	// [[FBRoute GET:@"/wda/element/:uuid/getVisibleCells"] respondWithTarget:self action:@selector(handleFindVisibleCells:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/wda/element", we.id, "/getVisibleCells"); err != nil {
		return nil, err
	}
	var elementIDs []string
	if elementIDs, err = rawResp.valueConvertToElementIDs(); err != nil {
		if errors.Is(err, errNoSuchElement) {
			return nil, fmt.Errorf("%w: unable to find a cell element in this element", err)
		}
		return nil, err
	}
	elements = make([]WebElement, len(elementIDs))
	for i := range elementIDs {
		elements[i] = &wdaElement{parent: we.parent, id: elementIDs[i]}
	}
	return
}

func (we wdaElement) Rect() (rect Rect, err error) {
	// [[FBRoute GET:@"/element/:uuid/rect"] respondWithTarget:self action:@selector(handleGetRect:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/element", we.id, "/rect"); err != nil {
		return Rect{}, err
	}
	reply := new(struct{ Value struct{ Rect } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Rect{}, err
	}
	rect = reply.Value.Rect
	return
}

func (we wdaElement) Location() (Point, error) {
	rect, err := we.Rect()
	if err != nil {
		return Point{}, err
	}
	return rect.Point, nil
}

func (we wdaElement) Size() (Size, error) {
	rect, err := we.Rect()
	if err != nil {
		return Size{}, err
	}
	return rect.Size, nil
}

func (we wdaElement) Text() (text string, err error) {
	// [[FBRoute GET:@"/element/:uuid/text"] respondWithTarget:self action:@selector(handleGetText:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/element", we.id, "/text"); err != nil {
		return "", err
	}
	if text, err = rawResp.valueConvertToString(); err != nil {
		return "", err
	}
	return
}

func (we wdaElement) Type() (elemType string, err error) {
	// [[FBRoute GET:@"/element/:uuid/name"] respondWithTarget:self action:@selector(handleGetName:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/element", we.id, "/name"); err != nil {
		return "", err
	}
	if elemType, err = rawResp.valueConvertToString(); err != nil {
		return "", err
	}
	return
}

func (we wdaElement) IsEnabled() (enabled bool, err error) {
	// [[FBRoute GET:@"/element/:uuid/enabled"] respondWithTarget:self action:@selector(handleGetEnabled:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/element", we.id, "/enabled"); err != nil {
		return false, err
	}
	if enabled, err = rawResp.valueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (we wdaElement) IsDisplayed() (displayed bool, err error) {
	// [[FBRoute GET:@"/element/:uuid/displayed"] respondWithTarget:self action:@selector(handleGetDisplayed:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/element", we.id, "/displayed"); err != nil {
		return false, err
	}
	if displayed, err = rawResp.valueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (we wdaElement) IsSelected() (selected bool, err error) {
	// [[FBRoute GET:@"/element/:uuid/selected"] respondWithTarget:self action:@selector(handleGetSelected:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/element", we.id, "/selected"); err != nil {
		return false, err
	}
	if selected, err = rawResp.valueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (we wdaElement) IsAccessible() (accessible bool, err error) {
	// [[FBRoute GET:@"/wda/element/:uuid/accessible"] respondWithTarget:self action:@selector(handleGetAccessible:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/wda/element", we.id, "/accessible"); err != nil {
		return false, err
	}
	if accessible, err = rawResp.valueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (we wdaElement) IsAccessibilityContainer() (isAccessibilityContainer bool, err error) {
	// [[FBRoute GET:@"/wda/element/:uuid/accessibilityContainer"] respondWithTarget:self action:@selector(handleGetIsAccessibilityContainer:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/wda/element", we.id, "/accessibilityContainer"); err != nil {
		return false, err
	}
	if isAccessibilityContainer, err = rawResp.valueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (we wdaElement) GetAttribute(attr ElementAttribute) (value string, err error) {
	// [[FBRoute GET:@"/element/:uuid/attribute/:name"] respondWithTarget:self action:@selector(handleGetAttribute:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/element", we.id, "/attribute", attr.getAttributeName()); err != nil {
		return "", err
	}
	if value, err = rawResp.valueConvertToString(); err != nil {
		return "", err
	}
	return
}

func (we wdaElement) UID() (uid string) {
	return we.id
}

func (we wdaElement) Screenshot() (raw *bytes.Buffer, err error) {
	// W3C element screenshot
	// [[FBRoute GET:@"/element/:uuid/screenshot"] respondWithTarget:self action:@selector(handleElementScreenshot:)]
	// JSONWP element screenshot
	// [[FBRoute GET:@"/screenshot/:uuid"] respondWithTarget:self action:@selector(handleElementScreenshot:)]
	var rawResp rawResponse
	if rawResp, err = we.parent.httpGET("/session", we.parent.sessionId, "/element", we.id, "/screenshot"); err != nil {
		return nil, err
	}
	if raw, err = rawResp.valueDecodeAsBase64(); err != nil {
		return nil, err
	}
	return
}
