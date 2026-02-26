package gpu

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	v1alpha1Types "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	gpuclassv1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/gpuclass/v1alpha1"
	appgroup "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/appgroup/clientset/versioned"
	gpuclassclientset "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/gpuclass/clientset/versioned"
	helmclientSet "github.com/k3s-io/helm-controller/pkg/generated/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types2 "k8s.io/apimachinery/pkg/types"
)

const K3SGPU = "k3s.gpu"

type GpuResponse struct {
	GpuOperatorIsSuccess bool `json:"gpuOperatorIsSuccess"`
	HamiIsSuccess        bool `json:"hamiIsSuccess"`

	GpuOperatorIsDeployed bool `json:"gpuOperatorIsDeployed"`
	HamiIsDeployed        bool `json:"hamiIsDeployed"`

	CanInstallGpuOperator bool `json:"canInstallGpuOperator"`
	CanInstallHami        bool `json:"canInstallHami"`

	GpuIsEnabled  bool `json:"gpuEnabled"`
	CanEnabledGpu bool `json:"canEnabledGpu"`

	GpuOperatorMode string `json:"gpuOperatorMode"`
	HamiMode        string `json:"hamiMode"`
	VgpuMode        string `json:"vgpuMode"`
}

type GpuManager struct {
	sdk            *k8s.Sdk
	helmchartsdk   *helmclientSet.Clientset
	helmReleaseApi *k8s.Helm
	gpuClassApi    gpuclassclientset.Interface
	gpuClassName   string
	namespace      string
	appgroupClient *appgroup.Clientset
}

func InitK3sGpu(sdk *k8s.Sdk) error {
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k3s.gpu",
			Namespace: "kube-system",
		},
		Data: map[string]string{"enabled": "false", "gpu-operator-mode": "0", "hami-mode": "0"},
	}
	_, err := sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Create(sdk.Ctx, configmap, metav1.CreateOptions{})
	if err != nil {
		// slog.Error("Failed to create ConfigMap: %v", "err", err)
		return err
	}
	return err
}
func PatchK3sGpu(sdk *k8s.Sdk, data map[string]string) error {
	patchData := map[string]interface{}{
		"data": data,
	}

	// 5. 将 patchData 转换为 JSON 格式
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		slog.Error("Failed to marshal patch data: %v", "err", err)
		return err
	}
	// 6. 执行 Patch 操作
	_, err = sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Patch(
		sdk.Ctx,
		"k3s.gpu",
		types2.MergePatchType, // 使用 Merge Patch 类型
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		slog.Error("Failed to patch ConfigMap: %v", "err", err)
	}
	return err
}

func NewGpuManager(sdk *k8s.Sdk, namespace string, gpuClassName string) (*GpuManager, error) {
	config, err := sdk.ToRESTConfig()
	if err != nil {
		slog.Error("failed to create rest config", "error", err)
		return nil, err
	}
	hsdk, err := helmclientSet.NewForConfig(config)
	if err != nil {
		slog.Error("failed to create helm client set", "error", err)
		return nil, err
	}
	gpuClassApi, err := gpuclassclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	if namespace == "" {
		namespace = "default"
	}
	if gpuClassName == "" {
		gpuClassName = "nvidia"
	}
	appgroupClient, err := appgroup.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &GpuManager{
		sdk:            sdk,
		helmchartsdk:   hsdk,
		helmReleaseApi: k8s.NewHelm(sdk),
		gpuClassApi:    gpuClassApi,
		gpuClassName:   gpuClassName,
		appgroupClient: appgroupClient,
	}, nil
}

func (g *GpuManager) ToJsonStruct() *GpuResponse {
	return &GpuResponse{
		GpuOperatorIsSuccess:  g.GpuOperatorIsSuccess(), //是否安装成功
		HamiIsSuccess:         g.HamiIsSuccess(),
		GpuOperatorIsDeployed: g.GpuOperatorIsDeployed(), //是否已安装
		HamiIsDeployed:        g.HamiIsDeployed(),

		CanInstallGpuOperator: g.CanInstallGpuOperator(), //是否可以安装gpu-operator
		CanInstallHami:        g.CanInstallHami(),        //是否可以安装hami

		GpuIsEnabled:  g.GpuIsEnabled(),  //是否已经启用GPU
		CanEnabledGpu: g.CanEnabledGpu(), //是否可以启用GPU

		GpuOperatorMode: g.GpuOperatorMode(),
		HamiMode:        g.HamiMode(),
		VgpuMode:        g.HamiMode(),
	}
}

