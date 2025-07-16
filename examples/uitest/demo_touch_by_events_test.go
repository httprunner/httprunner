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

// ParseTouchEventsOptimized parses touch event data using EventTime and DownTime
func ParseTouchEventsOptimized(data string) ([]types.TouchEvent, error) {
	lines := strings.Split(strings.TrimSpace(data), "\n")
	events := make([]types.TouchEvent, 0, len(lines))

	for lineNum, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, ",")
		event := types.TouchEvent{}
		var err error

		if len(parts) == 13 {
			// Legacy format: first field is EventTime for backward compatibility
			if event.EventTime, err = strconv.ParseInt(parts[0], 10, 64); err != nil {
				return nil, fmt.Errorf("line %d: invalid eventTime: %v", lineNum+1, err)
			}
			if event.X, err = strconv.ParseFloat(parts[1], 64); err != nil {
				return nil, fmt.Errorf("line %d: invalid x: %v", lineNum+1, err)
			}
			if event.Y, err = strconv.ParseFloat(parts[2], 64); err != nil {
				return nil, fmt.Errorf("line %d: invalid y: %v", lineNum+1, err)
			}
			if event.DeviceID, err = strconv.Atoi(parts[3]); err != nil {
				return nil, fmt.Errorf("line %d: invalid deviceId: %v", lineNum+1, err)
			}
			if event.Pressure, err = strconv.ParseFloat(parts[4], 64); err != nil {
				return nil, fmt.Errorf("line %d: invalid pressure: %v", lineNum+1, err)
			}
			if event.Size, err = strconv.ParseFloat(parts[5], 64); err != nil {
				return nil, fmt.Errorf("line %d: invalid size: %v", lineNum+1, err)
			}
			if event.RawX, err = strconv.ParseFloat(parts[6], 64); err != nil {
				return nil, fmt.Errorf("line %d: invalid rawX: %v", lineNum+1, err)
			}
			if event.RawY, err = strconv.ParseFloat(parts[7], 64); err != nil {
				return nil, fmt.Errorf("line %d: invalid rawY: %v", lineNum+1, err)
			}
			if event.DownTime, err = strconv.ParseInt(parts[8], 10, 64); err != nil {
				return nil, fmt.Errorf("line %d: invalid downTime: %v", lineNum+1, err)
			}
			// Skip parts[9] (duplicate EventTime)
			if event.ToolType, err = strconv.Atoi(parts[10]); err != nil {
				return nil, fmt.Errorf("line %d: invalid toolType: %v", lineNum+1, err)
			}
			if event.Flag, err = strconv.Atoi(parts[11]); err != nil {
				return nil, fmt.Errorf("line %d: invalid flag: %v", lineNum+1, err)
			}
			if event.Action, err = strconv.Atoi(parts[12]); err != nil {
				return nil, fmt.Errorf("line %d: invalid action: %v", lineNum+1, err)
			}
		} else {
			return nil, fmt.Errorf("line %d: expected 13 fields, got %d", lineNum+1, len(parts))
		}

		events = append(events, event)
	}

	return events, nil
}

// ValidateGestureSequence validates the touch event sequence
func ValidateGestureSequence(events []types.TouchEvent) error {
	if len(events) == 0 {
		return fmt.Errorf("empty event sequence")
	}

	downTime := events[0].DownTime
	prevEventTime := int64(0)

	for i, event := range events {
		// Check DownTime consistency
		if event.DownTime != downTime {
			return fmt.Errorf("event %d: DownTime mismatch", i)
		}

		// Check EventTime ordering
		if event.EventTime < prevEventTime {
			return fmt.Errorf("event %d: EventTime should be increasing", i)
		}
		prevEventTime = event.EventTime

		// Validate action sequence
		if i == 0 && event.Action != 0 {
			return fmt.Errorf("first event should be ACTION_DOWN")
		}
		if i == len(events)-1 && event.Action != 1 {
			return fmt.Errorf("last event should be ACTION_UP")
		}
	}

	return nil
}

