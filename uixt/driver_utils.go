package uixt

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func convertToAbsoluteScope(driver IDriver, opts ...option.ActionOption) []option.ActionOption {
	actionOptions := option.NewActionOptions(opts...)

	// convert relative scope to absolute scope
	if len(actionOptions.AbsScope) != 4 && len(actionOptions.Scope) == 4 {
		scope := actionOptions.Scope
		x1, y1, x2, y2, err := convertToAbsoluteCoordinates(
			driver, scope[0], scope[1], scope[2], scope[3])
		if err != nil {
			log.Error().Err(err).Msg("convert absolute scope failed")
			return opts
		}
		actionOptions.AbsScope = []int{int(x1), int(y1), int(x2), int(y2)}
	}

	return actionOptions.Options()
}

func convertToAbsolutePoint(driver IDriver, x, y float64) (absX, absY float64, err error) {
	// absolute coordinates
	if x > 1 || y > 1 {
		return x, y, nil
	}

	// relative coordinates
	if assertRelative(x) && assertRelative(y) {
		windowSize, err := driver.WindowSize()
		if err != nil {
			err = errors.Wrap(code.DeviceGetInfoError, err.Error())
			return 0, 0, err
		}

		absX = builtin.RoundToOneDecimal(float64(windowSize.Width) * x)
		absY = builtin.RoundToOneDecimal(float64(windowSize.Height) * y)
		return absX, absY, nil
	}

	// invalid coordinates
	err = errors.Wrap(code.InvalidCaseError,
		fmt.Sprintf("invalid coordinates x(%f), y(%f)", x, y))
	return
}

func convertToAbsoluteCoordinates(driver IDriver, fromX, fromY, toX, toY float64) (
	absFromX, absFromY, absToX, absToY float64, err error,
) {
	// absolute coordinates
	if fromX > 1 || toX > 1 || fromY > 1 || toY > 1 {
		return fromX, fromY, toX, toY, nil
	}

	// relative coordinates
	if assertRelative(fromX) && assertRelative(fromY) &&
		assertRelative(toX) && assertRelative(toY) {
		windowSize, err := driver.WindowSize()
		if err != nil {
			err = errors.Wrap(code.DeviceGetInfoError, err.Error())
			return 0, 0, 0, 0, err
		}
		width := windowSize.Width
		height := windowSize.Height

		absFromX = float64(width) * fromX
		absFromY = float64(height) * fromY
		absToX = float64(width) * toX
		absToY = float64(height) * toY

		return absFromX, absFromY, absToX, absToY, nil
	}

	// invalid coordinates
	err = errors.Wrap(code.InvalidCaseError,
		fmt.Sprintf("invalid coordinates fromX(%f), fromY(%f), toX(%f), toY(%f)",
			fromX, fromY, toX, toY))
	return
}

func assertRelative(p float64) bool {
	return p >= 0 && p <= 1
}

func (dExt *XTDriver) Setup() error {
	// unlock device screen
	err := dExt.Unlock()
	if err != nil {
		log.Error().Err(err).Msg("unlock device screen failed")
		return err
	}

	return nil
}

func (dExt *XTDriver) assertOCR(text, assert string) error {
	var opts []option.ActionOption
	opts = append(opts, option.WithScreenShotFileName(fmt.Sprintf("assert_ocr_%s", text)))

	switch assert {
	case option.AssertionEqual:
		_, err := dExt.FindScreenText(text, opts...)
		if err != nil {
			return errors.Wrap(err, "assert ocr equal failed")
		}
	case option.AssertionNotEqual:
		_, err := dExt.FindScreenText(text, opts...)
		if err == nil {
			return errors.New("assert ocr not equal failed")
		}
	case option.AssertionExists:
		opts = append(opts, option.WithRegex(true))
		_, err := dExt.FindScreenText(text, opts...)
		if err != nil {
			return errors.Wrap(err, "assert ocr exists failed")
		}
	case option.AssertionNotExists:
		opts = append(opts, option.WithRegex(true))
		_, err := dExt.FindScreenText(text, opts...)
		if err == nil {
			return errors.New("assert ocr not exists failed")
		}
	default:
		return fmt.Errorf("unexpected assert method %s", assert)
	}
	return nil
}

func (dExt *XTDriver) assertForegroundApp(appName, assert string) error {
	app, err := dExt.ForegroundInfo()
	if err != nil {
		log.Warn().Err(err).Msg("get foreground app failed, skip app assertion")
		return nil // Notice: ignore error when get foreground app failed
	}

	switch assert {
	case option.AssertionEqual:
		if app.PackageName != appName {
			return errors.Wrap(err, "assert foreground app equal failed")
		}
	case option.AssertionNotEqual:
		if app.PackageName == appName {
			return errors.New("assert foreground app not equal failed")
		}
	default:
		return fmt.Errorf("unexpected assert method %s", assert)
	}
	return nil
}

