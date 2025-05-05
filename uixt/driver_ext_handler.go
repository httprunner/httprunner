package uixt

import (
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/rs/zerolog/log"
)

func handlerTapAbsXY(driver IDriver, rawX, rawY float64, opts ...option.ActionOption) (
	x, y, duration float64, err error) {

	actionOptions := option.NewActionOptions(opts...)
	x, y = actionOptions.ApplyTapOffset(rawX, rawY)

	// mark UI operation
	if actionOptions.MarkOperationEnabled {
		if markErr := MarkUIOperation(driver, ACTION_TapAbsXY, []float64{x, y}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark tap operation")
		}
	}

	duration = 100.0
	if actionOptions.PressDuration > 0 {
		duration = actionOptions.PressDuration * 1000 // convert to ms
	}

	return x, y, duration, nil
}
