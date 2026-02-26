package logic

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/app/zpk/logic/types"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	convert "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	helm "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	zpktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
	"gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	"github.com/samber/lo"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Install struct {
	sdk      *k8s.Sdk
	pk       types.Package
	groupApi *appgroup.AppGroupApi
}

func NewInstall(sdk *k8s.Sdk, pk types.Package) *Install {
	groupApi, err := appgroup.NewAppGroupApi(sdk)
	if err != nil {
		return nil
	}
	return &Install{sdk: sdk, pk: pk, groupApi: groupApi}
}

func (z *Install) InstallOrUpgrade(name, namespace string) error {
	_, err := z.Get(name, namespace)
	if err != nil {
		return z.Install(name, namespace)
	} else {
		return z.Upgrade(name, namespace)
	}
}

func (z *Install) createHelmJob(myPack *types.PackageApp, shellType types.ShellType, children []*types.PackageApp) (v1alpha1.DeployItem, []*batchv1.Job, error) {
	jobs := []*batchv1.Job{}
	job := toHelmInstallJob(myPack, children)
	jobs = append(jobs, job)
	shell := myPack.GetShellByType(string(shellType))
	var shellJob *batchv1.Job
	if shell != nil {
		shellJob = convert.ToShellJob2(myPack, myPack, string(shellType))
		jobs = append(jobs, shellJob)
	}
	info := v1alpha1.ResourceInfo{
		Name:         job.Name,
		Namespace:    myPack.GetNamespace(),
		Kind:         "Job",
		ApiVersion:   "batch/v1",
		DeployStatus: v1alpha1.StatusDeploying,
		DeployTitle:  "helm安装",
	}

	installResult := v1alpha1.DeployItem{
		Identifie:    myPack.GetIdentifie(),
		Title:        myPack.GetTitle(),
		ResourceList: []v1alpha1.ResourceInfo{info},
		DeployStatus: v1alpha1.StatusDeploying,
	}
	if shellJob != nil {
		shellInfo := v1alpha1.ResourceInfo{
			Name:         shellJob.Name,
			Namespace:    myPack.GetNamespace(),
			Kind:         "Job",
			ApiVersion:   "batch/v1",
			DeployStatus: v1alpha1.StatusDeploying,
			DeployTitle:  shell.GetTitle(),
		}
		installResult.ResourceList = append(installResult.ResourceList, shellInfo)
	}
	return installResult, jobs, nil
}

func (z *Install) InstallUseJob(name, namespace string, shellType types.ShellType) error {

	items := []v1alpha1.DeployItem{}
	jobs := []*batchv1.Job{}
	groups := []*v1alpha1.AppGroup{}
	// microApps := []*microapp.MicroApp{}
	root := z.pk.Root
	rootItem, rootjobs, err := z.createHelmJob(root, shellType, z.pk.Children)
	if err != nil {
		slog.Error("create helm job error", slog.String("error", err.Error()))
		return err
	}
	// rootMicro := convert.ToMicroApp(root)
	// if rootMicro != nil {
	// 	microApps = append(microApps, rootMicro)
	// }

	items = append(items, rootItem)
	jobs = append(jobs, rootjobs...)

	for _, child := range z.pk.Children {
		if !child.IsHelm() {
			continue
		}
		if child.Replicas == 0 {
			continue
		}
		childItem, childJobs, err := z.createHelmJob(child, shellType, []*types.PackageApp{})
		if err != nil {
			slog.Error("create child helm job error", slog.String("error", err.Error()))
			return err
		}
		for _, childJob := range childJobs {
			childJob.Labels["w7.cc/parent"] = name
		}
		items = append(items, childItem)
		jobs = append(jobs, childJobs...)
		chidlGroup := convert.ToAppGroup(child, []v1alpha1.DeployItem{childItem})
		chidlGroup.Labels["w7.cc/parent"] = name
		groups = append(groups, chidlGroup)
	}
	rootGroup := convert.ToAppGroup(root, items)
	rootGroup.Spec.UpgradingVersion = rootGroup.Spec.Version

	// if (root.Parent != nil)  { //package app getLabel 判断了Parent is nil 就不需要设置parent
	// 	continue
	// }
	groups = append(groups, rootGroup)

	// sigClient, err := z.sdk.ToSigClient()
	// if err != nil {
	// 	slog.Error("create sig client error", slog.String("error", err.Error()))
	// 	return err
	// }

	for _, group := range groups {
		err = z.persistGroup(group)
		if err != nil {
			slog.Error("update group error", slog.String("error", err.Error()))
			return err
		}
	}

	// if !z.pk.IsHelm() {
	// 	for _, microApp := range microApps {
	// 		clone := microApp.DeepCopy()
	// 		_, err = controllerutil.CreateOrUpdate(z.sdk.Ctx, sigClient, clone, func() error {
	// 			clone.Spec = microApp.Spec
	// 			return nil
	// 		})
	// 		if err != nil {
	// 			slog.Error("create microapp error", slog.String("error", err.Error()))
	// 			return err
	// 		}
	// 	}
	// }
	for _, job := range jobs {
		_, err = z.sdk.ClientSet.BatchV1().Jobs(root.GetNamespace()).Create(z.sdk.Ctx, job, metav1.CreateOptions{})
		if err != nil {
			slog.Error("create job error", slog.String("error", err.Error()))
			return err
		}
	}

	return nil
}

