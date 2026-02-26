package types

import (
	"net/url"
	"strconv"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	zpktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
)

type DeployItem struct {
	DeployId        int    `json:"deployId"`
	DeployItemId    int    `json:"deployItemId"`
	SiteKey         int    `json:"siteKey"`
	SiteUrl         string `json:"siteUrl"`
	SiteName        string `json:"siteName"`
	DeployUrl       string `json:"deployUrl"`
	CanDeploy       bool   `json:"canDeploy"`
	ReleaseName     string `json:"releaseName"`
	Enabled         bool   `json:"enabled"`
	IsServiceExpire bool   `json:"isServiceExpire"`
	IsDemoExpire    bool   `json:"isDemoExpire"`
}
type ManifestPackage struct {
	Manifest                 Manifest                    `json:"manifest"`
	ZpkUrl                   string                      `json:"zpkUrl"`
	HelmUrl                  string                      `json:"helmUrl"`
	ZipUrl                   string                      `json:"zipUrl"`
	OciUrl                   string                      `json:"ociUrl"`
	WebZipUrl                map[string]string           `json:"webZipUrl"`
	Version                  Version                     `json:"version"`
	Children                 map[string]*ManifestPackage `json:"children"`
	RequireInstall           bool                        `json:"requireInstall"` //是否一定需要安装
	ConsoleReleaseName       string                      `json:"consoleReleaseName"`
	Parent                   *ManifestPackage            `json:"parent"`                   //正常情况下，父节点是空的。只有在父子关系中才会有值
	RequireParentReleaseName bool                        `json:"requireParentReleaseName"` //是否需要父节点发布名
	DeployItems              []DeployItem                `json:"deployItems"`
	IconUrl                  string                      `json:"iconUrl"`
	Ticket                   string                      `json:"ticket"`
}

func (p *ManifestPackage) GetChartAnnotations(releaseName string) map[string]string {
	helmConfig := p.Manifest.Platform.Helm
	anno := map[string]string{
		zpktypes.HELM_RELEASE_SOURCE:   "zpk",
		zpktypes.HELM_REPOSITORY_URL:   helmConfig.Repository,
		zpktypes.HELM_CHART_NAME:       helmConfig.ChartName,
		zpktypes.HELM_CHART_VERSION:    helmConfig.Version,
		zpktypes.HELM_ZPK_VERSION:      p.GetVersion(),
		zpktypes.HELM_ZPK_URL:          p.ZpkUrl,
		zpktypes.HELM_LOGO:             p.Manifest.Application.Icon,
		zpktypes.HELM_INDENTIFIE:       strings.Replace(p.Manifest.Application.Identifie, "_", "-", -1),
		zpktypes.HELM_APPLICATION_TYPE: p.Manifest.Application.Type,
		zpktypes.HELM_TITLE:            p.Manifest.Application.Name,
		zpktypes.HELM_STATIC_URL:       p.GetStaticUrl(releaseName),
	}
	return anno
}

func (p *ManifestPackage) GetIcon() string {
	if p.Manifest.Application.Icon != "" {
		return p.Manifest.Application.Icon
	}
	if p.IconUrl != "" {
		return p.IconUrl
	}
	zpkUri, err := url.Parse(p.ZpkUrl)
	if err != nil {
		return ""
	}
	zpkUri.Path = "/zip/icon/" + p.Manifest.Application.Identifie
	return zpkUri.String()
}

func (p *ManifestPackage) GetTitle() string {

	if p.Manifest.Application.Name != "" {
		return p.Manifest.Application.Name
	}
	return p.GetName()
}

func (p *ManifestPackage) GetRootTitle() string {
	if p.Manifest.Platform.Container.BaseInfo.Name != "" {
		return p.Manifest.Platform.Container.BaseInfo.Name
	}
	return p.GetTitle()
}

func (p *ManifestPackage) GetDescription() string {
	if p.Manifest.Application.Description != "" {
		return p.Manifest.Application.Description
	}
	return p.GetName()
}

func (p *ManifestPackage) GetRootDescription() string {
	if p.Manifest.Platform.Container.BaseInfo.Description != "" {
		return p.Manifest.Platform.Container.BaseInfo.Description
	}
	return p.GetDescription()
}

func (p *ManifestPackage) GetStaticUrl(releaseName string) string {
	if (p.Manifest.WebApp.Url != "") && strings.HasPrefix(p.Manifest.WebApp.Url, "http") {
		return p.Manifest.WebApp.Url
	}

	return "/ui/microapp/" + releaseName + "/index.html"
	// zpkUri, err := url.Parse(p.ZpkUrl)
	// if err != nil {
	// 	return ""
	// }
	// zpkUri.Path = "/static/" + p.Manifest.Application.Identifie + "/index.html"
	// return zpkUri.String()
}

