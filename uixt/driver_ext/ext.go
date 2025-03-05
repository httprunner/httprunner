package driver_ext

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

var (
	_ IStubDriver = (*StubAndroidDriver)(nil)
	_ IStubDriver = (*StubIOSDriver)(nil)
	_ IStubDriver = (*StubBrowserDriver)(nil)
)

type IStubDriver interface {
	GetDriver() uixt.IDriver
	LoginNoneUI(packageName, phoneNumber, captcha, password string) (info AppLoginInfo, err error)
	LogoutNoneUI(packageName string) error
}

func NewStubXTDriver(stubDriver IStubDriver, opts ...ai.AIServiceOption) *StubXTDriver {
	services := ai.NewAIService(opts...)
	driverExt := &StubXTDriver{
		XTDriver: &uixt.XTDriver{
			IDriver:    stubDriver.GetDriver(),
			CVService:  services.ICVService,
			LLMService: services.ILLMService,
		},
		IStubDriver: stubDriver,
	}
	return driverExt
}

type StubXTDriver struct {
	*uixt.XTDriver
	IStubDriver
}

func (dExt *StubXTDriver) InstallByUrl(url string, opts ...option.InstallOption) error {
	appPath, err := uixt.DownloadFileByUrl(url)
	if err != nil {
		return err
	}
	err = dExt.Install(appPath, opts...)
	if err != nil {
		return err
	}
	return nil
}

func (dExt *StubXTDriver) Install(filePath string, opts ...option.InstallOption) error {
	if _, ok := dExt.GetDevice().(*uixt.AndroidDevice); ok {
		stopChan := make(chan struct{})
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					go func() {
						_ = dExt.TapByOCR("^(.*无视风险安装|正在扫描.*|我知道了|稍后继续|稍后提醒|继续安装|知道了|确定|继续|完成|点击继续安装|继续安装旧版本|替换|.*正在安装|安装|授权本次安装|重新安装|仍要安装|更多详情|我知道了|已了解此应用未经检测.)$", option.WithRegex(true), option.WithIgnoreNotFoundError(true))
						//_ = dExt.IDriver.TapByHierarchy("^(.*无视风险安装|正在扫描.*|我知道了|稍后继续|稍后提醒|继续安装|知道了|确定|继续|完成|点击继续安装|继续安装旧版本|替换|.*正在安装|安装|授权本次安装|重新安装|仍要安装|更多详情|我知道了|已了解此应用未经检测.)$", option.WithRegex(true), option.WithIgnoreNotFoundError(true))
					}()
				case <-stopChan:
					log.Info().Msg("install complete")
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
