package controller

import (
	// "archive/zip"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	"helm.sh/helm/v3/pkg/registry"
)

type Helm struct {
	controller.Abstract
}

func (self Helm) List(http *gin.Context) {
	type ParamsValidate struct {
		Namespace     string `form:"namespace"`
		LabelSelector string `form:"labelSelector"`
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
	releases, err := helmApi.List(params.Namespace, params.LabelSelector)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponse(http, releases, nil, 200)
}

func (self Helm) Info(http *gin.Context) {
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
	releases, err := helmApi.Info(http.Param("name"), params.Namespace)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponse(http, releases, nil, 200)
}

func (self Helm) UnInstall(http *gin.Context) {
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
	releases, err := helmApi.UnInstall(http.Param("name"), params.Namespace)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponse(http, releases, nil, 200)
}

func (self Helm) InstallUseRepo(http *gin.Context) {
	type ParamsValidate struct {
		Namespace  string                 `form:"namespace" binding:"required`
		Repository string                 `form:"repository"`
		ChartName  string                 `form:"chartName" binding:"required`
		Vals       map[string]interface{} `form:"vals" binding:"required`
		ChartType  string                 `form:"chartType"`
		Version    string                 `form:"version"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	releaseName := http.Param("name")
	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	helmApi := k8s.NewHelm(client)
	registry, err := registry.NewClient(registry.ClientOptPlainHTTP())
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	chart, err := k8s.LocateChart(params.Repository, params.ChartName, true, registry, params.Version)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	releases, err := helmApi.Install(client.Ctx, chart, params.Vals, releaseName, params.Namespace, nil)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponse(http, releases, nil, 200)
}

func (self Helm) AppInfo(http *gin.Context) {
	helmVerison := facade.Config.GetString("app.helm_version")
	helmReleaseName := facade.Config.GetString("app.helm_release_name")
	helmNamespace := facade.Config.GetString("app.helm_namespace")
	metadataName := facade.Config.GetString("app.metadata_name")
	deploymentName := facade.Config.GetString("app.deployment_name")
	self.JsonResponse(http, gin.H{
		"helmVersion":     helmVerison,
		"helmReleaseName": helmReleaseName,
		"helmNamespace":   helmNamespace,
		"metadataName":    metadataName,
		"deploymentName":  deploymentName,
	}, nil, 200)

}

func (self Helm) ReUseValues(http *gin.Context) {
	type ParamsValidate struct {
		Namespace string                 `form:"namespace" binding:"required`
		Vals      map[string]interface{} `form:"vals" binding:"required`
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
	releases, err := helmApi.ReUseValues(client.Ctx, params.Vals, http.Param("name"), params.Namespace)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponse(http, releases, nil, 200)
}
