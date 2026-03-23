package controller

import (

	// "github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"

	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/common/service/k8s/pid"
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
	pidResult, err := pidObj.Handle(params)
	if err != nil {
		self.JsonResponseWithoutError(http, err)
		return
	}
	result := pidResult.ToArray()
	result["webdavToken"] = token

	self.JsonResponseWithoutError(http, result)
}

func (self Pid) EtcPasswd(http *gin.Context) {
}
