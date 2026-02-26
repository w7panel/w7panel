package controller

import (
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type User struct {
	controller.Abstract
}

func (self Auth) Add(http *gin.Context) {
	type ParamsValidate struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	client := k8s.NewK8sClient()
	err := client.ResetPassword(params.Username, params.Password, "normal")

	if err != nil {

		self.JsonResponseWithError(http, err, 500)
		return
	}
	self.JsonSuccessResponse(http)
	return
}
