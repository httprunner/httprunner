package uixt

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

type InstallResult struct {
	Result    int    `json:"result"`
	ErrorCode int    `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

func (dExt *XTDriver) InstallByUrl(url string, opts ...option.InstallOption) error {
	// 获取当前目录
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// 将文件保存到当前目录
	appPath := filepath.Join(cwd, fmt.Sprint(time.Now().UnixNano())) // 替换为你想保存的文件名
	err = builtin.DownloadFile(appPath, url)
	if err != nil {
		log.Error().Err(err).Msg("download file failed")
		return err
	}

	err = dExt.Install(appPath, opts...)
	if err != nil {
		log.Error().Err(err).Msg("install app failed")
		return err
	}
	return nil
}

func (dExt *XTDriver) Install(filePath string, opts ...option.InstallOption) error {
	if _, ok := dExt.GetDevice().(*AndroidDevice); ok {
		stopChan := make(chan struct{})
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					actions := []TapTextAction{
						{
							Text: "^.*无视风险安装$",
							Options: []option.ActionOption{
								option.WithTapOffset(100, 0),
								option.WithRegex(true),
								option.WithIgnoreNotFoundError(true),
							},
						},
						{
							Text: "^已了解此应用未经检测.*",
							Options: []option.ActionOption{
								option.WithTapOffset(-450, 0),
								option.WithRegex(true),
								option.WithIgnoreNotFoundError(true),
							},
						},
					}
					_ = dExt.TapByTexts(actions...)

					_ = dExt.TapByOCR(
						"^(.*无视风险安装|确定|继续|完成|点击继续安装|继续安装旧版本|替换|授权本次安装|稍后提醒|继续安装|重新安装|安装)$",
						option.WithRegex(true),
						option.WithIgnoreNotFoundError(true),
					)
				case <-stopChan:
					log.Info().Msg("Ticker stopped")
					return
				}
			}
		}()
		defer func() {
			close(stopChan)
		}()
	}

	return dExt.GetDevice().Install(filePath, opts...)
}

func (dExt *XTDriver) Uninstall(packageName string, opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)
	err := dExt.GetDevice().Uninstall(packageName)
	if err != nil {
		log.Warn().Err(err).Msg("failed to uninstall")
	}
	if actionOptions.IgnoreNotFoundError {
		return nil
	}
	return err
}
