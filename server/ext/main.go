package server_ext

import (
	"github.com/httprunner/httprunner/v5/server"
)

type RouterExt struct {
	*server.Router
}

func NewExtRouter() *RouterExt {
	router := &RouterExt{
		Router: server.NewRouter(),
	}
	router.Setup()
	return router
}

func (r *RouterExt) Setup() {
	r.Router.Init()
	apiV1PlatformSerial := r.Group("/api/v1").Group("/:platform").Group("/:serial")

	apiV1PlatformSerial.GET("/stub/source", r.sourceHandler)
	apiV1PlatformSerial.POST("/stub/login", r.loginHandler)
	apiV1PlatformSerial.POST("/stub/logout", r.logoutHandler)
	apiV1PlatformSerial.POST("/app/install", r.installAppHandler)
}
