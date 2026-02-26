package appgroup

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	appv1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	clientset "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/appgroup/clientset/versioned"
	v1alpha1Lister "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/appgroup/listers/appgroup/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetAppgroupUseSdk(name, namespace string, sdk *k8s.Sdk) (*appv1.AppGroup, error) {

	client, err := sdk.ToSigClient()
	if err != nil {
		return nil, err
	}

	return GetAppgroup(name, namespace, client)
}
func GetAppgroup(name, namespace string, client sigclient.Client) (*appv1.AppGroup, error) {
	group := &appv1.AppGroup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "appgroup.w7.cc/v1alpha1",
			Kind:       "AppGroup",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := client.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, group, &sigclient.GetOptions{})
	if err != nil {
		return nil, err
	}
	return group, nil
}

type AppGroupApi struct {
	sdk       *k8s.Sdk
	clientset *clientset.Clientset
	lister    v1alpha1Lister.AppGroupLister
}

func NewAppGroupApi(sdk *k8s.Sdk) (*AppGroupApi, error) {
	config, err := sdk.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &AppGroupApi{
		sdk:       sdk,
		clientset: clientset,
	}, nil
}

func (a *AppGroupApi) SetLister(lister v1alpha1Lister.AppGroupLister) {
	a.lister = lister
}

func (a *AppGroupApi) GetAppGroupNamespaced(namespace string) v1alpha1Lister.AppGroupNamespaceLister {
	return a.lister.AppGroups(namespace)
}

func (a *AppGroupApi) GetAppGroupList(namespace string) (*appv1.AppGroupList, error) {
	// return a.lister.AppGroups(namespace).List(la)
	return a.clientset.AppgroupV1alpha1().AppGroups(namespace).List(a.sdk.Ctx, metav1.ListOptions{})
}

func (a *AppGroupApi) GetAppGroupListByIdentifie(namespace string, identifie string) (*appv1.AppGroupList, error) {
	// return a.lister.AppGroups(namespace).List(la)
	return a.clientset.AppgroupV1alpha1().AppGroups(namespace).List(a.sdk.Ctx, metav1.ListOptions{LabelSelector: "w7.cc/identifie=" + identifie})
}

func (a *AppGroupApi) GetAppGroup(namespace string, name string) (*appv1.AppGroup, error) {
	return a.clientset.AppgroupV1alpha1().AppGroups(namespace).Get(a.sdk.Ctx, name, metav1.GetOptions{})
}

func (a *AppGroupApi) UpdateAppGroup(namespace string, group *appv1.AppGroup) (*appv1.AppGroup, error) {
	a.filterEmpty(group)
	return a.clientset.AppgroupV1alpha1().AppGroups(namespace).Update(a.sdk.Ctx, group, metav1.UpdateOptions{})
}

func (a *AppGroupApi) CreateGroup(namespace string, group *appv1.AppGroup) (*appv1.AppGroup, error) {
	a.filterEmpty(group)
	return a.clientset.AppgroupV1alpha1().AppGroups(namespace).Create(a.sdk.Ctx, group, metav1.CreateOptions{})
}

func (a *AppGroupApi) filterEmpty(app *appv1.AppGroup) {
	// if app.Spec.Domains == nil || len(app.Spec.Domains) == 0 {
	// 	app.Spec.Domains = []string{""}
	// }

	if (app.Status.Items == nil) || (len(app.Status.Items) == 0) {
		app.Status.Items = make([]appv1.AppGroupItemStatus, 0)
	}
	if (app.Status.DeployItems == nil) || (len(app.Status.DeployItems) == 0) {
		app.Status.DeployItems = make([]appv1.DeployItem, 0)
	}

}

func (a *AppGroupApi) GetAppGroupRO(namespace string, name string) (*appv1.AppGroup, error) {
	group, err := a.lister.AppGroups(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	return group.DeepCopy(), nil
}

/*
*

	DeleteAppGroup 删除应用
*/
func (a *AppGroupApi) DeleteAppGroup(namespace string, name string) error {
	return a.clientset.AppgroupV1alpha1().AppGroups(namespace).Delete(a.sdk.Ctx, name, metav1.DeleteOptions{})
}

func (a *AppGroupApi) UninstallStorePanel(ns string) error {
	groups, err := a.GetAppGroupList(ns)
	if err != nil {
		slog.Error("GetAppGroupList error", "error", err)
		return err
	}
	for _, group := range groups.Items {
		if !group.Spec.IsHelm && !group.DeletionTimestamp.IsZero() {
			continue
		}
		if group.Spec.Identifie == "w7panel" && group.Name != "w7panel" {
			err = a.DeleteAppGroup(group.Namespace, group.Name)
			if err != nil {
				slog.Error("DeleteAppGroup error", "error", err)
				return err
			}
		}
	}
	return nil
}

func (a *AppGroupApi) Persist(wapper *appgroupWrapper) (*appv1.AppGroup, error) {
	if wapper.exists {
		if wapper.IsEmpty() {
			if !wapper.Spec.IsHelm && wapper.IsDeleteEnabled() {
				return nil, a.DeleteAppGroup(wapper.Namespace, wapper.Name)
			}
			// return nil, nil
		}
		if wapper.changed {
			_, err := a.UpdateAppGroup(wapper.Namespace, wapper.AppGroup)
			if err != nil {
				return nil, err
			}
			if wapper.parent != nil {
				a.Persist(wapper.parent)
			}
		}
		return wapper.AppGroup, nil
	} else {
		return a.CreateGroup(wapper.Namespace, wapper.AppGroup)
	}
}

func (a *AppGroupApi) getAppGroupUseReadOnly(namespace string, name string) (*appv1.AppGroup, error) {
	if a.lister != nil {
		return a.GetAppGroupRO(namespace, name)
	}
	return a.GetAppGroup(namespace, name)
}

func (a *AppGroupApi) GetAppGroupWrapper(namespace string, name string, spec appv1.AppGroupSpec) *appgroupWrapper {
	group, err := a.getAppGroupUseReadOnly(namespace, name)
	exites := true
	if err != nil {
		if errors.IsNotFound(err) {
			exites = false
			group = &appv1.AppGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Finalizers: []string{"appgroup.w7.cc/finalizers"},
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "AppGroup",
					APIVersion: appv1.SchemeGroupVersion.String(),
				},
				Status: appv1.AppGroupStatus{
					Items: make([]appv1.AppGroupItemStatus, 0),
				},
				Spec: spec,
			}
		}
	}
	if group == nil {
		return nil
	}
	wapper := NewAppGroupWrapper(group, exites)
	parentName, ok := group.Labels["w7.cc/parent"]
	if ok {
		parent, err := a.getAppGroupUseReadOnly(namespace, parentName)
		if err != nil {
			slog.Error("Get parent err", "err", err)
		}
		if parent != nil {
			wapper.SetParent(NewAppGroupWrapper(parent, true))
		}
	}
	return wapper
}
