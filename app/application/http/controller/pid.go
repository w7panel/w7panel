package controller

import (

	// "github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/pid"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Pid struct {
	controller.Abstract
}

func (self Pid) GetPid(http *gin.Context) {
	params := pid.PidParam{}
	if !self.Validate(http, &params) {
		return
	}
	token := http.MustGet("k8s_token").(string)
	pidObj, err := pid.NewPid(token)
	if err != nil {
		self.JsonResponseWithoutError(http, err)
		return
	}
	result, err := pidObj.Handle(params)
	if err != nil {
		self.JsonResponseWithoutError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, result.ToArray())
}

func (self Pid) EtcPasswd(http *gin.Context) {
}
