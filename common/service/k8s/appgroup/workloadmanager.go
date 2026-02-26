package appgroup

import (
	"context"
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"

	// "gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk"
	"gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	appv1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	microapp "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/microapp/v1alpha1"
	v1alpha1Lister "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/appgroup/listers/appgroup/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	appsv1lister "k8s.io/client-go/listers/apps/v1"
	batchv1lister "k8s.io/client-go/listers/batch/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
)

type WorkloadManager struct {
	groupApi          *AppGroupApi
	helm              *k8s.Helm
	sdk               *k8s.Sdk
	AppGroupLister    v1alpha1Lister.AppGroupLister
	DeploymentLister  appsv1lister.DeploymentLister
	DaemonSetLister   appsv1lister.DaemonSetLister
	StatefulSetLister appsv1lister.StatefulSetLister
	JobLister         batchv1lister.JobLister
	SecretLister      corev1lister.SecretLister
	helmworkload      *HelmWorkload
}

func NewWorkLoadTestManager() *WorkloadManager {
	sdk := k8s.NewK8sClientInner()
	helm := k8s.NewHelm(sdk)
	groupApi, err := NewAppGroupApi(sdk)
	if err != nil {
		panic(fmt.Errorf("init appgroup api error: %v", err))
	}
	manager := NewWorkloadManager(groupApi, helm)
	return manager

}

func NewWorkloadManager(groupApi *AppGroupApi, helm *k8s.Helm) *WorkloadManager {
	return &WorkloadManager{
		groupApi: groupApi,
		helm:     helm,
		sdk:      groupApi.sdk,
	}
}

func (d *WorkloadManager) SetGroupLister(lister v1alpha1Lister.AppGroupLister) {
	d.groupApi.SetLister(lister)
	d.AppGroupLister = lister
}

func NewWorkloadManagerWithLister(groupApi *AppGroupApi, helm *k8s.Helm,
	groupLister v1alpha1Lister.AppGroupLister,
	deployment appsv1lister.DeploymentLister,
	statefulset appsv1lister.StatefulSetLister,
	daemonset appsv1lister.DaemonSetLister,
	job batchv1lister.JobLister,
	secret corev1lister.SecretLister,
) *WorkloadManager {
	return &WorkloadManager{
		groupApi:          groupApi,
		helm:              helm,
		sdk:               groupApi.sdk,
		AppGroupLister:    groupLister,
		DeploymentLister:  deployment,
		StatefulSetLister: statefulset,
		DaemonSetLister:   daemonset,
		JobLister:         job,
		SecretLister:      secret,
	}
}

func (d *WorkloadManager) GetGroupName(ds WorkloadWrapperInterface) string {
	name := ds.Name()
	parentName := ds.ReleaseName()
	if parentName != "" {
		name = parentName
	}
	lb := ds.Labels()
	managerBy, ok := lb["app.kubernetes.io/managed-by"]
	if ok {
		if managerBy != "Helm" {
			deployment, err := d.DeploymentLister.Deployments(ds.Namespace()).Get(managerBy)
			if err != nil {
				slog.Error("get deployment error", "error", err)
				return name
			}
			releaseName, ok := deployment.Labels["app.kubernetes.io/instance"]
			if ok {
				name = releaseName
			}
		}
	}
	anno := ds.Annotations()
	if releaseName, ok := anno["meta.helm.sh/release-name"]; ok {
		name = releaseName
	}
	return name
}

func (d *WorkloadManager) GetFromRO(kind, namespace, name string) (WorkloadWrapperInterface, error) {

	switch kind {
	case "Deployment":
		deployment, err := d.DeploymentLister.Deployments(namespace).Get(name)
		if err != nil {
			return nil, err
		}
		return NewWorkloadWrapper(deployment), nil
	case "StatefulSet":
		statefulset, err := d.StatefulSetLister.StatefulSets(namespace).Get(name)
		if err != nil {
			return nil, err
		}
		return NewWorkloadWrapper(statefulset), nil
	case "DaemonSet":
		daemonset, err := d.DaemonSetLister.DaemonSets(namespace).Get(name)
		if err != nil {
			return nil, err
		}
		return NewWorkloadWrapper(daemonset), nil
	case "Job":
		job, err := d.JobLister.Jobs(namespace).Get(name)
		if err != nil {
			return nil, err
		}
		return NewWorkloadWrapper(job), nil
	default:

		return nil, fmt.Errorf("unknown kind: %s", kind)
	}
}

