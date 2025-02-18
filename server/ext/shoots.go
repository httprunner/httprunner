package server_ext

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/driver_ext"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/server"
)

func NewExtServer(port int) error {
	router := server.NewRouter()
	apiV1PlatformSerial := router.Group("/api/v1").Group("/:platform").Group("/:serial")

	// shoots operations
	apiV1PlatformSerial.GET("/shoots/source", handleDeviceContext(), sourceHandler)
	apiV1PlatformSerial.POST("/shoots/login", handleDeviceContext(), loginHandler)
	apiV1PlatformSerial.POST("/shoots/logout", handleDeviceContext(), logoutHandler)

	err := router.Engine.Run(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Err(err).Msg("failed to start http server")
		return err
	}
	return nil
}

func sourceHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	app, err := dExt.ForegroundInfo()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get foreground app", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			server.HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	source, err := dExt.Source(option.WithProcessName(app.PackageName))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get source %s", c.HandlerName(), app.PackageName))
		c.JSON(http.StatusInternalServerError,
			server.HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, server.HttpResponse{Result: source})
}

func loginHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		log.Error().Err(err).Msg("validate request failed")
		c.JSON(http.StatusBadRequest, server.HttpResponse{
			Code:    code.GetErrorCode(code.InvalidParamError),
			Message: fmt.Sprintf("validate request param failed: %s", err.Error()),
		})
		c.Abort()
		return
	}

	var info driver_ext.AppLoginInfo
	platform := c.Param("platform")
	if platform == "android" {
		info, err = dExt.(*driver_ext.ShootsAndroidDriver).LoginNoneUI(
			loginReq.PackageName, loginReq.PhoneNumber, loginReq.Captcha, loginReq.Password)
	} else {
		// ios
		info, err = dExt.(*driver_ext.ShootsIOSDriver).LoginNoneUI(
			loginReq.PackageName, loginReq.PhoneNumber, loginReq.Captcha, loginReq.Password)
	}
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to login", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			server.HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, server.HttpResponse{Code: 0, Message: "success", Result: info})
}

func logoutHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var logoutReq LogoutRequest
	if err := c.ShouldBindJSON(&logoutReq); err != nil {
		log.Error().Err(err).Msg("validate request failed")
		c.JSON(http.StatusBadRequest, server.HttpResponse{
			Code:    code.GetErrorCode(code.InvalidParamError),
			Message: fmt.Sprintf("validate request param failed: %s", err.Error()),
		})
		c.Abort()
		return
	}

	platform := c.Param("platform")
	if platform == "android" {
		err = dExt.(*driver_ext.ShootsAndroidDriver).LogoutNoneUI(logoutReq.PackageName)
	} else {
		// ios
		err = dExt.(*driver_ext.ShootsIOSDriver).LogoutNoneUI(logoutReq.PackageName)
	}
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to login", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			server.HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, server.HttpResponse{Code: 0, Message: "success"})
}

type LoginRequest struct {
	PackageName string `json:"packageName"`
	PhoneNumber string `json:"phoneNumber"`
	Captcha     string `json:"captcha"`
	Password    string `json:"password"`
}

type LogoutRequest struct {
	PackageName string `json:"packageName"`
}

var uiClients = make(map[string]uixt.IDriver) // UI automation clients for iOS and Android, key is udid/serial

func handleDeviceContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		serial := c.Param("serial")
		if serial == "" {
			log.Error().Str("platform", platform).Msg(fmt.Sprintf("[%s]: serial is empty", c.HandlerName()))
			c.JSON(http.StatusBadRequest, server.HttpResponse{
				Code:    code.GetErrorCode(code.InvalidParamError),
				Message: "serial is empty",
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
			device, err := uixt.NewAndroidDevice(
				option.WithSerialNumber(serial))
			if err != nil {
				log.Error().Err(err).Str("platform", platform).Str("serial", serial).
					Msg("android device not found")
				c.JSON(http.StatusBadRequest,
					server.HttpResponse{
						Code:    code.GetErrorCode(err),
						Message: err.Error(),
					},
				)
				c.Abort()
				return
			}

			driver, err := driver_ext.NewShootsAndroidDriver(device)
			if err != nil {
				log.Error().Err(err).Str("platform", platform).Str("serial", serial).
					Msg("failed to init shoots android driver")
				c.JSON(http.StatusInternalServerError,
					server.HttpResponse{
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
			c.JSON(http.StatusBadRequest, server.HttpResponse{
				Code:    code.GetErrorCode(code.InvalidParamError),
				Message: fmt.Sprintf("unsupported platform %s", platform),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func getContextDriver(c *gin.Context) (uixt.IDriver, error) {
	driverObj, exists := c.Get("driver")
	if !exists {
		log.Error().Msg("init device driver failed")
		c.JSON(http.StatusInternalServerError,
			server.HttpResponse{
				Code:    code.GetErrorCode(code.MobileUIDriverError),
				Message: "init driver failed",
			},
		)
		c.Abort()
		return nil, fmt.Errorf("driver not found")
	}
	dExt := driverObj.(uixt.IDriver)
	return dExt, nil
}
