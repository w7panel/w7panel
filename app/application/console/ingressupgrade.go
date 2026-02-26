package console

import (
	"log/slog"
	"os"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	"golang.org/x/mod/semver"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressUpgrade struct {
	console2.Abstract
}

type ingressupOption struct {
	cmd     string
	srcPath string
	toPath  string
	pid     string
}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443
var ingressupOp = ingressupOption{}

func (c IngressUpgrade) GetName() string {
	return "ingress-add-group"
}

func (c IngressUpgrade) Configure(cmd *cobra.Command) {

}

func (c IngressUpgrade) GetDescription() string {
	return "升级ingress信息到新版"
}

func (c IngressUpgrade) Handle(cmd *cobra.Command, args []string) {
	deploymentName, ok := os.LookupEnv("DEPLOYMENT_NAME")
	if !ok {
		deploymentName = "w7panel"
	}
	slog.Info("开始升级ingress信息到新版")
	sdk := k8s.NewK8sClient().Sdk
	now := time.Now()
	for true {
		deployment, err := sdk.ClientSet.AppsV1().Deployments(sdk.GetNamespace()).Get(sdk.Ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			slog.Error("获取面板信息失败", slog.String("error", err.Error()))
			time.Sleep(3 * time.Second)
			continue
		}
		if c.IsReady(deployment) {
			slog.Info("面板 已就绪")
			if time.Now().Sub(now).Seconds() > 300 {
				slog.Info("升级超时退出")
				os.Exit(0)
			}
			time.Sleep(3 * time.Second)
			break
		}
		time.Sleep(3 * time.Second)
	}
	old, err := appgroup.NewOldUpgrade(sdk)
	if err != nil {
		slog.Error("新版升级失败", slog.String("error", err.Error()))
		return
	}
	old.Upgrade()
}

func (c IngressUpgrade) IsReady(deployment *appsv1.Deployment) bool {
	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas && deployment.Generation == deployment.Status.ObservedGeneration {
		envs := deployment.Spec.Template.Spec.Containers[0].Env
		for _, env := range envs {
			if env.Name == "HELM_VERSION" {
				version := env.Value
				slog.Info("current面板版本", slog.String("version", version))
				if semver.Compare("v"+version, "v1.0.39") >= 0 {
					slog.Info("面板版本大于1.0.39")
					return true
				}
			}
		}
		// return true
	}
	return false
}
