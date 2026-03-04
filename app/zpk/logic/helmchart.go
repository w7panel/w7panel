package logic

import (
	"bytes"
	"errors"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	zpktypes "gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress"
	convert "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	helm "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	v1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	"github.com/samber/lo"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	yaml "sigs.k8s.io/yaml/goyaml.v3"
	// applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	// appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
)

func LocateChartByHelmZpk(repository, chartName, zipurl, version string) (*chart.Chart, error) {
	if repository == "" && chartName == "" {
		loader := helm.NewZipHelmChartLoader(zipurl)
		return loader.Load()
	} else {
		return LocateChartByHelm(repository, chartName, version)
	}
}

func LocateChartByHelm(respo, chartName, version string) (*chart.Chart, error) {
	client, err := registry.NewClient(registry.ClientOptPlainHTTP())
	if err != nil {
		return nil, err
	}
	return k8s.LocateChart(respo, chartName, true, client, version)
}

type HelmChart struct {
	zpktypes.Package
	zpktypes.ShellType
}

func NewHelmChart(pack zpktypes.Package, shellType zpktypes.ShellType) *HelmChart {
	return &HelmChart{
		pack,
		shellType,
	}
}

func (h *HelmChart) toChartMetaYaml() ([]byte, error) {
	return yaml.Marshal(h.Package.PackageChartMetadata)
}

func (h *HelmChart) GetValues() (map[string]interface{}, error) {
	settings := cli.New()
	optValues := &values.Options{}
	for _, params := range h.Root.Manifest.Platform.Container.StartParams {
		optValues.Values = append(optValues.Values, params.Name+"="+params.ValuesText)
	}
	// for _, env := range h.Root.Manifest.Platform.Container.Env {
	// 	optValues.Values = append(optValues.Values, env.Name+"="+env.Value)
	// }

	provider := getter.All(settings)
	vals, err := optValues.MergeValues(provider)
	if err != nil {
		return nil, err
	}
	return vals, nil
}

// 废弃
func (h *HelmChart) ToHelmChart() (*chart.Chart, error) {

	if h.Package.IsHelm() {
		return h.helmChart()
	}
	_, chart, err := h.convertManifestToChart()
	return chart, err
}

func (h *HelmChart) ToHelmChartWithGroup() (*v1alpha1.AppGroup, *chart.Chart, error) {
	return h.convertManifestToChart()
}

