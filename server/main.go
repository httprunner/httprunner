package server

import (
	"fmt"
	"time"

	"github.com/httprunner/httprunner/v5/pkg/uixt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func NewRouter() *Router {
	router := &Router{
		Engine:            gin.Default(),
		IRouterBaseMethod: &RouterBaseMethod{},
	}
	router.Init()
	return router
}

type Router struct {
	*gin.Engine
	IRouterBaseMethod
}

type RouterBaseMethod struct {
}

type IRouterBaseMethod interface {
	GetDriver(c *gin.Context) (driver uixt.IXTDriver, err error)
}

func (r *Router) Init() {
	r.Engine.Use(teardown())
	r.Engine.GET("/ping", pingHandler)
	r.Engine.GET("/", pingHandler)
	r.Engine.POST("/", pingHandler)
	r.Engine.GET("/api/v1/devices", listDeviceHandler)
	r.Engine.POST("/api/v1/browser/create_browser", createBrowserHandler)

	apiV1PlatformSerial := r.Group("/api/v1").Group("/:platform").Group("/:serial")

	// UI operations
	apiV1PlatformSerial.POST("/ui/tap", tapHandler)
	apiV1PlatformSerial.POST("/ui/double_tap", doubleTapHandler)
	apiV1PlatformSerial.POST("/ui/drag", dragHandler)
	apiV1PlatformSerial.POST("/ui/input", inputHandler)
	apiV1PlatformSerial.POST("/ui/home", homeHandler)
	apiV1PlatformSerial.POST("/ui/upload", uploadHandler)
	apiV1PlatformSerial.POST("/ui/hover", hoverHandler)
	apiV1PlatformSerial.POST("/ui/scroll", scrollHandler)

	// Key operations
	apiV1PlatformSerial.POST("/key/unlock", r.unlockHandler)
	apiV1PlatformSerial.POST("/key/home", r.homeHandler)
	apiV1PlatformSerial.POST("/key/backspace", r.backspaceHandler)
	apiV1PlatformSerial.POST("/key", r.keycodeHandler)

	// APP operations
	apiV1PlatformSerial.GET("/app/foreground", r.foregroundAppHandler)
	apiV1PlatformSerial.GET("/app/appInfo", r.appInfoHandler)
	apiV1PlatformSerial.POST("/app/clear", r.clearAppHandler)
	apiV1PlatformSerial.POST("/app/launch", r.launchAppHandler)
	apiV1PlatformSerial.POST("/app/terminal", r.terminalAppHandler)
	apiV1PlatformSerial.POST("/app/uninstall", r.uninstallAppHandler)

	// Device operations
	apiV1PlatformSerial.GET("/screenshot", screenshotHandler)
	apiV1PlatformSerial.DELETE("/close_browser", deleteBrowserHandler)
	apiV1PlatformSerial.GET("/video", videoHandler)
	apiV1PlatformSerial.POST("/device/push_image", pushImageHandler)
	apiV1PlatformSerial.POST("/device/clear_image", clearImageHandler)
	apiV1PlatformSerial.GET("/adb/source", adbSourceHandler)

	// uixt operations
	apiV1PlatformSerial.POST("/uixt/action", r.uixtActionHandler)
	apiV1PlatformSerial.POST("/uixt/actions", r.uixtActionsHandler)
}

func (r *Router) Run(port int) error {
	err := r.Engine.Run(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Err(err).Msg("failed to start http server")
		return err
	}
	return nil
}

func (r *Router) pingHandler(c *gin.Context) {
	RenderSuccess(c, true)
}

func teardown() gin.HandlerFunc {
	return func(c *gin.Context) {
		logID := c.Request.Header.Get("x-tt-logid")
		startTime := time.Now()
		// 结束处理后打印日志
		fmt.Printf("[GIN] %s | %s | %-7s \"%s\"\n",
			startTime.Format("2006/01/02 - 15:04:05"),
			logID,
			c.Request.Method,
			c.Request.URL.Path,
		)
		// 执行请求处理器
		c.Next()

		driverObj, exists := c.Get("driver")
		if exists {
			if driver, ok := driverObj.(*uixt.XTDriver); ok {
				_ = driver.TearDown()
			}
		}

		deviceObj, exists := c.Get("device")
		if exists {
			if device, ok := deviceObj.(*uixt.IOSDevice); ok {
				err := device.Teardown()
				if err != nil {
					log.Error().Err(err)
				}
			}
		}

		// 处理请求后获取结束时间
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 获取请求的状态码、客户端IP等信息
		statusCode := c.Writer.Status()

		// 结束处理后打印日志
		fmt.Printf("[GIN] %s | %d | %v | %s | %-7s \"%s\"\n",
			endTime.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			logID,
			c.Request.Method,
			c.Request.URL.Path,
		)
		c.Writer.Flush()
	}
}
