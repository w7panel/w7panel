package console

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/longhorn"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

const repo = "https://zpk.w7.cc/respo/info/w7panel_metrics"
const chartName = "https://cdn.w7.cc/w7panel/charts/metrics/k8s-offline-metrics-1.0.10.tgz"

type MetricsUpgrade struct {
	console2.Abstract
}

func (c MetricsUpgrade) GetName() string {
	return "metrics:upgrade"
}

func (c MetricsUpgrade) Handle(cmd *cobra.Command, args []string) {
	sdk := k8s.NewK8sClient().Sdk
	helmApi := k8s.NewHelm(sdk)
	_, err := helmApi.Info("vm-operator", "w7-system")
	if err != nil {
		slog.Error("默认监控未安装跳过", "err", err)
		return
	}
	slog.Info("start uninstall vm-operator")
	_, err = helmApi.UnInstall("vm-operator", "w7-system")
	if err != nil {
		slog.Error("uninstall vm-operator error", "err", err)
		return
	}
	slog.Info("开始安装面板监控")

	longhornClient, err := longhorn.NewLonghornClient(sdk)
	if err != nil {
		slog.Error("new longhorn client error", "err", err)
		return
	}
	list, err := longhornClient.GetK8sStorageClassList()
	if err != nil {
		slog.Error("get k8s storage class list error", "err", err)
		return
	}
	defaultStorageClass := ""
	for _, v := range list.Items {
		if v.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			defaultStorageClass = v.Name
		}
	}
	if defaultStorageClass == "" {
		slog.Error("default storage class not found")
		return
	}
	vals := map[string]interface{}{
		"storage.storageClassName": defaultStorageClass,
	}
	_, err = helmApi.InstallRaw(cmd.Context(), "", chartName, "w7panel-metrics", "default", vals, nil)
	if err != nil {
		slog.Error("install w7panel-metrics error", "err", err)
		return
	}

}
