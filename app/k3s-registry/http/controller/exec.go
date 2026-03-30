package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Exec struct {
	controller.Abstract
}

// Run 在容器内执行命令
func (c Exec) Run(ctx *gin.Context) {

}