func (dExt *XTDriver) assertSelector(selector, assert string) error {
	driver, ok := dExt.IDriver.(*BrowserDriver)
	if !ok {
		return errors.New("assert selector only supports browser driver")
	}
	switch assert {
	case option.AssertionExists:
		_, err := driver.IsElementExistBySelector(selector)
		if err != nil {
			return errors.Wrap(err, "assert ocr exists failed")
		}
	case option.AssertionNotExists:
		_, err := driver.IsElementExistBySelector(selector)
		if err == nil {
			return errors.New("assert ocr not exists failed")
		}
	default:
		return fmt.Errorf("unexpected assert method %s", assert)
	}
	return nil
}

func (dExt *XTDriver) DoValidation(check, assert, expected string, message ...string) (aiResult *AIExecutionResult, err error) {
	switch check {
	case option.SelectorOCR:
		err = dExt.assertOCR(expected, assert)
	case option.SelectorAI:
		aiResult, err = dExt.AIAssert(expected)
	case option.SelectorForegroundApp:
		err = dExt.assertForegroundApp(expected, assert)
	case option.SelectorSelector:
		err = dExt.assertSelector(expected, assert)
	default:
		return nil, fmt.Errorf("validator %s not implemented", check)
	}

	if err != nil {
		// Technical error (not assertion failure)
		if message == nil {
			message = []string{""}
		}
		log.Error().Err(err).Str("assert", assert).Str("expect", expected).
			Str("msg", message[0]).Msg("validate failed")
		return nil, err
	} else if aiResult != nil {
		// Check assertion result instead of relying on error
		if !aiResult.AssertionResult.Pass {
			return aiResult, errors.New(aiResult.AssertionResult.Thought)
		}
		log.Info().Str("check", check).Str("assert", assert).
			Str("expect", expected).
			Interface("ai_assertion_result", aiResult.AssertionResult).
			Msg("ai assertion passed")
		return aiResult, nil
	} else {
		log.Info().Str("check", check).Str("assert", assert).
			Str("expect", expected).Msg("validate success")
		return nil, nil
	}
}

type SleepConfig struct {
	StartTime    time.Time `json:"start_time"`
	Seconds      float64   `json:"seconds,omitempty"`
	Milliseconds int64     `json:"milliseconds,omitempty"`
}

// getSimulationDuration returns simulation duration by given params (in seconds)
func getSimulationDuration(params []float64) (milliseconds int64) {
	if len(params) == 1 {
		// given constant duration time
		return int64(params[0] * 1000)
	}

	if len(params) == 2 {
		// given [min, max], missing weight
		// append default weight 1
		params = append(params, 1.0)
	}

	var sections []struct {
		min, max, weight float64
	}
	totalProb := 0.0
	for i := 0; i+3 <= len(params); i += 3 {
		min := params[i]
		max := params[i+1]
		weight := params[i+2]
		totalProb += weight
		sections = append(sections,
			struct{ min, max, weight float64 }{min, max, weight},
		)
	}

	if totalProb == 0 {
		log.Warn().Msg("total weight is 0, skip simulation")
		return 0
	}

	r := rand.Float64()
	accProb := 0.0
	for _, s := range sections {
		accProb += s.weight / totalProb
		if r < accProb {
			milliseconds := int64((s.min + rand.Float64()*(s.max-s.min)) * 1000)
			log.Info().Int64("random(ms)", milliseconds).
				Interface("strategy_params", params).Msg("get simulation duration")
			return milliseconds
		}
	}

	log.Warn().Interface("strategy_params", params).
		Msg("get simulation duration failed, skip simulation")
	return 0
}

// sleepStrict sleeps for strict duration with optional start time correction
// If startTime is zero, acts as normal context-aware sleep
// If startTime is provided, corrects sleep duration by subtracting elapsed time
// ctx allows for cancellation during sleep
func sleepStrict(ctx context.Context, startTime time.Time, strictMilliseconds int64) {
	var elapsed int64
	if !startTime.IsZero() {
		elapsed = time.Since(startTime).Milliseconds()
	}
	dur := strictMilliseconds - elapsed

	// if elapsed time is greater than given duration, skip sleep to reduce deviation caused by process time
	if dur <= 0 {
		log.Warn().
			Int64("elapsed(ms)", elapsed).
			Int64("strictSleep(ms)", strictMilliseconds).
			Msg("elapsed >= simulation duration, skip sleep")
		return
	}

	log.Info().Int64("sleepDuration(ms)", dur).
		Int64("elapsed(ms)", elapsed).
		Int64("strictSleep(ms)", strictMilliseconds).
		Msg("sleep remaining duration time")

	// Use context-aware sleep instead of blocking time.Sleep
	select {
	case <-time.After(time.Duration(dur) * time.Millisecond):
		// Normal completion
		log.Debug().Int64("duration_ms", dur).Msg("strict sleep completed normally")
	case <-ctx.Done():
		// Interrupted by context cancellation (e.g., CTRL+C)
		log.Info().Int64("planned_duration_ms", dur).
			Msg("strict sleep interrupted by context cancellation")
		return
	}
}

