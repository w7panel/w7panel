package controller

import (
	"gitee.com/we7coreteam/k8s-offline/app/k3s-registry/logic"
	"gitee.com/we7coreteam/k8s-offline/app/k3s-registry/model"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Commit struct {
	controller.Abstract
}

var commitLogic = logic.NewCommitLogic()

// Run 提交容器为新镜像
func (c Commit) Run(ctx *gin.Context) {
	id := ctx.Param("id")

	var req model.CommitRequest
	if !c.Validate(ctx, &req) {
		return
	}

	resp, err := commitLogic.Run(ctx, id, req)
	if err != nil {
		c.JsonResponseWithServerError(ctx, err)
		return
	}

	c.JsonResponseWithoutError(ctx, resp)
}