func (d *WorkloadManager) GetSecretFromRO(kind, namespace, name string) (*corev1.Secret, error) {
	return d.SecretLister.Secrets(namespace).Get(name)
}

func (d *WorkloadManager) GetAppGroupFromRO(namespace, name string) (*v1alpha1.AppGroup, error) {
	return d.AppGroupLister.AppGroups(namespace).Get(name)
}

func (d *WorkloadManager) GetAppGroupWrapper(ds WorkloadWrapperInterface) *appgroupWrapper {
	name := d.GetGroupName(ds)
	spec := appv1.AppGroupSpec{
		Type:        appv1.CUSTOM,
		Identifie:   ds.Identifie(),
		Version:     "",
		Logo:        "",
		Description: "",
		ZpkUrl:      "",
		Suffix:      ds.Labels()["w7.cc/suffix"],
		Title:       ds.Title(),
	}

	return d.groupApi.GetAppGroupWrapper(ds.Namespace(), name, spec)
}

func (d *WorkloadManager) HandleQueue(key interface{}) error {
	evt, err := ParseEvent(key)
	if err != nil {
		slog.Error("parse key error", "error", err)
		return nil
	}

	if evt.Namespace != "default" {
		return nil
	}
	// if (ev)
	if evt.Kind == "Event" {
		return nil
	}
	if evt.Kind == "Secret" {
		secret, err := d.GetSecretFromRO(evt.Kind, evt.Namespace, evt.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				slog.Error("get from ro error", "error", err)
				// return d.HandleSecret(secret, true)
			}
			// slog.Error("get from ro error", "error", err)
			return nil
		}
		return d.HandleSecret(secret, false)
	}
	if evt.Kind == "AppGroup" {
		group, err := d.GetAppGroupFromRO(evt.Namespace, evt.Name)
		if err != nil {
			slog.Error("get from ro error", "error", err)
			return nil
		}
		return d.HandleAppGroup(group, false, evt.IsInit)
	}

	ds, err := d.GetFromRO(evt.Kind, evt.Namespace, evt.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			workload := NewWorkloadFromEvent(evt)
			return d.HandleWorkload(NewWorkloadWrapper(workload), true)
		}
		slog.Error("get from ro error", "error", err)
		return nil
	}
	return d.HandleWorkload(ds, evt.EventType == "delete")
}

func (d *WorkloadManager) HandleWorkload(ds WorkloadWrapperInterface, delete bool) error {
	slog.Debug("handle workload", "workload", ds.Name())
	group := d.GetAppGroupWrapper(ds)
	// if !group.exists && ds.IsHelm() {
	// 	return nil
	// }
	itemStatus := ds.ToItemStatus()
	if delete {
		if group.IsExists() {
			group.RemoveStatusItem(itemStatus)
			_, err := d.groupApi.Persist(group)
			if err != nil {
				slog.Error("delete group error", "error", err)
				return err
			}
			return nil
		}
		return nil
	}
	if itemStatus.Kind == "Job" {
		if !group.exists {
			return nil
		}
		group.FixDeployItem(itemStatus)
		if group.changed {
			_, err := d.groupApi.Persist(group)
			if err != nil {
				slog.Error("update group error", "error", err)
				return err
			}
		}
	} else {
		// 如果是helm应用 helmworkload 需要预先新建应用组appgroup
		if ds.IsHelm() && !group.exists {
			return nil
		}
		lb := ds.Labels()
		//兼容 operator 管理的应用 同步到appgroup
		managerBy, ok := lb["app.kubernetes.io/managed-by"]
		if ok && managerBy != "Helm" && !group.exists {
			return nil
		}

		group.AddStatusItem(itemStatus)
		_, err := d.groupApi.Persist(group)
		if err != nil {
			slog.Error("update group error", "error", err)
			return err
		}
	}
	if itemStatus.Kind != "Job" {
		d.fixSvc(ds, delete)
	}
	return nil

}

// secret 只修改spec 中logo isHelm 字段
func (d *WorkloadManager) HandleSecret(secret *corev1.Secret, delete bool) error {
	if secret == nil {
		return fmt.Errorf("secret is nil")
	}
	// if helper.IsChildAgent() && helper.IsK3kVirtual() && secret.Type != "helm.sh/release.v1" {
	// 	k3k.SyncSecretHttp(secret)
	// }
	// if !helper.IsChildAgent() {// 因为只watch default namespace 使用webhook 处理
	// 	k3k.SyncToChildSecret(secret) // 主集群同步到子集群的secret
	// }
	if secret.Type != "helm.sh/release.v1" {
		return nil
	}
	if delete {
		slog.Debug("delete secret", slog.String("name", secret.Name))
		return nil
	}
	// 全量同步helm 应用组信息
	err := d.helmworkload.Sync()
	if err != nil {
		slog.Error("helm sync error", "error", err)
		return err
	}
	return err

}

