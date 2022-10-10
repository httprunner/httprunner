package uixt

import (
	"strconv"
	"strings"
)

type W3CActions []map[string]interface{}

func NewW3CActions(capacity ...int) *W3CActions {
	if len(capacity) == 0 || capacity[0] <= 0 {
		capacity = []int{8}
	}
	tmp := make(W3CActions, 0, capacity[0])
	return &tmp
}

func (act *W3CActions) SendKeys(text string) *W3CActions {
	keyboard := make(map[string]interface{})
	keyboard["type"] = "key"
	keyboard["id"] = "keyboard" + strconv.FormatInt(int64(len(*act)+1), 10)

	ss := strings.Split(text, "")
	type KeyEvent struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	actOptKey := make([]KeyEvent, 0, len(ss)+1)
	for i := range ss {
		actOptKey = append(
			actOptKey,
			KeyEvent{Type: "keyDown", Value: ss[i]},
			KeyEvent{Type: "keyUp", Value: ss[i]},
		)
	}
	keyboard["actions"] = actOptKey
	*act = append(*act, keyboard)
	return act
}

func (act *W3CActions) _newFinger() map[string]interface{} {
	pointer := make(map[string]interface{})
	pointer["type"] = "pointer"
	pointer["id"] = "finger" + strconv.FormatInt(int64(len(*act)+1), 10)
	pointer["parameters"] = map[string]string{"pointerType": "touch"}
	return pointer
}

func (act *W3CActions) FingerAction(fingerAct *FingerAction, fActs ...*FingerAction) *W3CActions {
	fActs = append([]*FingerAction{fingerAct}, fActs...)
	for i := range fActs {
		pointer := act._newFinger()
		pointer["actions"] = *fActs[i]
		*act = append(*act, pointer)
	}
	return act
}

type FingerAction []map[string]interface{}

func NewFingerAction(capacity ...int) *FingerAction {
	if len(capacity) == 0 || capacity[0] <= 0 {
		capacity = []int{8}
	}
	tmp := make(FingerAction, 0, capacity[0])
	return &tmp
}

type FingerMove map[string]interface{}

func NewFingerMove() FingerMove {
	return map[string]interface{}{"type": "pointerMove"}
}

func (fm FingerMove) WithXY(x, y int) FingerMove {
	fm["x"] = x
	fm["y"] = y
	return fm
}

func (fm FingerMove) WithXYFloat(x, y float64) FingerMove {
	fm["x"] = x
	fm["y"] = y
	return fm
}

func (fm FingerMove) WithOrigin(element WebElement) FingerMove {
	fm["origin"] = element.UID()
	return fm
}

func (fm FingerMove) WithDuration(second float64) FingerMove {
	fm["duration"] = second
	return fm
}

func (fa *FingerAction) Move(fm FingerMove) *FingerAction {
	*fa = append(*fa, fm)
	return fa
}

func (fa *FingerAction) Down() *FingerAction {
	*fa = append(*fa, map[string]interface{}{"type": "pointerDown"})
	return fa
}

func (fa *FingerAction) Up() *FingerAction {
	*fa = append(*fa, map[string]interface{}{"type": "pointerUp"})
	return fa
}

func (fa *FingerAction) Pause(second ...float64) *FingerAction {
	if len(second) == 0 || second[0] < 0 {
		second = []float64{0.5}
	}
	tmp := map[string]interface{}{
		"type":     "pause",
		"duration": second[0] * 1000,
	}
	*fa = append(*fa, tmp)
	return fa
}

func (act *W3CActions) Tap(x, y int, element ...WebElement) *W3CActions {
	fm := NewFingerMove().WithXY(x, y)
	if len(element) != 0 {
		fm.WithOrigin(element[0])
	}
	fingerAction := NewFingerAction().
		Move(fm).
		Down().
		Pause(0.1).
		Up()
	return act.FingerAction(fingerAction)
}

func (act *W3CActions) DoubleTap(x, y int, element ...WebElement) *W3CActions {
	fm := NewFingerMove().WithXY(x, y)
	if len(element) != 0 {
		fm.WithOrigin(element[0])
	}
	fingerAction := NewFingerAction().
		Move(fm).
		Down().
		Pause(0.1).
		Up().
		Pause(0.04).
		Down().
		Pause(0.1).
		Up()
	return act.FingerAction(fingerAction)
}

func (act *W3CActions) Press(x, y int, second float64, element ...WebElement) *W3CActions {
	fm := NewFingerMove().WithXY(x, y)
	if len(element) != 0 {
		fm.WithOrigin(element[0])
	}
	fingerAction := NewFingerAction().
		Move(fm).
		Down().
		Pause(second).
		Up()
	return act.FingerAction(fingerAction)
}

func (act *W3CActions) Swipe(fromX, fromY, toX, toY int, element ...WebElement) *W3CActions {
	fmFrom := NewFingerMove().WithXY(fromX, fromY)
	fmTo := NewFingerMove().WithXY(toX, toY)
	if len(element) != 0 {
		fmFrom.WithOrigin(element[0])
		fmTo.WithOrigin(element[0])
	}
	fingerAction := NewFingerAction().
		Move(fmFrom).
		Down().
		Pause(0.25).
		Move(fmTo).
		Pause(0.25).
		Up()
	return act.FingerAction(fingerAction)
}

