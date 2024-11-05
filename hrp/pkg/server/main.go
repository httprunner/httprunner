package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func NewServer(port int) error {
	router := gin.Default()
	router.GET("/ping", pingHandler)

	apiV1Platform := router.Group("/api/v1").Group("/:platform")
	apiV1Platform.GET("/devices", listDeviceHandler)

	apiV1PlatformSerial := apiV1Platform.Group("/:serial")
	// UI operations
	apiV1PlatformSerial.POST("/ui/tap", handleDeviceContext(), tapHandler)
	apiV1PlatformSerial.POST("/ui/drag", handleDeviceContext(), dragHandler)
	apiV1PlatformSerial.POST("/ui/input", handleDeviceContext(), inputHandler)
	// Key operations
	apiV1PlatformSerial.POST("/key/unlock", handleDeviceContext(), unlockHandler)
	apiV1PlatformSerial.POST("/key/home", handleDeviceContext(), homeHandler)
	apiV1PlatformSerial.POST("/key", handleDeviceContext(), keycodeHandler)
	// App operations
	apiV1PlatformSerial.GET("/app/foreground", handleDeviceContext(), foregroundAppHandler)
	apiV1PlatformSerial.POST("/app/clear", handleDeviceContext(), clearAppHandler)
	apiV1PlatformSerial.POST("/app/launch", handleDeviceContext(), launchAppHandler)
	apiV1PlatformSerial.POST("/app/terminal", handleDeviceContext(), terminalAppHandler)
	// get screen info
	apiV1PlatformSerial.GET("/screenshot", handleDeviceContext(), screenshotHandler)
	apiV1PlatformSerial.POST("/screenresult", handleDeviceContext(), screenResultHandler)
	apiV1PlatformSerial.GET("/stub/source", handleDeviceContext(), sourceHandler)
	apiV1PlatformSerial.GET("/adb/source", handleDeviceContext(), adbSourceHandler)
	// Stub operations
	apiV1PlatformSerial.POST("/stub/login", handleDeviceContext(), loginHandler)
	apiV1PlatformSerial.POST("/stub/logout", handleDeviceContext(), logoutHandler)

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
