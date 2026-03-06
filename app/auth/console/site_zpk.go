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
	AppName           string
	ContainerName     string
	Namespace         string
}

// ./runtime/main site:register-zpk --thirdPartyCDToken=qEINzTKqtPUYKi7f --host=w7job.test.w7.com --releaseName=app-nfohievs0w --AppName=w7-pros-28692-app-nfohievs0w --namespace=default
var siteroZpk = siteZpkOption{}

func (c SiteZpk) GetName() string {
	return "site:register-zpk"
}

func (c SiteZpk) Configure(cmd *cobra.Command) {
	cmd.Flags().StringVar(&siteroZpk.ThirdPartyCDToken, "thirdPartyCDToken", "", "交付系统token")
	cmd.Flags().StringVar(&siteroZpk.Host, "host", "", "域名")
	cmd.Flags().StringVar(&siteroZpk.ReleaseName, "releaseName", "", "安装name")
	cmd.Flags().StringVar(&siteroZpk.AppName, "appName", "", "deployment名字")
	cmd.Flags().StringVar(&siteroZpk.ContainerName, "containerName", "", "containerName名字")
	cmd.Flags().StringVar(&siteroZpk.Namespace, "namespace", "", "namespace")
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
	secret, err := console.RegisterSiteZpk(siteroZpk.ThirdPartyCDToken, siteroZpk.ReleaseName, siteroZpk.Host)
	if err != nil {
		slog.Error("注册站点失败", "err", err)
		os.Exit(1)
	}

	slog.Info("注册站点成功", "secret", secret)
	sdk := k8s.NewK8sClientInner()
	err = console.PatchAppId(sdk, secret, siteroZpk.AppName, siteroZpk.Namespace, siteroZpk.ContainerName)
	if err != nil {
		slog.Error("更新appid失败", "err", err)
		os.Exit(1)
	}

	slog.Info("站点注册完成")
	os.Exit(0)
}
