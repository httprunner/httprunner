package uixt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

type uiaElement struct {
	parent *uiaDriver
	id     string
}

func (e *uiaElement) Text() (text string, err error) {
	// register(getHandler, new GetText("/wd/hub/session/:sessionId/element/:id/text"))
	var rawResp rawResponse
	if rawResp, err = e.parent.httpGET("/session", e.parent.sessionId, "/element", e.id, "/text"); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	text = reply.Value
	return
}

func (e *uiaElement) GetAttribute(name string) (attribute string, err error) {
	// register(getHandler, new GetElementAttribute("/wd/hub/session/:sessionId/element/:id/attribute/:name"))
	var rawResp rawResponse
	if rawResp, err = e.parent.httpGET("/session", e.parent.sessionId, "/element", e.id, "/attribute", name); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	attribute = reply.Value
	return
}

func (e *uiaElement) ContentDescription() (name string, err error) {
	// register(getHandler, new GetName("/wd/hub/session/:sessionId/element/:id/name"))
	var rawResp rawResponse
	if rawResp, err = e.parent.httpGET("/session", e.parent.sessionId, "/element", e.id, "/name"); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	name = reply.Value
	return
}

func (e *uiaElement) Size() (size Size, err error) {
	// register(getHandler, new GetSize("/wd/hub/session/:sessionId/element/:id/size"))
	var rawResp rawResponse
	if rawResp, err = e.parent.httpGET("/session", e.parent.sessionId, "/element", e.id, "/size"); err != nil {
		return Size{-1, -1}, err
	}
	reply := new(struct{ Value Size })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Size{-1, -1}, err
	}
	size = reply.Value
	return
}

func (e *uiaElement) Rect() (rect Rect, err error) {
	// register(getHandler, new GetRect("/wd/hub/session/:sessionId/element/:id/rect"))
	var rawResp rawResponse
	if rawResp, err = e.parent.httpGET("/session", e.parent.sessionId, "/element", e.id, "/rect"); err != nil {
		return Rect{}, err
	}
	reply := new(struct{ Value Rect })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Rect{}, err
	}
	rect = reply.Value
	return
}