func (z *Install) IsHelm() bool {
	return z.pk.IsHelm()
}

func (z *Install) NeedHelmInstall() bool {
	// return z.pk.IsHelm()
	return z.pk.IsHelm() || z.pk.Root.HelmUrl != ""
}

func (z *Install) Install(name, namespace string) error {
	go downStatic(z.pk.Root)
	if z.NeedHelmInstall() {
		//为啥helm 单独走一条线， 如果helmjob 当作一个helmchart安装的花，导致helm更新时候判断currentRelease只有一个job,比对不出来需要更新的资源
		//导致pvc每次都重建 数据丢失
		return z.InstallUseJob(name, namespace, types.ShellInstall)
	}

	helmchart := NewHelmChart(z.pk, types.ShellInstall)
	group, chart, err := helmchart.ToHelmChartWithGroup()
	if err != nil {
		return err
	}
	err = z.CreateOrUpdateGroup(namespace, name, group.Status.DeployItems)
	if err != nil {
		slog.Error("create or update group error", slog.String("error", err.Error()))
		return err
	}
	vals := map[string]interface{}{}
	// if z.pk.IsHelm() {
	vals, err = helmchart.GetValues()
	if err != nil {
		return err
	}
	// }
	helmApi := k8s.NewHelm(z.sdk)
	labels := z.fillLabelAndAnnation(chart)
	_, err = helmApi.Install(z.sdk.Ctx, chart, vals, name, namespace, labels)
	if err != nil {
		return err
	}
	return nil
}

