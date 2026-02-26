package controller

import (
	"errors"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/gin-gonic/gin"

	// "github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Namespaces struct {
	controller.Abstract
}

func (self Namespaces) GetList(http *gin.Context) {
	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	slog.Info("client is nil", "client", client)
	namespaces, err := client.ClientSet.CoreV1().Namespaces().List(client.Ctx, metav1.ListOptions{})
	if err != nil {
		defaultNamespace := client.GetNamespace() // facade.GetConfig().GetString("k8s.default_namespace")
		if defaultNamespace == "" {
			self.JsonResponseWithServerError(http, errors.New("k8s default_namespace not found"))
			return
		}
		namespaces = &v1.NamespaceList{
			Items: []v1.Namespace{
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: defaultNamespace,
					},
				},
			},
		}
	}

	self.JsonResponseWithoutError(http, namespaces)
}
