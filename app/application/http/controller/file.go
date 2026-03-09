package controller

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/procpath"
	"gitee.com/we7coreteam/k8s-offline/common/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type File struct {
	controller.Abstract
}

func (self File) Download(http *gin.Context) {
	r := http.Request
	filename := filepath.Join(facade.GetConfig().GetString("s3.base_dir"),
		strings.TrimPrefix(r.URL.Path, "/panel-api/v1/download/"),
	)
	fs, err := os.Stat(filename)
	if os.IsNotExist(err) {
		self.JsonResponseWithError(http, fmt.Errorf("file not found"), 404)
		return
	}
	if fs.IsDir() {
	}
	file, err := os.Open(filename)
	if err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	defer file.Close()

	http.Header("Content-Type", "application/octet-stream")
	http.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name()))
	http.File(file.Name())
}

func (self File) Upload(http *gin.Context) {
	server := s3.NewS3Server(facade.Config.GetString("s3.base_dir"), "/tmp/metadata", "s3bucket")
	server.Server().ServeHTTP(http.Writer, http.Request)
}

func (self File) CpPidFile(http *gin.Context) {
	baseDir := facade.Config.GetString("s3.base_dir")
	type ParamsValidate struct {
		From   string `form:"from"      binding:"required"`
		To     string `form:"to"        binding:"required"`
		Upload string `form:"upload"    binding:"required"`
		Pid    string `form:"pid"       binding:"required"`
	}

	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	if params.Upload == "1" {
		params.From = filepath.Join(baseDir, params.From)
		params.To = procpath.ConvertToLocalPath(params.To)
	} else {
		params.To = filepath.Join(baseDir, params.To)
		params.From = procpath.ConvertToLocalPath(params.From)
	}
	err := os.Mkdir(filepath.Dir(params.To), 0755)
	if err != nil && !os.IsExist(err) {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	slog.Info("cp", "from", params.From, "to", params.To)
	if err = exec.Command("cp", "-r", params.From, params.To).Run(); err != nil {
		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.JsonSuccessResponse(http)
}
