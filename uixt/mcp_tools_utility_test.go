package uixt

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestToolSleep_ConvertActionToCallToolRequest(t *testing.T) {
	tool := &ToolSleep{}

	tests := []struct {
		name         string
		action       option.MobileAction
		expectedArgs map[string]any
		shouldError  bool
	}{
		{
			name: "json.Number parameter",
			action: option.MobileAction{
				Method: option.ACTION_Sleep,
				Params: json.Number("3.5"),
			},
			expectedArgs: map[string]any{"seconds": float64(3.5)},
			shouldError:  false,
		},
		{
			name: "float64 parameter",
			action: option.MobileAction{
				Method: option.ACTION_Sleep,
				Params: float64(5.2),
			},
			expectedArgs: map[string]any{"seconds": float64(5.2)},
			shouldError:  false,
		},
		{
			name: "int64 parameter",
			action: option.MobileAction{
				Method: option.ACTION_Sleep,
				Params: int64(5),
			},
			expectedArgs: map[string]any{"seconds": float64(5)},
			shouldError:  false,
		},
		{
			name: "SleepConfig with startTime",
			action: option.MobileAction{
				Method: option.ACTION_Sleep,
				Params: SleepConfig{
					StartTime: time.UnixMilli(1691234567890),
					Seconds:   2.5,
				},
			},
			expectedArgs: map[string]any{
				"seconds":       2.5,
				"start_time_ms": int64(1691234567890),
			},
			shouldError: false,
		},
		{
			name: "invalid parameter type",
			action: option.MobileAction{
				Method: option.ACTION_Sleep,
				Params: "invalid",
			},
			expectedArgs: nil,
			shouldError:  true,
		},
		{
			name: "json.Number with integer value",
			action: option.MobileAction{
				Method: option.ACTION_Sleep,
				Params: json.Number("10"),
			},
			expectedArgs: map[string]any{"seconds": float64(10)},
			shouldError:  false,
		},
		{
			name: "json.Number with decimal value",
			action: option.MobileAction{
				Method: option.ACTION_Sleep,
				Params: json.Number("1.25"),
			},
			expectedArgs: map[string]any{"seconds": float64(1.25)},
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := tool.ConvertActionToCallToolRequest(tt.action)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				args := request.GetArguments()
				for key, expectedValue := range tt.expectedArgs {
					assert.Equal(t, expectedValue, args[key], "Argument %s mismatch", key)
				}
			}
		})
	}
}

func TestToolSleepMS_ConvertActionToCallToolRequest(t *testing.T) {
	tool := &ToolSleepMS{}

	tests := []struct {
		name         string
		action       option.MobileAction
		expectedArgs map[string]any
		shouldError  bool
	}{
		{
			name: "json.Number parameter",
			action: option.MobileAction{
				Method: option.ACTION_SleepMS,
				Params: json.Number("1500"),
			},
			expectedArgs: map[string]any{"milliseconds": int64(1500)},
			shouldError:  false,
		},
		{
			name: "int64 parameter",
			action: option.MobileAction{
				Method: option.ACTION_SleepMS,
				Params: int64(2000),
			},
			expectedArgs: map[string]any{"milliseconds": int64(2000)},
			shouldError:  false,
		},
		{
			name: "float64 parameter",
			action: option.MobileAction{
				Method: option.ACTION_SleepMS,
				Params: float64(2500.7),
			},
			expectedArgs: map[string]any{"milliseconds": int64(2500)},
			shouldError:  false,
		},
		{
			name: "SleepConfig with startTime",
			action: option.MobileAction{
				Method: option.ACTION_SleepMS,
				Params: SleepConfig{
					StartTime:    time.UnixMilli(1691234567890),
					Milliseconds: 3000,
				},
			},
			expectedArgs: map[string]any{
				"milliseconds":  int64(3000),
				"start_time_ms": int64(1691234567890),
			},
			shouldError: false,
		},
		{
			name: "json.Number with decimal value",
			action: option.MobileAction{
				Method: option.ACTION_SleepMS,
				Params: json.Number("1234.56"),
			},
			expectedArgs: map[string]any{"milliseconds": int64(1234)},
			shouldError:  false,
		},
		{
			name: "invalid parameter type",
			action: option.MobileAction{
				Method: option.ACTION_SleepMS,
				Params: "invalid",
			},
			expectedArgs: nil,
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := tool.ConvertActionToCallToolRequest(tt.action)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				args := request.GetArguments()
				for key, expectedValue := range tt.expectedArgs {
					assert.Equal(t, expectedValue, args[key], "Argument %s mismatch", key)
				}
			}
		})
	}
}

func TestSleepStrictTiming(t *testing.T) {
	// Test that strict sleep properly adjusts for elapsed time
	startTime := time.Now()

	// Simulate some processing time
	time.Sleep(50 * time.Millisecond)

	ctx := context.Background()

	// Test sleepStrict with the start time
	testStart := time.Now()
	sleepStrict(ctx, startTime, 200) // 200ms total duration
	actualElapsed := time.Since(testStart)

	// Should sleep approximately 150ms (200ms - 50ms already elapsed)
	// Allow some tolerance for timing variations
	expectedSleep := 150 * time.Millisecond
	assert.Greater(t, actualElapsed, expectedSleep/2, "Sleep too short")
	assert.Less(t, actualElapsed, expectedSleep*2, "Sleep too long")
}

func TestSleepCancellation(t *testing.T) {
	// Test that sleep respects context cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	sleepStrict(ctx, time.Time{}, 500) // Try to sleep 500ms
	elapsed := time.Since(start)

	// Should be cancelled after ~50ms, not sleep full 500ms
	assert.Less(t, elapsed, 200*time.Millisecond, "Sleep was not properly cancelled")
}

func TestSleepStrictWithZeroTime(t *testing.T) {
	// Test sleepStrict behaves like normal sleep when startTime is zero
	ctx := context.Background()

	start := time.Now()
	sleepStrict(ctx, time.Time{}, 100) // 100ms, no start time
	elapsed := time.Since(start)

	// Should sleep full duration
	expectedSleep := 100 * time.Millisecond
	assert.Greater(t, elapsed, expectedSleep/2, "Sleep too short")
	assert.Less(t, elapsed, expectedSleep*2, "Sleep too long")
}

func TestSleepStrictWithPastStartTime(t *testing.T) {
	// Test sleepStrict skips sleep when elapsed time exceeds duration
	startTime := time.Now().Add(-300 * time.Millisecond) // 300ms ago
	ctx := context.Background()

	start := time.Now()
	sleepStrict(ctx, startTime, 200) // Want 200ms total, but 300ms already elapsed
	elapsed := time.Since(start)

	// Should skip sleep entirely
	assert.Less(t, elapsed, 50*time.Millisecond, "Should have skipped sleep")
}

func TestJsonNumberHandling(t *testing.T) {
	// Test that json.Number is correctly handled in different scenarios

	// Test float json.Number
	floatNumber := json.Number("3.14")
	floatVal, err := floatNumber.Float64()
	assert.NoError(t, err)
	assert.Equal(t, 3.14, floatVal)

	// Test int json.Number
	intNumber := json.Number("1500")
	intVal, err := intNumber.Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(1500), intVal)

	// Test invalid json.Number
	invalidNumber := json.Number("invalid")
	_, err = invalidNumber.Float64()
	assert.Error(t, err)
}
