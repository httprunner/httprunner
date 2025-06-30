package option

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnifiedActionRequest_Options(t *testing.T) {
	// Test TapXY request conversion
	unifiedReq := &ActionOptions{
		Platform:      "android",
		Serial:        "device123",
		X:             0.5,
		Y:             0.7,
		Duration:      1.0,
		MaxRetryTimes: 3,
		ScreenOptions: ScreenOptions{
			ScreenFilterOptions: ScreenFilterOptions{
				Regex: true,
			},
		},
	}

	actionOpts := unifiedReq.Options()

	assert.Equal(t, 1.0, unifiedReq.Duration)
	assert.Equal(t, 3, unifiedReq.MaxRetryTimes)
	assert.True(t, unifiedReq.Regex)
	assert.NotEmpty(t, actionOpts)
}

func TestUnifiedActionRequest_GetMCPOptions(t *testing.T) {
	unifiedReq := &ActionOptions{
		Platform: "android",
		Serial:   "device123",
	}

	// Test TapXY options
	tapOptions := unifiedReq.GetMCPOptions(ACTION_TapXY)
	assert.NotEmpty(t, tapOptions)

	// Test TapByOCR options
	ocrOptions := unifiedReq.GetMCPOptions(ACTION_TapByOCR)
	assert.NotEmpty(t, ocrOptions)

	// Test unknown action (should return empty options)
	unknownOptions := unifiedReq.GetMCPOptions("unknown_action")
	assert.Empty(t, unknownOptions)
}

func TestUnifiedActionRequest_SwipeDirection(t *testing.T) {
	unifiedReq := &ActionOptions{
		Platform:      "android",
		Serial:        "device123",
		Direction:     "up",
		Duration:      2.0,
		PressDuration: 0.5,
	}

	opts := unifiedReq.Options()
	assert.Equal(t, "up", unifiedReq.Direction)
	assert.Equal(t, 2.0, unifiedReq.Duration)
	assert.Equal(t, 0.5, unifiedReq.PressDuration)
	assert.NotEmpty(t, opts)
}

func TestUnifiedActionRequest_SwipeCoordinate(t *testing.T) {
	params := []float64{0.2, 0.8, 0.2, 0.2}

	unifiedReq := &ActionOptions{
		Platform:  "android",
		Serial:    "device123",
		Direction: params,
	}

	opts := unifiedReq.Options()
	assert.Equal(t, params, unifiedReq.Direction)
	assert.NotEmpty(t, opts)
}

func TestUnifiedActionRequest_ScreenOptions(t *testing.T) {
	uiTypes := []string{"button", "text"}

	unifiedReq := &ActionOptions{
		Platform: "android",
		Serial:   "device123",
		ScreenOptions: ScreenOptions{
			ScreenShotOptions: ScreenShotOptions{
				ScreenShotWithOCR:     true,
				ScreenShotWithUpload:  true,
				ScreenShotWithUITypes: uiTypes,
			},
		},
	}

	opts := unifiedReq.Options()
	assert.True(t, unifiedReq.ScreenShotWithOCR)
	assert.True(t, unifiedReq.ScreenShotWithUpload)
	assert.Equal(t, uiTypes, unifiedReq.ScreenShotWithUITypes)
	assert.NotEmpty(t, opts)
}

func TestUnifiedActionRequest_NilPointerSafety(t *testing.T) {
	// Test with nil pointers
	unifiedReq := &ActionOptions{
		Platform: "android",
		Serial:   "device123",
		// All pointer fields are nil
	}

	opts := unifiedReq.Options()
	assert.Equal(t, 0, unifiedReq.MaxRetryTimes)
	assert.Equal(t, 0.0, unifiedReq.Duration)
	assert.Equal(t, 0.0, unifiedReq.PressDuration)
	assert.False(t, unifiedReq.Regex)
	assert.False(t, unifiedReq.TapRandomRect)
	// When all fields are default values, Options() may return empty slice
	// This is expected behavior
	assert.NotNil(t, opts)
}

func TestUnifiedActionRequest_CustomOptions(t *testing.T) {
	customData := map[string]interface{}{
		"custom_key": "custom_value",
		"number":     42,
	}

	unifiedReq := &ActionOptions{
		Platform: "android",
		Serial:   "device123",
		Custom:   customData,
	}

	opts := unifiedReq.Options()
	assert.Equal(t, customData, unifiedReq.Custom)
	assert.NotEmpty(t, opts)
}

func TestUnifiedActionRequest_BasicTypeFields(t *testing.T) {
	// Test basic type fields (no longer pointers)
	unifiedReq := &ActionOptions{
		Platform: "android",
		Serial:   "device123",
		Count:    5,
		Keycode:  123,
		Delta:    10,
		Width:    800,
		Height:   600,
		TabIndex: 3,
	}

	// Test direct field access (no need for Getter methods)
	assert.Equal(t, 5, unifiedReq.Count)
	assert.Equal(t, 123, unifiedReq.Keycode)
	assert.Equal(t, 10, unifiedReq.Delta)
	assert.Equal(t, 800, unifiedReq.Width)
	assert.Equal(t, 600, unifiedReq.Height)
	assert.Equal(t, 3, unifiedReq.TabIndex)

	// Test zero value detection
	emptyReq := &ActionOptions{}
	assert.Equal(t, 0, emptyReq.Count)
	assert.Equal(t, 0, emptyReq.Keycode)
	assert.Equal(t, 0, emptyReq.Delta)
	assert.Equal(t, 0, emptyReq.Width)
	assert.Equal(t, 0, emptyReq.Height)
	assert.Equal(t, 0, emptyReq.TabIndex)
}