func (z *Install) Upgrade(name, namespace string) error {
	go downStatic(z.pk.Root)
	if z.NeedHelmInstall() {
		return z.InstallUseJob(name, namespace, types.ShellUpgrade)
	}
	helmchart := NewHelmChart(z.pk, types.ShellUpgrade)
	group, chart, err := helmchart.ToHelmChartWithGroup()
	if err != nil {
		return err
	}
	err = z.CreateOrUpdateGroup(namespace, name, group.Status.DeployItems)
	if err != nil {
		slog.Error("create or update group error", slog.String("error", err.Error()))
		return err
	}
	helmApi := k8s.NewHelm(z.sdk)
	labels := z.fillLabelAndAnnation(chart)
	vals := map[string]interface{}{}
	if z.pk.IsHelm() {
		vals, err = helmchart.GetValues()
		if err != nil {
			return err
		}
	}
	_, err = helmApi.Upgrade(z.sdk.Ctx, chart, vals, name, namespace, labels)
	if err != nil {
		return err
	}
	return nil
}
func (z *Install) persistGroup(group *v1alpha1.AppGroup) error {
	namespace := group.Namespace
	fetchGroup, err := z.groupApi.GetAppGroup(namespace, group.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = z.groupApi.CreateGroup(namespace, group)
			if err != nil {
				slog.Error("create group error", slog.String("error", err.Error()))
				return err
			}
		}
		return err
	}
	oldVersion := fetchGroup.Spec.Version
	fetchGroup.Spec = group.Spec
	keepAnnoKey := []string{"w7.cc/domains", "w7.cc/ports", "w7.cc/default-domain", "w7.cc/create-svc"}
	for k, v := range group.Annotations {
		if !lo.Contains(keepAnnoKey, k) {
			fetchGroup.Annotations[k] = v
		}
	}
	//更新中的版本
	fetchGroup.Annotations = group.Annotations
	fetchGroup.Labels = group.Labels

	fetchGroup.Spec.Version = oldVersion
	fetchGroup.Spec.UpgradingVersion = group.Spec.Version

	fetchGroup.Status.DeployItems = []v1alpha1.DeployItem{}
	fetchGroup.Status.DeployStatus = v1alpha1.StatusDeploying
	fetchGroup.Status.DeployItems = group.Status.DeployItems
	parentName, ok := fetchGroup.Labels["w7.cc/parent"]
	if ok {
		fetchGroup.Labels["w7.cc/parent"] = parentName
	}
	_, err = z.groupApi.UpdateAppGroup(namespace, fetchGroup)
	if err != nil {
		slog.Error("update group error", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (z *Install) CreateOrUpdateGroup(namespace, name string, items []v1alpha1.DeployItem) error {
	group, err := z.groupApi.GetAppGroup(namespace, name)
	if err != nil {
		group2 := helm.ToAppGroup(z.pk.Root, items)
		// group2.Spec.IsHelm = z.pk.IsHelm()
		_, err = z.groupApi.CreateGroup(namespace, group2)
		if err != nil {
			return err
		}
		return nil
	}
	if group != nil {
		group3 := helm.ToAppGroup(z.pk.Root, items)
		keepAnnoKey := []string{"w7.cc/domains", "w7.cc/ports", "w7.cc/default-domain", "w7.cc/create-svc"}
		for k, v := range group3.Annotations {
			if !lo.Contains(keepAnnoKey, k) {
				group.Annotations[k] = v
			}
		}
		oldVersion := group.Spec.Version
		group.Spec = group3.Spec
		group.Spec.Version = oldVersion
		group.Spec.UpgradingVersion = group3.Spec.Version
		group.Status.DeployItems = []v1alpha1.DeployItem{}
		group.Status.DeployStatus = v1alpha1.StatusDeploying
		group.Status.DeployItems = append(group.Status.DeployItems, items...)
		_, err := z.groupApi.UpdateAppGroup(namespace, group)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (z *Install) UnInstall(name, namespace string) error {
	helmApi := k8s.NewHelm(z.sdk)
	_, err := helmApi.UnInstall(namespace, name)
	if err != nil {
		return err
	}
	return nil
}

func (z *Install) Get(name, namespace string) (*release.Release, error) {
	helmApi := k8s.NewHelm(z.sdk)
	return helmApi.Info(name, namespace)
}

func (z *Install) fillLabelAndAnnation(chart *chart.Chart) map[string]string {
	return z.GetLabels()
}

// func (z *Install) GetAnnotations() map[string]string {
// 	helmConfig := z.pk.Root.Manifest.Platform.Helm
// 	anno := map[string]string{
// 		zpktypes.HELM_RELEASE_SOURCE:   "zpk",
// 		zpktypes.HELM_REPOSITORY_URL:   helmConfig.Repository,
// 		zpktypes.HELM_CHART_NAME:       helmConfig.ChartName,
// 		zpktypes.HELM_CHART_VERSION:    helmConfig.Version,
// 		zpktypes.HELM_ZPK_VERSION:      z.pk.Root.GetVersion(),
// 		zpktypes.HELM_ZPK_URL:          z.pk.Root.ZpkUrl,
// 		zpktypes.HELM_LOGO:             z.pk.Root.Manifest.Application.Icon,
// 		zpktypes.HELM_INDENTIFIE:       z.pk.Root.GetIdentifie(),
// 		zpktypes.HELM_APPLICATION_TYPE: z.pk.Root.Manifest.Application.Type,
// 		zpktypes.HELM_TITLE:            z.pk.Root.Manifest.Application.Name,
// 	}
// 	return anno
// }

func (z *Install) GetLabels() map[string]string {
	label := map[string]string{
		zpktypes.HELM_RELEASE_SOURCE: "zpk",
		zpktypes.HELM_INDENTIFIE:     z.pk.Root.GetIdentifie(),
	}
	return label
}
