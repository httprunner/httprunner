package server_ext

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/server"
)

func (r *RouterExt) installAppHandler(c *gin.Context) {
	var appInstallReq AppInstallRequest
	if err := c.ShouldBindJSON(&appInstallReq); err != nil {
		server.RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.InstallByUrl(appInstallReq.AppUrl)
	if err != nil {
		server.RenderError(c, err)
		return
	}
	if androidDevice, ok := driver.GetDevice().(*uixt.AndroidDevice); ok {
		_ = driver.Home()
		if appInstallReq.MappingUrl == "" || appInstallReq.ResourceMappingUrl == "" {
			server.RenderSuccess(c, true)
			return
		}
		localMappingPath, err := uixt.DownloadFileByUrl(appInstallReq.MappingUrl)
		if err != nil {
			server.RenderError(c, err)
		}
		defer func() {
			_ = os.Remove(localMappingPath)
		}()
		if err = androidDevice.PushFile(
			localMappingPath,
			fmt.Sprintf("/data/local/tmp/%s_map.txt", appInstallReq.PackageName)); err != nil {
			server.RenderError(c, err)
			return
		}
		localResourceMappingPath, err := uixt.DownloadFileByUrl(
			appInstallReq.ResourceMappingUrl)
		if err != nil {
			server.RenderError(c, err)
		}
		defer func() {
			_ = os.Remove(localResourceMappingPath)
		}()
		if err = androidDevice.PushFile(localResourceMappingPath,
			fmt.Sprintf("/data/local/tmp/%s_resmap.txt", appInstallReq.PackageName)); err != nil {
			server.RenderError(c, err)
			return
		}
	}
	server.RenderSuccess(c, true)
}