func (p *ManifestPackage) GetVersion() string {
	return p.Version.Name
}

func (p *ManifestPackage) GetName() string {
	return p.Manifest.Application.Name
}

type ParentAddConfig struct {
	ParentTitle     string `json:"parentTitle"`
	ParentIdentifie string `json:"parentIdentifie"`
}

type PackageAddConfig struct {
	RequireBuild             bool                 `json:"requireBuild"`
	Namespace                string               `json:"namespace"`
	RequirePvc               bool                 `json:"requirePvc"`
	RequireDomain            bool                 `json:"requireDomain"`
	RequireDomainHttps       bool                 `json:"requireDomainHttps"`
	RequireDomainForce       bool                 `json:"requireDomainForce"`
	StartParams              []StartParams        `json:"startParams"`
	Identifie                string               `json:"identifie"`
	Name                     string               `json:"name"`
	Ingress                  []Ingress            `json:"ingress"`
	Icon                     string               `json:"icon"`
	Version                  string               `json:"version"`
	RequireInstall           bool                 `json:"requireInstall"`
	ReleaseName              string               `json:"releaseName"`
	OutModuleNames           []string             `json:"outModuleNames"`
	DependsOn                []Depends            `json:"dependsOnes"`
	RequireParentReleaseName bool                 `json:"requireParentReleaseName"`
	ParentTitle              string               `json:"parentTitle"`
	ParentIdentifie          string               `json:"parentIdentifie"`
	DeployName               string               `json:"deployName"` //部署名称，非必须字段
	IsHelm                   bool                 `json:"isHelm"`     //是否helm应用 非必须字段
	DeployItems              []DeployItem         `json:"deployItems"`
	IsConsole                bool                 `json:"isConsole,omitempty"`
	ZipURL                   string               `json:"zipURL"` //zip包地址 非必须字段 创通应用 重新部署需要
	RequireLimit             bool                 `json:"requireLimit"`
	VolumeMounts             []corev1.VolumeMount `json:"volumesMounts"`
	Volumes                  []corev1.Volume      `json:"volumes"`
	IsUpgrade                bool                 `json:"isUpgrade"`
}

func (p *ManifestPackage) ToPackageAddConfig(releaseName string, requireLimit bool) PackageAddConfig {

	if releaseName == "" {
		releaseName = p.ConsoleReleaseName
	}
	//20260130 + 更新时候不知道为啥releaseName为空，所以这里加一个判断
	if p.ConsoleReleaseName != "" {
		releaseName = p.ConsoleReleaseName
	}

	result := PackageAddConfig{
		Identifie:                p.Manifest.Application.Identifie,
		Name:                     p.Manifest.Application.Name,
		RequireBuild:             p.Manifest.RequireBuild(),
		RequirePvc:               p.Manifest.requrirePvc(),
		RequireDomain:            p.Manifest.RequireDomain(),
		RequireDomainHttps:       p.Manifest.RequireDomainHttps(),
		RequireDomainForce:       p.Manifest.RequireDomainForce(),
		StartParams:              p.Manifest.Platform.Container.StartParams,
		Ingress:                  p.Manifest.Platform.Ingress,
		Icon:                     p.GetIcon(),
		Version:                  p.GetVersion(),
		RequireInstall:           p.RequireInstall,
		ReleaseName:              releaseName,           //p.ConsoleReleaseName(20260130为啥用ConsoleReleaseName),
		OutModuleNames:           p.GetOutModuleNames(), //外部模块名称，不包括自己和子应用 非必须字段
		DependsOn:                p.Manifest.GetOutDepends(),
		RequireParentReleaseName: p.RequireParentReleaseName,
		IsHelm:                   p.Manifest.IsHelm(),
		DeployItems:              p.DeployItems,
		IsConsole:                (p.ConsoleReleaseName != ""),
		ZipURL:                   p.ZipUrl,
		RequireLimit:             requireLimit,
		VolumeMounts:             p.GetVolumeMounts("%PVCNAME%", releaseName, nil),
		Volumes:                  p.GetVolumes("%PVCNAME%"),
		// IsUpgrade: p,

		// DependsOns:         p.Manifest.DependsOnes,
	}
	if p.RequireParentReleaseName && p.Parent != nil {
		result.ParentIdentifie = p.Parent.Manifest.Application.Identifie
		result.ParentTitle = p.Parent.Manifest.Application.Name
	}
	if releaseName != "" {
		if p.Manifest.IsHelm() {
			result.DeployName = releaseName
		} else {
			//根据DeployName 获取上次的安装参数
			suffix := getSuffix(releaseName)
			result.DeployName = GetDeployName(p.Manifest.Application.Identifie, suffix)
		}
	}
	if p.Manifest.IsHelm() {
		result.RequireLimit = false
	}

	return result
}

