package console

import (
	"log/slog"
	"os"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MetricsInstall struct {
	console2.Abstract
}

type installOption struct {
	vmHelmname string
	namespace  string
}

// ./runtime/main cluster:register --thirdPartyCDToken=ywA2N3ImkVo0tPOn --registerCluster=true --offlineUrl=http://118.25.145.25:9090 --apiServerUrl=https://118.25.145.25:6443
var inOp = installOption{}

func (c MetricsInstall) GetName() string {
	return "metrics:install"
}

func (c MetricsInstall) Configure(cmd *cobra.Command) {
	// username password register
	cmd.Flags().StringVar(&inOp.vmHelmname, "metricsHelmname", "vm-operator", "安装的name")
	cmd.Flags().StringVar(&inOp.namespace, "namespace", "vm-operator", "安装vm operator的命名空间")
	// cmd.Flags().StringVar(&inOp.pid, "pid", "", "pid")
}

func (c MetricsInstall) GetDescription() string {
	return "安装metrics组件和kubeblocks crd组件"
}

func (c MetricsInstall) Handle(cmd *cobra.Command, args []string) {
	sdk := k8s.NewK8sClientInner()
	helm := k8s.NewHelm(sdk)
	release, err := helm.Info(inOp.vmHelmname, inOp.namespace)
	isInstalled := false
	if err != nil {
		slog.Error("helm release not found", slog.String("error", err.Error()))
		isInstalled = false
	}
	if release != nil {
		if release.Info.Status == "deployed" || release.Info.Status == "pending-install" || release.Info.Status == "pending-upgrade" {
			isInstalled = true
		}
		if release.Info.Status == "unknow" || release.Info.Status == "failed" {
			_, err := helm.UnInstall(inOp.vmHelmname, inOp.namespace)
			if err != nil {
				slog.Error("helm release uninstall error", slog.String("error", err.Error()))
			}
		}
	}

	baseDir, ok := os.LookupEnv("KO_DATA_PATH")
	if !ok {
		baseDir = "./kodata"
	}
	err = c.Apply(baseDir, "/crds/victoria-metrics-operator-0.43.0.yaml")
	if err != nil {
		slog.Error("apply crd error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = c.Apply(baseDir, "/yaml/victoria-metrics")
	if err != nil {
		slog.Error("apply vm error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if !isInstalled {

		args := []string{
			"install",
			inOp.vmHelmname,
			baseDir + "/charts/victoria-metrics-operator-0.43.0.tgz",
			"--namespace",
			inOp.namespace,
			"--create-namespace",
			"--set",
			"crds.enabled=false",
			"--set",
			"crds.plain=false",
			"--set",
			"crds.cleanup.enabled=true",
		}
		successstr, errstr, err := helper.Runsh("helm", args...)
		if err != nil {
			slog.Error("helm install error", slog.String("error", errstr))
			os.Exit(1)
		}
		if err == nil {
			print(successstr)
		}

	}
	_, err = sdk.ClientSet.CoreV1().Namespaces().Get(sdk.Ctx, inOp.namespace, metav1.GetOptions{})
	if err != nil {
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: inOp.namespace,
			},
		}

		_, err := sdk.ClientSet.CoreV1().Namespaces().Create(sdk.Ctx, namespace, metav1.CreateOptions{})
		if err != nil {
			slog.Error("create namespace error", slog.String("error", err.Error()))
		}
	}

}

func (c MetricsInstall) Apply(baseDir, file string) error {
	kubectlArgs := []string{
		"apply",
		"--validate=false",
		"-f",
		baseDir + file,
	}
	sstr, estr, err := helper.Runsh("kubectl", kubectlArgs...)
	if err != nil {
		slog.Error("kubectl apply error", slog.String("error", estr))
		return err
	}
	print(sstr)
	return nil
}
