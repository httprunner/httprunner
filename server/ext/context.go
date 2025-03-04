package server_ext

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/driver_ext"
	"github.com/httprunner/httprunner/v5/server"
)

func (r *RouterExt) GetDriver(c *gin.Context) (driverExt *driver_ext.StubXTDriver, err error) {
	var device uixt.IDevice
	var driver driver_ext.IStubDriver
	deviceObj, exists := c.Get("device")
	if !exists {
		device, err = r.GetDevice(c)
		if err != nil {
			return nil, err
		}
	} else {
		device = deviceObj.(uixt.IDevice)
	}
	platform := c.Param("platform")
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

	driverExt = driver_ext.NewStubXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
	c.Set("driver", driverExt)
	return driverExt, nil
}