func fillHelmSet(packageApp *types.PackageApp, childName string, ignore []string, fillfullName bool) string {
	set := ""
	for _, params := range packageApp.Manifest.Platform.Container.StartParams {
		if lo.Contains(ignore, params.Name) {
			continue
		}
		set += " --set " + childName + params.Name + "='" + params.ValuesText + "'"
	}
	for _, env := range packageApp.Manifest.Platform.Container.Env {
		set += " --set " + childName + env.Name + "='" + env.Value + "'"
	}

	if packageApp.PvcName != "" {
		set += " --set PVC_NAME=" + (packageApp.PvcName)
	}

	set += " --set " + "replicas=" + strconv.Itoa(int(packageApp.Replicas))
	// 完全交给helm 外部不干预 只提供PVC_NAME
	// if packageApp.GetVolumeMounts() != nil && len(packageApp.GetVolumeMounts()) > 0 {
	// 	jsonstr, err := helper.ToJson(packageApp.GetVolumeMounts())
	// 	if err != nil {
	// 		slog.Error("helm install job", "error", err)
	// 	} else {
	// 		set += " --set-json '" + childName + "volumeMounts=" + jsonstr + "'"
	// 	}
	// }
	// if packageApp.GetVolumes() != nil && len(packageApp.GetVolumes()) > 0 {
	// 	jsonstr, err := helper.ToJson(packageApp.GetVolumes())
	// 	if err != nil {
	// 		slog.Error("helm install job", "error", err)
	// 	} else {
	// 		set += " --set-json '" + childName + "volumes=" + jsonstr + "'"
	// 	}
	// }
	if fillfullName {
		set += " --set " + childName + "fullnameOverride=" + packageApp.GetName()
	}
	return set
}
func toHelmInstallJob(packageApp *types.PackageApp, children []*types.PackageApp) *batchv1.Job {
	// releaseName := packageApp.GetReleaseName()
	releaseName := packageApp.GetReleaseName()
	// if !packageApp.IsHelm() {
	// packageApp.Manifest.Platform.Helm.ChartName = packageApp.HelmUrl
	// }
	packageApp.Manifest.Platform.Helm.ChartName = packageApp.HelmUrl //统一使用新的helmUrl
	helmConfig := packageApp.Manifest.Platform.Helm
	labels := packageApp.GetLabels()
	anno := packageApp.GetAnnotations()
	shellCmd := "/ko-app/k8s-offline helmgo --chartName=" + helmConfig.ChartName + " --namespace=" + packageApp.Namespace + " --repository=" + helmConfig.Repository + " --zipUrl=" + packageApp.ZipUrl + " --releaseName=" + releaseName + ""
	atomic := false
	set := fillHelmSet(packageApp, "", []string{"HELM_ATOMIC", "DOMAIN_URL"}, false) //pvc 站点管理 会新建一个名字出来

	if !packageApp.IsHelm() {
		for _, child := range children {
			cname := child.Manifest.Application.Identifie
			cname = strings.ReplaceAll(cname, "_", "-")
			newSet := fillHelmSet(child, cname+".", []string{"HELM_ATOMIC", "DOMAIN_URL"}, true)
			set += newSet
		}
	}
	if !packageApp.IsHelm() {

		if packageApp.IngressHost != "" {
			shellCmd += " --set ingressHost=" + packageApp.IngressHost
			shellCmd += " --set DOMAIN_URL=" + (packageApp.IngressHost)
			// shellCmd += " --set DOMAIN_URL=" + (packageApp.IngressHost)
		}
		if packageApp.IngressClassName != "" {
			shellCmd += " --set ingressClassName=" + packageApp.IngressClassName
		}
		if packageApp.IngressForceHttps {
			shellCmd += " --set ingressForceHttps=" + helper.BoolToString(packageApp.IngressForceHttps)
		}

		if packageApp.IngressSeletorName != "" {
			shellCmd += " --set ingressSelectorName=" + (packageApp.IngressSeletorName)
		}
		if packageApp.GetVolumeMounts() != nil && len(packageApp.GetVolumeMounts()) > 0 {
			jsonstr, err := helper.ToJson(packageApp.GetVolumeMounts())
			if err != nil {
				slog.Error("helm install job", "error", err)
			} else {
				shellCmd += " --set-json 'volumeMounts=" + jsonstr + "'"
			}
		}
		if packageApp.GetVolumes() != nil && len(packageApp.GetVolumes()) > 0 {
			jsonstr, err := helper.ToJson(packageApp.GetVolumes())
			if err != nil {
				slog.Error("helm install job", "error", err)
			} else {
				shellCmd += " --set-json 'volumes=" + jsonstr + "'"
			}
		}
		shellCmd += " --set 'backend_identifier=" + packageApp.GetName() + "'"
		shellCmd += " --set 'backend_identifie=" + packageApp.GetName() + "'"
	}

	for _, env := range packageApp.Manifest.Platform.Container.Env {
		if (env.Name == "HELM_ATOMIC") && env.Value == "true" {
			atomic = true
		}
	}

	if atomic {
		shellCmd += " --atomic"
	}
	labelstr := ""
	for k, v := range labels {
		labelstr += " --labels " + k + "='" + v + "'"
	}
	annostr := ""
	for k, v := range anno {
		// 注解字段异常导致安装失败
		// if v == "" || k == "" {
		// 	continue
		// }
		if k == "w7.cc/shells" { //临时处理
			continue
		}
		annostr += " --anno " + k + "=\"" + v + "\""
	}
	if len(set) > 0 {
		shellCmd += set
	}
	if len(labelstr) > 0 {
		shellCmd += labelstr
	}
	if len(annostr) > 0 {
		shellCmd += annostr
	}
	if len(helmConfig.Version) > 0 {
		shellCmd += " --version=" + helmConfig.Version
	}
	shell := &types.Shell{
		Title: "helm安装" + packageApp.GetTitle(),
		Type:  "helm",
		Shell: shellCmd,
	}
	slog.Debug("helm install job", "shellCmd", shellCmd)
	job := helm.ToHelmShellJob(packageApp, shell)
	return job
}

