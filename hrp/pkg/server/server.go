package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gadb"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func NewServer(port int) error {
	router := gin.Default()
	router.GET("/ping", pingHandler)

	apiV1Platform := router.Group("/api/v1").Group("/:platform")
	apiV1Platform.GET("/devices", listDeviceHandler)

	apiV1PlatformSerial := apiV1Platform.Group("/:serial")
	// UI operations
	apiV1PlatformSerial.POST("/ui/tap", parseDeviceInfo(), tapHandler)
	apiV1PlatformSerial.POST("/ui/drag", parseDeviceInfo(), dragHandler)
	apiV1PlatformSerial.POST("/ui/input", parseDeviceInfo(), inputHandler)
	// Key operations
	apiV1PlatformSerial.POST("/key/unlock", parseDeviceInfo(), unlockHandler)
	apiV1PlatformSerial.POST("/key/home", parseDeviceInfo(), homeHandler)
	apiV1PlatformSerial.POST("/key", parseDeviceInfo(), keycodeHandler)
	// App operations
	apiV1PlatformSerial.GET("/app/foreground", parseDeviceInfo(), foregroundAppHandler)
	apiV1PlatformSerial.POST("/app/clear", parseDeviceInfo(), clearAppHandler)
	apiV1PlatformSerial.POST("/app/launch", parseDeviceInfo(), launchAppHandler)
	apiV1PlatformSerial.POST("/app/terminal", parseDeviceInfo(), terminalAppHandler)
	// get screen info
	apiV1PlatformSerial.GET("/screenshot", parseDeviceInfo(), screenshotHandler)
	apiV1PlatformSerial.GET("/stub/source", parseDeviceInfo(), sourceHandler)
	apiV1PlatformSerial.GET("/adb/source", parseDeviceInfo(), adbSourceHandler)
	// Stub operations
	apiV1PlatformSerial.POST("/stub/login", parseDeviceInfo(), loginHandler)
	apiV1PlatformSerial.POST("/stub/logout", parseDeviceInfo(), logoutHandler)

	err := router.Run(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Err(err).Msg("failed to start http server")
		return err
	}
	return nil
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func listDeviceHandler(c *gin.Context) {
	platform := c.Param("platform")
	switch strings.ToLower(platform) {
	case "android":
		{
			client, err := gadb.NewClient()
			if err != nil {
				log.Err(err).Msg("failed to init adb client")
				c.JSON(http.StatusInternalServerError,
					HttpResponse{
						Code:    code.GetErrorCode(err),
						Message: err.Error(),
					},
				)
				c.Abort()
				return
			}
			devices, err := client.DeviceList()
			if err != nil && strings.Contains(err.Error(), "no android device found") {
				c.JSON(http.StatusOK, HttpResponse{Result: nil})
				return
			} else if err != nil {
				log.Err(err).Msg("failed to list devices")
				c.JSON(http.StatusInternalServerError,
					HttpResponse{
						Code:    code.GetErrorCode(err),
						Message: err.Error(),
					},
				)
				c.Abort()
				return
			}
			var deviceList []interface{}
			for _, device := range devices {
				brand, err := device.Brand()
				if err != nil {
					log.Err(err).Msg("failed to get device brand")
					c.JSON(http.StatusInternalServerError,
						HttpResponse{
							Code:    code.GetErrorCode(err),
							Message: err.Error(),
						},
					)
					c.Abort()
					return
				}
				model, err := device.Model()
				if err != nil {
					log.Err(err).Msg("failed to get device model")
					c.JSON(http.StatusInternalServerError,
						HttpResponse{
							Code:    code.GetErrorCode(err),
							Message: err.Error(),
						},
					)
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
				Code:    InvalidParamErrorCode,
				Message: fmt.Sprintf(InvalidParamErrorMsg, "platform"),
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
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}

	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	var actionOptions []uixt.ActionOption
	if tapReq.Options != nil {
		actionOptions = tapReq.Options.Options()
	}

	dExt := driverObj.(*uixt.DriverExt)
	if tapReq.Text != "" {
		err := dExt.TapByOCR(tapReq.Text, actionOptions...)
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to tap text %s", c.HandlerName(), tapReq.Text))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	} else if tapReq.X < 1 && tapReq.Y < 1 {
		err := dExt.TapXY(tapReq.X, tapReq.Y, actionOptions...)
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to tap %f, %f", c.HandlerName(), tapReq.X, tapReq.Y))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	} else {
		err := dExt.TapAbsXY(tapReq.X, tapReq.Y, actionOptions...)
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to tap %f, %f", c.HandlerName(), tapReq.X, tapReq.Y))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func dragHandler(c *gin.Context) {
	var dragReq DragRequest
	if err := c.ShouldBindJSON(&dragReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	if dragReq.FromX < 1 && dragReq.FromY < 1 && dragReq.ToX < 1 && dragReq.ToY < 1 {
		err := dExt.SwipeRelative(dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY, uixt.WithPressDuration(dragReq.Duration))
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to drag from %f, %f to %f, %f", c.HandlerName(), dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	} else {
		err := dExt.Driver.SwipeFloat(dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY, uixt.WithPressDuration(dragReq.Duration))
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to drag from %f, %f to %f, %f", c.HandlerName(), dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func inputHandler(c *gin.Context) {
	var inputReq InputRequest
	if err := c.ShouldBindJSON(&inputReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.SendKeys(inputReq.Text, uixt.WithFrequency(inputReq.Frequency))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to input text %s", c.HandlerName(), inputReq.Text))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func unlockHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.Unlock()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to unlick screen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func homeHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.Homescreen()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to enter homescreen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func keycodeHandler(c *gin.Context) {
	var keycodeReq KeycodeRequest
	if err := c.ShouldBindJSON(&keycodeReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.PressKeyCode(uixt.KeyCode(keycodeReq.Keycode))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to input keycode %d", c.HandlerName(), keycodeReq.Keycode))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func foregroundAppHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	appInfo, err := dExt.Driver.GetForegroundApp()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to unlick screen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: appInfo})
}

func clearAppHandler(c *gin.Context) {
	var appClearReq AppClearRequest
	if err := c.ShouldBindJSON(&appClearReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}

	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: false, Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.Clear(appClearReq.PackageName)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to unlick screen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func launchAppHandler(c *gin.Context) {
	var appLaunchReq AppLaunchRequest
	if err := c.ShouldBindJSON(&appLaunchReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.AppLaunch(appLaunchReq.PackageName)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to launch app %s", c.HandlerName(), appLaunchReq.PackageName))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func terminalAppHandler(c *gin.Context) {
	var appTerminalReq AppTerminalRequest
	if err := c.ShouldBindJSON(&appTerminalReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	success, err := dExt.Driver.AppTerminate(appTerminalReq.PackageName)
	if !success {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to launch app %s", c.HandlerName(), appTerminalReq.PackageName))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func screenshotHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	raw, err := dExt.Driver.Screenshot()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get screenshot", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, HttpResponse{Result: base64.StdEncoding.EncodeToString(raw.Bytes())})
}

func sourceHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	app, err := dExt.Driver.GetForegroundApp()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get foreground app", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	source, err := dExt.Driver.Source(uixt.NewSourceOption().WithProcessName(app.PackageName))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get source %s", c.HandlerName(), app.PackageName))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, HttpResponse{Result: source})
}

func adbSourceHandler(c *gin.Context) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}
	driver := driverObj.(*uixt.DriverExt)
	source, err := driver.Driver.Source()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get adb source", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: source})
}

func loginHandler(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}

	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.LoginNoneUI(loginReq.PackageName, loginReq.PhoneNumber, loginReq.Captcha)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to login", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func logoutHandler(c *gin.Context) {
	var logoutReq LogoutRequest
	if err := c.ShouldBindJSON(&logoutReq); err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: Invalid Request", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "request")})
		c.Abort()
		return
	}

	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg(fmt.Sprintf("[%s]: Driver Not exsit", c.HandlerName()))
		c.JSON(http.StatusBadRequest, HttpResponse{Result: "", Code: InvalidParamErrorCode, Message: fmt.Sprintf(InvalidParamErrorMsg, "driver")})
		c.Abort()
		return
	}

	dExt := driverObj.(*uixt.DriverExt)
	err := dExt.Driver.LogoutNoneUI(logoutReq.PackageName)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to login", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

var uiClients = make(map[string]*uixt.DriverExt) // UI automation clients for iOS and Android, key is udid/serial

func parseDeviceInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		serial := c.Param("serial")
		if serial == "" {
			log.Error().Str("platform", platform).Msg(fmt.Sprintf("[%s]: serial is empty", c.HandlerName()))
			c.JSON(http.StatusBadRequest, HttpResponse{
				Code:    InvalidParamErrorCode,
				Message: fmt.Sprintf(InvalidParamErrorMsg, "serial"),
			})
			c.Abort()
			return
		}
		// get cached driver
		if driver, ok := uiClients[serial]; ok {
			c.Set("driver", driver)
			c.Next()
			return
		}

		switch strings.ToLower(platform) {
		case "android":
			device, err := uixt.NewAndroidDevice(uixt.WithSerialNumber(serial), uixt.WithStub(true))
			if err != nil {
				log.Error().Err(err).Str("platform", platform).Str("serial", serial).Msg(fmt.Sprintf("[%s]: Device Not Found", c.HandlerName()))
				c.JSON(http.StatusBadRequest, HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				})
				c.Abort()
				return
			}
			driver, err := device.NewDriver(uixt.WithDriverImageService(true), uixt.WithDriverResultFolder(true))
			if err != nil {
				log.Error().Err(err).Str("platform", platform).Str("serial", serial).Msg(fmt.Sprintf("[%s]: Failed New Driver", c.HandlerName()))
				c.JSON(http.StatusInternalServerError,
					HttpResponse{
						Code:    code.GetErrorCode(err),
						Message: err.Error(),
					},
				)
				c.Abort()
				return
			}
			c.Set("driver", driver)
			// cache driver
			uiClients[serial] = driver
		default:
			c.JSON(http.StatusBadRequest, HttpResponse{
				Code:    InvalidParamErrorCode,
				Message: fmt.Sprintf(InvalidParamErrorMsg, "platform"),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
