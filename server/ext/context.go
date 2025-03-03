package server_ext

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/driver_ext"
	"github.com/httprunner/httprunner/v5/server"
)

func (p RouterBaseMethodExt) GetDriver(c *gin.Context) (driverExt uixt.IXTDriver, err error) {
	platform := c.Param("platform")
	deviceObj, exists := c.Get("device")
	var device uixt.IDevice
	var driver driver_ext.IStubDriver
	if !exists {
		device, err = p.GetDevice(c)
		if err != nil {
			return nil, err
		}
	} else {
		device = deviceObj.(uixt.IDevice)
	}
	switch strings.ToLower(platform) {
	case "android":
		driver, err = driver_ext.NewStubAndroidDriver(device.(*uixt.AndroidDevice))
	case "ios":
		driver, err = driver_ext.NewStubIOSDriver(device.(*uixt.IOSDevice))
	case "browser":
		driver, err = driver_ext.NewStubBrowserDriver(device.(*uixt.BrowserDevice))
	}
	if err != nil {
		server.RenderErrorInitDriver(c, err)
		return
	}
	driverExt = driver_ext.NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
	c.Set("driver", driverExt)
	return driverExt, nil
}
