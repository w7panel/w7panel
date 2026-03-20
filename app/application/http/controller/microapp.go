package controller

import (

	// "github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"

	"github.com/gin-gonic/gin"
	"github.com/w7panel/w7panel/common/service/k8s/microapp"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type MicroApp struct {
	controller.Abstract
}

func (self MicroApp) List(http *gin.Context) {
	token := http.MustGet("k8s_token").(string)
	list, err := microapp.ListTop(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, list)

}
