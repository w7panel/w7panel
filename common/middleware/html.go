package middleware

import (
	"io"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type Html struct {
	middleware.Abstract
}

func (self Html) Process(ctx *gin.Context) {
	path := ctx.Request.URL.Path

	// API 路由不处理
	if strings.HasPrefix(path, "/panel-api/") ||
		strings.HasPrefix(path, "/k8s-proxy/") ||
		strings.HasPrefix(path, "/k8s/") ||
		strings.HasPrefix(path, "/api/") {
		ctx.Status(404)
		ctx.Writer.Write([]byte(`{"code":404,"msg":"Not Found"}`))
		ctx.Abort()
		return
	}

	// 非 API 路由返回 index.html
	staticPath := facade.Config.GetString("app.static_path")
	data, err := os.Open(staticPath + "/index.html")
	if err != nil {
		ctx.Status(500)
		return
	}
	defer data.Close()

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Status(200)
	io.Copy(ctx.Writer, data)
	ctx.Abort()
}