/*
*

	外部模块名称，不包括自己和子应用
*/
func (p *ManifestPackage) GetOutModuleNames() []string {
	result := p.Manifest.GetStartParamsModuleNames()
	ignoreModule := []string{p.Manifest.Application.Identifie}
	for _, child := range p.Children {
		result = append(result, child.Manifest.GetStartParamsModuleNames()...)
		ignoreModule = append(ignoreModule, child.Manifest.Application.Identifie)
	}
	result = lo.Uniq(result)
	return lo.Filter(result, func(item string, index int) bool {
		return !lo.Contains(ignoreModule, item)
	})
}

func (p *ManifestPackage) GetChildren(moduleName string) (child *ManifestPackage) {
	if moduleName == p.Manifest.Application.Identifie {
		return p
	}
	child, ok := p.Children[(moduleName)]
	if !ok {
		return nil
	}
	return child
}

func (p *ManifestPackage) PutKv(kv EnvKv) {
	for i, param := range p.Manifest.Platform.Container.StartParams {
		if param.Name == kv.Name {
			p.Manifest.Platform.Container.StartParams[i].ValuesText = kv.Value
		}
	}
	// print(p.Manifest.Platform.Container.StartParams)
}

func (p *ManifestPackage) PutEnvs(kv []EnvKv) {
	for _, param := range kv {
		p.PutKv(param)
	}
}

func (p *ManifestPackage) GetKey(key string) StartParams {
	params := p.Manifest.Platform.Container.StartParams
	for _, param := range params {
		if param.Name == key {
			return param
		}
	}
	var emptyParam StartParams
	return emptyParam
}

func (p *ManifestPackage) GetSvcName(suffix, namespace string) string {
	return helper.ClusterDomain(strings.ReplaceAll(p.Manifest.Application.Identifie+"-"+suffix, "_", "-"), namespace)
	// return strings.ReplaceAll(p.Manifest.Application.Identifie+"-"+releaseName, "_", "-") + "." + namespace + ".svc.cluster.local"
}

func (p *ManifestPackage) GetFirstPort() string {
	port := p.Manifest.GetFirstPort()
	if port == 0 {
		return "80"
	}
	return strconv.Itoa(int(port))
}

