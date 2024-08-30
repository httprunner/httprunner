package uixt

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/env"
)

// GetScreenShot takes screenshot and saves image file to $CWD/screenshots/ folder
func (dExt *DriverExt) GetScreenShot(fileName string) (raw *bytes.Buffer, path string, err error) {
	// iOS 优先使用 MJPEG 流进行截图，性能最优
	// 如果 MJPEG 流未开启，则使用 WebDriver 的截图接口
	if dExt.frame != nil {
		return dExt.frame, "", nil
	}
	if raw, err = dExt.Driver.Screenshot(); err != nil {
		log.Error().Err(err).Msg("capture screenshot data failed")
		return nil, "", err
	}

	// compress image data
	compressed, err := compressImageBuffer(raw)
	if err != nil {
		log.Error().Err(err).Msg("compress screenshot data failed")
		return nil, "", err
	}

	// save screenshot to file
	path = filepath.Join(env.ScreenShotsPath, fileName)
	path, err = saveScreenShot(compressed, path)
	if err != nil {
		log.Error().Err(err).Msg("save screenshot file failed")
		return nil, "", err
	}
	dExt.cacheStepData.screenShots = append(dExt.cacheStepData.screenShots, path)
	return compressed, path, nil
}

func compressImageBuffer(raw *bytes.Buffer) (compressed *bytes.Buffer, err error) {
	// 解码原始图像数据
	img, format, err := image.Decode(raw)
	if err != nil {
		return nil, err
	}

	// 创建一个用来保存压缩后数据的buffer
	var buf bytes.Buffer

	switch format {
	// Convert to jpeg uniformly and compress with a compression rate of 95
	case "jpeg", "png":
		jpegOptions := &jpeg.Options{Quality: 95}
		err = jpeg.Encode(&buf, img, jpegOptions)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	// 返回压缩后的图像数据
	return &buf, nil
}

// saveScreenShot saves image file with file name
func saveScreenShot(raw *bytes.Buffer, fileName string) (string, error) {
	// notice: screenshot data is a stream, so we need to copy it to a new buffer
	copiedBuffer := &bytes.Buffer{}
	if _, err := copiedBuffer.Write(raw.Bytes()); err != nil {
		log.Error().Err(err).Msg("copy screenshot buffer failed")
	}

	img, format, err := image.Decode(copiedBuffer)
	if err != nil {
		return "", errors.Wrap(err, "decode screenshot image failed")
	}

	// The default format uses jpeg for compression
	screenshotPath := filepath.Join(fmt.Sprintf("%s.%s", fileName, "jpeg"))
	file, err := os.Create(screenshotPath)
	if err != nil {
		return "", errors.Wrap(err, "create screenshot image file failed")
	}
	defer func() {
		_ = file.Close()
	}()

	switch format {
	case "jpeg", "png":
		jpegOptions := &jpeg.Options{}
		err = jpeg.Encode(file, img, jpegOptions)
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return "", errors.Wrap(err, "encode screenshot image failed")
	}

	log.Info().Str("path", screenshotPath).Msg("save screenshot file success")
	return screenshotPath, nil
}
