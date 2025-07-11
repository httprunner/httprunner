package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/mcphost"
	"github.com/httprunner/httprunner/v5/uixt"
)

func NewRouter() *Router {
	router := &Router{
		Engine: gin.Default(),
	}
	router.Init()
	return router
}

type Router struct {
	*gin.Engine
	mcpHost *mcphost.MCPHost
}

func (r *Router) InitMCPHost(configPath string) error {
	mcpHost, err := mcphost.NewMCPHost(configPath, true)
	if err != nil {
		log.Error().Err(err).Str("configPath", configPath).
			Msg("init MCP host failed")
		return err
	}
	r.mcpHost = mcpHost
	return nil
}

func (r *Router) Init() {
	r.Engine.Use(r.teardown())
	r.Engine.GET("/ping", r.pingHandler)
	r.Engine.GET("/", r.pingHandler)
	r.Engine.POST("/", r.pingHandler)

	apiV1PlatformSerial := r.Group("/api/v1").Group("/:platform").Group("/:serial")

	// tool operations
	apiV1PlatformSerial.POST("/tool/invoke", r.invokeToolHandler)

	// uixt operations
	apiV1PlatformSerial.POST("/uixt/action", r.uixtActionHandler)
	apiV1PlatformSerial.POST("/uixt/actions", r.uixtActionsHandler)
}

func (r *Router) Run(port int) error {
	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", port),
		Handler: r.Engine,
	}

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Info().Int("port", port).Msg("Starting hrp server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("HTTP server failed to start")
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Info().Msg("Shutting down hrp server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown MCP host first if it exists
	if r.mcpHost != nil {
		log.Info().Msg("Shutting down MCP host...")
		r.mcpHost.Shutdown()
	}

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("hrp server forced to shutdown")
		return err
	}

	log.Info().Msg("hrp server exited")
	return nil
}

func (r *Router) pingHandler(c *gin.Context) {
	RenderSuccess(c, true)
}

func (r *Router) teardown() gin.HandlerFunc {
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
			if device, ok := deviceObj.(uixt.IDevice); ok {
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