func (g *GpuManager) GetGpuClass() (*gpuclassv1alpha1.GpuClass, error) {
	gc, err := g.gpuClassApi.GpuclassV1alpha1().GpuClasses(g.namespace).Get(g.sdk.Ctx, g.gpuClassName, metav1.GetOptions{})
	return gc, err
}

func (g *GpuManager) GetAppGroup(releaseName string) (*v1alpha1Types.AppGroup, error) {
	gc, err := g.appgroupClient.AppgroupV1alpha1().AppGroups(g.namespace).Get(g.sdk.Ctx, releaseName, metav1.GetOptions{})
	return gc, err
}

func (g *GpuManager) GetAppGroupByIndentifie(identifie string) (*v1alpha1Types.AppGroup, error) {
	gclist, err := g.appgroupClient.AppgroupV1alpha1().AppGroups(g.namespace).List(g.sdk.Ctx, metav1.ListOptions{LabelSelector: "w7.cc/identifie=" + identifie})
	if err != nil {
		return nil, err
	}
	if len(gclist.Items) == 0 {
		return nil, fmt.Errorf("app group %s not found", identifie)
	}
	return &gclist.Items[0], nil
}

func (g *GpuManager) GetAppGroupUseGpuClass() (*v1alpha1Types.AppGroup, error) {
	gpuclass, err := g.GetGpuClass()
	if err != nil {
		slog.Error("failed to get gpu class", "error", err)
		return nil, err
	}
	if gpuclass.Spec.AppName == "" {
		return nil, fmt.Errorf("gpu class %s not found", g.gpuClassName)
	}
	gc, err := g.appgroupClient.AppgroupV1alpha1().AppGroups(g.namespace).Get(g.sdk.Ctx, gpuclass.Spec.AppName, metav1.GetOptions{})
	return gc, err
}

func (g *GpuManager) GpuOperatorIsSuccess() bool {
	group, err := g.GetAppGroupByIndentifie(NVIDIA_IDENTIFIE)
	if err != nil {
		return false
	}
	return group.Status.DeployStatus == v1alpha1Types.StatusDeployed
}

func (g *GpuManager) HamiIsSuccess() bool {
	group, err := g.GetAppGroupByIndentifie(HAMI_IDENTIFIE)
	if err != nil {
		return false
	}
	return group.Status.DeployStatus == v1alpha1Types.StatusDeployed
}

// 安装中
func (g *GpuManager) GpuOperatorIsDeployed() bool {
	group, err := g.GetAppGroupByIndentifie(NVIDIA_IDENTIFIE)
	if err != nil {
		return false
	}
	return group.Status.DeployStatus == v1alpha1Types.StatusDeploying
}

func (g *GpuManager) HamiIsDeployed() bool {
	group, err := g.GetAppGroupByIndentifie(HAMI_IDENTIFIE)
	if err != nil {
		return false
	}
	return group.Status.DeployStatus == v1alpha1Types.StatusDeploying
}

func (g *GpuManager) GpuIsEnabled() bool {
	configmap, err := g.sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Get(g.sdk.Ctx, "k3s.gpu", metav1.GetOptions{})
	if err != nil {
		return false
	}
	return configmap.Data["enabled"] == "true"
}

// 0 未安装 1 已安装但未成功 2 已成功 3 不能安装
func (g *GpuManager) GpuOperatorMode() string {
	// return "0"
	if g.GpuOperatorIsSuccess() {
		return "2"
	}
	if g.GpuOperatorIsDeployed() {
		return "1"
	}
	return "0"
}

