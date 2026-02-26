package appgroup

import (
	"log/slog"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	zpktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
	appv1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	v1alpha1Types "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	"gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/appgroup/listers/appgroup/v1alpha1"
	v1alpha1Lister "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/appgroup/listers/appgroup/v1alpha1"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
)

type HelmWorkload struct {
	appGroupLister v1alpha1.AppGroupLister
	helmApi        *k8s.Helm
	groupApi       *AppGroupApi
	queue          *EventQueue
}

func NewHelmWorkload(helm *k8s.Helm, appGroupLister v1alpha1Lister.AppGroupLister, queue *EventQueue, groupApi *AppGroupApi) *HelmWorkload {
	if appGroupLister == nil {
		slog.Error("app group lister is nil")
		panic("app group lister is nil")
	}
	return &HelmWorkload{
		appGroupLister: appGroupLister,
		helmApi:        helm,
		queue:          queue,
		groupApi:       groupApi,
	}
}

func (h *HelmWorkload) Sync() error {

	releases, err := h.helmApi.ListRaw("default", "")
	if err != nil {
		return err
	}
	groups := map[string]*v1alpha1Types.AppGroup{}
	releasesMap := map[string]*release.Release{}
	list, err := h.appGroupLister.AppGroups("default").List(labels.Everything())
	if err != nil {
		return err
	}
	for _, release := range releases {
		releasesMap[release.Name] = release
	}
	//relase 没有的，需要删除appgroup
	for _, r := range list {
		groups[r.Name] = r
		// helm卸载会触发helm workloadmanager eventqueue 删除事件，这里不处理了
		// if _, ok := releasesMap[r.Name]; !ok {
		// 	group, err := h.appGroupLister.AppGroups(r.Namespace).Get(r.Name)
		// 	if err != nil {
		// 		slog.Warn("get app group error", "err", err)
		// 		continue
		// 	}
		// 	if group.Spec.Type == "helm" || group.Spec.Type == "zpk" {
		// 		if group.Status.Items == nil || len(r.Status.Items) == 0 {
		// 			err := h.groupApi.DeleteAppGroup(r.Namespace, group.Name)
		// 			if err != nil {
		// 				slog.Warn("delete app group error", "err", err)
		// 			}
		// 		}
		// 	}
		// }
	}
	//release 有的，但是appgroup 没有的，需要创建appgroup
	for _, release := range releases {
		_, ok := groups[release.Name]
		if !ok {
			err := h.handleWorkload(release)
			if err != nil {
				slog.Warn("handle helm workload error", "err", err)
			}
		}
	}

	return nil
}

func (h *HelmWorkload) handleWorkload(release *release.Release) error {
	if release.Info.Status != "deployed" {
		return nil
	}
	group := h.releaseToAppGroup(release)
	_, err := h.groupApi.CreateGroup(group.Namespace, group)
	if err != nil {
		slog.Error("create appgroup error", "err", err)
	}
	h.resourceToQueue(release)
	return nil
}

func (h *HelmWorkload) resourceToQueue(release *release.Release) (bool, error) {
	resourceList, err := h.helmApi.BuildResourceList([]byte(release.Manifest), false)
	if err != nil {
		slog.Error("build resource list error", "err", err)
		return false, err
	}
	hasApp := false
	for _, k := range resourceList {
		kind := k.Mapping.GroupVersionKind.Kind
		if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" {
			meta := metav1.ObjectMeta{
				Name:      k.Name,
				Namespace: k.Namespace,
			}
			typemeta := metav1.TypeMeta{
				Kind:       kind,
				APIVersion: "apps/v1",
			}
			event := NewK8sResourceEvent(typemeta, meta, "update", false)
			// h.queue.AddEvent(event)
			h.queue.AddAfter(event, time.Minute)
			// w := NewWorkloadWrapper(NewWorkloadMock("apps/v1", kind, k.Name, k.Namespace, release.Name))
			// h.queue.Push(w.queueKey())
			hasApp = true
		}

	}
	return hasApp, nil
}

func (h *HelmWorkload) releaseToAppGroup(release *release.Release) *v1alpha1Types.AppGroup {
	annotations := release.Chart.Metadata.Annotations
	source, ok := annotations[zpktypes.HELM_RELEASE_SOURCE]
	isHelm := true
	wtype := "helm"
	if ok && source == "zpk" {
		wtype = "zpk"
		isHelm = false
		manifestAppType := annotations[zpktypes.HELM_APPLICATION_TYPE]
		if manifestAppType == "helm" {
			isHelm = true
		}
	}
	identifie, ok1 := annotations[zpktypes.HELM_INDENTIFIE]
	if !ok1 {
		identifie = release.Name
	}
	logo, ok2 := annotations[zpktypes.HELM_LOGO]
	if !ok2 {
		logo = release.Chart.Metadata.Icon
	}
	title, ok3 := annotations[zpktypes.HELM_TITLE]
	if !ok3 {
		title = release.Name
	}
	version := release.Chart.Metadata.AppVersion
	zpkVersion, ok4 := annotations[zpktypes.HELM_ZPK_VERSION]
	if ok4 {
		version = zpkVersion
	}
	// if version == "" {
	// 	version = release.Chart.Metadata.Version
	// }
	groupSpec := appv1.AppGroupSpec{
		Type:        wtype,
		Identifie:   identifie,
		Version:     version,
		Logo:        logo,
		Description: release.Chart.Metadata.Description,
		ZpkUrl:      release.Chart.Metadata.Annotations[zpktypes.HELM_ZPK_URL],
		Suffix:      release.Name,
		// Annotations: release.Chart.Metadata.Annotations,
		Title: title,
		HelmConfig: v1alpha1Types.HelmConfig{
			ChartName:  release.Chart.Metadata.Annotations[zpktypes.HELM_CHART_NAME],
			Repository: release.Chart.Metadata.Annotations[zpktypes.HELM_REPOSITORY_URL],
			Version:    release.Chart.Metadata.AppVersion,
		},
		IsHelm: isHelm,
	}
	group := CreateAppGroup(release.Name, release.Namespace)
	group.Spec = groupSpec
	group.Status = v1alpha1Types.AppGroupStatus{
		Items: []v1alpha1Types.AppGroupItemStatus{},
		Ready: true,
	}
	return group
}

// for _, r := range releases {
// 	element := &k8s.ReleaseElement{
