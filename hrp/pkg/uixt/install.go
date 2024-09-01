package uixt

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

type InstallOptions struct {
	Reinstall       bool
	GrantPermission bool
	Downgrade       bool
	RetryTime       int
}

type InstallOption func(o *InstallOptions)

func NewInstallOptions(options ...InstallOption) *InstallOptions {
	installOptions := &InstallOptions{}
	for _, option := range options {
		option(installOptions)
	}
	return installOptions
}

func WithReinstall(reinstall bool) InstallOption {
	return func(o *InstallOptions) {
		o.Reinstall = reinstall
	}
}

func WithGrantPermission(grantPermission bool) InstallOption {
	return func(o *InstallOptions) {
		o.GrantPermission = grantPermission
	}
}

func WithDowngrade(downgrade bool) InstallOption {
	return func(o *InstallOptions) {
		o.Downgrade = downgrade
	}
}

func WithRetryTime(retryTime int) InstallOption {
	return func(o *InstallOptions) {
		o.RetryTime = retryTime
	}
}

type InstallResult struct {
	Result    int    `json:"result"`
	ErrorCode int    `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

func (dExt *DriverExt) InstallByUrl(url string, opts *InstallOptions) error {
	// 获取当前目录
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// 将文件保存到当前目录
	appPath := filepath.Join(cwd, fmt.Sprint(time.Now().UnixNano())) // 替换为你想保存的文件名
	err = builtin.DownloadFile(appPath, url)
	if err != nil {
		return err
	}

	err = dExt.Install(appPath, opts)
	if err != nil {
		return err
	}
	return nil
}

func (dExt *DriverExt) Install(filePath string, opts *InstallOptions) error {
	if _, ok := dExt.Device.(*AndroidDevice); ok {
		stopChan := make(chan struct{})
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					actions := []TapTextAction{
						{Text: "^.*无视风险安装$", Options: []ActionOption{WithTapOffset(100, 0), WithRegex(true), WithIgnoreNotFoundError(true)}},
						{Text: "^已了解此应用未经检测.*", Options: []ActionOption{WithTapOffset(-450, 0), WithRegex(true), WithIgnoreNotFoundError(true)}},
					}
					_ = dExt.Driver.TapByTexts(actions...)
					_ = dExt.TapByOCR("^(.*无视风险安装|确定|继续|完成|点击继续安装|继续安装旧版本|替换|安装|授权本次安装|继续安装|重新安装)$", WithRegex(true), WithIgnoreNotFoundError(true))
				case <-stopChan:
					fmt.Println("Ticker stopped")
					return
				}
			}
		}()
		defer func() {
			close(stopChan)
		}()
	}

	return dExt.Device.Install(filePath, opts)
}

func (dExt *DriverExt) Uninstall(packageName string, options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)
	err := dExt.Device.Uninstall(packageName)
	if err != nil {
		log.Warn().Err(err).Msg("failed to uninstall")
	}
	if actionOptions.IgnoreNotFoundError {
		return nil
	}
	return err
}
