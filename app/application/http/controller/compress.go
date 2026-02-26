package controller

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/compress"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type CompressAgent struct {
	controller.Abstract
}

func (c CompressAgent) Compress(http *gin.Context) {
	pid := http.Param("pid")
	subpid := http.Param("subpid")

	type Params struct {
		Sources []string `form:"sources" json:"sources" binding:"required"`
		Output  string   `form:"output"  json:"output"  binding:"required"`
	}

	params := Params{}
	if err := http.ShouldBind(&params); err != nil {
		c.JsonResponseWithError(http, err, 400)
		return
	}

	slog.Info("compress request", "pid", pid, "subpid", subpid, "sources", params.Sources, "output", params.Output)

	targetPid := pid
	if subpid != "" {
		targetPid = subpid
	}

	compressor := compress.NewCompressor(targetPid)
	if err := compressor.Compress(params.Sources, params.Output); err != nil {
		slog.Error("compress failed", "error", err, "pid", targetPid)
		c.JsonResponseWithError(http, err, 500)
		return
	}

	c.JsonSuccessResponse(http)
}

func (c CompressAgent) Extract(http *gin.Context) {
	pid := http.Param("pid")
	subpid := http.Param("subpid")

	type Params struct {
		Source string `form:"source" json:"source" binding:"required"`
		Target string `form:"target" json:"target" binding:"required"`
	}

	params := Params{}
	if err := http.ShouldBind(&params); err != nil {
		c.JsonResponseWithError(http, err, 400)
		return
	}

	slog.Info("extract request", "pid", pid, "subpid", subpid, "source", params.Source, "target", params.Target)

	targetPid := pid
	if subpid != "" {
		targetPid = subpid
	}

	compressor := compress.NewCompressor(targetPid)
	if err := compressor.Extract(params.Source, params.Target); err != nil {
		slog.Error("extract failed", "error", err, "pid", targetPid)
		c.JsonResponseWithError(http, err, 500)
		return
	}

	c.JsonSuccessResponse(http)
}