func (p *ManifestPackage) ReplaceStartParams(suffix string, root *ManifestPackage, namespace string) string {
	params := p.Manifest.Platform.Container.StartParams
	for i, param := range params {
		if param.ModuleName != "" {
			child, ok := root.Children[param.ModuleName]
			if ok {
				replaceVal := strings.ReplaceAll(param.ValuesText, "%HOST%", child.GetSvcName(suffix, namespace))
				replaceVal = strings.ReplaceAll(replaceVal, "%PORT%", child.GetFirstPort())
				replaceVal = p.ReplaceVal(replaceVal, child.Manifest.Platform.Container.StartParams)
				p.Manifest.Platform.Container.StartParams[i].ValuesText = replaceVal
			}
		}
	}
	return ""
}
func (p *ManifestPackage) ReplaceDefault() {
	params := p.Manifest.Platform.Container.StartParams
	for i, param := range params {
		valuesTxt := param.ValuesText
		valuesTxt = strings.ReplaceAll(valuesTxt, "%RANDOM%", helper.RandomString(10))
		valuesTxt = strings.ReplaceAll(valuesTxt, "%LARAVEL_APP_KEY%", helper.LaravelAppKey(32))
		p.Manifest.Platform.Container.StartParams[i].ValuesText = valuesTxt
	}
	envs := p.Manifest.Platform.Container.Env
	for k, env := range envs {
		valuesTxt := env.Value
		valuesTxt = strings.ReplaceAll(valuesTxt, "%RANDOM%", helper.RandomString(10))
		valuesTxt = strings.ReplaceAll(valuesTxt, "%LARAVEL_APP_KEY%", helper.LaravelAppKey(32))
		p.Manifest.Platform.Container.Env[k].Value = valuesTxt
	}
}
func (p *ManifestPackage) ReplaceCommon(host string) {
	params := p.Manifest.Platform.Container.StartParams
	for i, param := range params {
		valuesTxt := param.ValuesText
		if p.Manifest.IsHelm() {
			valuesTxt = strings.ReplaceAll(valuesTxt, "%DOMAIN_URL%", host)
			valuesTxt = strings.ReplaceAll(valuesTxt, "%DOMAIN_SSL_URL%", host)
		} else {
			valuesTxt = strings.ReplaceAll(valuesTxt, "%DOMAIN_URL%", "http://"+host)
			valuesTxt = strings.ReplaceAll(valuesTxt, "%DOMAIN_SSL_URL%", "https://"+host)
		}
		valuesTxt = strings.ReplaceAll(valuesTxt, "%DOMAIN_HOST%", host)
		valuesTxt = strings.ReplaceAll(valuesTxt, "%RANDOM%", helper.RandomString(10))
		valuesTxt = strings.ReplaceAll(valuesTxt, "%LARAVEL_APP_KEY%", helper.LaravelAppKey(32))
		p.Manifest.Platform.Container.StartParams[i].ValuesText = valuesTxt
	}
	envs := p.Manifest.Platform.Container.Env
	for k, env := range envs {
		valuesTxt := env.Value
		valuesTxt = strings.ReplaceAll(valuesTxt, "%DOMAIN_URL%", "http://"+host)
		valuesTxt = strings.ReplaceAll(valuesTxt, "%DOMAIN_SSL_URL%", "https://"+host)
		valuesTxt = strings.ReplaceAll(valuesTxt, "%DOMAIN_HOST%", host)
		valuesTxt = strings.ReplaceAll(valuesTxt, "%RANDOM%", helper.RandomString(10))
		valuesTxt = strings.ReplaceAll(valuesTxt, "%LARAVEL_APP_KEY%", helper.LaravelAppKey(32))
		p.Manifest.Platform.Container.Env[k].Value = valuesTxt
	}
}

func (p *ManifestPackage) ReplaceVal(fval string, startParams []StartParams) string {
	newval := fval
	for _, param := range startParams {
		val := param.ValuesText
		key := param.Name
		newval = strings.ReplaceAll(newval, "%"+key+"%", val)
	}
	return newval
}

func (p *ManifestPackage) GetVolumes(pvcName string) []corev1.Volume {
	volumes := make([]corev1.Volume, 0)
	hasEmptyStorage := false
	hasHostStorage := false
	hastDiskStorage := false
	for _, volume := range p.Manifest.Platform.Container.Volumes {
		if volume.Type == "emptyStorage" && !hasEmptyStorage {
			volumes = append(volumes, corev1.Volume{
				Name: volume.Name,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
			hasEmptyStorage = true
		}
		if volume.Type == "hostStorage" && !hasHostStorage {
			volumes = append(volumes, corev1.Volume{
				Name: volume.Name,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: volume.HostPath.Path,
						Type: (*corev1.HostPathType)(&volume.HostPath.Type),
					},
				},
			})
			// hasHostStorage = true
		}
		if volume.Type == "diskStorage" && pvcName != "" && !hastDiskStorage {
			volumes = append(volumes, corev1.Volume{
				Name: pvcName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvcName,
					},
				},
			})
			hastDiskStorage = true
		}
	}
	return volumes
}

// Bug 修复：将所有的 lower 改为 strings.ToLower
func (p *ManifestPackage) GetVolumeMounts(pvcName string, releaseName string, preSubPath map[string]string) []corev1.VolumeMount {
	volumeMounts := make([]corev1.VolumeMount, 0)
	for _, volume := range p.Manifest.Platform.Container.Volumes {
		defaultSubPath := helper.StringToMD5(volume.MountPath + p.GetName())
		if preSubPath != nil {
			prePath, ok := preSubPath[volume.MountPath]
			if ok {
				volume.SubPath = prePath
			}
		}
		if volume.SubPath == "" {
			// volume.SubPath = defaultSubPath

		}
		volume.SubPath = strings.ReplaceAll(volume.SubPath, "%RANDOM_DIR%", defaultSubPath)
		volume.SubPath = strings.ReplaceAll(volume.SubPath, "%RELEASE_NAME%", releaseName)
		volumeName := volume.Name
		// Bug 修复：pvc volumes 必须使用唯一个volumeName
		if volume.Type == "diskStorage" && pvcName != "" {
			volumeName = pvcName
		}
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: volume.MountPath,
			SubPath:   volume.SubPath,
		})

	}
	return volumeMounts
}