func (act *W3CActions) SwipeFloat(fromX, fromY, toX, toY float64, element ...WebElement) *W3CActions {
	fmFrom := NewFingerMove().WithXYFloat(fromX, fromY)
	fmTo := NewFingerMove().WithXYFloat(toX, toY)
	if len(element) != 0 {
		fmFrom.WithOrigin(element[0])
		fmTo.WithOrigin(element[0])
	}
	fingerAction := NewFingerAction().
		Move(fmFrom).
		Down().
		Pause(0.25).
		Move(fmTo).
		Pause(0.25).
		Up()
	return act.FingerAction(fingerAction)
}

/* ---------------------------------------------------------------------------------------------------------------- */

type TouchActions []map[string]interface{}

func NewTouchActions(capacity ...int) *TouchActions {
	if len(capacity) == 0 || capacity[0] <= 0 {
		capacity = []int{8}
	}
	tmp := make(TouchActions, 0, capacity[0])
	return &tmp
}

func (act *TouchActions) MoveTo(opt TouchActionMoveTo) *TouchActions {
	tmp := map[string]interface{}{
		"action":  "moveTo",
		"options": opt,
	}
	*act = append(*act, tmp)
	return act
}

func (act *TouchActions) Tap(opt TouchActionTap) *TouchActions {
	tmp := map[string]interface{}{
		"action":  "tap",
		"options": opt,
	}
	*act = append(*act, tmp)
	return act
}

func (act *TouchActions) Press(opt TouchActionPress) *TouchActions {
	tmp := map[string]interface{}{
		"action":  "press",
		"options": opt,
	}
	*act = append(*act, tmp)
	return act
}

func (act *TouchActions) LongPress(opt TouchActionLongPress) *TouchActions {
	tmp := map[string]interface{}{
		"action":  "longPress",
		"options": opt,
	}
	*act = append(*act, tmp)
	return act
}

func (act *TouchActions) Wait(second ...float64) *TouchActions {
	if len(second) == 0 || second[0] < 0 {
		second = []float64{0.5}
	}
	tmp := map[string]interface{}{
		"action":  "wait",
		"options": map[string]interface{}{"ms": second[0] * 1000},
	}
	*act = append(*act, tmp)
	return act
}

func (act *TouchActions) Release() *TouchActions {
	tmp := map[string]interface{}{"action": "release"}
	*act = append(*act, tmp)
	return act
}

func (act *TouchActions) Cancel() *TouchActions {
	tmp := map[string]interface{}{"action": "cancel"}
	*act = append(*act, tmp)
	return act
}

type TouchActionMoveTo map[string]interface{}

func NewTouchActionMoveTo() TouchActionMoveTo {
	return make(map[string]interface{})
}

func (opt TouchActionMoveTo) WithXY(x, y int) TouchActionMoveTo {
	opt["x"] = x
	opt["y"] = y
	return opt
}

func (opt TouchActionMoveTo) WithXYFloat(x, y float64) TouchActionMoveTo {
	opt["x"] = x
	opt["y"] = y
	return opt
}

func (opt TouchActionMoveTo) WithElement(element WebElement) TouchActionMoveTo {
	opt["element"] = element.UID()
	return opt
}

type TouchActionTap map[string]interface{}

func NewTouchActionTap() TouchActionTap {
	return make(map[string]interface{})
}

func (opt TouchActionTap) WithXY(x, y int) TouchActionTap {
	opt["x"] = x
	opt["y"] = y
	return opt
}

func (opt TouchActionTap) WithXYFloat(x, y float64) TouchActionTap {
	opt["x"] = x
	opt["y"] = y
	return opt
}

func (opt TouchActionTap) WithElement(element WebElement) TouchActionTap {
	opt["element"] = element.UID()
	return opt
}

func (opt TouchActionTap) WithCount(count int) TouchActionTap {
	opt["count"] = count
	return opt
}

type TouchActionPress map[string]interface{}

func NewTouchActionPress() TouchActionPress {
	return make(map[string]interface{})
}

func (opt TouchActionPress) WithXY(x, y int) TouchActionPress {
	opt["x"] = x
	opt["y"] = y
	return opt
}

func (opt TouchActionPress) WithXYFloat(x, y float64) TouchActionPress {
	opt["x"] = x
	opt["y"] = y
	return opt
}

func (opt TouchActionPress) WithElement(element WebElement) TouchActionPress {
	opt["element"] = element.UID()
	return opt
}

func (opt TouchActionPress) WithPressure(pressure float64) TouchActionPress {
	opt["pressure"] = pressure
	return opt
}

type TouchActionLongPress map[string]interface{}

func NewTouchActionLongPress() TouchActionLongPress {
	return make(map[string]interface{})
}

func (opt TouchActionLongPress) WithXY(x, y int) TouchActionLongPress {
	opt["x"] = x
	opt["y"] = y
	return opt
}

func (opt TouchActionLongPress) WithXYFloat(x, y float64) TouchActionLongPress {
	opt["x"] = x
	opt["y"] = y
	return opt
}

func (opt TouchActionLongPress) WithElement(element WebElement) TouchActionLongPress {
	opt["element"] = element.UID()
	return opt
}
