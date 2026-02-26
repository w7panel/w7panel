package console

import (
	"log"
	"os"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

type UninstallStorePanel struct {
	console2.Abstract
}

// 删除应用市场安装的面板
func (c UninstallStorePanel) GetName() string {
	return "uninstall-store-panel"
}

func (c UninstallStorePanel) Configure(cmd *cobra.Command) {

}

func (c UninstallStorePanel) GetDescription() string {
	return "执行shell命令"
}

func (c UninstallStorePanel) Handle(cmd *cobra.Command, args []string) {

	sdk := k8s.NewK8sClientInner()
	groupApi, err := appgroup.NewAppGroupApi(sdk)
	if err != nil {
		log.Println(err)
		return
	}

	ns, ok := os.LookupEnv("NAMESPACE")
	if !ok {
		ns = "default"
	}
	groupApi.UninstallStorePanel(ns)

}
