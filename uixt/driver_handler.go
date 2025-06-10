package uixt

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/rs/zerolog/log"
)

func preHandler_TapAbsXY(driver IDriver, options *option.ActionOptions, rawX, rawY float64) (
	x, y float64, err error) {

	// Call MCP action tool if anti-risk is enabled
	if options.AntiRisk {
		arguments := getAntiRisk_SetTouchInfoList_Arguments(driver, []ai.PointF{
			{X: rawX, Y: rawY},
		})
		if arguments != nil {
			callMCPActionTool(driver, "evalpkgs",
				string(option.ACTION_SetTouchInfoList), arguments)
		}
	}

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

	// Call MCP action tool if anti-risk is enabled
	if options.AntiRisk {
		arguments := getAntiRisk_SetTouchInfoList_Arguments(driver, []ai.PointF{
			{X: fromX, Y: fromY},
			{X: toX, Y: toY},
		})
		if arguments != nil {
			callMCPActionTool(driver, "evalpkgs",
				string(option.ACTION_SetTouchInfoList), arguments)
		}
	}

	// mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, option.ACTION_Drag, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark drag operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}

func preHandler_Swipe(driver IDriver, actionType option.ActionName,
	options *option.ActionOptions, rawFomX, rawFromY, rawToX, rawToY float64) (
	fromX, fromY, toX, toY float64, err error) {

	fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(driver, rawFomX, rawFromY, rawToX, rawToY)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	fromX, fromY, toX, toY = options.ApplySwipeOffset(fromX, fromY, toX, toY)

	// Call MCP action tool if anti-risk is enabled
	if options.AntiRisk {
		arguments := getAntiRisk_SetTouchInfoList_Arguments(driver, []ai.PointF{
			{X: fromX, Y: fromY},
			{X: toX, Y: toY},
		})
		if arguments != nil {
			callMCPActionTool(driver, "evalpkgs",
				string(option.ACTION_SetTouchInfoList), arguments)
		}
	}

	// save screenshot before action and mark UI operation
	if options.PreMarkOperation {
		if markErr := MarkUIOperation(driver, actionType, []float64{fromX, fromY, toX, toY}); markErr != nil {
			log.Warn().Err(markErr).Msg("Failed to mark swipe operation")
		}
	}

	return fromX, fromY, toX, toY, nil
}

func postHandler(driver IDriver, actionType option.ActionName, options *option.ActionOptions) error {
	if options.AntiRisk && actionType == option.ACTION_SetTouchInfo {
		arguments := getAntiRisk_SetTouchInfo_Arguments(driver)
		if arguments != nil {
			callMCPActionTool(driver, "evalpkgs", string(actionType), arguments)
		}
	}

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
			config.GetConfig().ScreenShotsPath(),
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

// callMCPActionTool calls MCP tool for the given action
func callMCPActionTool(driver IDriver,
	serverName, actionType string, arguments map[string]any) {
	// Get XTDriver from cache
	dExt := getXTDriverFromCache(driver)
	if dExt == nil {
		log.Warn().Msg("XTDriver not found in cache, skipping MCP tool call")
		return
	}

	// Create a context with timeout that can be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Debug().Str("server", serverName).Str("action", actionType).
		Interface("arguments", arguments).Msg("calling MCP action tool")

	// Call MCP tool with timeout context
	result, err := dExt.CallMCPTool(ctx, serverName, actionType, arguments)
	if err != nil {
		// Classify error types for better debugging
		if ctx.Err() == context.DeadlineExceeded {
			log.Warn().Str("server", serverName).Str("action", actionType).
				Msg("MCP action tool call timeout")
		} else {
			log.Warn().Err(err).Str("server", serverName).Str("action", actionType).
				Msg("MCP action tool call failed")
		}
		return
	}

	log.Debug().Str("server", serverName).Str("action", actionType).
		Interface("result", result).Msg("MCP action tool call succeeded")
}

// getAntiRisk_SetTouchInfo_Arguments gets arguments for SetTouchInfo MCP tool
func getAntiRisk_SetTouchInfo_Arguments(driver IDriver) map[string]interface{} {
	arguments := getCommonMCPArguments(driver)
	return arguments
}

// getAntiRisk_SetTouchInfoList_Arguments gets arguments for SetTouchInfoList MCP tool
func getAntiRisk_SetTouchInfoList_Arguments(driver IDriver, points []ai.PointF) map[string]interface{} {
	arguments := getCommonMCPArguments(driver)

	pointsList := make([]map[string]float64, len(points))
	for i, point := range points {
		pointsList[i] = map[string]float64{
			"x": point.X,
			"y": point.Y,
		}
	}

	arguments["points"] = pointsList
	arguments["clean"] = true

	return arguments
}

// getCommonMCPArguments gets common arguments for MCP tools
func getCommonMCPArguments(driver IDriver) map[string]interface{} {
	arguments := make(map[string]interface{})

	device := driver.GetDevice()

	// Get device model for Android devices
	if adbDevice, ok := device.(*AndroidDevice); ok {
		// Get device model
		if deviceModel, err := adbDevice.Device.Model(); err == nil {
			arguments["deviceModel"] = deviceModel
		}

		// Get device serial number
		arguments["deviceSerial"] = adbDevice.Device.Serial()
	}

	// Get current foreground app info
	if appInfo, err := driver.ForegroundInfo(); err == nil {
		arguments["packageName"] = appInfo.PackageName
	}

	return arguments
}
