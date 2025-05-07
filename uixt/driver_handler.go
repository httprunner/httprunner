package uixt

import (
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/rs/zerolog/log"
)

func handlerTapAbsXY(driver IDriver, rawX, rawY float64, opts ...option.ActionOption) (
	x, y float64, err error) {

	actionOptions := option.NewActionOptions(opts...)
	x, y = actionOptions.ApplyTapOffset(rawX, rawY)

	// mark UI operation
	if actionOptions.MarkOperationEnabled {
		if markErr := MarkUIOperation(driver, ACTION_TapAbsXY, []float64{x, y}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark tap operation")
		}
	}

	return x, y, nil
}

func handlerDoubleTap(driver IDriver, rawX, rawY float64, opts ...option.ActionOption) (
	x, y float64, err error) {

	x, y, err = convertToAbsolutePoint(driver, rawX, rawY)
	if err != nil {
		return 0, 0, err
	}

	actionOptions := option.NewActionOptions(opts...)
	x, y = actionOptions.ApplyTapOffset(x, y)

	// mark UI operation
	if actionOptions.MarkOperationEnabled {
		if markErr := MarkUIOperation(driver, ACTION_DoubleTapXY, []float64{x, y}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark double tap operation")
		}
	}

	return x, y, nil
}

func handlerDrag(driver IDriver, rawFomX, rawFromY, rawToX, rawToY float64, opts ...option.ActionOption) (
	fromX, fromY, toX, toY float64, err error) {

	actionOptions := option.NewActionOptions(opts...)
	fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(driver, rawFomX, rawFromY, rawToX, rawToY)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	fromX, fromY, toX, toY = actionOptions.ApplySwipeOffset(fromX, fromY, toX, toY)

	// mark UI operation
	if actionOptions.MarkOperationEnabled {
		if markErr := MarkUIOperation(driver, ACTION_Drag, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark drag operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}

func handlerSwipe(driver IDriver, rawFomX, rawFromY, rawToX, rawToY float64, opts ...option.ActionOption) (
	fromX, fromY, toX, toY float64, err error) {

	actionOptions := option.NewActionOptions(opts...)
	fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(driver, rawFomX, rawFromY, rawToX, rawToY)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	fromX, fromY, toX, toY = actionOptions.ApplySwipeOffset(fromX, fromY, toX, toY)

	// mark UI operation
	if actionOptions.MarkOperationEnabled {
		if markErr := MarkUIOperation(driver, ACTION_Swipe, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark swipe operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}
