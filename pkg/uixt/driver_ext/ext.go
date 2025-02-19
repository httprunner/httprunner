package driver_ext

import (
	"fmt"
	"time"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

type IStubDriver interface {
	uixt.IDriver

	LoginNoneUI(packageName, phoneNumber, captcha, password string) (info AppLoginInfo, err error)
	LogoutNoneUI(packageName string) error
}

func NewXTDriver(driver uixt.IDriver, opts ...ai.AIServiceOption) *XTDriver {
	services := ai.NewAIService(opts...)
	driverExt := &XTDriver{
		XTDriver: &uixt.XTDriver{
			IDriver:    driver,
			CVService:  services.ICVService,
			LLMService: services.ILLMService,
		},
	}
	return driverExt
}

type XTDriver struct {
	*uixt.XTDriver
}

func (dExt *XTDriver) InstallByUrl(url string, opts ...option.InstallOption) error {
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

func (dExt *XTDriver) Install(filePath string, opts ...option.InstallOption) error {
	if _, ok := dExt.GetDevice().(*uixt.AndroidDevice); ok {
		stopChan := make(chan struct{})
		go func() {
			ticker := time.NewTicker(8 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					_ = dExt.TapByOCR("^(.*无视风险安装|正在扫描.*|我知道了|稍后继续|稍后提醒|继续安装|知道了|确定|继续|完成|点击继续安装|继续安装旧版本|替换|.*正在安装|安装|授权本次安装|重新安装|仍要安装|更多详情|我知道了|已了解此应用未经检测.)$", option.WithRegex(true), option.WithIgnoreNotFoundError(true))
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

	return dExt.GetDevice().Install(filePath, opts...)
}