// +Declared
func (h *HelmChart) helmChart() (*chart.Chart, error) {
	helmConfig := h.Package.GetHelm()
	if helmConfig.Repository == "" && helmConfig.ChartName == "" {
		loader := helm.NewZipHelmChartLoader(h.Package.Root.ZipUrl)
		return loader.Load()
	} else {
		client, err := registry.NewClient(registry.ClientOptPlainHTTP())
		if err != nil {
			return nil, err
		}
		return k8s.LocateChart(helmConfig.Repository, helmConfig.ChartName, true, client, helmConfig.Version)
	}
}

func (h *HelmChart) convertManifestToChart() (*v1alpha1.AppGroup, *chart.Chart, error) {
	var files []*loader.BufferedFile
	chartYaml, err := h.toChartMetaYaml()
	if err != nil {
		return nil, nil, err
	}
	files = append(files, &loader.BufferedFile{Name: "Chart.yaml", Data: []byte(chartYaml)})
	deployItems := []v1alpha1.DeployItem{}
	//root app
	root := h.Root
	root.AppGroupInstallResult = &v1alpha1.DeployItem{
		ResourceList: make([]v1alpha1.ResourceInfo, 0),
		DeployStatus: v1alpha1.StatusDeploying,
		Title:        root.GetTitle(),
		Identifie:    root.Identifie,
	}
	parent := root.Parent
	if parent == nil {
		parent = root
	}
	// if !root.IsHelm() {
	convertFiles, err := h.toBufferFiles(root, parent, true)
	if err != nil {
		return nil, nil, err
	}
	files = append(files, convertFiles...)
	deployItems = append(deployItems, *root.AppGroupInstallResult)
	// }

	// 子应用

	if root.Parent == nil {
		for _, packageApp := range h.Children {
			if packageApp.IsHelm() {
				// continue
			}
			packageApp.AppGroupInstallResult = &v1alpha1.DeployItem{
				ResourceList: make([]v1alpha1.ResourceInfo, 0),
				DeployStatus: v1alpha1.StatusDeploying,
				Title:        packageApp.GetTitle(),
				Identifie:    packageApp.Identifie,
			}
			convertFiles, err := h.toBufferFiles(packageApp, root, false)
			if err != nil {
				return nil, nil, err
			}

			files = append(files, convertFiles...)
			deployItems = append(deployItems, *packageApp.AppGroupInstallResult)
		}
	}

	group := helm.ToAppGroup(h.Root, deployItems)
	// if root.Parent != nil {
	// 	group.Labels["w7.cc/parent"] = parent.GetName()
	// }

	chart, err := loader.LoadFiles(files)
	if err != nil {
		return nil, nil, err
	}
	return group, chart, nil
}

func (h *HelmChart) convertToYaml(obj runtime.Object, filename string) (*loader.BufferedFile, error) {
	yaml, err := h.toYaml(obj)
	if err != nil {
		return nil, err
	}
	return &loader.BufferedFile{Name: filepath.Join("templates", filename), Data: yaml}, nil
}

func (h *HelmChart) appendResourceInfo(packageApp *zpktypes.PackageApp, obj runtime.Object, status string) {
	if status == "" {
		status = v1alpha1.StatusDeploying
	}
	resourceInfo := v1alpha1.ResourceInfo{
		Name:         obj.DeepCopyObject().(metav1.Object).GetName(),
		Namespace:    packageApp.Namespace,
		Kind:         obj.GetObjectKind().GroupVersionKind().Kind,
		ApiVersion:   obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		DeployStatus: status,
		DeployTitle:  obj.DeepCopyObject().(metav1.Object).GetAnnotations()["w7.cc/deploy-title"],
	}
	packageApp.AppGroupInstallResult.ResourceList = append(packageApp.AppGroupInstallResult.ResourceList, resourceInfo)
}

