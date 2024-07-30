package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gadb"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func NewServer(port int) error {
	router := gin.Default()
	router.GET("/ping", pingHandler)
	router.GET("/api/v1/:platform/devices", listDeviceHandler)
	router.POST("/api/v1/:platform/:serial/ui/tap", parseDeviceInfo(), tapHandler)
	router.POST("/api/v1/:platform/:serial/ui/drag", parseDeviceInfo(), dragHandler)
	router.POST("/api/v1/:platform/:serial/ui/input", parseDeviceInfo(), inputHandler)
	router.POST("/api/v1/:platform/:serial/key/unlock", parseDeviceInfo(), unlockHandler)
	router.POST("/api/v1/:platform/:serial/key/home", parseDeviceInfo(), homeHandler)
	router.POST("/api/v1/:platform/:serial/key", parseDeviceInfo(), keycodeHandler)
	router.GET("/api/v1/:platform/:serial/app/foreground", parseDeviceInfo(), foregroundAppHandler)
	router.POST("/api/v1/:platform/:serial/app/clear", parseDeviceInfo(), clearAppHandler)
	router.POST("/api/v1/:platform/:serial/app/launch", parseDeviceInfo(), launchAppHandler)
	router.POST("/api/v1/:platform/:serial/app/terminal", parseDeviceInfo(), terminalAppHandler)
	router.GET("/api/v1/:platform/:serial/screenshot", parseDeviceInfo(), screenshotHandler)
	router.GET("/api/v1/:platform/:serial/shoots/source", parseDeviceInfo(), sourceHandler)
	router.GET("/api/v1/:platform/:serial/adb/source", parseDeviceInfo(), adbSourceHandler)
	router.POST("/api/v1/:platform/:serial/shoots/login", parseDeviceInfo(), loginHandler)
	router.POST("/api/v1/:platform/:serial/shoots/logout", parseDeviceInfo(), logoutHandler)

	err := router.Run(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Err(err).Msg("failed to start http server")
		return err
	}
	return nil
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func listDeviceHandler(c *gin.Context) {
	platform := c.Param("platform")
	switch strings.ToLower(platform) {
	case "android":
		{
			client, err := gadb.NewClient()
			if err != nil {
				log.Err(err).Msg("failed to init adb client")
				c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
				c.Abort()
				return
			}
			devices, err := client.DeviceList()
			if err != nil && strings.Contains(err.Error(), "no android device found") {
				c.JSON(http.StatusOK, HttpResponse{Result: nil})
				return
			} else if err != nil {
				log.Err(err).Msg("failed to list devices")
				c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
				c.Abort()
				return
			}
			var deviceList []interface{}
			for _, device := range devices {
				brand, err := device.Brand()
				if err != nil {
					log.Err(err).Msg("failed to get device brand")
					c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
					c.Abort()
					return
				}
				model, err := device.Model()
				if err != nil {
					log.Err(err).Msg("failed to get device model")
					c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
					c.Abort()
					return
				}
				deviceInfo := map[string]interface{}{
					"serial":   device.Serial(),
					"brand":    brand,
					"model":    model,
					"platform": "android",
				}
				deviceList = append(deviceList, deviceInfo)
			}
			c.JSON(http.StatusOK, HttpResponse{Result: deviceList})
			return
		}
	default:
		{
			c.JSON(http.StatusBadRequest, HttpResponse{
				ErrorCode: InvalidParamErrorCode,
				ErrorMsg:  fmt.Sprintf(InvalidParamErrorMsg, "platform"),
			})
			c.Abort()
			return
		}
	}
}