func (d *WorkloadManager) cleanGroupChildren(group *v1alpha1.AppGroup) error {
	req, err := labels.NewRequirement("w7.cc/parent", selection.In, []string{group.Name})
	if err != nil {
		slog.Error("failed to create requirement", slog.String("error", err.Error()))
		return err
	}
	mlabel := labels.NewSelector().Add(*req)
	children, err := d.groupApi.lister.List(mlabel)
	if err != nil {
		slog.Error("failed to list children", slog.String("error", err.Error()))
		return err
	}

	hasChildren := false
	for _, child := range children {
		if child.DeletionTimestamp == nil {
			hasChildren = true
			d.groupApi.DeleteAppGroup(child.Namespace, child.Name)
		}
	}
	if !hasChildren {
		slog.Debug("no children need to delete", "group name", group.Name, "namespace", group.Namespace)
		group.Finalizers = nil
		_, err := d.groupApi.UpdateAppGroup(group.Namespace, group)
		if err != nil {
			slog.Error("update group error", "error", err)
			return err
		}
	}

	return nil
}

func (d *WorkloadManager) HandleAppGroup(group *v1alpha1.AppGroup, delete bool, isInit bool) error {
	if group.DeletionTimestamp != nil {
		d.cleanAppGroup(group)
		parentName, isChild := group.Labels["w7.cc/parent"]
		if isChild {
			group.Finalizers = nil
			_, err := d.groupApi.UpdateAppGroup(group.Namespace, group)
			if err != nil {
				slog.Error("update group error", "error", err)
				return err
			}
			parentGroup, err := d.GetAppGroupFromRO(group.Namespace, group.Name)
			if err != nil {
				slog.Debug("cannot find parent group", slog.String("parentName", parentName), slog.String("error", err.Error()))
				return nil
			}
			return d.cleanGroupChildren(parentGroup)
		} else {
			NotifyDeleted(group)
			return d.cleanGroupChildren(group)
		}
	}
	// 如果没有删除 且 没有finalizer 则添加finalizer
	// if group.DeletionTimestamp == nil && (group.Finalizers == nil || len(group.Finalizers) == 0) {
	// 	group.Finalizers = []string{"appgroup.w7.cc/finalizer"}
	// 	_, err := d.groupApi.UpdateAppGroup(group.Namespace, group)
	// 	if err != nil {
	// 		slog.Error("update group error", "error", err)
	// 		return err
	// 	}
	// 	return nil
	// }
	changed := false
	if group.Spec.Suffix == "" {
		group.Spec.Suffix = group.Name
		changed = true
	}

	if group.Status.Ready {
		if NeedNotifyInstalled(group) {
			err := NotifyInstalled(group)
			if err != nil && err.Error() == "appgroup notify error" {
				// d.groupApi.DeleteAppGroup(group.Namespace, group.Name)
				return nil
			}
			if err == nil {
				group.Annotations["w7.cc/notify-installed"] = "true"
				changed = true
			}
		}

	}
	if changed {
		_, err := d.groupApi.UpdateAppGroup(group.Namespace, group)
		if err != nil {
			slog.Error("update group error", "error", err)
			return err
		}
	}

	if isInit {
		downStatic(group)
	}

	return nil
}

func (d *WorkloadManager) cleanHelm(group *v1alpha1.AppGroup, createJob bool) error {
	_, err := d.helm.UnInstall(group.Name, group.Namespace)
	if err != nil {
		slog.Error("failed helm uninstall to uninstall app", slog.String("error", err.Error()))
	}
	if !createJob {
		return err
	}
	uninstallJob := ToUninstallTmpJob(group, "uninstall")
	if uninstallJob != nil {
		_, err := d.sdk.ClientSet.BatchV1().Jobs(group.Namespace).Create(context.TODO(), uninstallJob, metav1.CreateOptions{})
		if err != nil {
			slog.Error("failed to create delete job", slog.String("error", err.Error()))
		}
	}
	return err

}

