package option

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestMigrationHelpers(t *testing.T) {
	// Test TapRequest migration
	oldTapReq := TapRequest{
		TargetDeviceRequest: TargetDeviceRequest{
			Platform: "android",
			Serial:   "device123",
		},
		X:        0.5,
		Y:        0.7,
		Duration: 1.0,
	}

	unifiedReq := MigrateTapRequestToUnified(oldTapReq)
	require.NotNil(t, unifiedReq.X)
	require.NotNil(t, unifiedReq.Y)
	require.NotNil(t, unifiedReq.Duration)
	assert.Equal(t, 0.5, *unifiedReq.X)
	assert.Equal(t, 0.7, *unifiedReq.Y)
	assert.Equal(t, 1.0, *unifiedReq.Duration)
	assert.Equal(t, "android", unifiedReq.Platform)
	assert.Equal(t, "device123", unifiedReq.Serial)

	// Test SwipeRequest migration
	oldSwipeReq := SwipeRequest{
		TargetDeviceRequest: TargetDeviceRequest{
			Platform: "ios",
			Serial:   "device456",
		},
		Direction:     "up",
		Duration:      2.0,
		PressDuration: 0.5,
	}

	unifiedSwipeReq := MigrateSwipeRequestToUnified(oldSwipeReq)
	require.NotNil(t, unifiedSwipeReq.Duration)
	require.NotNil(t, unifiedSwipeReq.PressDuration)
	assert.Equal(t, "up", unifiedSwipeReq.Direction)
	assert.Equal(t, 2.0, *unifiedSwipeReq.Duration)
	assert.Equal(t, 0.5, *unifiedSwipeReq.PressDuration)
	assert.Equal(t, "ios", unifiedSwipeReq.Platform)
	assert.Equal(t, "device456", unifiedSwipeReq.Serial)

	// Test TapByOCRRequest migration
	oldOCRReq := TapByOCRRequest{
		TargetDeviceRequest: TargetDeviceRequest{
			Platform: "android",
			Serial:   "device789",
		},
		Text:                "登录",
		IgnoreNotFoundError: true,
		MaxRetryTimes:       3,
		Index:               1,
		Regex:               true,
		TapRandomRect:       false,
	}

	unifiedOCRReq := MigrateTapByOCRRequestToUnified(oldOCRReq)
	require.NotNil(t, unifiedOCRReq.IgnoreNotFoundError)
	require.NotNil(t, unifiedOCRReq.MaxRetryTimes)
	require.NotNil(t, unifiedOCRReq.Index)
	require.NotNil(t, unifiedOCRReq.Regex)
	require.NotNil(t, unifiedOCRReq.TapRandomRect)
	assert.Equal(t, "登录", unifiedOCRReq.Text)
	assert.True(t, *unifiedOCRReq.IgnoreNotFoundError)
	assert.Equal(t, 3, *unifiedOCRReq.MaxRetryTimes)
	assert.Equal(t, 1, *unifiedOCRReq.Index)
	assert.True(t, *unifiedOCRReq.Regex)
	assert.False(t, *unifiedOCRReq.TapRandomRect)
	assert.Equal(t, "android", unifiedOCRReq.Platform)
	assert.Equal(t, "device789", unifiedOCRReq.Serial)
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
