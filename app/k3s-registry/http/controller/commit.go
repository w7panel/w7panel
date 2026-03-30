package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/app/k3s-registry/logic"
	"github.com/w7panel/w7panel/common/service/registry"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Commit struct {
	controller.Abstract
}

var commitLogic = logic.NewCommitLogic()

// Run 提交容器为新镜像
func (self Commit) Run(ctx *gin.Context) {
	id := ctx.Param("id")
	ref := ctx.GetString("ref")

	digest, err := registry.CommitToContainerD(ctx, ref, id)
	if err != nil {
		self.JsonResponseWithServerError(ctx, err)
		return
	}
	self.JsonResponseWithoutError(ctx, gin.H{"digest": digest})
	return
}
