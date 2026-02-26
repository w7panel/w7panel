package console

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterUpgrade struct {
	console2.Abstract
}

func (c ClusterUpgrade) GetName() string {
	return "cluster-upgrade"
}

func (c ClusterUpgrade) Configure(cmd *cobra.Command) {

}

func (c ClusterUpgrade) GetDescription() string {
	return "升级cluster"
}

func (c ClusterUpgrade) Handle(cmd *cobra.Command, args []string) {

	sdk := k8s.NewK8sClient()
	sigClient, err := sdk.ToSigClient()
	if err != nil {
		slog.Error("Failed to create sigclient", "error", err)
	}
	list := &v1alpha1.ClusterList{}
	err = sigClient.List(sdk.Ctx, list, &sigclient.ListOptions{})
	if err != nil {
		slog.Error("Failed to list cluster list", "error", err)
	}
	for _, cluster := range list.Items {
		if cluster.Spec.ServerArgs != nil {
			cluster.Spec.ServerArgs = []string{"--kubelet-arg=$cgroup_root", "--disable=traefik", "--embedded-registry", "--disable-network-policy"}
			err = sigClient.Update(sdk.Ctx, &cluster)
			if err != nil {
				slog.Error("Failed to update cluster", "error", err)
			}
		}
	}
}
