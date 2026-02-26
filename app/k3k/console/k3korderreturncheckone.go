package console

import (
	"errors"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/order"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K3kOrderReturnCheckOne struct {
	console2.Abstract
}

type shellOption struct {
	saName string
}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443
var shOp = shellOption{}

// go run main.go k3k-return-check-one --sa=console-303483
func (c K3kOrderReturnCheckOne) GetName() string {
	return "k3k-return-check-one"
}

func (c K3kOrderReturnCheckOne) Configure(cmd *cobra.Command) {
	cmd.Flags().StringVar(&shOp.saName, "sa", "", "用户名")
}

func (c K3kOrderReturnCheckOne) GetDescription() string {
	return "退款记录除了里"
}

func (c K3kOrderReturnCheckOne) Handle(cmd *cobra.Command, args []string) {

	sdk := k8s.NewK8sClient()
	// sigClient, err := sdk.ToSigClient()
	// if err != nil {
	// 	slog.Error("Failed to create sigclient", "error", err)
	// 	return
	// }

	sa, err := sdk.ClientSet.CoreV1().ServiceAccounts("default").Get(sdk.Ctx, shOp.saName, v1.GetOptions{})
	if err != nil {
		slog.Error("Failed to list configmaps", "error", err)
		return
	}
	orderApi, err := order.NewK3kOrderApi(sdk.Sdk)
	if err != nil {
		slog.Error("获取订单API失败", "error", err)
		return
	}
	err = fixReturn(sa, orderApi)
	if err != nil {
		slog.Warn("handle sa err", "err", err)
	}
}

func fixReturn(sa *corev1.ServiceAccount, orderApi *order.K3kOrderApi) error {
	k3kUser := types.NewK3kUser(sa)
	if !k3kUser.IsClusterUser() {
		return errors.New("not cluster user")
	}
	if !k3kUser.IsClusterReady() {
		// return errors.New("cluster not ready")
	}
	if k3kUser.HasProcessReturnOrder() {
		slog.Info("has process return order", "name", sa.Name)
		err := orderApi.ProcessReturnOrder(k3kUser)
		if err != nil {
			slog.Warn("处理退款记录失败1", "name", sa.Name, "err", err)
			return err
		}
	}
	slog.Info("check new return log")
	err := orderApi.ProcessReturnLastOrder(k3kUser, true)
	if err != nil {
		slog.Warn("处理退款记录失败2", "name", sa.Name, "err", err)
		return err
	}
	slog.Info("处理退款记录成功", "name", sa.Name)
	return nil
}
