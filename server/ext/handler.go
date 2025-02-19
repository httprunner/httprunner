package server_ext

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/uixt/driver_ext"
	"github.com/httprunner/httprunner/v5/server"
)

func (r *RouterExt) loginHandler(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		server.RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	info, err := driver.GetIDriver().(driver_ext.IStubDriver).
		LoginNoneUI(loginReq.PackageName, loginReq.PhoneNumber,
			loginReq.Captcha, loginReq.Password)
	if err != nil {
		server.RenderError(c, err)
		return
	}
	server.RenderSuccess(c, info)
}

func (r *RouterExt) logoutHandler(c *gin.Context) {
	var logoutReq LogoutRequest
	if err := c.ShouldBindJSON(&logoutReq); err != nil {
		server.RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.GetIDriver().(driver_ext.IStubDriver).
		LogoutNoneUI(logoutReq.PackageName)
	if err != nil {
		server.RenderError(c, err)
		return
	}
	server.RenderSuccess(c, true)
}

func (r *RouterExt) sourceHandler(c *gin.Context) {
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	source, err := driver.Source()
	if err != nil {
		log.Warn().Err(err).Msg("get source failed")
	}
	server.RenderSuccess(c, source)
}
