package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func NewRouter() *Router {
	router := &Router{}
	router.Init()
	return router
}

type Router struct {
	*gin.Engine
}

func (r *Router) Init() {
	r.Engine = gin.Default()
	r.Engine.GET("/ping", pingHandler)

	apiV1Platform := r.Engine.Group("/api/v1").Group("/:platform")
	apiV1Platform.GET("/devices", listDeviceHandler)

	apiV1PlatformSerial := apiV1Platform.Group("/:serial")
	// UI operations
	apiV1PlatformSerial.POST("/ui/tap", r.HandleDeviceContext(), tapHandler)
	apiV1PlatformSerial.POST("/ui/drag", r.HandleDeviceContext(), dragHandler)
	apiV1PlatformSerial.POST("/ui/input", r.HandleDeviceContext(), inputHandler)
	// Key operations
	apiV1PlatformSerial.POST("/key/unlock", r.HandleDeviceContext(), unlockHandler)
	apiV1PlatformSerial.POST("/key/home", r.HandleDeviceContext(), homeHandler)
	apiV1PlatformSerial.POST("/key", r.HandleDeviceContext(), keycodeHandler)
	// App operations
	apiV1PlatformSerial.GET("/app/foreground", r.HandleDeviceContext(), foregroundAppHandler)
	apiV1PlatformSerial.POST("/app/clear", r.HandleDeviceContext(), clearAppHandler)
	apiV1PlatformSerial.POST("/app/launch", r.HandleDeviceContext(), launchAppHandler)
	apiV1PlatformSerial.POST("/app/terminal", r.HandleDeviceContext(), terminalAppHandler)
	// get screen info
	apiV1PlatformSerial.GET("/screenshot", r.HandleDeviceContext(), screenshotHandler)
	apiV1PlatformSerial.POST("/screenresult", r.HandleDeviceContext(), screenResultHandler)
	apiV1PlatformSerial.GET("/adb/source", r.HandleDeviceContext(), adbSourceHandler)
	// run uixt actions
	apiV1PlatformSerial.POST("/uixt/action", r.HandleDeviceContext(), uixtActionHandler)
	apiV1PlatformSerial.POST("/uixt/actions", r.HandleDeviceContext(), uixtActionsHandler)
}

func (r *Router) Run(port int) error {
	err := r.Engine.Run(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Err(err).Msg("failed to start http server")
		return err
	}
	return nil
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}
