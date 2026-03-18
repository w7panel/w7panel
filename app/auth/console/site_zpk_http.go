package console

import (
	"log/slog"
	"os"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// username password register
type SiteZpkHttp struct {
	console2.Abstract
}

type siteZpkHttpOption struct {
	PanelUrl      string `json:"panelUrl"`
	Host          string `json:"host"`
	ReleaseName   string `json:"releaseName"`
	SiteIdentifie string `json:"siteIdentifie"`
	AppName       string `json:"appName"`
	ContainerName string `json:"containerName"`
	Namespace     string `json:"namespace"`
	InstallId     string `json:"installId"`
}

// ./runtime/main site:register-zpk --thirdPartyCDToken=qEINzTKqtPUYKi7f --siteIdentifie=test.txt --host=w7job.test.w7.com --releaseName=app-nfohievs0w --appName=w7-pros-28692-app-nfohievs0w --namespace=default
var siteroZpkHttp = siteZpkHttpOption{}

func (c SiteZpkHttp) GetName() string {
	return "site:register-zpk-http"
}

func (c SiteZpkHttp) Configure(cmd *cobra.Command) {
	cmd.Flags().StringVar(&siteroZpkHttp.PanelUrl, "panelUrl", "", "交付系统token")
	cmd.Flags().StringVar(&siteroZpkHttp.Host, "host", "", "域名")
	cmd.Flags().StringVar(&siteroZpkHttp.ReleaseName, "releaseName", "", "安装name")
	cmd.Flags().StringVar(&siteroZpkHttp.SiteIdentifie, "siteIdentifie", "", "站点标识")
	cmd.Flags().StringVar(&siteroZpkHttp.AppName, "appName", "", "deployment名字")
	cmd.Flags().StringVar(&siteroZpkHttp.ContainerName, "containerName", "", "containerName名字")
	cmd.Flags().StringVar(&siteroZpkHttp.Namespace, "namespace", "default", "namespace")
	cmd.Flags().StringVar(&siteroZpkHttp.InstallId, "installId", "", "namespace")
}

func (c SiteZpkHttp) GetDescription() string {
	return "站点注册http"
}

func (c SiteZpkHttp) Handle(cmd *cobra.Command, args []string) {
	c.registerSite()
}

// 检查TLS握手是否成功
func (c SiteZpkHttp) registerSite() {
	req := helper.RetryHttpClient().R()
	mapdata := make(map[string]string)
	mapdata["appName"] = siteroZpkHttp.AppName
	mapdata["containerName"] = siteroZpkHttp.ContainerName
	mapdata["host"] = siteroZpkHttp.Host
	mapdata["installId"] = siteroZpkHttp.InstallId
	mapdata["namespace"] = siteroZpkHttp.Namespace
	mapdata["releaseName"] = siteroZpkHttp.ReleaseName
	mapdata["siteIdentifie"] = siteroZpkHttp.SiteIdentifie
	url := siteroZpkHttp.PanelUrl + "/panel-api/v1/auth/console/register-zpk-site"
	slog.Info("register site", "url", url)
	res, err := req.SetFormData(mapdata).Post(url)
	slog.Info("register site", "statusCode", res.StatusCode(), "response", res.String())
	if err != nil {
		slog.Error("resiter site error", "error", err)
		os.Exit(0)
	}
	if res.StatusCode() > 299 {
		slog.Error("register site error", "statusCode", res.StatusCode(), "response", res.String())
		os.Exit(0)
	}

}
