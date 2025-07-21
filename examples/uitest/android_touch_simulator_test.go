package uitest

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/httprunner/httprunner/v5/uixt"
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

func TestAndroidTouchByEvents(t *testing.T) {
	device, err := uixt.NewAndroidDevice(
		option.WithSerialNumber(""),
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := uixt.NewUIA2Driver(device)
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
