package uixt

import "strings"

type touchGesture struct {
	Touch PointF  `json:"touch"`
	Time  float64 `json:"time"`
}

type TouchAction []touchGesture

func NewTouchAction(cap ...int) *TouchAction {
	if len(cap) == 0 || cap[0] <= 0 {
		cap = []int{8}
	}
	tmp := make(TouchAction, 0, cap[0])
	return &tmp
}

func (ta *TouchAction) Add(x, y int, startTime ...float64) *TouchAction {
	return ta.AddFloat(float64(x), float64(y), startTime...)
}

func (ta *TouchAction) AddFloat(x, y float64, startTime ...float64) *TouchAction {
	if len(startTime) == 0 {
		var tmp float64 = 0
		if len(*ta) != 0 {
			g := (*ta)[len(*ta)-1]
			tmp = g.Time + 0.05
		}
		startTime = []float64{tmp}
	}
	*ta = append(*ta, touchGesture{Touch: PointF{x, y}, Time: startTime[0]})
	return ta
}

func (ta *TouchAction) AddPoint(point Point, startTime ...float64) *TouchAction {
	return ta.AddFloat(float64(point.X), float64(point.Y), startTime...)
}

func (ta *TouchAction) AddPointF(point PointF, startTime ...float64) *TouchAction {
	return ta.AddFloat(point.X, point.Y, startTime...)
}

func (ud *uiaDriver) MultiPointerGesture(gesture1 *TouchAction, gesture2 *TouchAction, tas ...*TouchAction) (err error) {
	// Must provide coordinates for at least 2 pointers
	actions := make([]*TouchAction, 0)
	actions = append(actions, gesture1, gesture2)
	if len(tas) != 0 {
		actions = append(actions, tas...)
	}
	data := map[string]interface{}{
		"actions": actions,
	}
	// register(postHandler, new MultiPointerGesture("/wd/hub/session/:sessionId/touch/multi/perform"))
	_, err = ud.httpPOST(data, "/session", ud.sessionId, "/touch/multi/perform")
	return
}

type w3cGesture map[string]interface{}

func _newW3CGesture() w3cGesture {
	return make(w3cGesture)
}

func (g w3cGesture) _set(key string, value interface{}) w3cGesture {
	g[key] = value
	return g
}

func (g w3cGesture) pause(duration float64) w3cGesture {
	return g._set("type", "pause").
		_set("duration", duration)
}

func (g w3cGesture) keyDown(value string) w3cGesture {
	return g._set("type", "keyDown").
		_set("value", value)
}

func (g w3cGesture) keyUp(value string) w3cGesture {
	return g._set("type", "keyUp").
		_set("value", value)
}

func (g w3cGesture) pointerDown(button int) w3cGesture {
	return g._set("type", "pointerDown")._set("button", button)
}

func (g w3cGesture) pointerUp(button int) w3cGesture {
	return g._set("type", "pointerUp")._set("button", button)
}

func (g w3cGesture) pointerMove(x, y float64, origin string, duration float64, pressureAndSize ...float64) w3cGesture {
	switch len(pressureAndSize) {
	case 1:
		g._set("pressure", pressureAndSize[0])
	case 2:
		g._set("pressure", pressureAndSize[0])
		g._set("size", pressureAndSize[1])
	}
	return g._set("type", "pointerMove").
		_set("duration", duration).
		_set("origin", origin).
		_set("x", x).
		_set("y", y)
}

func (g w3cGesture) size(size ...float64) w3cGesture {
	if len(size) == 0 {
		size = []float64{1.0}
	}
	return g._set("size", size[0])
}

func (g w3cGesture) pressure(pressure ...float64) w3cGesture {
	if len(pressure) == 0 {
		pressure = []float64{1.0}
	}
	return g._set("pressure", pressure[0])
}

type W3CGestures []w3cGesture

func NewW3CGestures(cap ...int) *W3CGestures {
	if len(cap) == 0 || cap[0] <= 0 {
		cap = []int{8}
	}
	tmp := make(W3CGestures, 0, cap[0])
	return &tmp
}

func (g *W3CGestures) Pause(duration ...float64) *W3CGestures {
	if len(duration) == 0 || duration[0] < 0 {
		duration = []float64{0.5}
	}
	*g = append(*g, _newW3CGesture().pause(duration[0]*1000))
	return g
}

func (g *W3CGestures) KeyDown(value string) *W3CGestures {
	*g = append(*g, _newW3CGesture().keyDown(value))
	return g
}

func (g *W3CGestures) KeyUp(value string) *W3CGestures {
	*g = append(*g, _newW3CGesture().keyUp(value))
	return g
}

func (g *W3CGestures) SendKeys(text string) *W3CGestures {
	ss := strings.Split(text, "")
	for i := range ss {
		g.KeyDown(ss[i])
		g.KeyUp(ss[i])
	}
	return g
}