func (h2 *HelmChart) toBufferFiles(packageApp *zpktypes.PackageApp, root *zpktypes.PackageApp, isRoot bool) ([]*loader.BufferedFile, error) {

	var files []*loader.BufferedFile

	for _, ing := range packageApp.Manifest.Platform.Ingress {
		ing.ReplaceSvcName(packageApp)
	}
	if h2.ShellType == zpktypes.ShellUpgrade {
		packageApp.InstallOption.IsUpgrade = true
	}

	// if packageApp.IsHelm() && false { //暂不支持普通应用 包含helm 子应用
	if packageApp.IsHelm() && false {
		//暂不支持普通应用 包含helm 子应用 //如果helm安装了 appgroup新建后 helm安装命令又把这个appgroup 删除了 因为helm和appgroup同名
		// 如果不同名，还得做一层关联关系

		// if isRoot {
		// 	return []*loader.BufferedFile{}, nil
		// }
		cloneApp := packageApp

		shellType := h2.ShellType

		//helm shell job
		shell := cloneApp.GetShellByType(string(shellType))
		var shellJob *batchv1.Job
		if shell != nil {
			shellJob = convert.ToShellJob2(cloneApp, cloneApp, string(shellType))
			if shellJob != nil {
				h2.appendResourceInfo(cloneApp, shellJob, "")
			}
			file, err := h2.convertToYaml(shellJob, cloneApp.Identifie+"-"+string(shellType)+"-job.yaml")
			if err != nil {
				return nil, err
			}
			files = append(files, file)
			h2.appendResourceInfo(cloneApp, shellJob, "")
		}
		//安装helm job
		job := toHelmInstallJob(cloneApp, []*types.PackageApp{})
		file, err := h2.convertToYaml(job, cloneApp.Identifie+"-helm-job.yaml")
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		h2.appendResourceInfo(cloneApp, job, "") //必须是packageApp 不能用clone 对象 因为clone对象在后续会被修改 导致后续生成的appgroup 资源列表不对

		info := v1alpha1.ResourceInfo{
			Name:         job.Name,
			Namespace:    cloneApp.GetNamespace(),
			Kind:         "Job",
			ApiVersion:   "batch/v1",
			DeployStatus: v1alpha1.StatusDeploying,
			DeployTitle:  "helm安装",
		}

		installResult := v1alpha1.DeployItem{
			Identifie:    cloneApp.GetIdentifie(),
			Title:        cloneApp.GetTitle(),
			ResourceList: []v1alpha1.ResourceInfo{info},
			DeployStatus: v1alpha1.StatusDeploying,
		}
		if shellJob != nil {
			shellInfo := v1alpha1.ResourceInfo{
				Name:         shellJob.Name,
				Namespace:    cloneApp.GetNamespace(),
				Kind:         "Job",
				ApiVersion:   "batch/v1",
				DeployStatus: v1alpha1.StatusDeploying,
				DeployTitle:  shell.GetTitle(),
			}
			installResult.ResourceList = append(installResult.ResourceList, shellInfo)
		}
		// 如果是root
		if isRoot {
			return files, nil
		}
		group := helm.ToAppGroup(cloneApp, []v1alpha1.DeployItem{installResult})
		if !isRoot {
			group.Labels["w7.cc/parent"] = root.GetName()
		}

		groupfile, err := h2.convertToYaml(group, cloneApp.Identifie+"-appgroup.yaml")
		if err != nil {
			return nil, err
		}
		files = append(files, groupfile)

		return files, nil
	}

	shellfile, err := h2.convertJob(packageApp, h2.ShellType, true)
	if err == nil {
		files = append(files, shellfile)

	}

	if packageApp.RequireBuildImage() {
		buildJob := helm.ToBuildJob(packageApp, packageApp, string(h2.ShellType))
		file, err := h2.convertToYaml(buildJob, packageApp.Identifie+"-buildjob.yaml")
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		h2.appendResourceInfo(packageApp, buildJob, "")
	}
	// if packageApp.SupportMicroApp() {
	// 	microapp := helm.ToMicroApp(packageApp)
	// 	file, err := h2.convertToYaml(microapp, packageApp.Identifie+"-microapp.yaml")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	files = append(files, file)
	// }

	deployment := helm.ToDeployment(packageApp)
	file, err := h2.convertToYaml(deployment, packageApp.Identifie+"-deployment.yaml")
	if err != nil {
		return nil, err
	}
	files = append(files, file)
	h2.appendResourceInfo(packageApp, deployment, "")
	service := helm.ToService(packageApp)
	file, err = h2.convertToYaml(service, packageApp.Identifie+"-service.yaml")
	if err != nil {
		return nil, err
	}
	files = append(files, file)
	// h2.appendResourceInfo(packageApp, service)
	if len(packageApp.GetServiceLbPort()) > 0 {
		serviceLb := helm.ToLoadBalancerService(packageApp)
		file, err = h2.convertToYaml(serviceLb, packageApp.Identifie+"-service-lb.yaml")
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		// h2.appendResourceInfo(packageApp, serviceLb)
	}

	if packageApp.IngressHost != "" && higress.NeedCheckBeian() {
		// h2.convertJob()
		job := helm.ToBeianCheckJob(packageApp, packageApp.IngressHost)
		file, err := h2.convertToYaml(job, packageApp.Identifie+"-beiancheckjob.yaml")
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		h2.appendResourceInfo(packageApp, job, "")
	}

	ingresses := helm.ToIngresses(packageApp)
	for key, ingress := range ingresses {
		keyStr := strconv.Itoa(key)
		file, err = h2.convertToYaml(&ingress, packageApp.Identifie+"-"+keyStr+"-ingress.yaml")
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		if key == 0 { //只返回第一个ingress的resourceinfo
			h2.appendResourceInfo(packageApp, &ingress, v1alpha1.StatusDeployed) //域名暂不考虑解析状态， 否则应用长时间处于部署中状态
			continue
		}
		// h2.appendResourceInfo(packageApp, &ingress)
	}

	return files, nil

}

