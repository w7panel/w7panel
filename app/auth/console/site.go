package console

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

// username password register
type Site struct {
	console2.Abstract
}

type siteOption struct {
	ThirdPartyCDToken string
	Host              string
	ReleaseName       string
	DeploymentName    string
	Namespace         string
}

// ./runtime/main site:register --thirdPartyCDToken=qEINzTKqtPUYKi7f --host=w7job.test.w7.com --releaseName=app-nfohievs0w --deploymentName=w7-pros-28692-app-nfohievs0w --namespace=default
var sitero = siteOption{}
var stopCh = make(chan struct{})

func (c Site) GetName() string {
	return "site:register"
}

func (c Site) Configure(cmd *cobra.Command) {
	cmd.Flags().StringVar(&sitero.ThirdPartyCDToken, "thirdPartyCDToken", "", "交付系统token")
	cmd.Flags().StringVar(&sitero.Host, "host", "", "域名")
	cmd.Flags().StringVar(&sitero.ReleaseName, "releaseName", "", "安装name")
	cmd.Flags().StringVar(&sitero.DeploymentName, "deploymentName", "", "deployment名字")
	cmd.Flags().StringVar(&sitero.Namespace, "namespace", "", "namespace")
}

func (c Site) GetDescription() string {
	return "站点注册"
}

func (c Site) Handle(cmd *cobra.Command, args []string) {
	slog.Info("域名验证中...")

	// 创建一个定时器，每30秒检查一次TLS握手
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 设置10分钟的总超时时间
	timeout := time.After(10 * time.Minute)

	// 立即进行第一次检查
	if c.checkTLSHandshake() {
		c.registerSite()
		return
	}

	// 每30秒检查一次，直到成功或超时
	for {
		select {
		case <-ticker.C:
			slog.Info("尝试验证域名TLS握手...")
			if c.checkTLSHandshake() {
				c.registerSite()
				return
			}
		case <-timeout:
			slog.Error("域名验证超时，可能是域名未备案或者没有正确解析到公网IP")
			os.Exit(1)
		}
	}
}

// 检查TLS握手是否成功
func (c Site) checkTLSHandshake() bool {
	url := fmt.Sprintf("https://%s", sitero.Host)
	slog.Info("正在验证TLS握手", "url", url)

	// 创建一个自定义的HTTP客户端，正确验证证书
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// 不跳过证书验证，确保证书有效
				ServerName: sitero.Host, // 确保验证的是正确的域名
			},
		},
	}

	// 发送HEAD请求，只检查连接，不下载内容
	resp, err := client.Head(url)
	if err != nil {
		slog.Info("TLS握手失败或证书无效,30s后重试中...")
		slog.Info(err.Error())
		return false
	}
	defer resp.Body.Close()

	slog.Info("TLS握手成功且证书有效", "status", resp.Status)
	return true
}

// 注册站点并更新相关信息
func (c Site) registerSite() {
	slog.Info("证书验证成功，开始注册站点...")
	secret, err := console.RegisterSite(sitero.ThirdPartyCDToken, sitero.ReleaseName, sitero.Host)
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
