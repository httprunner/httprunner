package uixt

import (
	"time"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/rs/zerolog/log"
)

// Call custom function, used for pre/post hook for actions
func (dExt *XTDriver) Call(desc string, fn func()) error {
	startTime := time.Now()
	fn()
	log.Info().Str("desc", desc).
		Int64("duration(ms)", time.Since(startTime).Milliseconds()).
		Msg("function called")

	return nil
}

func preHandler_TapAbsXY(driver IDriver, options *option.ActionOptions, rawX, rawY float64) (
	x, y float64, err error) {

	if options.PreHook != nil {
		options.PreHook()
	}

	x, y = options.ApplyTapOffset(rawX, rawY)

	// mark UI operation
	if options.MarkOperationEnabled {
		if markErr := MarkUIOperation(driver, ACTION_TapAbsXY, []float64{x, y}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark tap operation")
		}
	}

	return x, y, nil
}

func preHandler_DoubleTap(driver IDriver, options *option.ActionOptions, rawX, rawY float64) (
	x, y float64, err error) {

	if options.PreHook != nil {
		options.PreHook()
	}

	x, y, err = convertToAbsolutePoint(driver, rawX, rawY)
	if err != nil {
		return 0, 0, err
	}

	x, y = options.ApplyTapOffset(x, y)

	// mark UI operation
	if options.MarkOperationEnabled {
		if markErr := MarkUIOperation(driver, ACTION_DoubleTapXY, []float64{x, y}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark double tap operation")
		}
	}

	return x, y, nil
}

func preHandler_Drag(driver IDriver, options *option.ActionOptions, rawFomX, rawFromY, rawToX, rawToY float64) (
	fromX, fromY, toX, toY float64, err error) {

	if options.PreHook != nil {
		options.PreHook()
	}

	fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(driver, rawFomX, rawFromY, rawToX, rawToY)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	fromX, fromY, toX, toY = options.ApplySwipeOffset(fromX, fromY, toX, toY)

	// mark UI operation
	if options.MarkOperationEnabled {
		if markErr := MarkUIOperation(driver, ACTION_Drag, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark drag operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}

func preHandler_Swipe(driver IDriver, options *option.ActionOptions, rawFomX, rawFromY, rawToX, rawToY float64) (
	fromX, fromY, toX, toY float64, err error) {

	if options.PreHook != nil {
		options.PreHook()
	}

	fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(driver, rawFomX, rawFromY, rawToX, rawToY)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	fromX, fromY, toX, toY = options.ApplySwipeOffset(fromX, fromY, toX, toY)

	// mark UI operation
	if options.MarkOperationEnabled {
		if markErr := MarkUIOperation(driver, ACTION_Swipe, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark swipe operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}

func preHandler_AppLaunch(_ IDriver, options *option.ActionOptions) (err error) {
	if options.PreHook != nil {
		options.PreHook()
	}

	return nil
}

func preHandler_AppTerminate(_ IDriver, options *option.ActionOptions) (err error) {
	if options.PreHook != nil {
		options.PreHook()
	}

	return nil
}

func postHandler(_ IDriver, options *option.ActionOptions) {
	if options.PostHook != nil {
		options.PostHook()
	}
}