func (h *HelmChart) toYaml(obj runtime.Object) ([]byte, error) {
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	buf := bytes.NewBuffer([]byte{})
	if err := s.Encode(obj, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h *HelmChart) toAppGroup(installResult []v1alpha1.DeployItem) *v1alpha1.AppGroup {
	return helm.ToAppGroup(h.Root, installResult)

}

func (h2 *HelmChart) convertJob(packageApp *types.PackageApp, shellType zpktypes.ShellType, appendResource bool) (*loader.BufferedFile, error) {
	shellJob := helm.ToShellJob2(packageApp, packageApp, string(shellType))
	if shellJob != nil {
		if shellType == zpktypes.ShellRequireInstall { //安装前执行安装crd or other
			// shellJob.Annotations["helm.sh/hook"] = "pre-install, pre-upgrade"
		}
		if shellType == zpktypes.ShellUninstall {
			shellJob.Annotations["helm.sh/hook"] = "post-delete"
		}

		shellJob.Annotations["helm.sh/resource-policy"] = "keep"
		// shellJob.Annotations["helm.sh/hook-delete-policy"] = "before-hook-creation"
		file, err := h2.convertToYaml(shellJob, packageApp.Identifie+"-"+string(shellType)+"job.yaml")
		if err != nil {
			return nil, err
		}
		if appendResource {
			h2.appendResourceInfo(packageApp, shellJob, "")
		}
		// files = append(files, file)
		return file, nil
	}
	return nil, errors.New("no shell job")
}
