package console

import (
	"log/slog"
	"os"

	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// username password register
type SiteZpk struct {
	console2.Abstract
}

type siteZpkOption struct {
	ThirdPartyCDToken string
	Host              string
	ReleaseName       string
	DeploymentName    string
	Namespace         string
}

// ./runtime/main site:register --thirdPartyCDToken=qEINzTKqtPUYKi7f --host=w7job.test.w7.com --releaseName=app-nfohievs0w --deploymentName=w7-pros-28692-app-nfohievs0w --namespace=default
var siteroZpk = siteZpkOption{}
var stopChZpk = make(chan struct{})

func (c SiteZpk) GetName() string {
	return "site:register"
}

func (c SiteZpk) Configure(cmd *cobra.Command) {
	cmd.Flags().StringVar(&sitero.ThirdPartyCDToken, "thirdPartyCDToken", "", "交付系统token")
	cmd.Flags().StringVar(&sitero.Host, "host", "", "域名")
	cmd.Flags().StringVar(&sitero.ReleaseName, "releaseName", "", "安装name")
	cmd.Flags().StringVar(&sitero.DeploymentName, "deploymentName", "", "deployment名字")
	cmd.Flags().StringVar(&sitero.Namespace, "namespace", "", "namespace")
}

func (c SiteZpk) GetDescription() string {
	return "站点注册"
}

func (c SiteZpk) Handle(cmd *cobra.Command, args []string) {
	c.registerSite()
}

// 检查TLS握手是否成功
func (c SiteZpk) registerSite() {
	slog.Info("证书验证成功，开始注册站点...")
	secret, err := console.RegisterSiteZpk(sitero.ThirdPartyCDToken, sitero.ReleaseName, sitero.Host)
	if err != nil {
		slog.Error("注册站点失败", "err", err)
		os.Exit(1)
	}

	slog.Info("注册站点成功", "secret", secret)
	sdk := k8s.NewK8sClientInner()
	err = console.PatchAppId(sdk, secret, sitero.DeploymentName, sitero.Namespace)
	if err != nil {
		slog.Error("更新appid失败", "err", err)
		os.Exit(1)
	}

	err = os.WriteFile("/tmp-w7-data/APP_ID", []byte(secret.AppId), 0644)
	if err != nil {
		slog.Error("写入APP_ID失败", "err", err)
		os.Exit(1)
	}

	err = os.WriteFile("/tmp-w7-data/APP_SECRET", []byte(secret.AppSecret), 0644)
	if err != nil {
		slog.Error("写入APP_SECRET失败", "err", err)
		os.Exit(1)
	}

	slog.Info("站点注册完成")
	os.Exit(0)
}
