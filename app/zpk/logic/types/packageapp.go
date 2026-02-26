package types

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	helm "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
	v1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

type EnvKv struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// 主应用是helm 子应用是普通应用 名字拼接规则
func GetDeployName(identifie, suffix string) string {
	identifie = strings.ToLower(strings.Replace(identifie, "_", "-", -1))
	return strings.ToLower(strings.ReplaceAll(identifie+"-"+suffix, "_", "-"))
}

func getSuffix(releaseName string) string {
	suffix := releaseName
	rp := strings.Split(suffix, "-")
	if len(rp) > 1 {
		suffix = rp[len(rp)-1] //获取最后一个字段作为后缀
	}
	return strings.ToLower(suffix)
}

type ModuleName string

type ShellType string

const (
	ShellInstall        ShellType = "install"        //安装
	ShellUpgrade        ShellType = "upgrade"        //升级
	ShellUninstall      ShellType = "uninstall"      //卸载后触发脚本
	ShellRequireInstall ShellType = "requireinstall" //安装前检查
	ShellCustom         ShellType = "custom"         //自定义脚本
)

// 安装参数
type InstallOption struct {
	Identifie                string               `json:"identifie"`
	PvcName                  string               `json:"pvcname"`
	DockerRegistry           types.DockerRegistry `json:"registry"`
	DockerRegistrySecretName string               `json:"dockerRegistrySecretName"`
	Namespace                string               `json:"namespace"`
	InstallId                string               `json:"installId"`   //安装的id
	ReleaseName              string               `json:"releaseName"` //安装name
	Suffix                   string               `json:"suffix"`      //安装后缀 //releasename = root.Identifie+"-"+root.Suffix
	EnvKv                    []EnvKv              `json:"envkv"`
	IngressSeletorName       string               `json:"ingressSelectorName"`  //IngressSelector前端选择的名称
	IngressHost              string               `json:"ingressHost"`          //域名
	IngressClassName         string               `json:"ingressClassName"`     //ingressclassname
	IngressForceHttps        bool                 `json:"ingressForceHttps"`    //ingressclassname
	Replicas                 int32                `json:"replicas"`             // 可选安装的数量为0
	IsChildApp               bool                 `json:"isChild"`              //是否IsChildApp
	IsUpgrade                bool                 `json:"isUpgrade"`            //是否更新模式
	Annotations              map[string]string    `json:"annotations"`          //注解
	ServiceAccountName       string               `json:"serviceAccountName"`   //ServiceAccountName
	BuildImageSuccessUrl     string               `json:"buildImageSuccessUrl"` //ServiceAccountName
	ParentReleaseName        string               `json:"parentReleaseName"`    // 父节点发布名
	PreSubPath               map[string]string    `json:"preSubPath"`           // 上次安装的子路径
	Cpu                      string               `json:"cpu"`
	Memory                   string               `json:"memory"`
	Volumes                  []corev1.Volume      `json:"volumes"`
	VolumesMounts            []corev1.VolumeMount `json:"volumesMounts"`
	K8sToken                 *k8s.K8sToken
	// K3kMode                  string               `json:"k3kMode"`              // 子集群模式
}