func (e *uiaElement) Screenshot() (raw *bytes.Buffer, err error) {
	// W3C endpoint
	// register(getHandler, new GetElementScreenshot("/wd/hub/session/:sessionId/element/:id/screenshot"))
	// JSONWP endpoint
	// register(getHandler, new GetElementScreenshot("/wd/hub/session/:sessionId/screenshot/:id"))
	var rawResp rawResponse
	if rawResp, err = e.parent.httpGET("/session", e.parent.sessionId, "/element", e.id, "/screenshot"); err != nil {
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

func (e *uiaElement) Location() (point Point, err error) {
	// register(getHandler, new Location("/wd/hub/session/:sessionId/element/:id/location"))
	var rawResp rawResponse
	if rawResp, err = e.parent.httpGET("/session", e.parent.sessionId, "/element", e.id, "/location"); err != nil {
		return Point{-1, -1}, err
	}
	reply := new(struct{ Value Point })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Point{-1, -1}, err
	}
	point = reply.Value
	return
}

func (e *uiaElement) Click() (err error) {
	// register(postHandler, new Click("/wd/hub/session/:sessionId/element/:id/click"))
	_, err = e.parent.httpPOST(nil, "/session", e.parent.sessionId, "/element", e.id, "/click")
	return
}

func (e *uiaElement) Clear() (err error) {
	// register(postHandler, new Clear("/wd/hub/session/:sessionId/element/:id/clear"))
	_, err = e.parent.httpPOST(nil, "/session", e.parent.sessionId, "/element", e.id, "/clear")
	return
}

func (e *uiaElement) SendKeys(text string, isReplace ...bool) (err error) {
	if len(isReplace) == 0 {
		isReplace = []bool{true}
	}
	// register(postHandler, new SendKeysToElement("/wd/hub/session/:sessionId/element/:id/value"))
	// https://github.com/appium/appium-uiautomator2-server/blob/master/app/src/main/java/io/appium/uiautomator2/handler/SendKeysToElement.java#L76-L85
	data := map[string]interface{}{
		"text":    text,
		"replace": isReplace[0],
	}
	_, err = e.parent.httpPOST(data, "/session", e.parent.sessionId, "/element", e.id, "/value")
	return
}

func (e *uiaElement) FindElements(by AndroidBySelector) (elements []*uiaElement, err error) {
	method, selector := by.getMethodAndSelector()
	return e.parent._findElements(method, selector, e.id)
}

func (e *uiaElement) FindElement(by AndroidBySelector) (elem *uiaElement, err error) {
	method, selector := by.getMethodAndSelector()
	return e.parent._findElement(method, selector, e.id)
}

func (e *uiaElement) Swipe(startX, startY, endX, endY int, steps ...int) (err error) {
	return e.SwipeFloat(float64(startX), float64(startY), float64(endX), float64(endY), steps...)
}

func (e *uiaElement) SwipeFloat(startX, startY, endX, endY float64, steps ...int) (err error) {
	if len(steps) == 0 {
		steps = []int{12}
	}
	return e.parent._swipe(startX, startY, endX, endY, steps[0], e.id)
}

func (e *uiaElement) SwipePoint(startPoint, endPoint Point, steps ...int) (err error) {
	return e.Swipe(startPoint.X, startPoint.Y, endPoint.X, endPoint.Y, steps...)
}

func (e *uiaElement) SwipePointF(startPoint, endPoint PointF, steps ...int) (err error) {
	return e.SwipeFloat(startPoint.X, startPoint.Y, endPoint.X, endPoint.Y, steps...)
}

func (e *uiaElement) Drag(endX, endY int, steps ...int) (err error) {
	return e.DragFloat(float64(endX), float64(endY), steps...)
}

func (e *uiaElement) DragFloat(endX, endY float64, steps ...int) error {
	if len(steps) == 0 {
		steps = []int{12 * 10}
	} else {
		steps[0] = 12 * 10
	}
	data := map[string]interface{}{
		"elementId": e.id,
		"endX":      endX,
		"endY":      endY,
		"steps":     steps[0],
	}
	return e.parent._drag(data)
}

func (e *uiaElement) DragPoint(endPoint Point, steps ...int) error {
	return e.Drag(endPoint.X, endPoint.Y, steps...)
}

func (e *uiaElement) DragPointF(endPoint PointF, steps ...int) (err error) {
	return e.DragFloat(endPoint.X, endPoint.Y, steps...)
}

func (e *uiaElement) DragTo(destElem *uiaElement, steps ...int) error {
	if len(steps) == 0 {
		steps = []int{12}
	}
	data := map[string]interface{}{
		"elementId": e.id,
		"destElId":  destElem.id,
		"steps":     steps[0],
	}
	return e.parent._drag(data)
}

func (e *uiaElement) Flick(xOffset, yOffset, speed int) (err error) {
	data := map[string]interface{}{
		legacyWebElementIdentifier: e.id,
		webElementIdentifier:       e.id,
		"xoffset":                  xOffset,
		"yoffset":                  yOffset,
		"speed":                    speed,
	}
	return e.parent._flick(data)
}

func (e *uiaElement) ScrollTo(by AndroidBySelector, maxSwipes ...int) (err error) {
	if len(maxSwipes) == 0 {
		maxSwipes = []int{0}
	}
	method, selector := by.getMethodAndSelector()
	return e.parent._scrollTo(method, selector, maxSwipes[0], e.id)
}

func (e *uiaElement) ScrollToElement(element *uiaElement) (err error) {
	// register(postHandler, new ScrollToElement("/wd/hub/session/:sessionId/appium/element/:id/scroll_to/:id2"))
	_, err = e.parent.httpPOST(nil, "/session", e.parent.sessionId, "/appium/element", e.id, "/scroll_to", element.id)
	return
}
