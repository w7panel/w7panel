package controller

import (
	"gitee.com/we7coreteam/k8s-offline/app/k3s-registry/logic"
	"gitee.com/we7coreteam/k8s-offline/app/k3s-registry/model"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Exec struct {
	controller.Abstract
}

var execLogic = logic.NewExecLogic()

// Run 在容器内执行命令
func (c Exec) Run(ctx *gin.Context) {
	id := ctx.Param("id")

	var req model.ExecRequest
	if !c.Validate(ctx, &req) {
		return
	}

	resp, err := execLogic.Run(ctx, id, req)
	if err != nil {
		c.JsonResponseWithServerError(ctx, err)
		return
	}

	c.JsonResponseWithoutError(ctx, resp)
}
