package server_ext

import (
	"github.com/httprunner/httprunner/v5/server"
)

func NewExtRouter() *server.Router {
	router := server.NewRouter()
	apiV1PlatformSerial := router.Group("/api/v1").Group("/:platform").Group("/:serial")

	apiV1PlatformSerial.GET("/stub/source", sourceHandler)
	apiV1PlatformSerial.POST("/stub/login", loginHandler)
	apiV1PlatformSerial.POST("/stub/logout", logoutHandler)
	apiV1PlatformSerial.POST("/app/install", installAppHandler)
	return router
}
