package k8s

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
)

var DefaultChartYaml = `
apiVersion: v2
appVersion: {{ .Chart.Version }}
description: {{ .Chart.Description }}
name: {{ .Chart.Name }}
version: {{ .Chart.Version }}
icon: {{ .Chart.Icon }}
`

type Helm struct {
	sdk    *Sdk
	atomic bool
}

type releaseElement struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Namespace   string `json:"namespace"`
	Revision    string `json:"revision"`
	Updated     string `json:"updated"`
	Status      string `json:"status"`
	Chart       string `json:"chart"`
	AppVersion  string `json:"app_version"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Created     string `json:"created"`
	ZpkUrl      string `json:"zpk_url"`
}

func NewHelm(sdk *Sdk) *Helm {
	return &Helm{sdk: sdk}
}

func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func LocateChartByHelm(respo, chartName, version string) (*chart.Chart, error) {
	client, err := registry.NewClient(registry.ClientOptPlainHTTP())
	if err != nil {
		return nil, err
	}
	return LocateChart(respo, chartName, true, client, version)
}

// 根据仓库地址获取chart
func LocateChart(respo string, chartname string, dependencyUpdate bool, registry *registry.Client, version string) (*chart.Chart, error) {
	settings := cli.New()
	pathopton := action.ChartPathOptions{RepoURL: respo, Version: version}
	cp, err := pathopton.LocateChart(chartname, settings)
	if err != nil {
		return nil, err
	}
	p := getter.All(settings)
	// vals, err := values.MergeValues(p)
	// if err != nil {
	// 	return nil, err
	// }
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		slog.Warn("This chart is deprecated")
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			err = errors.Wrap(err, "An error occurred while checking for chart dependencies. You may need to run `helm dependency build` to fetch missing dependencies")
			if dependencyUpdate {
				man := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        cp,
					Keyring:          pathopton.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
					RegistryClient:   registry,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(cp); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}
	return chartRequested, nil
}
func (self *Helm) Atomic(val bool) {
	self.atomic = val
}
func (self *Helm) InstallRaw(context context.Context, respo, chartName, releaseName, namespace string, vals map[string]interface{}, labels map[string]string) (*release.Release, error) {
	chart, err := LocateChartByHelm(respo, chartName, "")
	if err != nil {
		return nil, err
	}
	return self.Install(context, chart, vals, releaseName, namespace, labels)

}
func (self *Helm) Install(context context.Context, chart *chart.Chart, vals map[string]interface{}, releaseName string, namespace string, labels map[string]string) (*release.Release, error) {
	// 创建一个新的action配置
	cfg, err := self.getConfiguration(namespace)
	if err != nil {
		return nil, err
	}
	installAction := action.NewInstall(cfg)
	installAction.ReleaseName = releaseName
	installAction.Namespace = namespace
	installAction.Replace = true
	installAction.Atomic = self.atomic
	// installAction
	installAction.Timeout = time.Duration(300 * time.Second)
	installAction.IsUpgrade = true
	installAction.Labels = labels
	installAction.CreateNamespace = true
	// installAction.Replace = true
	result, err := installAction.RunWithContext(context, chart, vals)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *Helm) Upgrade(context context.Context, chart *chart.Chart, vals map[string]interface{}, releaseName string, namespace string, labels map[string]string) (*release.Release, error) {
	// 创建一个新的action配置
	cfg, err := self.getConfiguration(namespace)
	if err != nil {
		return nil, err
	}
	action := action.NewUpgrade(cfg)
	action.Namespace = namespace
	// action.Force = true
	// action.Recreate = true
	// action.Force = true
	// action.Recreate = false
	action.TakeOwnership = true
	action.Atomic = self.atomic
	action.Timeout = time.Duration(300 * time.Second)
	action.Labels = labels
	result, err := action.RunWithContext(context, releaseName, chart, vals)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *Helm) ReUseValues(context context.Context, vals map[string]interface{}, releaseName string, namespace string) (*release.Release, error) {
	// 创建一个新的action配置
	release, err := self.Info(releaseName, namespace)
	if err != nil {
		return nil, err
	}
	//createdAt modifiedAt
	labels := map[string]string{}
	for k, v := range release.Labels {
		if k == "createdAt" || k == "modifiedAt" || k == "owner" || k == "status" || k == "version" || k == "name" {
			continue
		}
		labels[k] = v
	}

	// labels := release.Labels
	// delete(labels, "createdAt")
	// delete(labels, "modifiedAt")
	return self.Upgrade(context, release.Chart, vals, releaseName, namespace, labels)
}

func (self *Helm) UnInstall(releaseName, namespace string) (*release.UninstallReleaseResponse, error) {
	// 创建一个新的action配置
	cfg, err := self.getConfiguration(namespace)
	if err != nil {
		return nil, err
	}
	uninstall := action.NewUninstall(cfg)
	uninstall.Wait = true
	uninstall.IgnoreNotFound = true
	// uninstall.DeletionPropagation = "background"
	uninstall.Timeout = time.Duration(600 * time.Second)
	result, err := uninstall.Run(releaseName)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *Helm) Info(releaseName string, namespace string) (*release.Release, error) {
	// 创建一个新的action配置
	cfg, err := self.getConfiguration(namespace)
	if err != nil {
		return nil, err
	}
	getAction := action.NewGet(cfg)
	result, err := getAction.Run(releaseName)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *Helm) ListRaw(namespace string, labelSelector string) ([]*release.Release, error) {
	// 创建一个新的action配置
	cfg, err := self.getConfiguration(namespace)
	if err != nil {
		return nil, err
	}

	// 创建一个列出所有已部署的Helm release的action
	listAction := action.NewList(cfg)
	listAction.Selector = labelSelector
	if namespace == "" {
		listAction.AllNamespaces = true
	}

	// 执行列出操作
	releases, err := listAction.Run()
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func (self *Helm) Exists(namespace string, labelSelector string) bool {
	// 创建一个新的action配置
	releaseList, err := self.ListRaw(namespace, labelSelector)
	if err != nil {
		return false
	}
	return len(releaseList) > 0
}

func (self *Helm) List(namespace string, labelSelector string) ([]releaseElement, error) {
	// 创建一个新的action配置
	cfg, err := self.getConfiguration(namespace)
	if err != nil {
		return nil, err
	}

	// 创建一个列出所有已部署的Helm release的action
	listAction := action.NewList(cfg)
	listAction.Selector = labelSelector
	if namespace == "" {
		listAction.AllNamespaces = true
	}

	// 执行列出操作
	releases, err := listAction.Run()
	if err != nil {
		return nil, err
	}

	elements := make([]releaseElement, 0, len(releases))
	for _, r := range releases {
		title := r.Chart.Metadata.Annotations["w7.cc/title"]
		if title == "" {
			title = r.Chart.Metadata.Name
		}

		element := releaseElement{
			Name:        r.Name,
			Namespace:   r.Namespace,
			Revision:    strconv.Itoa(r.Version),
			Status:      r.Info.Status.String(),
			Chart:       formatChartname(r.Chart),
			AppVersion:  formatAppVersion(r.Chart),
			Icon:        r.Chart.Metadata.Icon,
			Description: r.Chart.Metadata.Description,
			Title:       title,
			Created:     r.Info.FirstDeployed.String(),
			ZpkUrl:      r.Chart.Metadata.Annotations["w7.cc/zpk-url"],
		}

		t := "-"
		if tspb := r.Info.LastDeployed; !tspb.IsZero() {
			// if timeFormat != "" {
			// 	t = tspb.Format(timeFormat)
			// } else {
			// 	t = tspb.String()
			// }
			t = tspb.String()
		}
		element.Updated = t

		elements = append(elements, element)
	}

	return elements, nil
}

func formatChartname(c *chart.Chart) string {
	if c == nil || c.Metadata == nil {
		// This is an edge case that has happened in prod, though we don't
		// know how: https://github.com/helm/helm/issues/1347
		return "MISSING"
	}
	return fmt.Sprintf("%s-%s", c.Name(), c.Metadata.Version)
}

func formatAppVersion(c *chart.Chart) string {
	if c == nil || c.Metadata == nil {
		// This is an edge case that has happened in prod, though we don't
		// know how: https://github.com/helm/helm/issues/1347
		return "MISSING"
	}
	return c.AppVersion()
}

func (self *Helm) getConfiguration(namespace string) (*action.Configuration, error) {
	cfg := new(action.Configuration)
	if namespace == "" {
		namespace = self.sdk.GetNamespace()
	}
	err := cfg.Init(self.sdk, namespace, "secret", log.Printf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize helm client: %w", err)
	}
	// cfg.KubernetesClientSet()
	// cfg.KubernetesClientSet() = self.sdk
	return cfg, nil
}

func (self *Helm) IsReachable(namespace string) error {
	cfg := new(action.Configuration)
	if namespace == "" {
		namespace = self.sdk.GetNamespace()
	}
	err := cfg.Init(self.sdk, namespace, "secret", log.Printf)
	if err != nil {
		return fmt.Errorf("failed to initialize helm client: %w", err)
	}
	return cfg.KubeClient.IsReachable()
}

func (self *Helm) BuildResourceList(manifest []byte, validator bool) (kube.ResourceList, error) {
	cfg := new(action.Configuration)
	err := cfg.Init(self.sdk, self.sdk.GetNamespace(), "secret", log.Printf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize helm client: %w", err)
	}
	reader := bytes.NewReader(manifest)
	return cfg.KubeClient.Build(reader, validator)
}
