package console

import (
	"log/slog"
	"os"
	"time"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
)

type HelmCmd struct {
	console2.Abstract
}

type helmOption struct {
	version     string
	releaseName string
	repository  string
	chartName   string
	namespace   string
	set         []string
	setJson     []string
	atomic      bool              // 原子安装
	zipUrl      string            // 压缩包地址 应用市场用zip 存helm包
	annotations map[string]string // 注解信息
	labels      map[string]string // 标签数据
}

var helmOp = helmOption{}

func (c HelmCmd) GetName() string {
	return "helmgo"
}

func (c HelmCmd) Configure(cmd *cobra.Command) {
	// username password register
	cmd.Flags().StringVar(&helmOp.version, "version", "", "version")
	cmd.Flags().StringVar(&helmOp.releaseName, "releaseName", "", "安装的名称")
	cmd.Flags().StringVar(&helmOp.repository, "repository", "", "helm仓库地址")
	cmd.Flags().StringVar(&helmOp.namespace, "namespace", "", "namespace")
	cmd.Flags().StringVar(&helmOp.chartName, "chartName", "", "目标路径")
	cmd.Flags().StringArrayVar(&helmOp.set, "set", []string{}, "set参数")
	cmd.Flags().StringArrayVar(&helmOp.setJson, "set-json", []string{}, "set-json参数")
	cmd.Flags().StringVar(&helmOp.zipUrl, "zipUrl", "", "zip包地址")
	cmd.Flags().BoolVar(&helmOp.atomic, "atomic", false, "原子安装")
	// cmd.Flags().StringVar(&helmOp.Identifie, "identifie", "", "应用标识")
	cmd.Flags().StringToStringVar(&helmOp.annotations, "anno", map[string]string{}, "annotations参数")
	cmd.Flags().StringToStringVar(&helmOp.labels, "labels", map[string]string{}, "labels参数")
	// cmd.MarkFlagRequired("cmd")
	// cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("releaseName")
	// cmd.MarkFlagRequired("identifie")
}

func (c HelmCmd) Handle(cmd *cobra.Command, args []string) {
	slog.Info("helmgo start")
	time.Sleep(time.Second * 5) //appgroup未创建完成，延迟执行
	sdk := k8s.NewK8sClient().Sdk
	helmApi := k8s.NewHelm(sdk)
	chart, err := logic.LocateChartByHelmZpk(helmOp.repository, helmOp.chartName, helmOp.zipUrl, helmOp.version)
	if err != nil {
		slog.Error("locate chart error", "error", err)
		os.Exit(1)
		return
	}
	if (chart.Metadata.Annotations == nil) || (len(helmOp.annotations) == 0) {
		chart.Metadata.Annotations = make(map[string]string)
	}
	if helmOp.annotations != nil && len(helmOp.annotations) > 0 {
		for k, v := range helmOp.annotations {
			chart.Metadata.Annotations[k] = v
		}
	}

	settings := cli.New()
	optValues := &values.Options{}
	for _, val := range helmOp.set {
		optValues.StringValues = append(optValues.StringValues, val)
	}
	for _, val := range helmOp.setJson {
		optValues.JSONValues = append(optValues.JSONValues, val)
	}
	slog.Info("abcdefg merge values", "vals", optValues.JSONValues)
	provider := getter.All(settings)
	vals, err := optValues.MergeValues(provider)
	if err != nil {
		slog.Error("merge values error", "err", err)
		os.Exit(1)
		return
	}

	info, err := helmApi.Info(helmOp.releaseName, helmOp.namespace)
	helmApi.Atomic(helmOp.atomic)
	if info != nil {
		_, err := helmApi.Upgrade(sdk.Ctx, chart, vals, helmOp.releaseName, helmOp.namespace, helmOp.labels)
		if err != nil {
			slog.Error("upgrade error", "err", err)
			os.Exit(1)
			return
		}
	} else {
		_, err := helmApi.Install(sdk.Ctx, chart, vals, helmOp.releaseName, helmOp.namespace, helmOp.labels)
		if err != nil {
			slog.Error("install error", "err", err)
			os.Exit(1)
			return
		}
	}
	slog.Info("helmgo install success")

}