// global file lock
var (
	fileLocks sync.Map
)

func DownloadFileByUrl(fileUrl string) (filePath string, err error) {
	hash := md5.Sum([]byte(fileUrl))
	fileName := fmt.Sprintf("%x", hash)
	filePath = filepath.Join(config.GetConfig().DownloadsPath(), fileName)

	// get or create file lock
	lockI, _ := fileLocks.LoadOrStore(filePath, &sync.Mutex{})
	lock := lockI.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	if builtin.FileExists(filePath) {
		return filePath, nil
	}

	log.Info().Str("fileUrl", fileUrl).Str("filePath", filePath).Msg("downloading file")

	// Create an HTTP client with default settings.
	client := &http.Client{}

	// Build the HTTP GET request.
	req, err := http.NewRequest("GET", fileUrl, nil)
	if err != nil {
		return "", errors.Wrap(code.NetworkError, err.Error())
	}

	// Perform the request.
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(code.NetworkError, err.Error())
	}
	defer resp.Body.Close()

	// Check the HTTP status code.
	if resp.StatusCode != http.StatusOK {
		return "", errors.Wrap(code.NetworkError, fmt.Errorf("failed to download file: %s", resp.Status).Error())
	}

	// Create the output file.
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", errors.Wrap(code.MobileUIDriverError, err.Error())
	}
	defer outFile.Close()

	// Copy the response body to the file.
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", errors.Wrap(code.NetworkError, err.Error())
	}

	log.Info().Str("filePath", filePath).Msg("download file success")
	return filePath, nil
}

var (
	VEDEM_UPLOAD_URL        = os.Getenv("VEDEM_UPLOAD_URL")
	VEDEM_UPLOAD_ACCESS_KEY = os.Getenv("VEDEM_UPLOAD_ACCESS_KEY")
	VEDEM_UPLOAD_TOKEN      = os.Getenv("VEDEM_UPLOAD_TOKEN")
)

// uploadScreenshot uploads a screenshot to the server and returns the URL
func uploadScreenshot(imagePath string, imageBuffer *bytes.Buffer) (string, error) {
	if VEDEM_UPLOAD_URL == "" || VEDEM_UPLOAD_ACCESS_KEY == "" || VEDEM_UPLOAD_TOKEN == "" {
		return "", errors.Wrap(code.ConfigureError, "upload service env not configured")
	}

	// Create a new buffer for the multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create a form file field
	fileField, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return "", errors.Wrap(err, "failed to create form file")
	}

	// Copy the image buffer to the form file field
	if _, err := io.Copy(fileField, bytes.NewReader(imageBuffer.Bytes())); err != nil {
		return "", errors.Wrap(err, "failed to copy image data")
	}

	// Close the multipart writer
	if err := writer.Close(); err != nil {
		return "", errors.Wrap(err, "failed to close multipart writer")
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", VEDEM_UPLOAD_URL, &requestBody)
	if err != nil {
		return "", errors.Wrap(code.UploadFailed, err.Error())
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("accessKey", VEDEM_UPLOAD_ACCESS_KEY)
	req.Header.Set("token", VEDEM_UPLOAD_TOKEN)

	// Create HTTP client with HTTP/1.1 support
	client := &http.Client{
		Transport: &http.Transport{
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		},
	}

	// Send the request
	log.Debug().Str("url", VEDEM_UPLOAD_URL).Str("imagePath", imagePath).Msg("uploading screenshot")
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(code.UploadFailed, err.Error())
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(code.UploadFailed, err.Error())
	}

	// Parse the response JSON
	var result struct {
		StatusCode int         `json:"StatusCode"`
		Data       interface{} `json:"Data"`
		URL        string      `json:"URL"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Warn().Err(err).Str("response", string(respBody)).Msg("failed to parse upload response")
		return "", errors.Wrap(code.UploadFailed, "failed to parse response JSON")
	}

	// Check if the upload was successful
	if result.StatusCode != 0 {
		return "", fmt.Errorf("upload failed with status code: %d", result.StatusCode)
	}

	log.Debug().Str("url", result.URL).Msg("screenshot uploaded successfully")
	return result.URL, nil
}
