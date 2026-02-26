package console

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/order"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K3kOrderReturnCheck struct {
	console2.Abstract
}

func (c K3kOrderReturnCheck) GetName() string {
	return "k3k-return-check"
}

func (c K3kOrderReturnCheck) Configure(cmd *cobra.Command) {

}

func (c K3kOrderReturnCheck) GetDescription() string {
	return "退款记录除了里"
}

func (c K3kOrderReturnCheck) Handle(cmd *cobra.Command, args []string) {

	sdk := k8s.NewK8sClient()
	// sigClient, err := sdk.ToSigClient()
	// if err != nil {
	// 	slog.Error("Failed to create sigclient", "error", err)
	// 	return
	// }

	serviceAccounts, err := sdk.ClientSet.CoreV1().ServiceAccounts("default").List(context.Background(), v1.ListOptions{})
	if err != nil {
		slog.Error("Failed to list configmaps", "error", err)
		return
	}
	err = c.handleSa(serviceAccounts, sdk.Sdk)
	if err != nil {
		slog.Warn("handle sa err", "err", err)
	}
}

func (K3kOrderReturnCheck) handleSa(serviceAccounts *corev1.ServiceAccountList, sdk *k8s.Sdk) error {
	orderApi, err := order.NewK3kOrderApi(sdk)
	if err != nil {
		return err
	}
	for _, sa := range serviceAccounts.Items {
		err = fixReturn(&sa, orderApi)
		if err != nil {
			slog.Warn("处理退款记录失败", "name", sa.Name, "err", err)
			continue
		}
	}
	return nil
}
