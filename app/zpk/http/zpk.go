package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic"
	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	zpkk8s "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	zpkk8stypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

type Zpk struct {
	controller.Abstract
}

func (self Zpk) GetConfig(http *gin.Context) {
	type ParamsValidate struct {
		RepoUrl           string `form:"repoUrl" binding:"required"`
		ThirdpartyCDToken string `form:"thirdpartyCDToken"` // 域名选择业务名称
		ReleaseName       string `form:"releaseName"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	repoUrl := params.RepoUrl

	if repoUrl == "" {
		self.JsonResponseWithServerError(http, errors.New("repo url is empty"))
		return
	}

	if params.ThirdpartyCDToken == "" {
		self.JsonResponseWithServerError(http, errors.New("thirdpartyCDToken is required, please login first or refresh the page"))
		return
	}

	var config []types.PackageAddConfig

	// slog.Info("repoUrl", repoUrl)
	// slog.Info("ThirdpartyCDToken", params.ThirdpartyCDToken)
	repo := logic.NewRepo(repoUrl, params.ThirdpartyCDToken, "")
	// repo.SetCheckUpgrade(true)
	token := http.MustGet("k8s_token").(string)
	repo.SetPanelToken(token)
	k8sToken := k8s.NewK8sToken(token)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	params.ReleaseName = strings.ToLower(params.ReleaseName)
	appgroupObj, err := appgroup.GetAppgroupUseSdk(params.ReleaseName, client.GetNamespace(), client)
	// helmApi := k8s.NewHelm(client)
	// _, err = helmApi.Info(params.ReleaseName, client.GetNamespace())
	var dependsEnv *logic.DependEnv
	upgrade := false
	if err == nil {
		repo.SetUpgrade(true)
		repo.SetCurVersion(appgroupObj.Spec.Version)
		upgrade = true
		dependsEnv = logic.NewDependEnv(client)
	}
	mPackage, err := repo.Load()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	mPackage.ReplaceDefault()
	rootConfig := mPackage.ToPackageAddConfig(params.ReleaseName, k8sToken.IsShared())
	rootConfig.IsUpgrade = upgrade
	if dependsEnv != nil {
		rootConfig = dependsEnv.ReplacePackageAddConfig(mPackage, rootConfig)
	}
	config = append(config, rootConfig)
	for _, p := range mPackage.Children {
		p.ReplaceDefault()
		childConfig := p.ToPackageAddConfig(params.ReleaseName, k8sToken.IsShared())
		childConfig.IsUpgrade = upgrade
		if dependsEnv != nil {
			childConfig = dependsEnv.ReplacePackageAddConfig(p, childConfig)
		}
		config = append(config, childConfig)
	}
	self.JsonResponseWithoutError(http, config)
}

// 更新的话 覆盖上一次安装的环境变量

func (self Zpk) List(http *gin.Context) {
	type ParamsValidate struct {
		Namespace string `form:"namespace"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	helmApi := k8s.NewHelm(client)
	releases, err := helmApi.List(params.Namespace, "source=zpk")
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponse(http, releases, nil, 200)
}

func (self Zpk) Install(http *gin.Context) {
	type ParamsValidate struct {
		Namespace          string                `json:"namespace" binding:"required"`
		RepoUrl            string                `json:"repoUrl" binding:"required"`
		ReleaseName        string                `json:"releaseName" binding:"required"`
		InstallOptions     []types.InstallOption `json:"installOptions" binding:"required"`
		IngressHost        string                `json:"ingressHost"`        // 域名
		IngressSeletorName string                `json:"ingressSeletorName"` // 域名选择业务名称
		IngressClassName   string                `json:"ingressClass"`       // 域名选择业务名称
		IngressForceHttps  bool                  `json:"ingressForceHttps"`  // forceHttps
		ThirdpartyCDToken  string                `json:"thirdpartyCDToken"`  // 域名选择业务名称
		ClusterId          string                `json:"clusterId"`          // 集群ID
		IsTrandition       bool                  `json:"isTrandition"`       // 是否传统应用
		ZipUrl             string                `json:"zipUrl"`             // 代码包地址

	}

	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	params.IngressHost = strings.ToLower(params.IngressHost)
	repoUrl := params.RepoUrl
	if repoUrl == "" {
		self.JsonResponseWithServerError(http, errors.New("repo url is empty"))
		return
	}
	token := http.MustGet("k8s_token").(string)
	k8sToken := k8s.NewK8sToken(token)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	params.ReleaseName = strings.ToLower(params.ReleaseName)
	// helmApi := k8s.NewHelm(client)
	appgroupObj, err := appgroup.GetAppgroupUseSdk(params.ReleaseName, client.GetNamespace(), client)

	repo := logic.NewRepo(repoUrl, params.ThirdpartyCDToken, "")
	if err == nil {
		repo.SetUpgrade(true)
		repo.SetCurVersion(appgroupObj.Spec.Version)
	}
	repo.SetPanelToken(token)
	// os.Setenv("KUBERNETES_SERVICE_HOST", "172.16.1.13")
	// os.Setenv("KUBERNETES_SERVICE_PORT", "6443")
	namespace := params.Namespace
	mPackage, err := repo.Load()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	if mPackage.Manifest.IsHelm() {
		helmNs := mPackage.Manifest.GetHelmNamespce()
		if helmNs != "" {
			namespace = helmNs
		}
	}

	_, errns := client.CreateNamespace(namespace)
	if errns != nil {
		slog.Warn("create namespace error", "err", err)
		// return
	}

	releaseName := strings.ToLower(params.ReleaseName)
	if mPackage.Manifest.IsOnce() {
		releaseName = strings.ToLower(strings.ReplaceAll(mPackage.Manifest.Application.Identifie, "_", "-"))
	}
	appSecret := &console.AppSecret{}
	if repo.IsConsole {
		// if (params.ClusterId == "") {
		// 	params.ClusterId = config.MainW7Config.ClusterId
		// }
		preinstall, err := repo.PreInstall(params.ClusterId)
		if err != nil {
			self.JsonResponseWithServerError(http, err)
			return
		}
		releaseName = preinstall.ReleaseName
		zipUrl := preinstall.ZipURL
		mPackage.ZipUrl = zipUrl

		// cdClient := console.NewConsoleCdClient(params.ThirdpartyCDToken)
		// appSecret, err = cdClient.CreateSite(params.IngressHost, releaseName)
		// if err != nil {
		// 	slog.Warn("create site error may not need secret", "err", err)
		// }
	}
	if params.IsTrandition {
		mPackage.ZipUrl = params.ZipUrl
		// entry := ""
		// if len(params.InstallOptions) > 0 {
		// 	for _, v := range params.InstallOptions[0].EnvKv {
		// 		if v.Name == "ENTRY" {
		// 			entry = v.Value
		// 		}
		// 	}
		// }
		// mPackage.Manifest.Platform.Container.Shells = append(mPackage.Manifest.Platform.Container.Shells, types.Shell{
		// 	Type:  "install",
		// 	Title: "部署",
		// 	Shell: "/home/createsite.sh %CODE_ZIP_URL% " + entry,
		// })
	}

	//随机k8s deployment name
	installId := helper.RandomString(5)
	// releaseName = params.ReleaseName

	packageApps := types.NewPackage(mPackage, params.InstallOptions, releaseName, installId, namespace,
		params.IngressHost, params.IngressSeletorName, params.IngressClassName)
	packageApps.ForceHttps(params.IngressForceHttps)
	// packageApps.Root.K3kMode = k8sToken.K3kMode()
	isVirtual := k8sToken.IsVirtual()
	if isVirtual {
		// 虚拟集群也使用traefik
		// packageApps.Root.IngressClassName = "traefik"
	}

	sa := client.GetServiceAccountName()
	packageApps.Root.ServiceAccountName = sa
	packageApps.Root.K8sToken = k8sToken
	for _, child := range packageApps.Children {
		child.ServiceAccountName = sa
		if isVirtual {
			child.IngressClassName = packageApps.Root.IngressClassName
		}
		child.K8sToken = k8sToken
		// child.K3kMode = k8sToken.K3kMode()

	}

	// 微擎有安装脚本需要预先获取appid和secret
	if appSecret != nil && appSecret.AppId != "" && appSecret.AppSecret != "" {
		appId := types.Env{}
		appId.Name = "APP_ID"
		appId.Value = appSecret.AppId
		secret := types.Env{}
		secret.Name = "APP_SECRET"
		secret.Value = appSecret.AppSecret
		packageApps.Root.Manifest.Platform.Container.Env = append(packageApps.Root.Manifest.Platform.Container.Env, appId, secret)
	}
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	completeUrl := "http://" + http.Request.Host + "/api/v1/zpk/build-image-success?namespace=" + params.Namespace + "&releaseName=" + releaseName +
		"&domainHost=" + params.IngressHost + "&deploymentName=" + packageApps.Root.GetName() + "&thirdpartyCDToken=" + params.ThirdpartyCDToken + "&api-token=" + token

	packageApps.Root.InstallOption.BuildImageSuccessUrl = completeUrl
	packageApps.Root.ThirdpartyCDToken = params.ThirdpartyCDToken

	// 父节点发布名不为空 且 需要父节点发布名 且 有父节点 则设置父节点
	// 20250321
	// PackageApp label annnoations 添加 ["w7.cc/parent"]
	if packageApps.Root.ManifestPackage.RequireParentReleaseName &&
		packageApps.Root.ManifestPackage.Parent != nil &&
		packageApps.Root.ParentReleaseName != "" {
		parent := types.NewPackageApp(packageApps.Root.ManifestPackage.Parent, &types.InstallOption{ReleaseName: packageApps.Root.ParentReleaseName})
		packageApps.Root.Parent = parent
	}
	install := logic.NewInstall(client, packageApps)

	err = install.InstallOrUpgrade(releaseName, namespace)
	if err != nil {
		// panic(err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	result := map[string]interface{}{
		"releaseName": releaseName,
		"installId":   installId,
		"namespace":   namespace,
	}
	// mPackage.GetOutModuleNames()
	self.JsonResponse(http, result, nil, 200)

}
func (self Zpk) BuildJobFail(http *gin.Context) {

}

func (self Zpk) BuildImageSuccess(http *gin.Context) {

	type ParamsValidate struct {
		Namespace         string `form:"namespace" binding:"required"`
		ReleaseName       string `form:"releaseName" binding:"required"`
		DomainHost        string `form:"domainHost" binding:"required"`
		DeploymentName    string `form:"deploymentName" binding:"required"` //deploymentName
		ThirdpartyCDToken string `form:"thirdpartyCDToken"`                 // 控制台token zpk不需要这个token
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	if true {
		return
	}
	if params.ThirdpartyCDToken == "" {
		self.JsonSuccessResponse(http)
		return
	}
	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	cdClient := console.NewConsoleCdClient(params.ThirdpartyCDToken)
	appSecret, err := cdClient.CreateSite(params.DomainHost, params.ReleaseName)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	if appSecret.AppId == "" || appSecret.AppSecret == "" {
		self.JsonResponseWithServerError(http, errors.New("app id or secret is empty"))
		return
	}
	// appSecret := &console.AppSecret{
	// 	AppId:     "123",
	// 	AppSecret: "456",
	// }

	patchData := `{
		"spec": {
			"template": {
				"spec": {
					"containers": [
						{
							"name": "%s",
							"env": [
								{
									"name": "APP_ID",
									"value": "%s"
								},
								{
									"name": "APP_SECRET",
									"value": "%s"
								}
							]
						}
					]
				}
			}
		}
	}`
	patchData = fmt.Sprintf(patchData, params.DeploymentName, appSecret.AppId, appSecret.AppSecret)
	//deployment 修改env
	//patch deployment

	_, err = client.ClientSet.
		AppsV1().
		Deployments(params.Namespace).
		Patch(client.Ctx, params.DeploymentName, k8stypes.StrategicMergePatchType, []byte(patchData), metav1.PatchOptions{})
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonSuccessResponse(http)
}

func (self Zpk) Test(http *gin.Context) {

	type ParamsValidate struct {
		Namespace string `form:"namespace" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	helmApi := k8s.NewHelm(client)
	err = helmApi.IsReachable(params.Namespace)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonSuccessResponse(http)
}

func (self Zpk) UpgradeInfo(http *gin.Context) {
	type ParamsValidate struct {
		Namespace         string `form:"namespace" binding:"required"`
		ReleaseName       string `form:"releaseName" binding:"required"`
		ThirdpartyCDToken string `form:"thirdpartyCDToken"`
	}

	if !facade.GetConfig().GetBool("app.upgrade_enable") {
		self.JsonResponse(http, logic.NotUpgrade, nil, 200)
		return
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	upgradeCheck := logic.NewUpgradeCheck(client)
	upgradeCheck.WithCDToken(params.ThirdpartyCDToken)
	result := upgradeCheck.Check(params.Namespace, params.ReleaseName)

	self.JsonResponse(http, result, nil, 200)
	return

}

/*
传统应用列表
*/
func (self Zpk) TranditionList(http *gin.Context) {
	type t struct {
		ZpkUrl    string `json:"zpkUrl" binding:"required"`
		Icon      string `json:"icon" binding:"required"`
		Title     string `json:"title" binding:"required"`
		DetailUrl string `json:"detailUrl" binding:"required"`
	}
	result := map[string]t{
		"php7.2": t{
			ZpkUrl:    "https://zpk.w7.cc/respo/info/tradition_php72",
			Icon:      "https://zpk.w7.cc/zip/icon/tradition_php72",
			DetailUrl: "https://zpk.w7.cc/respo/detail/tradition_php72",
			Title:     "php7.2环境",
		},
		"php7.3": t{
			ZpkUrl:    "https://zpk.w7.cc/respo/info/tradition_php73",
			Icon:      "https://zpk.w7.cc/zip/icon/tradition_php73",
			DetailUrl: "https://zpk.w7.cc/respo/detail/tradition_php73",
			Title:     "php7.3环境",
		},
		"php7.4": t{
			ZpkUrl:    "https://zpk.w7.cc/respo/info/tradition_php74",
			Icon:      "https://zpk.w7.cc/zip/icon/tradition_php74",
			DetailUrl: "https://zpk.w7.cc/respo/detail/tradition_php74",
			Title:     "php7.4环境",
		},
		"php8.0": t{
			ZpkUrl:    "https://zpk.w7.cc/respo/info/tradition_php80",
			Icon:      "https://zpk.w7.cc/zip/icon/tradition_php80",
			DetailUrl: "https://zpk.w7.cc/respo/detail/tradition_php80",
			Title:     "php8.0环境",
		},
		"php8.1": t{
			ZpkUrl:    "https://zpk.w7.cc/respo/info/tradition_php81",
			Icon:      "https://zpk.w7.cc/zip/icon/tradition_php81",
			DetailUrl: "https://zpk.w7.cc/respo/detail/tradition_php81",
			Title:     "php8.1环境",
		},
	}
	self.JsonResponse(http, result, nil, 200)
}

/*
* 安装传统应用
 */
func (self Zpk) InstallTrandition(http *gin.Context) {
	type ParamsValidate struct {
		Namespace   string `form:"namespace" binding:"required"`
		ReleaseName string `form:"releaseName" binding:"required"`
		ZpkUrl      string `form:"zpkUrl" binding:"required"`
		ZipURL      string `form:"zipUrl" binding:"required"`
		PvcName     string `form:"pvcName" binding:"required"`
		EntryRoot   string `form:"entryRoot" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	pk, err := logic.LoadPackage(params.ZpkUrl)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	pk.ZipUrl = params.ZipURL
	pk.Manifest.Platform.Container.Shells = []types.Shell{
		{
			Type:  "install",
			Title: "传统应用安装",
			Shell: "/home/createsite.sh %CODE_ZIP_URL% " + params.EntryRoot,
		},
	}
	intstallOptions := []types.InstallOption{
		{
			Identifie: pk.Manifest.Application.Identifie,
			PvcName:   params.PvcName,
			Replicas:  1,
			Namespace: params.Namespace,
		},
	}
	installId := helper.RandomString(5)
	packageApps := types.NewPackage(pk, intstallOptions, params.ReleaseName, installId, params.Namespace,
		"", "", "")

	token := http.MustGet("k8s_token").(string)
	// k8sToken := k8s.NewK8sToken(token)
	client, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	install := logic.NewInstall(client, packageApps)
	err = install.InstallOrUpgrade(params.ReleaseName, params.Namespace)
	if err != nil {
		// panic(err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	result := map[string]interface{}{
		"releaseName": params.ReleaseName,
		"installId":   installId,
		"namespace":   params.Namespace,
	}
	self.JsonResponse(http, result, nil, 200)
}

/*
*
外部依赖配置
*/
func (self Zpk) OutDependEnv(http *gin.Context) {
	type ParamsValidate struct {
		Namespace string `form:"namespace" binding:"required"`
		Identifie string `form:"identifie" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	result := logic.DependEnvResult{
		Installed: false,
		Envs:      make(map[string]string),
	}

	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	depend := logic.NewDependEnv(client)
	result2, err := depend.LoadEnv(params.Identifie, params.Namespace)
	if err != nil {
		self.JsonResponseWithoutError(http, result)
		return
	}

	self.JsonResponseWithoutError(http, result2)
}

func (self Zpk) GenHelmMemory(http *gin.Context) {

	params := logic.HelmMemory{}
	if !self.Validate(http, &params) {
		return
	}
	if (params.Repository == "") && params.ChartName != "" {
		// param.Repository = "https://charts.helm.sh/stable"
		// param.Version = "latest"
		parseBytes, err := helper.ExtractSingleFileFromTgz(params.ChartName, "Chart.yaml")
		if err != nil {
			self.JsonResponseWithServerError(http, fmt.Errorf("无法正确解析chart包: %v", err))
		}
		meta := chart.Metadata{}
		err = yaml.Unmarshal(parseBytes, &meta)
		if err != nil {
			self.JsonResponseWithServerError(http, fmt.Errorf("无法正确解析chart包: %v", err))
		}
		// params.Repository = ""
		params.Version = meta.AppVersion
		params.Identifie = meta.Name
		params.Icon = meta.Icon
		params.Title = meta.Name
	}
	params.Identifie = strings.ReplaceAll(params.Identifie, "_", "-")
	manifest := logic.HelmManifestApp(&params)
	logic.NewManifestSingleton().Put(params.Identifie, &manifest)

	self.JsonResponseWithoutError(http, gin.H{
		"identifie": params.Identifie,
		"zpkUrl":    "memory://" + params.Identifie,
	})
}
func (self Zpk) ChartYaml(http *gin.Context) {

	type ParamsValidate struct {
		TgzPath string `form:"tgzPath" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	data, err := helper.ExtractSingleFileFromTgz(params.TgzPath, "Chart.yaml")
	if err != nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("无法正确解析chart包: %v", err))
		return
	}

	data2, err := helper.YamlParse(data)
	if err != nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("无法正确解析chart包: %v", err))
		return
	}
	self.JsonResponseWithoutError(http, data2)
}