// 0 未安装 1 已安装但未成功 2 已成功 3 不能安装
func (g *GpuManager) HamiMode() string {

	// return "0"
	if g.HamiIsSuccess() {
		return "2"
	}
	if g.HamiIsDeployed() {
		return "1"
	}
	// if g.CanInstallHami() {
	// 	return "3"
	// }
	return "0"
}

func (g *GpuManager) CanInstallHami() bool {
	return g.GpuOperatorIsSuccess() && !g.HamiIsDeployed()
}

func (g *GpuManager) CanInstallGpuOperator() bool {
	return !g.GpuOperatorIsSuccess() && !g.GpuOperatorIsDeployed()
}

func (g *GpuManager) CanEnabledGpu() bool {
	return true
	// return g.GpuOperatorIsSuccess() && g.HamiIsSuccess()
}

// 暂时不验证是否安装gpu-operator
func (g *GpuManager) GpuEnabled(ok bool) error {
	return PatchK3sGpu(g.sdk, map[string]string{
		"enabled": strconv.FormatBool(ok),
	})
}

func (g *GpuManager) UninstallHami() error {
	err := g.helmchartsdk.HelmV1().HelmCharts("kube-system").Delete(g.sdk.Ctx, "hami", metav1.DeleteOptions{})
	if err != nil {
		slog.Error("uninstall hami charts error", "err", err)
	}
	return err
	// err := g.helmReleaseApi.Uninstall("hami", "hami-system")
	// if err != nil {
	// 	slog.Error("uninstall hami error", "err", err)
	// }
	// return err
}

func (g *GpuManager) UninstallGpuOperator() error {
	err := g.helmchartsdk.HelmV1().HelmCharts("kube-system").Delete(g.sdk.Ctx, "gpu-operator", metav1.DeleteOptions{})
	if err != nil {
		slog.Error("uninstall gpu-operator charts error", "err", err)
	}
	return err
	// err := g.helmReleaseApi.Uninstall("gpu-operator", "gpu-operator")
	// if err != nil {
	// 	slog.Error("uninstall gpu-operator error", "err", err)
	// }
	// return err
}

func (g *GpuManager) InstallGpuOperator(driverEnabled bool, driverVerison string) error {
	g.UninstallGpuOperator()
	dstr := "true"
	if !driverEnabled {
		dstr = "false"
	}
	//判读目录是否存在
	yaml := `
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: gpu-operator
  namespace: kube-system
spec:
  chart: gpu-operator
  targetNamespace: gpu-operator
  createNamespace: true
  repo: https://helm.ngc.nvidia.com/nvidia
  set:
    driver.enabled: "` + dstr + `"
    devicePlugin.enabled: "false"
    toolkit.enabled: "true"
    toolkit.env[0].name: CONTAINERD_CONFIG
    toolkit.env[0].value: /var/lib/rancher/k3s/agent/etc/containerd/config.toml
    toolkit.env[1].name: CONTAINERD_SOCKET
    toolkit.env[1].value: /run/k3s/containerd/containerd.sock
    toolkit.env[2].name: CONTAINERD_RUNTIME_CLASS
    toolkit.env[2].value: nvidia`
	print(yaml)
	if driverVerison != "" {
		yaml = yaml + `    driver.version: "` + driverVerison + `"`
	}
	err := g.sdk.ApplyBytes([]byte(yaml), *k8s.NewApplyOptions("kube-system"))
	if err != nil {
		slog.Error("gpu operator error", "err", err)
	}
	return err

}

func (g *GpuManager) InstallHami(runtimeClassName string) error {
	g.UninstallHami()
	if runtimeClassName == "" {
		runtimeClassName = "nvidia"
	}
	yaml := `
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: hami
  namespace: kube-system
spec:
  chart: hami
  targetNamespace: hami-system
  createNamespace: true
  repo: https://project-hami.github.io/HAMi/
  set:
    devicePlugin.runtimeClassName: "` + runtimeClassName + `"
    devicePlugin.passDeviceSpecsEnabled: "false"`

	print(yaml)
	err := g.sdk.ApplyBytes([]byte(yaml), *k8s.NewApplyOptions("kube-system"))
	if err != nil {
		slog.Error("gpu operator error", "err", err)
	}
	return err
}
