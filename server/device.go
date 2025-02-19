package server

import (
	"os"
	"path"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/gadb"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/rs/zerolog/log"
)

func listDeviceHandler(c *gin.Context) {
	var deviceList []interface{}
	client, err := gadb.NewClient()
	if err == nil {
		androidDevices, err := client.DeviceList()
		if err == nil {
			for _, device := range androidDevices {
				brand, err := device.Brand()
				if err != nil {
					RenderError(c, err)
					return
				}
				model, err := device.Model()
				if err != nil {
					RenderError(c, err)
					return
				}
				version, err := device.SdkVersion()
				if err != nil {
					RenderError(c, err)
					return
				}
				deviceInfo := map[string]interface{}{
					"serial":   device.Serial(),
					"brand":    brand,
					"model":    model,
					"version":  version,
					"platform": "android",
				}
				deviceList = append(deviceList, deviceInfo)
			}
		}
	}
	iosDevices, err := ios.ListDevices()
	if err == nil {
		for _, dev := range iosDevices.DeviceList {
			device, err := uixt.NewIOSDevice(
				option.WithUDID(dev.Properties.SerialNumber))
			if err != nil {
				continue
			}
			properties := device.Properties
			err = ios.Pair(dev)
			if err != nil {
				log.Error().Err(err).Msg("failed to pair device")
				continue
			}
			version, err := ios.GetProductVersion(dev)
			if err != nil {
				continue
			}
			if version.LessThan(semver.MustParse("17.4.0")) &&
				version.GreaterThan(ios.IOS17()) {
				log.Warn().Msg("not support ios 17.0-17.3")
				continue
			}
			plist, err := ios.GetValuesPlist(dev)
			if err != nil {
				log.Error().Err(err).Msg("failed to get device info")
				continue
			}
			deviceInfo := map[string]interface{}{
				"udid":     properties.SerialNumber,
				"platform": "ios",
				"brand":    "apple",
				"model":    plist["ProductType"],
				"version":  plist["ProductVersion"],
			}
			deviceList = append(deviceList, deviceInfo)
		}
	}
	RenderSuccess(c, deviceList)
}

func createBrowserHandler(c *gin.Context) {
	var createBrowserReq CreateBrowserRequest
	if err := c.ShouldBindJSON(&createBrowserReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	browserInfo, err := uixt.CreateBrowser(createBrowserReq.Timeout)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, browserInfo)
	return
}

func deleteBrowserHandler(c *gin.Context) {
	driver, err := GetDriver(c)
	if err != nil {
		RenderError(c, err)
		return
	}
	err = driver.DeleteSession()
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func pushImageHandler(c *gin.Context) {
	var pushMediaReq PushMediaRequest
	if err := c.ShouldBindJSON(&pushMediaReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := GetDriver(c)
	if err != nil {
		return
	}
	imagePath, err := builtin.DownloadFileByUrl(pushMediaReq.ImageUrl)
	if path.Ext(imagePath) == "" {
		err = os.Rename(imagePath, imagePath+".png")
		if err != nil {
			RenderError(c, err)
			return
		}
		imagePath = imagePath + ".png"
	}
	if err != nil {
		RenderError(c, err)
		return
	}
	defer func() {
		_ = os.Remove(imagePath)
	}()
	err = driver.PushImage(imagePath)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func clearImageHandler(c *gin.Context) {
	driver, err := GetDriver(c)
	if err != nil {
		return
	}
	err = driver.ClearImages()
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func videoHandler(c *gin.Context) {
	RenderSuccess(c, "")
}