func tapHandler(c *gin.Context) {
	var tapReq TapRequest
	if err := c.ShouldBindJSON(&tapReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}

	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	if tapReq.X < 1 && tapReq.Y < 1 {
		err := dExt.TapXY(tapReq.X, tapReq.Y, uixt.WithPressDuration(tapReq.Duration))
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to tap %f, %f", c.HandlerName(), tapReq.X, tapReq.Y))
			c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
			c.Abort()
			return
		}
	} else {
		err := dExt.TapAbsXY(tapReq.X, tapReq.Y, uixt.WithPressDuration(tapReq.Duration))
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to tap %f, %f", c.HandlerName(), tapReq.X, tapReq.Y))
			c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
			c.Abort()
			return
		}
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func dragHandler(c *gin.Context) {
	var dragReq DragRequest
	if err := c.ShouldBindJSON(&dragReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	if dragReq.FromX < 1 && dragReq.FromY < 1 && dragReq.ToX < 1 && dragReq.ToY < 1 {
		err := dExt.SwipeRelative(dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY, uixt.WithPressDuration(dragReq.Duration))
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to drag from %f, %f to %f, %f", c.HandlerName(), dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY))
			c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
			c.Abort()
			return
		}
	} else {
		err := dExt.Driver.SwipeFloat(dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY, uixt.WithPressDuration(dragReq.Duration))
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to drag from %f, %f to %f, %f", c.HandlerName(), dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY))
			c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
			c.Abort()
			return
		}
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func inputHandler(c *gin.Context) {
	var inputReq InputRequest
	if err := c.ShouldBindJSON(&inputReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.SendKeys(inputReq.Text, uixt.WithFrequency(inputReq.Frequency))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to input text %s", c.HandlerName(), inputReq.Text))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func unlockHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.Unlock()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to unlick screen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func homeHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.Homescreen()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to enter homescreen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func keycodeHandler(c *gin.Context) {
	var keycodeReq KeycodeRequest
	if err := c.ShouldBindJSON(&keycodeReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.PressKeyCode(uixt.KeyCode(keycodeReq.Keycode))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to input keycode %d", c.HandlerName(), keycodeReq.Keycode))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func foregroundAppHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	appInfo, err := dExt.Driver.GetForegroundApp()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to unlick screen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: appInfo})
}

func clearAppHandler(c *gin.Context) {
	var appClearReq AppClearRequest
	if err := c.ShouldBindJSON(&appClearReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}

	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.Clear(appClearReq.PackageName)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to unlick screen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: false, ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func launchAppHandler(c *gin.Context) {
	var appLaunchReq AppLaunchRequest
	if err := c.ShouldBindJSON(&appLaunchReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.AppLaunch(appLaunchReq.PackageName)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to launch app %s", c.HandlerName(), appLaunchReq.PackageName))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: "", ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func terminalAppHandler(c *gin.Context) {
	var appTerminalReq AppTerminalRequest
	if err := c.ShouldBindJSON(&appTerminalReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	success, err := dExt.Driver.AppTerminate(appTerminalReq.PackageName)
	if !success {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to launch app %s", c.HandlerName(), appTerminalReq.PackageName))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: "", ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func screenshotHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	raw, err := dExt.Driver.Screenshot()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get screenshot", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: "", ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, HttpResponse{Result: base64.StdEncoding.EncodeToString(raw.Bytes())})
}

func sourceHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	app, err := dExt.Driver.GetForegroundApp()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get foreground app", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: "", ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	source, err := dExt.Driver.Source(uixt.NewSourceOption().WithProcessName(app.PackageName))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get source %s", c.HandlerName(), app.PackageName))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: "", ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, HttpResponse{Result: source})
}

func adbSourceHandler(c *gin.Context) {
	deviceObj, exists := c.Get("device")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "device")})
		c.Abort()
		return
	}
	device := deviceObj.(*uixt.AndroidDevice)
	driver, err := device.NewAdbDriver()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to new adb driver", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: "", ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	source, err := driver.Source()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get adb source", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: "", ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: source})
}

func loginHandler(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}

	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.LoginNoneUI(loginReq.PackageName, loginReq.PhoneNumber, loginReq.Captcha)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to login", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: "", ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func logoutHandler(c *gin.Context) {
	var logoutReq LogoutRequest
	if err := c.ShouldBindJSON(&logoutReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}

	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", ErrorCode: InvalidParamErrorCode, ErrorMsg: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.LogoutNoneUI(logoutReq.PackageName)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to login", c.HandlerName()))
		c.JSON(http.StatusInternalServerError, HttpResponse{Result: "", ErrorCode: InternalServerErrorCode, ErrorMsg: InternalServerErrorMsg})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: true})
}

func parseDeviceInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		switch strings.ToLower(platform) {
		case "android":
			serial := c.Param("serial")
			if serial == "" {
				log.Error().Str("platform", platform).Msg(fmt.Sprintf("[%s]: serial is empty", c.HandlerName()))
				c.JSON(http.StatusBadRequest, HttpResponse{
					ErrorCode: InvalidParamErrorCode,
					ErrorMsg:  fmt.Sprintf(InvalidParamErrorMsg, "serial"),
				})
				c.Abort()
				return
			}
			device, err := uixt.NewAndroidDevice(uixt.WithSerialNumber(serial), uixt.WithShoots(true))
			if err != nil {
				log.Error().Err(err).Str("platform", platform).Str("serial", serial).Msg(fmt.Sprintf("[%s]: Device Not Found", c.HandlerName()))
				c.JSON(http.StatusBadRequest, HttpResponse{
					ErrorCode: DeviceNotFoundCode,
					ErrorMsg:  fmt.Sprintf(DeviceNotFoundMsg, serial),
				})
				c.Abort()
				return
			}
			c.Set("device", device)
			driver, err := device.NewDriver(uixt.WithDriverImageService(false), uixt.WithDriverResultFolder(false))
			if err != nil {
				log.Error().Err(err).Str("platform", platform).Str("serial", serial).Msg(fmt.Sprintf("[%s]: Failed New Driver", c.HandlerName()))
				c.JSON(http.StatusInternalServerError, HttpResponse{
					ErrorCode: InternalServerErrorCode,
					ErrorMsg:  err.Error(),
				})
				c.Abort()
				return
			}
			c.Set("driver", driver)
		default:
			c.JSON(http.StatusBadRequest, HttpResponse{
				ErrorCode: InvalidParamErrorCode,
				ErrorMsg:  fmt.Sprintf(InvalidParamErrorMsg, "platform"),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