func TestTouchByEventsWithOptimizedTimeSystem(t *testing.T) {
	device, err := uixt.NewAndroidDevice(
		option.WithSerialNumber(""),
	)
	if err != nil {
		t.Skip("Android device not available")
	}

	driver, err := uixt.NewUIA2Driver(device)
	if err != nil {
		t.Skip("UIA2 driver not available")
	}
	defer driver.TearDown()

	// Optimized touch event data using EventTime and DownTime
	// This represents a swipe gesture from top to bottom
	touchEventData := `111586196,401.20703,1191.3164,2,1.0,0.03529412,457.20703,1359.3164,111586196,111586196,1,0,0
111586236,402.913,1185.0792,2,1.0,0.039215688,458.913,1353.0792,111586196,111586236,1,0,2
111586250,410.60825,1164.3806,2,1.0,0.03529412,466.60825,1332.3806,111586196,111586250,1,0,2
111586270,437.7335,1093.1417,2,1.0,0.039215688,493.7335,1261.1417,111586196,111586270,1,0,2
111586287,463.5786,1018.01746,2,1.0,0.039215688,519.5786,1186.0175,111586196,111586287,1,0,2
111586304,487.56482,948.9773,2,1.0,0.03529412,543.5648,1116.9773,111586196,111586304,1,0,2
111586320,511.81476,881.6183,2,1.0,0.039215688,567.81476,1049.6183,111586196,111586320,1,0,2
111586337,543.4369,811.4982,2,1.0,0.03529412,599.4369,979.4982,111586196,111586337,1,0,2
111586354,577.1632,747.4512,2,1.0,0.039215688,633.1632,915.4512,111586196,111586354,1,0,2
111586370,610.1538,691.72034,2,1.0,0.03529412,666.1538,859.72034,111586196,111586370,1,0,2
111586387,639.1683,642.6914,2,1.0,0.03529412,695.1683,810.6914,111586196,111586387,1,0,2
111586404,658.9832,605.90857,2,1.0,0.03529412,714.9832,773.90857,111586196,111586404,1,0,2
111586420,672.21954,581.1634,2,1.0,0.03529412,728.21954,749.1634,111586196,111586420,1,0,2
111586434,680.7687,566.1778,2,1.0,0.03529412,736.7687,734.1778,111586196,111586434,1,0,2
111586450,688.0894,554.2295,2,1.0,0.03529412,744.0894,722.2295,111586196,111586450,1,0,2
111586466,694.542,544.7783,2,1.0,0.03529412,750.542,712.7783,111586196,111586466,1,0,2
111586483,700.60645,537.2637,2,1.0,0.039215688,756.60645,705.2637,111586196,111586483,1,0,2
111586500,705.08887,531.1406,2,1.0,0.039215688,761.08887,699.1406,111586196,111586500,1,0,2
111586517,708.1211,527.8008,2,1.0,0.039215688,764.1211,695.8008,111586196,111586517,1,0,2
111586533,709.43945,524.46094,2,1.0,0.039215688,765.43945,692.46094,111586196,111586533,1,0,2
111586537,709.1758,523.34766,2,1.0,0.03529412,765.1758,691.34766,111586196,111586537,1,33554432,2
111586546,709.1758,523.34766,2,1.0,0.03529412,765.1758,691.34766,111586196,111586546,1,0,1`

	// Parse touch events
	events, err := ParseTouchEventsOptimized(touchEventData)
	if err != nil {
		t.Fatalf("ParseTouchEventsOptimized failed: %v", err)
	}

	t.Logf("Parsed %d touch events", len(events))

	// Validate gesture sequence
	if err := ValidateGestureSequence(events); err != nil {
		t.Fatalf("Gesture validation failed: %v", err)
	}

	// Analyze gesture timing
	analyzeGestureTiming(t, events)

	// Execute touch events using optimized time system
	err = driver.TouchByEvents(events)
	if err != nil {
		t.Fatalf("TouchByEvents failed: %v", err)
	}

	t.Logf("Successfully executed %d touch events", len(events))
}

func TestEventTimeCalculations(t *testing.T) {
	// Test data with clear timing relationships
	testData := `111586000,100.0,100.0,2,1.0,0.5,100.0,100.0,111586000,111586000,1,0,0
111586050,120.0,120.0,2,1.0,0.5,120.0,120.0,111586000,111586050,1,0,2
111586100,140.0,140.0,2,1.0,0.5,140.0,140.0,111586000,111586100,1,0,2
111586150,160.0,160.0,2,1.0,0.5,160.0,160.0,111586000,111586150,1,0,1`

	events, err := ParseTouchEventsOptimized(testData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(events) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(events))
	}

	// Test timing calculations
	gestureDuration := events[3].EventTime - events[0].DownTime
	expectedDuration := int64(150) // 111586150 - 111586000
	if gestureDuration != expectedDuration {
		t.Errorf("Gesture duration: expected %d, got %d", expectedDuration, gestureDuration)
	}

	// Test event intervals
	expectedIntervals := []int64{50, 50, 50} // Each event is 50ms apart
	for i := 1; i < len(events); i++ {
		interval := events[i].EventTime - events[i-1].EventTime
		if interval != expectedIntervals[i-1] {
			t.Errorf("Event %d interval: expected %d, got %d", i, expectedIntervals[i-1], interval)
		}
	}

	t.Logf("Timing calculations validated successfully")
}