/*
*
外部依赖配置
*/
func (self Zpk) LastVersionEnv(http *gin.Context) {
	type ParamsValidate struct {
		Namespace string `form:"namespace" binding:"required"`
		Name      string `form:"name" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	result := logic.DependEnvResult{
		Installed: false,
		Envs:      make(map[string]string),
	}

	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	depend := logic.NewDependEnv(client)
	if params.Namespace == "" {
		params.Namespace = "default"
	}
	result2, err := depend.LoadLastVersionEnv(params.Name, params.Namespace)
	if err != nil {
		self.JsonResponseWithoutError(http, result)
		return
	}

	self.JsonResponseWithoutError(http, result2)
}

func (self Zpk) OciDown(http *gin.Context) {
	ociRef := http.Param("oci")
	mediaType := http.Query("mediaType")
	if ociRef == "" {
		self.JsonResponseWithServerError(http, errors.New("OCI reference is required"))
		return
	}
	if strings.Contains(ociRef, "oci://") == false {
		self.JsonResponseWithServerError(http, errors.New("OCI reference is required"))
		return
	}

	if ociRef == "" {
		self.JsonResponseWithServerError(http, errors.New("OCI reference is required"))
		return
	}

	// 如果未指定媒体类型，默认使用代码压缩包类型
	if mediaType == "" {
		mediaType = logic.MediaTypeCodeZip
	}
	mediaType = logic.MediaToRealType(mediaType)

	// 创建临时文件来存储下载的内容
	tempFile, err := os.CreateTemp("", "oci-download-*")
	if err != nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("创建临时文件失败: %w", err))
		return
	}
	tempFilePath := tempFile.Name()
	tempFile.Close() // 关闭文件，让FetchArtifact可以写入

	// 在函数结束时删除临时文件
	defer os.Remove(tempFilePath)

	oci, err := logic.NewOCI(k8s.NewK8sClient().Sdk, ociRef)
	if err != nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("创建OCI客户端失败: %w", err))
		return
	}
	// 创建OCI客户端并获取文件
	err = oci.Download(http.Request.Context(), mediaType, tempFilePath)
	if err != nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("获取OCI制品失败: %w", err))
		return
	}

	// 打开下载的文件
	file, err := os.Open(tempFilePath)
	if err != nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("打开下载的文件失败: %w", err))
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("获取文件信息失败: %w", err))
		return
	}

	// 设置响应头
	fileName := filepath.Base(ociRef)
	if mediaType == logic.MediaTypeIcon {
		http.Header("Content-Type", "image/png")
		fileName = "icon.png"
	} else if mediaType == logic.MediaTypeFilesJson {
		http.Header("Content-Type", "application/json")
		fileName = "files.json"
	} else if mediaType == logic.MediaTypeCodeZip || mediaType == logic.MediaTypeWebCodeZip {
		http.Header("Content-Type", "application/zip")
		if mediaType == logic.MediaTypeCodeZip {
			fileName = "code.zip"
		} else {
			fileName = "web-code.zip"
		}
	}

	http.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	http.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// 将文件内容发送到响应
	http.Writer.WriteHeader(200)
	_, err = io.Copy(http.Writer, file)
	if err != nil {
		slog.Error("发送文件内容失败", "error", err)
		// 此时已经开始发送响应，无法再发送错误信息
	}
}

func (self Zpk) BuildImageJob(http *gin.Context) {

	params := zpkk8stypes.BuildImageParams{}
	if !self.Validate(http, &params) {
		return
	}
	if params.DockerRegistry.Host == "registry.local.w7.cc" {
		params.HostNetwork = true
		params.DockerRegistry.Username = "admin"
		params.DockerRegistry.Password = "w7-secret"
	}
	sdk := k8s.NewK8sClient()
	if params.DockerRegistrySecretName != "" {
		dockerSecret, err := sdk.ClientSet.CoreV1().Secrets("default").Get(sdk.Ctx, params.DockerRegistrySecretName, metav1.GetOptions{})
		if err != nil {
			slog.Error("获取docker registry secret失败", "error", err)
		} else {
			// 解析.dockerconfigjson字段，获取用户名和密码
			var dockerConfig struct {
				Auths map[string]struct {
					Username string `json:"username"`
					Password string `json:"password"`
					Auth     string `json:"auth"`
				} `json:"auths"`
			}
			if err := json.Unmarshal(dockerSecret.Data[".dockerconfigjson"], &dockerConfig); err != nil {
				slog.Error("解析docker config json失败", "error", err)
			} else {
				for key, auth := range dockerConfig.Auths {
					if auth.Username != "" && auth.Password != "" {
						params.DockerRegistry.Username = auth.Username
						params.DockerRegistry.Password = auth.Password
						break
					}
					if key == "registry.local.w7.cc" {
						params.HostNetwork = true
						break
					}
				}
			}
		}
	}
	job := zpkk8s.ToZpkBuildJob(&params)
	client, err := sdk.Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	job, err = client.ClientSet.BatchV1().Jobs("default").Create(client.Ctx, job, metav1.CreateOptions{})
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, job)
}

func (self Zpk) BuildImageCronJob(http *gin.Context) {

	params := zpkk8stypes.BuildImageParams{}
	if !self.Validate(http, &params) {
		return
	}
	job := zpkk8s.ToZpkBuildCronJob(&params, params.Schedule)
	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	job, err = client.ClientSet.BatchV1().CronJobs("default").Create(client.Ctx, job, metav1.CreateOptions{})
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, job)
}

// 本地安装制品库的域名
func (self Zpk) LocalZpkUrl(http *gin.Context) {

	sdk := k8s.NewK8sClient().Sdk
	ingressList, err := sdk.ClientSet.NetworkingV1().Ingresses("default").List(sdk.Ctx, metav1.ListOptions{LabelSelector: "app.kubernetes.io/instance=w7-zpkv2"})
	if err != nil || len(ingressList.Items) == 0 {
		self.JsonResponseWithoutError(http, gin.H{
			"host":       "",
			"hasIngress": "false",
		})
		return
	}

	ingress := ingressList.Items[0]
	deployments, err := sdk.ClientSet.AppsV1().Deployments(ingress.GetNamespace()).List(sdk.Ctx, metav1.ListOptions{LabelSelector: "app.kubernetes.io/instance=w7-zpkv2"})
	if err != nil || len(deployments.Items) == 0 {
		self.JsonResponseWithoutError(http, gin.H{
			"host":       "",
			"hasIngress": "false",
		})
		return
	}
	oauthToken := ""
	deployment := deployments.Items[0]
	container := deployment.Spec.Template.Spec.Containers[0]
	envs := container.Env
	for _, env := range envs {
		if env.Name == "OAUTH_TOKEN" {
			oauthToken = env.Value
		}
	}

	self.JsonResponseWithoutError(http, gin.H{
		"host":       ingressList.Items[0].Spec.Rules[0].Host,
		"hasIngress": len(ingressList.Items) > 0,
		"isHttps":    ingressList.Items[0].Spec.TLS != nil,
		"oauthToken": oauthToken,
	})
}
