package uixt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var errElementNotImplemented = errors.New("element method not implemented")

type uiaElement struct {
	parent *uiaDriver
	id     string
}

func (ue uiaElement) Click() (err error) {
	// register(postHandler, new Click("/wd/hub/session/:sessionId/element/:id/click"))
	_, err = ue.parent.httpPOST(nil, "/session", ue.parent.sessionId, "/element", ue.id, "/click")
	return
}

func (ue uiaElement) SendKeys(text string, options ...DataOption) (err error) {
	// register(postHandler, new SendKeysToElement("/wd/hub/session/:sessionId/element/:id/value"))
	// https://github.com/appium/appium-uiutomator2-server/blob/master/app/src/main/java/io/appium/uiutomator2/handler/SendKeysToElement.java#L76-L85
	data := map[string]interface{}{
		"text": text,
	}

	// new data options in post data for extra uiautomator configurations
	newData := NewData(data, options...)

	_, err = ue.parent.httpPOST(newData, "/session", ue.parent.sessionId, "/element", ue.id, "/value")
	return
}

func (ue uiaElement) Clear() (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) Tap(x, y int) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) TapFloat(x, y float64) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) DoubleTap() (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) TouchAndHold(second ...float64) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) TwoFingerTap() (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) TapWithNumberOfTaps(numberOfTaps, numberOfTouches int) (err error) {
	// Todo: implement
	log.Fatal().Msg("not support")
	return
}

func (ue uiaElement) ForceTouch(pressure float64, second ...float64) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) ForceTouchFloat(x, y, pressure float64, second ...float64) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) Drag(fromX, fromY, toX, toY int, steps ...float64) (err error) {
	return ue.DragFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), steps...)
}

func (ue uiaElement) DragFloat(fromX, fromY, toX, toY float64, steps ...float64) (err error) {
	if len(steps) == 0 {
		steps = []float64{12 * 10}
	} else {
		steps[0] = 12 * 10
	}
	data := map[string]interface{}{
		"elementId": ue.id,
		"endX":      toX,
		"endY":      toY,
		"steps":     steps[0],
	}
	return ue.parent._drag(data)
}

func (ue uiaElement) Swipe(fromX, fromY, toX, toY int) error {
	return ue.SwipeFloat(float64(fromX), float64(fromY), float64(toX), float64(toY))
}

func (ue uiaElement) SwipeFloat(fromX, fromY, toX, toY float64) error {
	options := []DataOption{
		WithDataSteps(12),
		WithCustomOption("elementId", ue.id),
	}
	return ue.parent._swipe(fromX, fromY, toX, toY, options...)
}

func (ue uiaElement) SwipeDirection(direction Direction, velocity ...float64) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) Pinch(scale, velocity float64) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) PinchToZoomOutByW3CAction(scale ...float64) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) Rotate(rotation float64, velocity ...float64) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) PickerWheelSelect(order PickerWheelOrder, offset ...int) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) scroll(data interface{}) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) ScrollElementByName(name string) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) ScrollElementByPredicate(predicate string) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) ScrollToVisible() (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) ScrollDirection(direction Direction, distance ...float64) (err error) {
	// TODO
	return errElementNotImplemented
}

func (ue uiaElement) FindElement(by BySelector) (element WebElement, err error) {
	method, selector := by.getMethodAndSelector()
	return ue.parent._findElement(method, selector, ue.id)
}

func (ue uiaElement) FindElements(by BySelector) (elements []WebElement, err error) {
	method, selector := by.getMethodAndSelector()
	return ue.parent._findElements(method, selector, ue.id)
}

func (ue uiaElement) FindVisibleCells() (elements []WebElement, err error) {
	// TODO
	return elements, errElementNotImplemented
}

func (ue uiaElement) Rect() (rect Rect, err error) {
	// register(getHandler, new GetRect("/wd/hub/session/:sessionId/element/:id/rect"))
	var rawResp rawResponse
	if rawResp, err = ue.parent.httpGET("/session", ue.parent.sessionId, "/element", ue.id, "/rect"); err != nil {
		return Rect{}, err
	}
	reply := new(struct{ Value Rect })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Rect{}, err
	}
	rect = reply.Value
	return
}

func (ue uiaElement) Location() (point Point, err error) {
	// register(getHandler, new Location("/wd/hub/session/:sessionId/element/:id/location"))
	var rawResp rawResponse
	if rawResp, err = ue.parent.httpGET("/session", ue.parent.sessionId, "/element", ue.id, "/location"); err != nil {
		return Point{-1, -1}, err
	}
	reply := new(struct{ Value Point })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Point{-1, -1}, err
	}
	point = reply.Value
	return
}

func (ue uiaElement) Size() (size Size, err error) {
	// register(getHandler, new GetSize("/wd/hub/session/:sessionId/element/:id/size"))
	var rawResp rawResponse
	if rawResp, err = ue.parent.httpGET("/session", ue.parent.sessionId, "/element", ue.id, "/size"); err != nil {
		return Size{-1, -1}, err
	}
	reply := new(struct{ Value Size })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Size{-1, -1}, err
	}
	size = reply.Value
	return
}

func (ue uiaElement) Text() (text string, err error) {
	// register(getHandler, new GetText("/wd/hub/session/:sessionId/element/:id/text"))
	var rawResp rawResponse
	if rawResp, err = ue.parent.httpGET("/session", ue.parent.sessionId, "/element", ue.id, "/text"); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	text = reply.Value
	return
}

func (ue uiaElement) Type() (elemType string, err error) {
	// TODO
	return elemType, errElementNotImplemented
}

func (ue uiaElement) IsEnabled() (enabled bool, err error) {
	// TODO
	return enabled, errElementNotImplemented
}

func (ue uiaElement) IsDisplayed() (displayed bool, err error) {
	// TODO
	return displayed, errElementNotImplemented
}

func (ue uiaElement) IsSelected() (selected bool, err error) {
	// TODO
	return selected, errElementNotImplemented
}

func (ue uiaElement) IsAccessible() (accessible bool, err error) {
	// TODO
	return accessible, errElementNotImplemented
}

func (ue uiaElement) IsAccessibilityContainer() (isAccessibilityContainer bool, err error) {
	// TODO
	return isAccessibilityContainer, errElementNotImplemented
}

func (ue uiaElement) GetAttribute(attr ElementAttribute) (value string, err error) {
	// register(getHandler, new GetElementAttribute("/wd/hub/session/:sessionId/element/:id/attribute/:name"))
	var rawResp rawResponse
	if rawResp, err = ue.parent.httpGET("/session", ue.parent.sessionId, "/element", ue.id, "/attribute", attr.getAttributeName()); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	value = reply.Value
	return
}

func (ue uiaElement) UID() (uid string) {
	return ue.id
}

func (ue uiaElement) Screenshot() (raw *bytes.Buffer, err error) {
	// W3C endpoint
	// register(getHandler, new GetElementScreenshot("/wd/hub/session/:sessionId/element/:id/screenshot"))
	// JSONWP endpoint
	// register(getHandler, new GetElementScreenshot("/wd/hub/session/:sessionId/screenshot/:id"))
	var rawResp rawResponse
	if rawResp, err = ue.parent.httpGET("/session", ue.parent.sessionId, "/element", ue.id, "/screenshot"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}

	var decodeStr []byte
	if decodeStr, err = base64.StdEncoding.DecodeString(reply.Value); err != nil {
		return nil, err
	}

	raw = bytes.NewBuffer(decodeStr)
	return
}
