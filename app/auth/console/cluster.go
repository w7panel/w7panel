package console

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// username password register
type Cluster struct {
	console2.Abstract
}

type clusterRegisterOption struct {
	ApiServerUrl      string
	ThirdPartyCDToken string
	OfflineUrl        string
	Username          string
	Password          string
	RoleName          string
	Namespace         string
	IsClusterRole     bool
	RegisterCluster   bool
	UserMode          string
}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443
var cro = clusterRegisterOption{}

func (c Cluster) GetName() string {
	return "cluster:register"
}

func (c Cluster) Configure(cmd *cobra.Command) {
	// username password register
	//
	cmd.Flags().StringVar(&cro.ApiServerUrl, "apiServerUrl", "", "apiServerUrl")
	cmd.Flags().StringVar(&cro.ThirdPartyCDToken, "thirdPartyCDToken", "", "交付系统token")
	cmd.Flags().StringVar(&cro.OfflineUrl, "offlineUrl", "", "离线地址")
	cmd.Flags().StringVar(&cro.Username, "username", "", "username")
	cmd.Flags().StringVar(&cro.Password, "password", "", "password")
	cmd.Flags().BoolVar(&cro.IsClusterRole, "isCclusterRole", true, "是否集群角色")
	cmd.Flags().StringVar(&cro.RoleName, "rolename", "cluster-admin", "角色")
	cmd.Flags().StringVar(&cro.Namespace, "namespace", "", "命名空间")
	cmd.Flags().BoolVar(&cro.RegisterCluster, "registerCluster", false, "是否注册集群")
	cmd.Flags().StringVar(&cro.UserMode, "usermode", "founder", "用户模式")
}

func (c Cluster) GetDescription() string {
	return "注册集群"
}

func (c Cluster) Handle(cmd *cobra.Command, args []string) {
	sdk := k8s.NewK8sClientInner()
	if cro.Namespace == "" {
		cro.Namespace = sdk.GetNamespace()
	}

	if len(cro.Username) > 0 && len(cro.Password) > 0 {
		err := sdk.Register(cro.Username, cro.Password, cro.Namespace, cro.RoleName, cro.IsClusterRole, cro.UserMode)
		if err != nil {
			slog.Error("err register user", "err", err)
		}
	}

	if cro.RegisterCluster {
		slog.Info("register cluster")
		// time.Sleep(66 * time.Second)
		c.RegisterCluster(sdk)
		slog.Info("register cluster end")

		// c.RegisterCluster(sdk)
	}

}
func (c Cluster) RegisterCluster(sdk *k8s.Sdk) error {
	kubeconfig, err := sdk.ToKubeconfig(cro.ApiServerUrl)
	if err != nil {
		slog.Error("err get kubeconfig", "err", err)
		return err
	}
	respo := config.NewW7ConfigRepository(sdk)
	client := console.NewClusterClient(respo, sdk, kubeconfig)
	client.SetThirdPartyCdTDken(cro.ThirdPartyCDToken)
	client.SetOfflineUrl(cro.OfflineUrl)
	err = client.RegisterUseCdToken(false, "")
	if err != nil {
		slog.Warn("err register cluster", "err", err)
		// return
		return err
	}
	return nil
}
