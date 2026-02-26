package console

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type QxUpgrade struct {
	console2.Abstract
}

func (c QxUpgrade) GetName() string {
	return "qx-upgrade"
}

func (c QxUpgrade) Configure(cmd *cobra.Command) {

}

func (c QxUpgrade) GetDescription() string {
	return "升级角色权限"
}

func (c QxUpgrade) Handle(cmd *cobra.Command, args []string) {

	sdk := k8s.NewK8sClient()
	sigClient, err := sdk.ToSigClient()
	if err != nil {
		slog.Error("Failed to create sigclient", "error", err)
		return
	}
	configmaps, err := sdk.ClientSet.CoreV1().ConfigMaps("default").List(context.Background(), v1.ListOptions{LabelSelector: "type=permission"})
	if err != nil {
		slog.Error("Failed to list configmaps", "error", err)
		return
	}
	c.handleConfigmaps(configmaps, sigClient)

	configmaps2, err := sdk.ClientSet.CoreV1().ConfigMaps("default").List(context.Background(), v1.ListOptions{LabelSelector: "type=quota"})
	if err != nil {
		slog.Error("Failed to list configmaps", "error", err)
		return
	}
	c.handleConfigmaps(configmaps2, sigClient)
}

func (QxUpgrade) handleConfigmaps(configmaps *corev1.ConfigMapList, sigClient sigclient.Client) {
	for _, configmap := range configmaps.Items {
		if configmap.Name == "k3k.permission.founder" {
			// configmap.Labels["w7.cc/role"] = "founder"
			configmap.Labels["typemode"] = "in"
			err := sigClient.Update(context.Background(), &configmap)
			if err != nil {
				slog.Error("Failed to update configmap", "error", err)
			}
			continue
		}
		if configmap.Labels != nil {
			_, ok := configmap.Labels["w7.cc/role"]
			if !ok && configmap.Labels["typemode"] != "in" {
				configmap.Labels["type"] = "permission"
				configmap.Labels["w7.cc/role"] = "normal"
				configmap.Labels["typemode"] = "custom"
				configmap.Annotations["title"] = "[普通用户]" + configmap.Annotations["title"]
				err := sigClient.Update(context.Background(), &configmap)
				if err != nil {
					slog.Error("Failed to update configmap", "error", err)
				}
			}
		}
	}
}
