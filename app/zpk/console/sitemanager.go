package console

import (
	// "go/types"
	"log/slog"
	"os"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic"
	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	"golang.org/x/mod/semver"
)

type SiteManagerCmd struct {
	console2.Abstract
}

func (c SiteManagerCmd) GetName() string {
	return "sitemanager-upgrade"
}

func (c SiteManagerCmd) Configure(cmd *cobra.Command) {

	// cmd.MarkFlagRequired("identifie")
}

// go run main.go helm-check --version=1.16.0 --releaseName=w7-zpkv2 --namespace=default
// 和 shell 配合 存在 exit 1 否则 0
func (c SiteManagerCmd) Handle(cmd *cobra.Command, args []string) {

	sdk := k8s.NewK8sClient().Sdk
	helmApi := k8s.NewHelm(sdk)
	api, err := appgroup.NewAppGroupApi(sdk)
	if err != nil {
		slog.Error("not find sitemanager api", "error", err)
		os.Exit(1)
		return
	}
	list, err := api.GetAppGroupListByIdentifie("", "w7-sitemanager")
	if err != nil {
		slog.Error("not found sitemanager", "error", err)
		os.Exit(1)
		return
	}

	// koDataPath := os.Getenv("KO_DATA_PATH")
	// if koDataPath == "" {
	// 	slog.Error("KO_DATA_PATH not set")
	// 	os.Exit(1)
	// 	return
	// }

	// chartPath := koDataPath + "/charts/site-manager-0.1.1.tgz"
	// chart, err := loader.Load(chartPath)
	// if err != nil {
	// 	slog.Error("load chart error", "error", err)
	// 	os.Exit(1)
	// 	return
	// }

	for _, item := range list.Items {
		compare := semver.Compare("v"+item.Spec.Version, "v1.0.19")
		if compare < 0 {
			releaseName := item.Name
			namespace := item.Namespace
			info, err := helmApi.Info(releaseName, namespace)
			if err != nil {
				slog.Error("cannot find helm info", "releaseName", releaseName)
				continue
			}
			repo := logic.NewRepo(item.Spec.ZpkUrl, "", "")
			repo.SetUpgrade(true)
			repo.SetCurVersion(item.Spec.Version)
			mPackage, err := repo.Load()
			if err != nil {
				slog.Error("can not load manifest")
				continue
			}
			configEnvs := helper.HelmValflattenMap(info.Config)
			slog.Info("config envs", "config", info.Config, "env", configEnvs)
			envs := []types.EnvKv{}
			for k, v := range configEnvs {
				env := types.EnvKv{
					Name:  k,
					Value: v,
				}
				envs = append(envs, env)
			}
			option := types.InstallOption{
				Identifie: "w7_sitemanager",
				EnvKv:     envs,
			}
			options := []types.InstallOption{option}
			packageApps := types.NewPackage(mPackage, options, releaseName, strings.ToLower(helper.RandomString(5)), namespace,
				"", "", "")
			packageApps.Root.ServiceAccountName = helper.ServiceAccountName()
			install := logic.NewInstall(sdk, packageApps)
			err = install.InstallOrUpgrade(releaseName, "default")
			if err != nil {
				slog.Error("upgrade sitemanager err", "releasename", releaseName)
			}
			// slog.Info("upgrading sitemanager", "releaseName", releaseName, "namespace", namespace, "currentVersion", item.Spec.Version)
			// info, err := helmApi.Info(releaseName, namespace)
			// if err != nil {
			// 	slog.Error("get helm info error")
			// 	continue
			// }
			// _, err = helmApi.Upgrade(sdk.Ctx, chart, info.Config, releaseName, namespace, nil)
			// if err != nil {
			// 	slog.Error("upgrade sitemanager error", "releaseName", releaseName, "error", err)
			// 	continue
			// }
			// slog.Info("upgrade sitemanager success", "releaseName", releaseName, "namespace", namespace)
		}
	}

	slog.Info("sitemanager upgrade completed")

}
