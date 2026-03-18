package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Static struct {
	controller.Abstract
}

func (self Static) StaticInfo(http *gin.Context) {

}

func (self Static) Download(http *gin.Context) {
	name := gin.Param["name"]
}