// helm chart metadata 结构体
type PackageChartMetadata struct {
	AppVersion  string                 `json:"appVersion"`
	ApiVersion  string                 `json:"apiVersion"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Icon        string                 `json:"icon"`
	Description string                 `json:"description"`
	Home        string                 `json:"home"`
	Annotations map[string]interface{} `json:"annotations"`
}

type Package struct {
	PackageChartMetadata
	Root          *PackageApp `json:"root"`
	Children      []*PackageApp
	ReleaseName   string          `json:"releaseName"`
	InstallId     string          `json:"installId"`
	Namespace     string          `json:"namespace"`
	PackageType   string          `json:"packageType"`
	InstallResult []installResult `json:"installResult"` //appgroup 安装结果集合
}

type installResult struct {
	Identifie         string `json:"identifie"`
	DeploymentName    string `json:"deploymentName"`
	JobName           string `json:"jobName"`
	Title             string `json:"title"`
	RequireBuildImage bool   `json:"RequireBuildImage"` //是否需要构建docker镜像
}

func (p Package) IsHelm() bool {
	return p.PackageType == "helm"
}

func (p Package) ForceHttps(forceHttps bool) {
	p.Root.IngressForceHttps = forceHttps
}

func (p Package) GetHelm() Helm {
	app := p.Root
	return app.Manifest.Platform.Helm
}

// type PackageApps []*PackageApp

func NewPackage(mPackage *ManifestPackage, installOptions []InstallOption,
	releaseName string, installId string, namespace string, ingressHost string,
	ingressSelectorName string, IngressClassName string) Package {
	var children []*PackageApp
	var rootApp *PackageApp
	releaseName = strings.ReplaceAll(strings.ToLower(releaseName), "_", "-")
	suffix := getSuffix(releaseName)
	for _, installOption := range installOptions {
		findPackage := mPackage.GetChildren(installOption.Identifie)
		//
		if reflect.ValueOf(findPackage).IsZero() {
			continue
		}
		installOption.ReleaseName = strings.ToLower(releaseName)
		// installOption.ServiceAccountName = releaseName
		installOption.InstallId = installId
		installOption.Namespace = namespace
		installOption.Suffix = suffix
		installApp := NewPackageApp(findPackage, &installOption)

		if (findPackage.Manifest.Application.Identifie == mPackage.Manifest.Application.Identifie) && rootApp == nil {
			rootApp = installApp
		} else if findPackage.Manifest.Application.Identifie != mPackage.Manifest.Application.Identifie {
			installApp.Parent = rootApp
			children = append(children, installApp)
		}

	}
	var installResults []installResult

	rootApp.ReplaceCommon(ingressHost)
	rootApp.Manifest.ReplaceShellJobName(rootApp.GetName())
	// rootApp.ReplaceStartParams(releaseName, rootApp.ManifestPackage, namespace)
	rootApp.InstallOption.IngressHost = ingressHost
	rootApp.InstallOption.IngressSeletorName = ingressSelectorName
	rootApp.InstallOption.IngressClassName = IngressClassName
	rootApp.InstallOption.Suffix = suffix
	rootR := installResult{
		Identifie:         rootApp.Manifest.Application.Identifie,
		Title:             rootApp.Manifest.Application.Name,
		DeploymentName:    rootApp.GetName(),
		JobName:           rootApp.GetBuildJobName(),
		RequireBuildImage: rootApp.RequireBuildImage(),
	}
	installResults = append(installResults, rootR)

	for _, app := range children {
		app.Manifest.ReplaceShellJobName(app.GetName())
		app.ReplaceCommon(ingressHost)
		app.ReplaceStartParams(suffix, rootApp.ManifestPackage, namespace)
		installResult := installResult{
			Identifie:         app.Manifest.Application.Identifie,
			Title:             app.Manifest.Application.Name,
			DeploymentName:    app.GetName(),
			JobName:           app.GetBuildJobName(),
			RequireBuildImage: app.RequireBuildImage(),
		}
		installResults = append(installResults, installResult)
	}
	// 二次遍历替换参数
	rootApp.ReplaceCommon(ingressHost)
	rootApp.ReplaceStartParams(suffix, rootApp.ManifestPackage, namespace)

	// resultJson, err := json.Marshal(installResults)
	// if err != nil {
	// 	slog.Warn("install result", "err", err)
	// }
	chartAnno := mPackage.GetChartAnnotations(releaseName)
	meta := PackageChartMetadata{
		Name:        mPackage.Manifest.Application.Identifie,
		Version:     mPackage.GetVersion(),
		Icon:        mPackage.GetIcon(),
		AppVersion:  mPackage.GetVersion(),
		ApiVersion:  "v2",
		Description: mPackage.Manifest.Application.Description,
		Home:        mPackage.ZpkUrl,
		Annotations: make(map[string]interface{}),
	}
	for k, v := range chartAnno {
		meta.Annotations[k] = v
	}
	return Package{
		Root:                 rootApp,
		Children:             children,
		ReleaseName:          releaseName,
		InstallId:            installId,
		Namespace:            namespace,
		PackageChartMetadata: meta,
		PackageType:          mPackage.Manifest.Application.Type,
	}
}

//keyBy

type PackageApp struct {
	Parent *PackageApp
	*ManifestPackage
	*InstallOption
	ThirdpartyCDToken     string
	AppGroupInstallResult *v1alpha1.DeployItem
}

func NewPackageApp(manifestPackage *ManifestPackage, installOption *InstallOption) *PackageApp {
	p := &PackageApp{ManifestPackage: manifestPackage, InstallOption: installOption}
	p.PutEnvs(installOption.EnvKv)
	return p
}

func (p *PackageApp) GetIdentifie() string {
	//_ convert -
	return strings.ToLower(strings.ReplaceAll(p.ManifestPackage.Manifest.Application.Identifie, "_", "-"))
}

// 后缀
func (p *PackageApp) GetSuffix() string {
	return p.Suffix
}

func (p *PackageApp) GetName() string {
	if p.Manifest.IsOnce() {
		return p.GetIdentifie()
	}
	return GetDeployName(p.GetIdentifie(), p.Suffix)
}

func (p *PackageApp) GetReleaseName() string {
	releaseName := strings.ToLower(strings.ReplaceAll(p.ReleaseName, "_", "-"))
	// if p.Parent != nil && p.IsHelm() {
	// 	releaseName = GetDeployName(p.Parent.GetReleaseName(), p.GetIdentifie()) //普通子应用判断Parent就错了
	// }
	return releaseName
}

// zpk controller 已经把once 安装一次处理过了
// w7.cc/group-name: w7-php
// 自应用是helm 的话 需要添加标识
// // 暂时不用
func (p *PackageApp) GetAppGroupName() string {
	if p.IsHelm() {
		if p.Parent != nil {
			return p.GetName()
		}
	}
	return p.GetReleaseName()
}

func (p *PackageApp) GetBuildJobName() string {
	//_ convert -
	return strings.ReplaceAll(p.GetName()+"-build", "_", "-") + "-" + p.InstallId
}

func (p *PackageApp) GetShellJobName(shellType string) string {
	//_ convert -
	return strings.ReplaceAll(p.GetName()+"-"+shellType, "_", "-") + "-" + helper.RandomString(5)
}

func (p *PackageApp) GetHelmInstallJobName(shellType string) string {
	//_ convert -
	return strings.ReplaceAll(p.GetName()+"-"+shellType, "_", "-") + "-" + p.InstallId
}

func (p *PackageApp) GetShellByType(shellType string) types.ShellInterface {

	return p.Manifest.GetShellByType(shellType)
}

func (p *PackageApp) GetNamespace() string {
	return p.Namespace
}

/*
*
helm.sh/chart: demo-0.1.0

	app.kubernetes.io/name: demo
	app.kubernetes.io/instance: demo3
	app.kubernetes.io/version: "1.16.0"
	app.kubernetes.io/managed-by: Helm
*/
func (p *PackageApp) GetLabels() map[string]string {
	result := map[string]string{
		"app.kubernetes.io/name":           p.GetIdentifie(),
		"app.kubernetes.io/instance":       p.ReleaseName,
		"app.kubernetes.io/managed-by":     "Helm",
		"app.kubernetes.io/managed-by-sub": "w7panel",
		"app.kubernetes.io/version":        p.GetVersion(),
		"app.kubernetes.io/created-by":     "zpk",
		"w7.cc/module-name":                p.GetIdentifie(),
		"w7.cc/identifie":                  p.GetIdentifie(),
		"w7.cc/name":                       p.GetName(),
		"w7.cc/release-name":               p.ReleaseName,
		"w7.cc/group-name":                 p.GetReleaseName(),
		"app":                              p.GetName(),
		"w7.cc/install-id":                 p.InstallId,
		"w7.cc/suffix":                     p.GetSuffix(),
		"w7.cc/manifest-version":           p.Manifest.V.String(),
	}
	if p.Parent != nil {
		result["w7.cc/parent"] = p.Parent.GetName()
		// result["parent"] = p.Parent.GetName()
		// result["parents"] = p.Parent.GetName()
	}
	shells := p.Manifest.Platform.Container.Shells
	if p.RequireBuildImage() || len(shells) > 0 {
		result["w7.cc/has-shell"] = "true"
	}
	if p.RequireBuildImage() {
		result["w7.cc/has-build"] = "true"
	}
	menuLabels := p.Manifest.MenuLabels()
	for k, v := range menuLabels {
		result[k] = v
	}
	if p.K8sToken != nil {
		saName, err := p.K8sToken.GetSaName()
		if err != nil {
			slog.Warn("get sa name", "err", err)
		}
		if saName != "" {
			result["w7.cc/create-username"] = saName
		}
		result["w7.cc/create-role"] = p.K8sToken.GetRole()
	}
	// if p.Manifest.Application.FrontType != nil {
	// 	for _, v := range p.Manifest.Application.FrontType {
	// 		if v == "thirdparty_cd" && p.Manifest.ShowMenuTop() {
	// 			result["w7.cc/menu-location"] = "top"
	// 		}
	// 	}

	// }
	return result

}

func (p *PackageApp) GetAnnotations() map[string]string {

	jsonParams, err := json.Marshal(p.Manifest.Platform.Container.StartParams)
	if err != nil {
		jsonParams = []byte{}
	}
	jsonPorts, err := json.Marshal(p.Manifest.Platform.Container.Ports)
	if err != nil {
		jsonPorts = []byte{}
	}
	bindingsJson := p.Manifest.GetBindsJson()

	frontTypejson, err := json.Marshal(p.Manifest.Application.FrontType)
	if err != nil {
		frontTypejson = []byte{}
	}
	domainsJson := []byte{}
	ingressDomainsJson := []byte{}
	scheme := "http://"
	if p.RequireDomainHttps() {
		scheme = "https://"
	}
	if p.IngressHost != "" {

		domainsJson, err = json.Marshal([]string{scheme + p.IngressHost})
		if err != nil {
			domainsJson = []byte{}
		}
		if p.IngressSeletorName != "" {
			// {"host":"abc.cc","auto_ssl":false,"ing_name":"ing-eovaklbc","config_name":"ollama"}
			routes := p.Manifest.GetRoutesByName(p.GetIngressSelectorName())
			ingName := ""
			if len(routes) > 0 {
				ingName = routes[0].Backend.IngName
			}
			arr := []map[string]interface{}{}
			arr = append(arr, map[string]interface{}{
				"host":        p.IngressHost,
				"auto_ssl":    !p.RequireDomainHttps(),
				"ing_name":    ingName,
				"config_name": p.GetIngressSelectorName(),
			}) //   w7.cc/menu-location: top
			ingressDomainsJson, _ = json.Marshal(arr)
		}
	}
	ingressConfigJson := []byte{}
	if p.Manifest.Platform.Ingress != nil {

		ingressConfigJson, err = json.Marshal(p.Manifest.Platform.Ingress)
		if err != nil {
			ingressConfigJson = []byte{}
		}
	}
	icon := p.GetIcon()
	if p.Parent != nil {
		icon = p.Parent.GetIcon()
	}
	shells := p.Manifest.Platform.Container.Shells
	if p.IsHelm() {
		shells = append(shells, Shell{
			Title:     "helm安装",
			Type:      "helm",
			Shell:     "helm install ",
			Image:     "",
			SearchJob: p.GetName() + "-helm",
		})
	}
	if p.RequireBuildImage() {
		shells = append(shells, Shell{
			Title:     "构建镜像",
			Type:      "build-image",
			Shell:     "/start.sh",
			Image:     p.GetBuildImage(),
			SearchJob: p.GetName() + "-build-image",
		})
	}
	shellsJson := []byte{}
	if len(shells) > 0 {
		shellsJson, _ = json.Marshal(shells)
	}

	result := map[string]string{
		"w7.cc/zpk-url":      p.ZpkUrl,
		"title":              p.Manifest.Application.Name,
		"w7.cc/title":        p.Manifest.Application.Name,
		"w7.cc/deploy-title": "创建应用",
		"w7.cc/description":  p.Manifest.Application.Description,
		"w7.cc/icon":         icon,
		"w7.cc/start-params": string(jsonParams),
		"w7.cc/install-id":   p.InstallId,
		"w7.cc/ports":        string(jsonPorts),
		"w7.cc/front-type":   string(frontTypejson),
		"w7.cc/bindings":     string(bindingsJson),
		"w7.cc/static-url":   p.GetStaticUrl(p.GetReleaseName()),
		"w7.cc/shells":       string(shellsJson),
		// "w7.cc/create-svc":   "true",
		// "w7.cc/domains":               string(domainsJson),
		"w7.cc/ingress-config": string(ingressConfigJson),
		// "w7.cc/ingress-domains":       string(ingressDomainsJson),
		"w7.cc/ingress-selector-name":    p.GetIngressSelectorName(),
		"w7.cc/identifie":                p.GetIdentifie(),
		"w7.cc/release-name":             p.GetReleaseName(),
		"w7.cc/group-name":               p.GetReleaseName(),
		"app":                            p.GetName(),
		"w7.cc/app":                      p.GetName(),
		"w7.cc/ticket":                   p.Ticket,
		"meta.helm.sh/release-name":      p.GetReleaseName(),
		"meta.helm.sh/release-namespace": p.GetNamespace(),
	}
	// if !p.IsUpgrade() {
	result["w7.cc/domains"] = string(domainsJson)
	result["w7.cc/ingress-domains"] = string(ingressDomainsJson)
	if p.IngressHost != "" {
		result["w7.cc/default-domain"] = scheme + p.IngressHost
	}
	if p.RequireBuildImage() {
		result["w7.cc/has-build"] = "true"
	}

	for k, v := range p.InstallOption.Annotations {
		result[k] = v
	}
	if (p.Manifest.Application.Annotation != nil) && (len(p.Manifest.Application.Annotation) > 0) {
		for k, v := range p.Manifest.Application.Annotation {
			result[k] = v
		}
	}

	return result
}

func (p *PackageApp) GetMatchLabels() map[string]string {
	return map[string]string{
		"app": p.GetName(),
		// "app.kubernetes.io/instance": p.ReleaseName,
		// "app.kubernetes.io/name":     p.GetIdentifie(),
		// app.kubernetes.io/name: k8s-offline"
		// "install-id": p.InstallId,
	}
}

func (p *PackageApp) GetImage() string {
	if p.IsHelm() {
		return strings.TrimSpace(helper.SelfImage())
	}
	if p.RequireBuildImage() {
		return p.GetBuildImage()
	}
	return strings.TrimSpace(p.Manifest.Platform.Container.Image)
}

func (p *PackageApp) RequireBuildImage() bool {
	return p.Manifest.Platform.Container.Image == ""
}

func (p *PackageApp) GetReplicas() int32 {
	return p.Replicas
}

func (p *PackageApp) GetCommand() []string {
	cmds := p.Manifest.Platform.Container.Cmd
	if len(cmds) == 0 {
		return []string{}
	}
	trimCmds := []string{}
	for _, cmd := range cmds {
		trimCmds = append(trimCmds, strings.TrimSpace(cmd))
	}
	return trimCmds
}

func (p *PackageApp) GetContainerPort() []corev1.ContainerPort {
	ports := make([]corev1.ContainerPort, 0)
	if len(p.Manifest.Platform.Container.Ports) == 0 {
		ports = append(ports, corev1.ContainerPort{
			Name:          "port-" + fmt.Sprint(p.Manifest.Platform.Container.ContainerPort),
			ContainerPort: (int32(p.Manifest.Platform.Container.ContainerPort)),
			Protocol:      corev1.ProtocolTCP,
		})
		return ports
	}
	for _, port := range p.Manifest.Platform.Container.Ports {
		ports = append(ports, corev1.ContainerPort{
			Name:          "port-" + fmt.Sprint(port.Port),
			ContainerPort: int32(port.Port),
			Protocol:      corev1.Protocol(port.Protocol),
		})
	}
	return ports
}

func (p *PackageApp) GetCpu() string {
	if p.Cpu != "" {
		return p.Cpu
	}
	return ""
	// return string(rune(p.Manifest.Platform.Container.CPU)) + "m"
}

func (p *PackageApp) GetMemory() string {
	if p.Memory != "" {
		return p.Memory
	}
	// Bug 修复：将 int 类型转换为字符串类型
	return "" //fmt.Sprint(p.Manifest.Platform.Container.Mem) + "Mi"
}

func (p *PackageApp) GetRuntimeClassName() string {
	return p.Manifest.Platform.Container.RuntimeClassName
}

func (p *PackageApp) GetPodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser:    &p.Manifest.Platform.Container.SecurityContext.RunAsUser,
		RunAsGroup:   &p.Manifest.Platform.Container.SecurityContext.RunAsGroup,
		RunAsNonRoot: &p.Manifest.Platform.Container.SecurityContext.RunAsNonRoot,
		FSGroup:      &p.Manifest.Platform.Container.SecurityContext.FsGroup,
	}
}

func (p *PackageApp) GetContainerSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		RunAsUser:    &p.Manifest.Platform.Container.SecurityContext.RunAsUser,
		RunAsGroup:   &p.Manifest.Platform.Container.SecurityContext.RunAsGroup,
		RunAsNonRoot: &p.Manifest.Platform.Container.SecurityContext.RunAsNonRoot,
		// Privileged:   &p.Manifest.Platform.Container.Privileged.Bool(),
		// FSGroup:      &p.Manifest.Platform.Container.SecurityContext.FsGroup,
	}
}

/*
*
volumes:

	-
	    hostPathType: DirectoryOrCreate
	    mountPath: /home/txt
	    subPath: /home
	    type: hostStorage
	-
	    mountPath: /tmp
	    subPath: /tmp
	    type: emptyStorage
	-
	    mountPath: /home/txt
	    subPath: '%RANDOM_DIR%'
	    type: diskStorage
*/
func (p *PackageApp) GetVolumes() []corev1.Volume {
	if p.InstallOption.Volumes != nil && len(p.InstallOption.Volumes) > 0 {
		for i, volume := range p.InstallOption.Volumes {
			vname := volume.Name
			vname = strings.ReplaceAll(vname, "%PVCNAME%", p.InstallOption.PvcName)
			volume.Name = vname
			if volume.PersistentVolumeClaim != nil {
				volume.PersistentVolumeClaim.ClaimName = vname
			}
			p.InstallOption.Volumes[i] = volume
		}
		return p.InstallOption.Volumes

	}
	return p.ManifestPackage.GetVolumes(p.InstallOption.PvcName)
}

// Bug 修复：将所有的 lower 改为 strings.ToLower
func (p *PackageApp) GetVolumeMounts() []corev1.VolumeMount {
	if p.InstallOption.VolumesMounts != nil && len(p.InstallOption.VolumesMounts) > 0 {
		for i, volumeMount := range p.InstallOption.VolumesMounts {
			vname := volumeMount.Name
			vname = strings.ReplaceAll(vname, "%PVCNAME%", p.InstallOption.PvcName)
			volumeMount.Name = vname
			p.InstallOption.VolumesMounts[i] = volumeMount
		}
		return p.InstallOption.VolumesMounts
	}
	return p.ManifestPackage.GetVolumeMounts(p.InstallOption.PvcName, p.ReleaseName, p.PreSubPath)
}

func (p *PackageApp) GetServicePort() []corev1.ServicePort {
	ports := make([]corev1.ServicePort, 0)
	for _, port := range p.Manifest.Platform.Container.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:     "name-" + strconv.Itoa(port.Port),
			Port:     int32(port.Port),
			Protocol: corev1.Protocol(port.Protocol),
		})
	}
	if len(ports) == 0 {
		cport := p.Manifest.Platform.Container.ContainerPort
		ports = append(ports, corev1.ServicePort{
			Name:     "name-" + strconv.Itoa(cport),
			Port:     int32(cport),
			Protocol: corev1.ProtocolTCP,
		})
	}
	return ports
}

func (p *PackageApp) GetServiceLbPort() []corev1.ServicePort {
	ports := make([]corev1.ServicePort, 0)
	manifestPorts := p.Manifest.Platform.Container.Ports
	if len(manifestPorts) == 0 {
		manifestPorts = append(manifestPorts, Ports{
			LbPort:   0,
			Port:     p.Manifest.Platform.Container.ContainerPort,
			Protocol: "TCP",
			Name:     "default",
		})
	}
	for _, port := range manifestPorts {

		var value int
		switch port.LbPort.(type) {
		case string:
			strVal := port.LbPort.(string)
			if strVal == "" || strVal == "0" {
				strVal = "0"
			}
			value, _ = (strconv.Atoi(strVal))
		case int:
			value = port.LbPort.(int)
		case float64:
			value = int(port.LbPort.(float64))
		}

		if value == 0 {
			continue
		}

		ports = append(ports, corev1.ServicePort{
			Name: "name-" + strconv.Itoa(port.Port),
			//string to int32
			// Port:     strconv.Atoi(port.LbPort),
			Port:     int32(value),
			Protocol: corev1.Protocol(port.Protocol),
		})
	}
	return ports
}

func (p *PackageApp) GetEnv() []corev1.EnvVar {
	envs := make([]corev1.EnvVar, 0)
	for _, env := range p.Manifest.Platform.Container.Env {
		envs = append(envs, corev1.EnvVar{
			Name:      env.Name,
			Value:     env.Value,
			ValueFrom: env.ValueFrom,
		})
	}

	for _, param := range p.Manifest.Platform.Container.StartParams {
		val := corev1.EnvVar{
			Name:  param.Name,
			Value: param.ValuesText,
		}

		if param.ModuleName == "k8s_field" {
			val.ValueFrom = &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  param.ValuesText, // Assuming you have a FieldPath field, instead of ValuesText
				},
			}
		}
		envs = append(envs, val)
	}
	// envs = append(envs, corev1.EnvVar{
	// 	Name:  "HELM_RELEASE_NAME",
	// 	Value: p.InstallOption.ReleaseName,
	// })
	envs = append(envs, corev1.EnvVar{
		Name:  "RELEASE_NAME_SUFFIX",
		Value: p.InstallOption.Suffix,
	})
	envs = append(envs, corev1.EnvVar{
		Name:  "METADATA_NAME",
		Value: p.GetName(),
	})
	return envs
}

func (p *PackageApp) GetFirstPort() int32 {
	if len(p.Manifest.Platform.Container.Ports) == 0 {
		return int32(p.Manifest.Platform.Container.ContainerPort)
	}
	return int32(p.Manifest.Platform.Container.Ports[0].Port)
}

// 获取构建镜像地址
func (p *PackageApp) GetBuildImage() string {
	return p.DockerRegistry.GetDockerImageWithNameAndTag(p.Identifie, p.GetVersion())
}

func (p *PackageApp) GetPushImage() string {
	return p.GetBuildImage()
}

func (p *PackageApp) RequireDomainHttps() bool {
	return p.Manifest.RequireDomainHttps() || p.IngressForceHttps
}

func (p *PackageApp) RequireDomain() bool {
	return p.Manifest.RequireDomain()
}

func (p *PackageApp) GetIngressSvcName(backendName string) string {
	if backendName == "current" || backendName == "" {
		backendName = strings.ToLower(p.Identifie)
	}
	backendName = strings.ReplaceAll(backendName, "_", "-")
	// return backendName + "-" + p.ReleaseName
	if p.Manifest.IsOnce() && backendName == p.GetIdentifie() {
		return p.GetIdentifie()
	}
	return GetDeployName(backendName, p.Suffix)
}

func (p *PackageApp) GetIngressHost() string {
	return p.IngressHost
}

func (p *PackageApp) GetIngressSelectorName() string {
	return p.IngressSeletorName
}

func (p *PackageApp) GetZipUrl() string {
	return p.ZipUrl
}

func (p *PackageApp) GetZpkUrl() string {
	return p.ZpkUrl
}
func (p *PackageApp) GetBuildContext() string {
	return p.Manifest.GetBuildContext()
}

func (p *PackageApp) GetRoutesByName(name string) []helm.ManifestRouteInterface {
	routes := p.Manifest.GetRoutesByName(name)
	var result []helm.ManifestRouteInterface
	for _, route := range routes {
		result = append(result, &route)
	}
	return result
}

func (p *PackageApp) GetDockerRegisty() types.DockerRegistry {
	return p.DockerRegistry
}

func (p *PackageApp) GetDockerfilePath() string {
	return p.Manifest.GetDockerfilePath()
}

// 获取镜像拉取secret
func (p *PackageApp) GetImagePullSecrets() []corev1.LocalObjectReference {

	if p.DockerRegistrySecretName == "" {
		return []corev1.LocalObjectReference{}
	}
	return []corev1.LocalObjectReference{corev1.LocalObjectReference{Name: p.DockerRegistrySecretName}}
}

func (p *PackageApp) GetIngressClassName() string {

	return p.IngressClassName
}

func (p *PackageApp) IsUpgrade() bool {
	return p.InstallOption.IsUpgrade
}

func (p *PackageApp) GetServiceAccountName() string {
	sa := p.InstallOption.ServiceAccountName
	if sa == "" {
		sa = "default"
	}
	return sa
}
func (p *PackageApp) IsPrivileged() bool {
	return p.Manifest.Platform.Container.IsPrivileged()
}

func (p *PackageApp) GetNotifyCompletionUrl() string {
	return p.BuildImageSuccessUrl
}

func (p *PackageApp) GetNotifyFailedUrl() string {
	return "/"
}

func (p *PackageApp) GetHostNetwork() bool {
	return p.GetDockerRegisty().Host == "registry.local.w7.cc"
}
func (p *PackageApp) GetHostAliases() []corev1.HostAlias {
	if !p.GetHostNetwork() {
		return []corev1.HostAlias{}
	}
	return []corev1.HostAlias{}
	// nodeIp := facade.Config.GetString("app.node_ip")
	// return []corev1.HostAlias{
	// 	{IP: nodeIp, Hostnames: []string{"registry.local.w7.cc"}},
	// }
}

func (p *PackageApp) RequireSite() bool {
	return p.Manifest.RequireSite()
}

func (p *PackageApp) GetThirdpartyCDToken() string {
	return p.ThirdpartyCDToken
}

func (p *PackageApp) RequireCreateDb() (bool, string) {
	return p.Manifest.RequireCreateDb()
}

func (p *PackageApp) RequireCreateDbUser() (bool, string, string) {
	return p.Manifest.RequireCreateDbUser()
}

func (p *PackageApp) GetLogo() string {
	return p.GetIcon()
}
func (p *PackageApp) GetVersion() string {
	return p.ManifestPackage.Version.Name
}

func (p *PackageApp) GetHelmConfig() types.HelmConfigInterface {
	return &p.Manifest.Platform.Helm
}

func (p *PackageApp) IsHelm() bool {
	return p.Manifest.Application.Type == "helm"
}

func (p *PackageApp) GetChartAnnotations() map[string]string {
	return p.ManifestPackage.GetChartAnnotations(p.GetReleaseName())
}
func (p *PackageApp) GetInstallConfig() map[string]string {
	result := make(map[string]string)
	for _, v := range p.EnvKv {
		result[v.Name] = v.Value
	}
	return result
}

func (p *PackageApp) GetMicroAppProps() map[string]string {
	result := make(map[string]string)
	for _, v := range p.EnvKv {
		result[v.Name] = v.Value
	}
	///k8s/v1/namespaces/default/services/w7-sitemanager-eeyidnlv-site-manager:8000/proxy-no
	// /ui/microapp/w7-sitemanager-eeyidnlv/index.html
	result["backendUrl"] = p.GetBackendUrl()
	result["frontendUrl"] = p.GetFrontendUrl()
	result["releaseName"] = p.GetReleaseName()
	result["group"] = p.GetReleaseName()
	return result
}

func (p *PackageApp) GetBackendUrl() string {
	port := p.GetFirstPort()
	if port == 80 || port == 0 {
		return "/k8s/v1/namespaces/default/services/" + p.GetName() + "/proxy-no"
	}
	return "/k8s/v1/namespaces/default/services/" + p.GetName() + ":" + strconv.Itoa(int(port)) + "/proxy-no"
}

func (p *PackageApp) GetFrontendUrl() string {
	return "/ui/microapp/" + p.GetName() + "/index.html"
}

func (p *PackageApp) SupportMicroApp() bool {
	return p.Manifest.SupportMicroApp()
}
