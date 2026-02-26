package logic

import (
	coreerr "errors"
	"log/slog"
	"strconv"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	zpktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
	typealpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	"github.com/samber/lo"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DependEnv struct {
	*k8s.Sdk
	*k8s.Helm
	*appgroup.AppGroupApi
}

func NewDependEnv(client *k8s.Sdk) *DependEnv {

	groupApi, err := appgroup.NewAppGroupApi(client)
	if err != nil {
		return nil
	}
	return &DependEnv{
		Sdk:         client,
		Helm:        k8s.NewHelm(client),
		AppGroupApi: groupApi,
		//Kubectl: k8s.NewKubectl(client),
	}
}

type DependEnvResult struct {
	Installed     bool              `json:"installed" binding:"required"`
	Name          string            `json:"name"`
	StaticUrl     string            `json:"staticUrl"` // 静态资源地址
	Envs          map[string]string `json:"envs" binding:"required"`
	PvcName       string            `json:"pvcName"`
	PreSubPath    map[string]string `json:"preSubPath"`
	PreVolumeName map[string]string `json:"preVolumeName"` //不需要前端处理
	Replicas      int32             `json:"replicas"`
	VolumesMounts []corev1.VolumeMount
	Volumes       []corev1.Volume
}

const (
	W7_V = "w7-pros-28692"
	W7_S = "w7-pros-28693"
	W7_X = "w7-pros-28694"
)

type deployName struct {
	Name      string
	Namespace string
}

func getSuffix(releaseName string) string {
	suffix := releaseName
	rp := strings.Split(suffix, "-")
	if len(rp) > 1 {
		suffix = rp[len(rp)-1] //获取最后一个字段作为后缀
	}
	return strings.ToLower(suffix)
}

func getMayBeNames(name string) []string {
	suffix := getSuffix(name)
	if strings.HasPrefix(name, W7_V) {
		return []string{W7_S + "-" + suffix, W7_X + "-" + suffix}
	}
	if strings.HasPrefix(name, W7_S) {
		return []string{W7_V + "-" + suffix, W7_X + "-" + suffix}
	}
	if strings.HasPrefix(name, W7_X) {
		return []string{W7_V + "-" + suffix, W7_S + "-" + suffix}
	}
	return []string{}
}

func (d *DependEnv) LoadLastVersionDeploymentEnv(name, namespace string) (*DependEnvResult, error) {
	deployment, err := d.Sdk.ClientSet.AppsV1().Deployments(namespace).Get(d.Ctx, name, metav1.GetOptions{})
	// if err = nil {
	// 	return nil, err
	// }
	//w7-pros-28694-jyvtanqm9x
	if err == nil {
		return d.GetDeploymentEnv(deployment)
	}
	if err != nil {
		if errors.IsNotFound(err) {
			maybeNames := getMayBeNames(name)
			for _, maybeName := range maybeNames {
				deployment, err = d.Sdk.ClientSet.AppsV1().Deployments(namespace).Get(d.Ctx, maybeName, metav1.GetOptions{})
				if err != nil {
					continue
				}
				result, err := d.GetDeploymentEnv(deployment)
				if err == nil && result.Installed {
					return result, nil
				}
			}
		}
	}
	return nil, coreerr.New("not found")

}

// 获取上一个版本的配置
func (d *DependEnv) LoadLastVersionEnv(name, namespace string) (*DependEnvResult, error) {

	deployResult, _ := d.LoadLastVersionDeploymentEnv(name, namespace)
	var helmResult *DependEnvResult
	helmInfo, err := d.Helm.Info(name, namespace)
	if helmInfo != nil {
		helmResult, _ = d.GetHelmEnv(helmInfo)
	}
	if deployResult != nil { //应用安装变量和helm变量合并
		if helmResult != nil {
			for k, v := range helmResult.Envs {
				deployResult.Envs[k] = v
			}
		}
		return deployResult, nil
	}
	if helmResult != nil {
		return helmResult, err
	}
	return nil, coreerr.New("not found")

}

func (d *DependEnv) LoadEnv(identifie, namespace string) (*DependEnvResult, error) {
	result, err := d.LoadDeploymentEnv(identifie, namespace)
	if err != nil {
		slog.Error("LoadDeploymentEnv", "err", err)
		return d.LoadAppGroupEnv(identifie, namespace)
	}
	if !result.Installed {
		return d.LoadAppGroupEnv(identifie, namespace)
	}
	return result, nil

}

func (d *DependEnv) LoadAppGroupEnv(identifie, namespace string) (*DependEnvResult, error) {
	identifie = strings.ReplaceAll(identifie, "_", "-")
	result := &DependEnvResult{
		Installed: false,
		Envs:      make(map[string]string),
	}
	appGroupList, err := d.AppGroupApi.GetAppGroupListByIdentifie(namespace, identifie)
	if err != nil {
		slog.Error("LoadAppGroupEnv", "err", err)
		return result, err
	}
	appGroup, ok := lo.Find(appGroupList.Items, func(item typealpha1.AppGroup) bool {
		return item.Status.DeployStatus == typealpha1.StatusDeployed && item.DeletionTimestamp == nil
	})
	if ok {
		result.Installed = true
		release, err := d.Helm.Info(appGroup.Name, appGroup.Namespace)
		if err != nil {
			slog.Error("LoadAppGroupEnv", "err", err)
			return result, err
		}
		return d.GetHelmEnv(release)
	}
	return result, nil
}

func (d *DependEnv) LoadHelmEnv(identifie, namespace string) (*DependEnvResult, error) {
	identifie = strings.ReplaceAll(identifie, "_", "-")
	result := &DependEnvResult{
		Installed: false,
		Envs:      make(map[string]string),
	}
	releaseList, err := d.Helm.ListRaw(namespace, "w7.cc/identifie="+identifie)
	if err != nil {
		slog.Error("LoadHelmEnv", "err", err)
		return result, err
	}
	if len(releaseList) == 0 {
		slog.Info("LoadHelmEnvRekease len=0", "identifie", identifie)
		return result, nil
	}
	deployList := lo.Filter[*release.Release](releaseList, func(item *release.Release, _ int) bool {
		return item.Info.Status == "deployed"
	})
	if len(deployList) == 0 {
		slog.Info("LoadHelmEnvDeploy len=0", "identifie", identifie)
		return result, nil
	}
	release := deployList[0]
	return d.GetHelmEnv(release)
}

func (d *DependEnv) LoadDeploymentEnv(identifie, namespace string) (*DependEnvResult, error) {
	identifie = strings.ReplaceAll(identifie, "_", "-")
	result := DependEnvResult{
		Installed: false,
		Envs:      make(map[string]string),
	}
	apps, err := d.Sdk.GetDeploymentAppByIdentifie(namespace, identifie)
	if err != nil {
		return nil, err
	}
	result.Installed = len(apps.Items) > 0
	if len(apps.Items) > 0 {
		return d.GetDeploymentEnv(&apps.Items[0])
	}
	return &result, nil
}

func (d *DependEnv) GetDeploymentEnv(first *v1.Deployment) (*DependEnvResult, error) {
	result := &DependEnvResult{
		Installed: true,
		Name:      first.Name,
		StaticUrl: first.Annotations[zpktypes.HELM_STATIC_URL],
		Envs:      make(map[string]string),
	}
	for _, env := range first.Spec.Template.Spec.Containers[0].Env {
		result.Envs[env.Name] = env.Value
	}
	svcName := helper.ClusterDomain(first.Name, first.Namespace) //first.Name + "." + first.Namespace + ".svc.cluster.local"
	result.Envs["HOST"] = svcName
	result.Envs["PORT"] = "80" //first.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort
	if (first.Spec.Template.Spec.Containers[0].Ports != nil) || len(first.Spec.Template.Spec.Containers[0].Ports) > 0 {
		result.Envs["PORT"] = strconv.Itoa(int(first.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort))
	}
	result.PvcName = ""
	if (first.Spec.Template.Spec.Volumes != nil) && len(first.Spec.Template.Spec.Volumes) > 0 {
		for _, v := range first.Spec.Template.Spec.Volumes {
			if v.PersistentVolumeClaim != nil {
				result.PvcName = v.PersistentVolumeClaim.ClaimName
			}
		}
	}

	identifie := first.Labels["w7.cc/identifie"]
	if (identifie == "w7-mysql") || (identifie == "w7-mysql5") {
		result.Envs["MYSQL_HOST"] = svcName
		result.Envs["MYSQL_PORT"] = result.Envs["PORT"]
	}
	if identifie == "w7-redis" {
		result.Envs["REDIS_HOST"] = svcName
		result.Envs["REDIS_PORT"] = result.Envs["PORT"]
	}
	container := first.Spec.Template.Spec.Containers[0]
	mounts := container.VolumeMounts
	result.PreSubPath = make(map[string]string)
	result.PreVolumeName = make(map[string]string)
	for _, mount := range mounts {
		result.PreSubPath[mount.MountPath] = mount.SubPath
		result.PreVolumeName[mount.MountPath] = mount.Name
	}

	result.VolumesMounts = container.VolumeMounts
	result.Volumes = first.Spec.Template.Spec.Volumes
	result.Replicas = *first.Spec.Replicas
	return result, nil
}

func (d *DependEnv) GetHelmEnv(release *release.Release) (*DependEnvResult, error) {
	result := &DependEnvResult{
		Installed: false,
		Name:      release.Name,
		StaticUrl: release.Chart.Metadata.Annotations[zpktypes.HELM_STATIC_URL],
		Envs:      make(map[string]string),
	}
	config := release.Config
	if release.Info.Status == "deployed" {
		result.Installed = true
	}

	var flattenMap func(prefix string, value interface{})
	flattenMap = func(prefix string, value interface{}) {
		switch v := value.(type) {
		case map[string]interface{}:
			for k, val := range v {
				newKey := k
				if prefix != "" {
					newKey = prefix + "." + k
				}
				flattenMap(newKey, val)
			}
		case []interface{}:
			for i, val := range v {
				newKey := prefix + "." + strconv.Itoa(i)
				flattenMap(newKey, val)
			}
		case string:
			result.Envs[prefix] = v
		case int:
			result.Envs[prefix] = strconv.Itoa(v)
		case float64:
			result.Envs[prefix] = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			result.Envs[prefix] = strconv.FormatBool(v)
		case nil:
			result.Envs[prefix] = ""
		default:
			result.Envs[prefix] = ""
		}
	}

	for key, value := range config {
		flattenMap(key, value)
	}
	return result, nil
}

func (d *DependEnv) ReplacePackageAddConfig(p *types.ManifestPackage, config types.PackageAddConfig) types.PackageAddConfig {
	name := config.ReleaseName
	if !config.IsHelm {
		suffix := getSuffix(name)
		name = types.GetDeployName(config.Identifie, suffix)
	}
	if config.Namespace == "" {
		config.Namespace = "default"
	}
	result, err := d.LoadLastVersionEnv(name, config.Namespace)
	if err != nil {
		slog.Error("ReplacePackageAddConfig", "err", err)
		return config
	}
	// fill params
	params := config.StartParams
	for k, param := range params {
		if val, ok := result.Envs[param.Name]; ok {
			has := lo.Contains(types.Z, param.ValuesText)
			if has { //有些占位不需要替换
				param.Lock = true
				params[k] = param
				continue
			}
			param.ValuesText = val
			param.ModuleName = ""
			param.Lock = true
			params[k] = param
		}
	}
	if !config.IsHelm {
		pvcName := result.PvcName
		p.Manifest.GenVolumesName(result.PreVolumeName) //重新找VolumesName
		vms := p.GetVolumeMounts(pvcName, config.ReleaseName, result.PreSubPath)
		old := result.VolumesMounts
		config.VolumeMounts = MergeVolumeMounts(old, vms)

		volumes := p.GetVolumes(pvcName)
		// config.Volumes = volumes
		config.Volumes = MergeVolumes(result.Volumes, volumes) //合并时候没法区分 改为直接前端传递传递
	}
	return config
}

func MergeVolumeMounts(existingMounts, newMounts []corev1.VolumeMount) []corev1.VolumeMount {
	result := make([]corev1.VolumeMount, len(existingMounts))
	copy(result, existingMounts)

	// 创建现有挂载名称的映射
	existingNames := make(map[string]bool)
	for _, mount := range existingMounts {
		existingNames[mount.MountPath] = true
	}

	// 只添加不重复的挂载
	for _, mount := range newMounts {
		if !existingNames[mount.MountPath] {
			result = append(result, mount)
			existingNames[mount.MountPath] = true
		}
	}

	return result
}

func MergeVolumes(baseVolumes, additionalVolumes []corev1.Volume) []corev1.Volume {
	volumeMap := make(map[string]corev1.Volume)

	// 先添加基础 volumes
	for _, volume := range baseVolumes {
		volumeMap[volume.Name] = volume
	}

	// 用新的 volumes 覆盖（如果有重复）
	for _, volume := range additionalVolumes {
		volumeMap[volume.Name] = volume
	}

	// 转换回切片
	result := make([]corev1.Volume, 0, len(volumeMap))
	for _, volume := range volumeMap {
		result = append(result, volume)
	}

	return result
}

/***
cluster:
  clusterDefinitionRef: mongodb
  clusterVersionRef: mongodb-6.0
  componentDefRef: mongodb
  storageClassName: longhorn
  storageSize: 2Gi
*/