func (d *WorkloadManager) cleanAppGroup(group *appv1.AppGroup) {

	defer func() {
		slog.Info("start delete helm")
		d.cleanHelm(group, true)
	}()
	defer helper.CleanStaticDir(group.Name)

	if group.Spec.IsHelm {
		// vm metrics opertor 会监听资源删除 如果helm uninstall 最后执行 会导致资源无法删除 helm清理不干净
		slog.Info("start delete helm ")
		d.cleanHelm(group, false)
	}

	slog.Info("start delete workload")
	for _, workload := range group.Status.Items {
		item := workload.Kind

		switch item {
		case "Deployment":
			err := d.sdk.ClientSet.AppsV1().Deployments(group.Namespace).Delete(context.TODO(), workload.Name, metav1.DeleteOptions{})
			if err != nil {
				slog.Error("failed to delete deployment", slog.String("error", err.Error()))
			}
		case "DaemonSet":
			err := d.sdk.ClientSet.AppsV1().DaemonSets(group.Namespace).Delete(context.TODO(), workload.Name, metav1.DeleteOptions{})
			if err != nil {
				slog.Error("failed to delete daemonset", slog.String("error", err.Error()))
			}
		case "StatefulSet":
			err := d.sdk.ClientSet.AppsV1().StatefulSets(group.Namespace).Delete(context.TODO(), workload.Name, metav1.DeleteOptions{})
			if err != nil {
				slog.Error("failed to delete statefulset", slog.String("error", err.Error()))
			}
		case "CronJob":
			err := d.sdk.ClientSet.BatchV1beta1().CronJobs(group.Namespace).Delete(context.TODO(), workload.Name, metav1.DeleteOptions{})
			if err != nil {
				slog.Error("failed to delete cronjob", slog.String("error", err.Error()))
			}
		case "Job":
			err := d.sdk.ClientSet.BatchV1().Jobs(group.Namespace).Delete(context.TODO(), workload.Name, metav1.DeleteOptions{})
			if err != nil {
				slog.Error("failed to delete job", slog.String("error", err.Error()))
			}
		}
		//清理service
		err := d.sdk.ClientSet.CoreV1().Services(group.Namespace).Delete(context.TODO(), workload.Name, metav1.DeleteOptions{})
		if err != nil {
			slog.Error("failed to delete service", slog.String("error", err.Error()))
		}
		err = d.sdk.ClientSet.CoreV1().Services(group.Namespace).Delete(context.TODO(), workload.Name+"-lb", metav1.DeleteOptions{})
		if err != nil {
			slog.Error("failed to delete servicelb", slog.String("error", err.Error()))
		}
	}

	jobs, err := d.sdk.ClientSet.BatchV1().Jobs(group.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("w7.cc/group-name=%s", group.Name)})
	if err != nil {
		slog.Error("failed to list jobs", slog.String("error", err.Error()))
		// return
	}
	if jobs != nil {
		for _, job := range jobs.Items {
			d.sdk.ClientSet.BatchV1().Jobs(group.Namespace).Delete(context.TODO(), job.Name, metav1.DeleteOptions{})
		}
	}

	cronJobs, err := d.sdk.ClientSet.BatchV1beta1().CronJobs(group.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("w7.cc/group-name=%s", group.Name)})
	if err != nil {
		slog.Error("failed to list cronjobs", slog.String("error", err.Error()))
		// return
	}
	if cronJobs != nil {
		for _, cronJob := range cronJobs.Items {
			d.sdk.ClientSet.BatchV1beta1().CronJobs(group.Namespace).Delete(context.TODO(), cronJob.Name, metav1.DeleteOptions{})
		}
	}

	slog.Info("start delete ingress")
	ingressLabel := fmt.Sprintf("group=%s", group.Name)
	ingress, err := d.sdk.ClientSet.NetworkingV1().Ingresses(group.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: ingressLabel})
	if err != nil {
		slog.Error("failed to list ingress", slog.String("error", err.Error()))
		return
	}
	for _, item := range ingress.Items {
		slog.Info("delete ingress", slog.String("ingressName", item.Name))
		d.sdk.ClientSet.NetworkingV1().Ingresses(group.Namespace).Delete(context.TODO(), item.Name, metav1.DeleteOptions{})
	}

	sigClient, err := d.sdk.ToSigClient()
	if err != nil {
		slog.Error("failed to get sig client", slog.String("error", err.Error()))
		return
	}
	mc := &microapp.MicroApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      group.Name,
			Namespace: group.Name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "MicroApp",
			APIVersion: "microapp.w7.cc/v1alpha1",
		},
	}
	err = sigClient.Delete(d.sdk.Ctx, mc, &sigclient.DeleteOptions{})
	if err != nil {
		slog.Error("failed to delete microapp", slog.String("error", err.Error()))
	}
}
