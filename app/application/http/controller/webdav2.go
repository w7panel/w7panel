package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/common/service/procpath"
	"github.com/w7panel/w7panel/common/service/webdav2"
	"golang.org/x/net/webdav"
)

func (c Webdav) handleWithPermissionPreservation2(ctx *gin.Context, prefix string, fs webdav.FileSystem, rootDir string) {
	// userGroup := webdav1.GetUserGroup(pid)

	relPath := ctx.Request.URL.Path[len(prefix):]
	if relPath == "" {
		relPath = "/"
	}
	dirStr := ""
	if dir, ok := fs.(webdav.Dir); ok {
		dirStr = string(dir)
	}
	webdavFileSystem := webdav2.NewWebDAVFileSystem(fs, dirStr)
	hander := webdav.Handler{
		Prefix:     prefix,
		FileSystem: webdavFileSystem,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				slog.Error("webdav", "error", err)
			}
		},
	}
	hander.ServeHTTP(ctx.Writer, ctx.Request)

}
func (c Webdav) HandlePid2(ctx *gin.Context) {
	pid := ctx.Param("pid")
	webDirPath := procpath.GetRootPath(pid)
	c.handleWithPermissionPreservation2(ctx,
		"/panel-api/v1/files/webdav-agent/"+pid+"/agent",
		webdav.Dir(webDirPath), webDirPath)
}

func (c Webdav) HandlePidSubPid2(ctx *gin.Context) {
	pid := ctx.Param("pid")
	subpid := ctx.Param("subpid")
	webDirPath := procpath.GetRootPathWithSubPid(pid, subpid)
	prefix := "/panel-api/v1/files/webdav-agent/" + pid + "/agent"
	if subpid != "" {
		prefix = "/panel-api/v1/files/webdav-agent/" + pid + "/subagent/" + subpid + "/agent"
	}
	c.handleWithPermissionPreservation2(ctx,
		prefix,
		webdav.Dir(webDirPath), webDirPath)
}

func (c Webdav) HandleTest(ctx *gin.Context) {
	c.handleWithPermissionPreservation2(ctx,
		"/panel-api/v1/files/webdav-test",
		webdav.Dir("/"), "/")
}
