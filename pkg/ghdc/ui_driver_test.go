package ghdc

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var (
	client Client
	device Device
	driver *UIDriver
)

func setUp(t *testing.T) {
	var err error
	client, err = NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := client.DeviceList()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("not found available device")
	}
	device = devices[0]
	driver, err = NewUIDriver(device)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_Touch(t *testing.T) {
	setUp(t)
	err := driver.Touch(1038, 798)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_Drag(t *testing.T) {
	setUp(t)
	err := driver.Drag(800, 1000, 200, 1000, 0.2)
	if err != nil {
		t.Fatal(err)
	}
}

type CaptureScreenCallback struct {
	count int
	mux   sync.Mutex
}

// OnData handles the data received
func (cb *CaptureScreenCallback) OnData(data []byte) {
	cb.mux.Lock()
	cb.count++
	cb.mux.Unlock()
	fmt.Printf("Data received: %s\n", string(data))
}

// onError handles the error received
func (cb *CaptureScreenCallback) OnError(err error) {
	fmt.Printf("Error received: %v\n", err)
}

func (cb *CaptureScreenCallback) startCounter() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cb.mux.Lock()
		fmt.Printf("Screen Data received count in the last second: %d\n", cb.count)
		cb.count = 0
		cb.mux.Unlock()
	}
}

func TestDevice_StartCaptureScreen(t *testing.T) {
	setUp(t)
	err := driver.StopCaptureScreen()
	if err != nil {
		t.Fatal(err)
	}
	callback := &CaptureScreenCallback{}
	err = driver.StartCaptureScreen(callback)
	if err != nil {
		t.Fatal(err)
	}
	callback.startCounter()
	time.Sleep(1 * time.Minute)
}

func TestDevice_StartCaptureUIAction(t *testing.T) {
	setUp(t)
	err := driver.StopCaptureUiAction()
	if err != nil {
		t.Fatal(err)
	}
	callback := &CaptureScreenCallback{}
	err = driver.StartCaptureUiAction(callback)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Minute)
}

func TestDevice_TouchDownMoveUp(t *testing.T) {
	setUp(t)

	for i := 0; i < 10000; i++ {
		// time.Sleep(1 * time.Second)
		debugLog(fmt.Sprintf("running... time %d", i))
		err := driver.TouchDown(225, 1700)
		if err != nil {
			debugLog(err.Error())
			// t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
		err = driver.TouchMove(325, 1700)
		if err != nil {
			debugLog(err.Error())
			// t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
		err = driver.TouchMove(425, 1700)
		if err != nil {
			debugLog(err.Error())
			// t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
		err = driver.TouchMove(525, 1700)
		if err != nil {
			debugLog(err.Error())
			// t.Fatal(err)
		}
		err = driver.TouchUp(625, 1700)
		if err != nil {
			debugLog(err.Error())
			// t.Fatal(err)
		}
	}
}

func TestDevice_TouchDownUp(t *testing.T) {
	setUp(t)
	err := driver.TouchDown(200, 2000)
	if err != nil {
		debugLog(err.Error())
	}
	err = driver.TouchUp(200, 2000)
	if err != nil {
		debugLog(err.Error())
		t.Fatal(err)
	}
}

func TestDevice_TouchDownMoveUpAsync(t *testing.T) {
	setUp(t)
	driver.TouchDown(225, 1700)
	time.Sleep(60 * time.Millisecond)
	driver.TouchMoveAsync(325, 1700)
	time.Sleep(60 * time.Millisecond)
	driver.TouchMoveAsync(425, 1700)
	time.Sleep(60 * time.Millisecond)
	driver.TouchMoveAsync(525, 1700)
	time.Sleep(60 * time.Millisecond)
	driver.TouchUpAsync(625, 1700)
	time.Sleep(4 * time.Second)
}

func TestDevice_GetDisplay(t *testing.T) {
	setUp(t)
	display, err := driver.GetDisplaySize()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(display)
}

func TestDevice_GetRotation(t *testing.T) {
	setUp(t)
	rotation, err := driver.GetDisplayRotation()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rotation)
}

func TestDevice_PressRecentApp(t *testing.T) {
	setUp(t)
	err := driver.PressRecentApp()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_PressBack(t *testing.T) {
	setUp(t)
	err := driver.PressBack()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_PressPowerKey(t *testing.T) {
	setUp(t)
	err := driver.PressPowerKey()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_UpVolume(t *testing.T) {
	setUp(t)
	err := driver.UpVolume()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_DownVolume(t *testing.T) {
	setUp(t)
	err := driver.DownVolume()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_CaptureLayout(t *testing.T) {
	setUp(t)
	layout, err := driver.CaptureLayout()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(layout)
}

func TestDevice_InputText(t *testing.T) {
	setUp(t)
	err := driver.InputText("abcdef")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_InjectPoint(t *testing.T) {
	setUp(t)
	err := driver.InjectGesture(NewGesture().Start(Point{800, 2000}).MoveTo(Point{200, 2000}, 2000))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)
}

func TestDevice_PressKey(t *testing.T) {
	setUp(t)
	err := driver.PressKey(KEYCODE_NUM_1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_PressKeys(t *testing.T) {
	setUp(t)
	err := driver.PressKeys([]KeyCode{KEYCODE_SHIFT_LEFT, KEYCODE_NUM_1})
	if err != nil {
		t.Fatal(err)
	}
}
