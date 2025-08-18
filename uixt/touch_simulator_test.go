//go:build localtest

package uixt

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

// ParseTouchEvents parses touch event data from comma-separated string format
func ParseTouchEvents(data string) ([]types.TouchEvent, error) {
	lines := strings.Split(strings.TrimSpace(data), "\n")
	events := make([]types.TouchEvent, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) != 13 {
			return nil, fmt.Errorf("invalid touch event data format: expected 13 fields, got %d", len(parts))
		}

		event := types.TouchEvent{}
		var err error

		if event.X, err = strconv.ParseFloat(parts[1], 64); err != nil {
			return nil, fmt.Errorf("invalid x coordinate: %v", err)
		}
		if event.Y, err = strconv.ParseFloat(parts[2], 64); err != nil {
			return nil, fmt.Errorf("invalid y coordinate: %v", err)
		}
		if event.DeviceID, err = strconv.Atoi(parts[3]); err != nil {
			return nil, fmt.Errorf("invalid device id: %v", err)
		}
		if event.Pressure, err = strconv.ParseFloat(parts[4], 64); err != nil {
			return nil, fmt.Errorf("invalid pressure: %v", err)
		}
		if event.Size, err = strconv.ParseFloat(parts[5], 64); err != nil {
			return nil, fmt.Errorf("invalid size: %v", err)
		}
		if event.RawX, err = strconv.ParseFloat(parts[6], 64); err != nil {
			return nil, fmt.Errorf("invalid raw x: %v", err)
		}
		if event.RawY, err = strconv.ParseFloat(parts[7], 64); err != nil {
			return nil, fmt.Errorf("invalid raw y: %v", err)
		}
		if event.DownTime, err = strconv.ParseInt(parts[8], 10, 64); err != nil {
			return nil, fmt.Errorf("invalid down time: %v", err)
		}
		if event.EventTime, err = strconv.ParseInt(parts[9], 10, 64); err != nil {
			return nil, fmt.Errorf("invalid event time: %v", err)
		}
		if event.ToolType, err = strconv.Atoi(parts[10]); err != nil {
			return nil, fmt.Errorf("invalid tool type: %v", err)
		}
		if event.Flag, err = strconv.Atoi(parts[11]); err != nil {
			return nil, fmt.Errorf("invalid flag: %v", err)
		}
		if event.Action, err = strconv.Atoi(parts[12]); err != nil {
			return nil, fmt.Errorf("invalid action: %v", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func TestIOSTouchByEvents(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithUDID(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := NewWDADriver(device)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.TearDown()

	// Example touch event data as provided
	touchEventData := `1752649131556,401.20703,1191.3164,2,1.0,0.03529412,457.20703,1359.3164,111586196,111586196,1,0,0
1752649131595,402.913,1185.0792,2,1.0,0.039215688,458.913,1353.0792,111586196,111586236,1,0,2
1752649131612,410.60825,1164.3806,2,1.0,0.03529412,466.60825,1332.3806,111586196,111586250,1,0,2
1752649131629,437.7335,1093.1417,2,1.0,0.039215688,493.7335,1261.1417,111586196,111586270,1,0,2
1752649131646,463.5786,1018.01746,2,1.0,0.039215688,519.5786,1186.0175,111586196,111586287,1,0,2
1752649131662,487.56482,948.9773,2,1.0,0.03529412,543.5648,1116.9773,111586196,111586304,1,0,2
1752649131679,511.81476,881.6183,2,1.0,0.039215688,567.81476,1049.6183,111586196,111586320,1,0,2
1752649131696,543.4369,811.4982,2,1.0,0.03529412,599.4369,979.4982,111586196,111586337,1,0,2
1752649131713,577.1632,747.4512,2,1.0,0.039215688,633.1632,915.4512,111586196,111586354,1,0,2
1752649131729,610.1538,691.72034,2,1.0,0.03529412,666.1538,859.72034,111586196,111586370,1,0,2
1752649131746,639.1683,642.6914,2,1.0,0.03529412,695.1683,810.6914,111586196,111586387,1,0,2
1752649131763,658.9832,605.90857,2,1.0,0.03529412,714.9832,773.90857,111586196,111586404,1,0,2
1752649131779,672.21954,581.1634,2,1.0,0.03529412,728.21954,749.1634,111586196,111586420,1,0,2
1752649131796,680.7687,566.1778,2,1.0,0.03529412,736.7687,734.1778,111586196,111586434,1,0,2
1752649131814,688.0894,554.2295,2,1.0,0.03529412,744.0894,722.2295,111586196,111586450,1,0,2
1752649131830,694.542,544.7783,2,1.0,0.03529412,750.542,712.7783,111586196,111586466,1,0,2
1752649131847,700.60645,537.2637,2,1.0,0.039215688,756.60645,705.2637,111586196,111586483,1,0,2
1752649131863,705.08887,531.1406,2,1.0,0.039215688,761.08887,699.1406,111586196,111586500,1,0,2
1752649131880,708.1211,527.8008,2,1.0,0.039215688,764.1211,695.8008,111586196,111586517,1,0,2
1752649131897,709.43945,524.46094,2,1.0,0.039215688,765.43945,692.46094,111586196,111586533,1,0,2
1752649131902,709.1758,523.34766,2,1.0,0.03529412,765.1758,691.34766,111586196,111586537,1,33554432,2
1752649131907,709.1758,523.34766,2,1.0,0.03529412,765.1758,691.34766,111586196,111586546,1,0,1`

	// Parse touch events
	events, err := ParseTouchEvents(touchEventData)
	if err != nil {
		t.Fatalf("ParseTouchEvents failed: %v", err)
	}

	// Check first event
	firstEvent := events[0]
	if firstEvent.Action != 0 { // ACTION_DOWN
		t.Errorf("Expected first event action to be 0 (ACTION_DOWN), got %d", firstEvent.Action)
	}

	// Check last event
	lastEvent := events[len(events)-1]
	if lastEvent.Action != 1 { // ACTION_UP
		t.Errorf("Expected last event action to be 1 (ACTION_UP), got %d", lastEvent.Action)
	}

	// Use TouchByEvents with parsed events
	err = driver.TouchByEvents(events)
	if err != nil {
		t.Fatalf("TouchByEvents failed: %v", err)
	}

	t.Logf("Successfully executed touch events: %d events processed", len(events))
}

func TestTouchEventParsing(t *testing.T) {
	// Test single touch event parsing
	singleEventData := "1752646457403,456.78418,1574.0195,7,1.0,0.016666668,504.78418,1721.0195,924451292,924451292,1,0,0"

	events, err := ParseTouchEvents(singleEventData)
	if err != nil {
		t.Fatalf("ParseTouchEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.X != 456.78418 {
		t.Errorf("Expected X 456.78418, got %f", event.X)
	}
	if event.Y != 1574.0195 {
		t.Errorf("Expected Y 1574.0195, got %f", event.Y)
	}
	if event.Action != 0 {
		t.Errorf("Expected Action 0, got %d", event.Action)
	}
	if event.Pressure != 1.0 {
		t.Errorf("Expected Pressure 1.0, got %f", event.Pressure)
	}
	if event.Size != 0.016666668 {
		t.Errorf("Expected Size 0.016666668, got %f", event.Size)
	}
}

func TestTouchEventParsingInvalidData(t *testing.T) {
	// Test with invalid data
	testCases := []struct {
		name string
		data string
	}{
		{
			name: "too few fields",
			data: "1752646457403,456.78418,1574.0195,7,1.0",
		},
		{
			name: "invalid timestamp",
			data: "invalid,456.78418,1574.0195,7,1.0,0.016666668,504.78418,1721.0195,924451292,924451292,1,0,0",
		},
		{
			name: "invalid x coordinate",
			data: "1752646457403,invalid,1574.0195,7,1.0,0.016666668,504.78418,1721.0195,924451292,924451292,1,0,0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseTouchEvents(tc.data)
			if err == nil {
				t.Errorf("Expected error for invalid data, but got none")
			}
		})
	}
}

func TestTouchEventSequenceValidation(t *testing.T) {
	// Test a complete touch sequence: DOWN -> MOVE -> MOVE -> UP
	sequenceData := `1752646457403,100.0,100.0,7,1.0,0.016666668,100.0,100.0,924451292,924451292,1,0,0
1752646457420,120.0,120.0,7,1.0,0.022058824,120.0,120.0,924451292,924451335,1,0,2
1752646457440,140.0,140.0,7,1.0,0.022058824,140.0,140.0,924451292,924451351,1,0,2
1752646457460,160.0,160.0,7,1.0,0.012254903,160.0,160.0,924451292,924451619,1,0,1`

	events, err := ParseTouchEvents(sequenceData)
	if err != nil {
		t.Fatalf("ParseTouchEvents failed: %v", err)
	}

	if len(events) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(events))
	}

	// Validate sequence: DOWN -> MOVE -> MOVE -> UP
	expectedActions := []int{0, 2, 2, 1} // ACTION_DOWN, ACTION_MOVE, ACTION_MOVE, ACTION_UP
	for i, event := range events {
		if event.Action != expectedActions[i] {
			t.Errorf("Event %d: expected action %d, got %d", i, expectedActions[i], event.Action)
		}
	}

	t.Logf("Touch sequence validation passed: %d events with correct action sequence", len(events))
}

func TestSwipeWithDirection(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithUDID(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := NewWDADriver(device)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.TearDown()

	// Test cases for different directions and distance configurations
	testCases := []struct {
		name        string
		direction   string
		startX      float64
		startY      float64
		minDistance float64
		maxDistance float64
	}{
		{
			name:        "随机距离上滑",
			direction:   "up",
			startX:      0.5,
			startY:      0.5,
			minDistance: 500.0,
			maxDistance: 500.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := driver.SIMSwipeWithDirection(
				tc.direction,
				tc.startX,
				tc.startY,
				tc.minDistance,
				tc.maxDistance,
			)
			if err != nil {
				t.Errorf("SwipeWithDirection failed: %v", err)
			} else {
				t.Logf("Successfully executed swipe: direction=%s, start=(%.1f,%.1f), distance=%.1f-%.1f",
					tc.direction, tc.startX, tc.startY, tc.minDistance, tc.maxDistance)
			}
		})
	}
}

func TestSwipeInArea(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithUDID(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := NewWDADriver(device)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.TearDown()

	// Test cases for different area configurations and directions
	testCases := []struct {
		name        string
		direction   string
		areaStartX  float64
		areaStartY  float64
		areaEndX    float64
		areaEndY    float64
		minDistance float64
		maxDistance float64
	}{
		{
			name:        "中心区域上滑_固定距离",
			direction:   "up",
			areaStartX:  0.2,
			areaStartY:  0.3,
			areaEndX:    0.8,
			areaEndY:    0.6,
			minDistance: 500.0,
			maxDistance: 700.0, // 固定距离
		},
	}

	for _, tc := range testCases {
		for i := 0; i < 3; i++ {
			t.Run(tc.name, func(t *testing.T) {
				err := driver.SIMSwipeInArea(
					tc.direction,
					tc.areaStartX,
					tc.areaStartY,
					tc.areaEndX,
					tc.areaEndY,
					tc.minDistance,
					tc.maxDistance,
				)
				if err != nil {
					t.Errorf("SwipeInArea failed: %v", err)
				} else {
					t.Logf("Successfully executed area swipe: direction=%s, area=(%.1f,%.1f)-(%.1f,%.1f), distance=%.1f-%.1f",
						tc.direction, tc.areaStartX, tc.areaStartY, tc.areaEndX, tc.areaEndY, tc.minDistance, tc.maxDistance)
				}
			})
		}
	}
}

func TestSwipeFromPointToPoint(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithUDID(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := NewWDADriver(device)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.TearDown()

	// Test cases for different point-to-point swipes
	testCases := []struct {
		name   string
		startX float64
		startY float64
		endX   float64
		endY   float64
	}{
		{
			name:   "对角线滑动_左上到右下",
			startX: 0.2,
			startY: 0.3,
			endX:   0.8,
			endY:   0.5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := driver.SIMSwipeFromPointToPoint(
				tc.startX,
				tc.startY,
				tc.endX,
				tc.endY,
			)
			if err != nil {
				t.Errorf("SwipeFromPointToPoint failed: %v", err)
			} else {
				t.Logf("Successfully executed point-to-point swipe: %s, from (%.1f,%.1f) to (%.1f,%.1f)",
					tc.name, tc.startX, tc.startY, tc.endX, tc.endY)
			}
		})
	}
}

func TestSwipeFromPointToPointInvalidInputs(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithUDID(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := NewWDADriver(device)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.TearDown()

	// Test same start and end point
	err = driver.SIMSwipeFromPointToPoint(0.5, 0.5, 0.5, 0.5)
	if err == nil {
		t.Error("Expected error for same start and end point, but got none")
	}

	// Test very close points (should result in distance too short)
	err = driver.SIMSwipeFromPointToPoint(0.5, 0.5, 0.501, 0.501)
	if err == nil {
		t.Error("Expected error for very close points, but got none")
	}

	t.Log("Point-to-point swipe invalid input validation tests passed")
}

func TestClickAtPoint(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithUDID(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := NewWDADriver(device)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.TearDown()

	// Test cases for different click positions
	testCases := []struct {
		name string
		x    float64
		y    float64
	}{
		{
			name: "屏幕中心点击",
			x:    0.5,
			y:    0.5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := driver.SIMClickAtPoint(tc.x, tc.y)
			if err != nil {
				t.Errorf("ClickAtPoint failed: %v", err)
			} else {
				t.Logf("Successfully executed click: %s at (%.1f, %.1f)",
					tc.name, tc.x, tc.y)
			}
		})
	}
}

func TestClickAtPointInvalidInputs(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithUDID(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := NewWDADriver(device)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.TearDown()

	// Test negative coordinates
	err = driver.SIMClickAtPoint(-0.1, 0.5)
	if err == nil {
		t.Error("Expected error for negative x coordinate, but got none")
	}

	err = driver.SIMClickAtPoint(0.5, -0.1)
	if err == nil {
		t.Error("Expected error for negative y coordinate, but got none")
	}

	// Test coordinates out of range (though these should be handled by convertToAbsolutePoint)
	err = driver.SIMClickAtPoint(1.5, 0.5)
	if err != nil {
		t.Logf("Out of range coordinates handled properly: %v", err)
	}

	t.Log("Click invalid input validation tests passed")
}

func TestSIMInput(t *testing.T) {
	device, err := NewIOSDevice(
		option.WithUDID(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := NewWDADriver(device)
	if err != nil {
		t.Fatal(err)
	}
	defer driver.TearDown()

	// Test cases for different text inputs
	testCases := []struct {
		name string
		text string
	}{
		{
			name: "长文本",
			text: "This is a very long text to test the performance of SIMInput function. 这是一个很长的文本用来测试SIMInput函数的性能。1234567890!@#$%^&*()英語の長い文字",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := driver.SIMInput(tc.text)
			// err := driver.Input(tc.text)
			if err != nil {
				t.Errorf("SIMInput failed: %v", err)
			} else {
				t.Logf("Successfully executed SIMInput: %s with text '%s'", tc.name, tc.text)
			}
		})
	}
}