func TestDownTimeConsistency(t *testing.T) {
	// Test data with consistent DownTime
	testData := `100000,10.0,10.0,1,1.0,0.5,10.0,10.0,100000,100000,1,0,0
100020,20.0,20.0,1,1.0,0.5,20.0,20.0,100000,100020,1,0,2
100040,30.0,30.0,1,1.0,0.5,30.0,30.0,100000,100040,1,0,1`

	events, err := ParseTouchEventsOptimized(testData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify DownTime consistency
	expectedDownTime := int64(100000)
	for i, event := range events {
		if event.DownTime != expectedDownTime {
			t.Errorf("Event %d: DownTime expected %d, got %d", i, expectedDownTime, event.DownTime)
		}
	}

	// Verify EventTime progression
	for i := 1; i < len(events); i++ {
		if events[i].EventTime <= events[i-1].EventTime {
			t.Errorf("Event %d: EventTime should be greater than previous event", i)
		}
	}

	t.Logf("DownTime consistency validated successfully")
}

func TestCoordinateSystem(t *testing.T) {
	// Test coordinate priority: rawX/rawY should be used preferentially
	testData := `100000,50.0,50.0,1,1.0,0.5,100.0,200.0,100000,100000,1,0,0
100010,60.0,60.0,1,1.0,0.5,110.0,210.0,100000,100010,1,0,1`

	events, err := ParseTouchEventsOptimized(testData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check that rawX/rawY values are parsed correctly
	if events[0].RawX != 100.0 || events[0].RawY != 200.0 {
		t.Errorf("Event 0: expected rawX=100.0, rawY=200.0, got rawX=%.1f, rawY=%.1f",
			events[0].RawX, events[0].RawY)
	}

	if events[1].RawX != 110.0 || events[1].RawY != 210.0 {
		t.Errorf("Event 1: expected rawX=110.0, rawY=210.0, got rawX=%.1f, rawY=%.1f",
			events[1].RawX, events[1].RawY)
	}

	t.Logf("Coordinate system validation passed")
}

func analyzeGestureTiming(t *testing.T, events []types.TouchEvent) {
	if len(events) == 0 {
		return
	}

	t.Logf("=== Gesture Timing Analysis ===")
	t.Logf("Total events: %d", len(events))
	t.Logf("DownTime: %d", events[0].DownTime)
	t.Logf("First EventTime: %d", events[0].EventTime)
	t.Logf("Last EventTime: %d", events[len(events)-1].EventTime)

	// Calculate total gesture duration using DownTime and final EventTime
	totalDuration := events[len(events)-1].EventTime - events[0].DownTime
	t.Logf("Total gesture duration: %dms", totalDuration)

	// Calculate average interval between events
	if len(events) > 1 {
		totalInterval := events[len(events)-1].EventTime - events[0].EventTime
		avgInterval := totalInterval / int64(len(events)-1)
		t.Logf("Average event interval: %dms", avgInterval)
	}

	// Verify DownTime consistency
	downTime := events[0].DownTime
	consistent := true
	for i, event := range events {
		if event.DownTime != downTime {
			t.Logf("Warning: Event %d has different DownTime: %d vs %d", i, event.DownTime, downTime)
			consistent = false
		}
	}
	t.Logf("DownTime consistency: %t", consistent)

	// Action sequence analysis
	actions := make([]string, len(events))
	for i, event := range events {
		switch event.Action {
		case 0:
			actions[i] = "DOWN"
		case 1:
			actions[i] = "UP"
		case 2:
			actions[i] = "MOVE"
		default:
			actions[i] = fmt.Sprintf("UNK(%d)", event.Action)
		}
	}
	t.Logf("Action sequence: %v", actions)
	t.Logf("================================")
}
