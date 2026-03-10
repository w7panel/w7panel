package console

import (
	"log/slog"
	"os"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	"golang.org/x/mod/semver"
)

type HelmCheckCmd struct {
	console2.Abstract
}

type helmCheckOption struct {
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

var helmCOp = helmCheckOption{}

func (c HelmCheckCmd) GetName() string {
	return "helm-check"
}

func (c HelmCheckCmd) Configure(cmd *cobra.Command) {
	// username password register
	cmd.Flags().StringVar(&helmCOp.version, "version", "", "version")
	cmd.Flags().StringVar(&helmCOp.releaseName, "releaseName", "", "安装的名称")
	cmd.Flags().StringVar(&helmCOp.namespace, "namespace", "", "namespace")

	// cmd.MarkFlagRequired("identifie")
}

// go run main.go helm-check --version=1.16.0 --releaseName=w7-zpkv2 --namespace=default
// 和shell配合 存在exit 1 否则0
func (c HelmCheckCmd) Handle(cmd *cobra.Command, args []string) {

	sdk := k8s.NewK8sClient().Sdk
	helmApi := k8s.NewHelm(sdk)
	helmInfo, err := helmApi.Info(helmCOp.releaseName, helmCOp.namespace)
	if err != nil {
		slog.Error("not found", "releasename", helmCOp.releaseName)
		return
	}
	version := helmInfo.Chart.AppVersion()
	pversion := helmCOp.version
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if !strings.HasPrefix(helmCOp.version, "v") {
		pversion = "v" + pversion
	}
	compare := semver.Compare(version, pversion)
	if compare <= 0 {
		slog.Info("find appversion will exit 1", "current-version", version, "param-version", pversion, "releasename", helmCOp.releaseName)
		os.Exit(1)
	}
	slog.Warn("not found release", "releasename", helmCOp.releaseName, "version", helmCOp.version)

}
