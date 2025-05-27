package option

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnifiedActionRequest_ToActionOptions(t *testing.T) {
	// Test TapXY request conversion
	x := 0.5
	y := 0.7
	duration := 1.0
	maxRetryTimes := 3
	regex := true

	unifiedReq := &UnifiedActionRequest{
		Platform:      "android",
		Serial:        "device123",
		X:             &x,
		Y:             &y,
		Duration:      &duration,
		MaxRetryTimes: &maxRetryTimes,
		Regex:         &regex,
	}

	actionOpts := unifiedReq.ToActionOptions()

	assert.Equal(t, 1.0, actionOpts.Duration)
	assert.Equal(t, 3, actionOpts.MaxRetryTimes)
	assert.True(t, actionOpts.Regex)
}

func TestUnifiedActionRequest_GetMCPOptions(t *testing.T) {
	unifiedReq := &UnifiedActionRequest{
		Platform: "android",
		Serial:   "device123",
	}

	// Test TapXY options
	tapOptions := unifiedReq.GetMCPOptions(ACTION_TapXY)
	assert.NotEmpty(t, tapOptions)

	// Test TapByOCR options
	ocrOptions := unifiedReq.GetMCPOptions(ACTION_TapByOCR)
	assert.NotEmpty(t, ocrOptions)

	// Test unknown action (should fallback to all fields)
	unknownOptions := unifiedReq.GetMCPOptions("unknown_action")
	assert.NotEmpty(t, unknownOptions)
}

func TestUnifiedActionRequest_SwipeDirection(t *testing.T) {
	duration := 2.0
	pressDuration := 0.5

	unifiedReq := &UnifiedActionRequest{
		Platform:      "android",
		Serial:        "device123",
		Direction:     "up",
		Duration:      &duration,
		PressDuration: &pressDuration,
	}

	actionOpts := unifiedReq.ToActionOptions()
	assert.Equal(t, "up", actionOpts.Direction)
	assert.Equal(t, 2.0, actionOpts.Duration)
	assert.Equal(t, 0.5, actionOpts.PressDuration)
}

func TestUnifiedActionRequest_SwipeCoordinate(t *testing.T) {
	params := []float64{0.2, 0.8, 0.2, 0.2}

	unifiedReq := &UnifiedActionRequest{
		Platform: "android",
		Serial:   "device123",
		Params:   params,
	}

	actionOpts := unifiedReq.ToActionOptions()
	assert.Equal(t, params, actionOpts.Direction)
}

func TestUnifiedActionRequest_ScreenOptions(t *testing.T) {
	ocrEnabled := true
	uploadEnabled := true
	uiTypes := []string{"button", "text"}

	unifiedReq := &UnifiedActionRequest{
		Platform:              "android",
		Serial:                "device123",
		ScreenShotWithOCR:     &ocrEnabled,
		ScreenShotWithUpload:  &uploadEnabled,
		ScreenShotWithUITypes: uiTypes,
	}

	actionOpts := unifiedReq.ToActionOptions()
	assert.True(t, actionOpts.ScreenShotWithOCR)
	assert.True(t, actionOpts.ScreenShotWithUpload)
	assert.Equal(t, uiTypes, actionOpts.ScreenShotWithUITypes)
}

func TestUnifiedActionRequest_NilPointerSafety(t *testing.T) {
	// Test with nil pointers
	unifiedReq := &UnifiedActionRequest{
		Platform: "android",
		Serial:   "device123",
		// All pointer fields are nil
	}

	actionOpts := unifiedReq.ToActionOptions()
	assert.Equal(t, 0, actionOpts.MaxRetryTimes)
	assert.Equal(t, 0.0, actionOpts.Duration)
	assert.Equal(t, 0.0, actionOpts.PressDuration)
	assert.False(t, actionOpts.Regex)
	assert.False(t, actionOpts.TapRandomRect)
}

func TestUnifiedActionRequest_CustomOptions(t *testing.T) {
	customData := map[string]interface{}{
		"custom_key": "custom_value",
		"number":     42,
	}

	unifiedReq := &UnifiedActionRequest{
		Platform: "android",
		Serial:   "device123",
		Custom:   customData,
	}

	actionOpts := unifiedReq.ToActionOptions()
	assert.Equal(t, customData, actionOpts.Custom)
}
