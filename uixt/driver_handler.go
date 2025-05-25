package uixt

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/rs/zerolog/log"
)

// Call custom function, used for pre/post action hook
func (dExt *XTDriver) Call(desc string, fn func(), opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)

	startTime := time.Now()
	defer func() {
		log.Info().Str("desc", desc).
			Int64("duration(ms)", time.Since(startTime).Milliseconds()).
			Msg("function called")
	}()

	if actionOptions.Timeout == 0 {
		// wait for function to finish
		fn()
		return nil
	}

	// set timeout for function execution
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
		// function completed within timeout
		return nil
	case <-time.After(time.Duration(actionOptions.Timeout) * time.Second):
		return fmt.Errorf("function execution exceeded timeout of %d seconds", actionOptions.Timeout)
	}
}

func preHandler_TapAbsXY(driver IDriver, options *option.ActionOptions, rawX, rawY float64) (
	x, y float64, err error) {

	x, y = options.ApplyTapOffset(rawX, rawY)

	// mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, option.ACTION_TapAbsXY, []float64{x, y}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark tap operation")
		}
	}

	return x, y, nil
}

func preHandler_DoubleTap(driver IDriver, options *option.ActionOptions, rawX, rawY float64) (
	x, y float64, err error) {

	x, y, err = convertToAbsolutePoint(driver, rawX, rawY)
	if err != nil {
		return 0, 0, err
	}

	x, y = options.ApplyTapOffset(x, y)

	// mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, option.ACTION_DoubleTapXY, []float64{x, y}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark double tap operation")
		}
	}

	return x, y, nil
}

func preHandler_Drag(driver IDriver, options *option.ActionOptions, rawFomX, rawFromY, rawToX, rawToY float64) (
	fromX, fromY, toX, toY float64, err error) {

	fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(driver, rawFomX, rawFromY, rawToX, rawToY)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	fromX, fromY, toX, toY = options.ApplySwipeOffset(fromX, fromY, toX, toY)

	// mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, option.ACTION_Drag, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark drag operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}

func preHandler_Swipe(driver IDriver, actionType option.ActionMethod,
	options *option.ActionOptions, rawFomX, rawFromY, rawToX, rawToY float64) (
	fromX, fromY, toX, toY float64, err error) {

	fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(driver, rawFomX, rawFromY, rawToX, rawToY)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	fromX, fromY, toX, toY = options.ApplySwipeOffset(fromX, fromY, toX, toY)

	// save screenshot before action and mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, actionType, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark swipe operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}

func postHandler(driver IDriver, actionType option.ActionMethod, options *option.ActionOptions) error {
	// save screenshot after action
	if options.PostMarkOperation {
		// get compressed screenshot buffer
		compressBufSource, err := getScreenShotBuffer(driver)
		if err != nil {
			return err
		}

		// save compressed screenshot to file
		timestamp := builtin.GenNameWithTimestamp("%d")
		imagePath := filepath.Join(
			config.GetConfig().ScreenShotsPath,
			fmt.Sprintf("action_%s_post_%s.png", timestamp, actionType),
		)

		go func() {
			err := saveScreenShot(compressBufSource, imagePath)
			if err != nil {
				log.Error().Err(err).Msg("save screenshot file failed")
			}
		}()
	}
	return nil
}
