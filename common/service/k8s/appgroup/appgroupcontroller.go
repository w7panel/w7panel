package appgroup

import (
	"context"
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	applicationversiond "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/appgroup/clientset/versioned"
	appInformer "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/appgroup/informers/externalversions"
	v1alpha1Lister "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/appgroup/listers/appgroup/v1alpha1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appsv1lister "k8s.io/client-go/listers/apps/v1"
	batchv1lister "k8s.io/client-go/listers/batch/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type AppController struct {
	*k8s.Sdk
	KubeInformerFactory        informers.SharedInformerFactory
	ApplicationInformerFactory appInformer.SharedInformerFactory
	HelmApi                    *k8s.Helm
	AppGroupClientSet          *applicationversiond.Clientset
	AppGroupApi                *AppGroupApi
	WorkloadManager            *WorkloadManager
	AppGroupInformer           cache.SharedIndexInformer
	AppGroupLister             v1alpha1Lister.AppGroupLister
	DeploymentInformer         cache.SharedIndexInformer
	DaemonSetInformer          cache.SharedIndexInformer
	StatefulSetInformer        cache.SharedIndexInformer
	JobInformer                cache.SharedIndexInformer
	SecretInformer             cache.SharedIndexInformer
	EventInformer              cache.SharedIndexInformer
	DeploymentLister           appsv1lister.DeploymentLister
	DaemonSetLister            appsv1lister.DaemonSetLister
	StatefulSetLister          appsv1lister.StatefulSetLister
	JobLister                  batchv1lister.JobLister
	SecretLister               corev1lister.SecretLister
	queue                      *EventQueue
	helmworkload               *HelmWorkload
}

func Watch() {
	sdk := k8s.NewK8sClient().GetSdk()
	appController, err := NewAppController(sdk)
	if err != nil {
		return
	}
	err = appController.Start()
	if err != nil {
		return
	}
}

func NewAppController(sdk *k8s.Sdk) (*AppController, error) {
	kubeInformerFactory := informers.NewSharedInformerFactory(sdk.ClientSet, 0)
	config, err := sdk.ToRESTConfig()
	if err != nil {
		slog.Error("failed to get rest config", slog.String("error", err.Error()))
		return nil, err
	}
	appClientset, err := applicationversiond.NewForConfig(config)
	if err != nil {
		slog.Error("failed to get app clientset", slog.String("error", err.Error()))
		return nil, err
	}
	// nodeLister := kubeInformerFactory.Core().V1().Nodes().Lister()
	applicationInformerFactory := appInformer.NewSharedInformerFactory(appClientset, 0)
	helmApi := k8s.NewHelm(sdk)

	gctrl := &AppController{
		Sdk:                        sdk,
		KubeInformerFactory:        kubeInformerFactory,
		HelmApi:                    helmApi,
		ApplicationInformerFactory: applicationInformerFactory,
		AppGroupClientSet:          appClientset,
	}

	groupApi, err := NewAppGroupApi(sdk)
	if err != nil {
		slog.Error("failed to get app api", slog.String("error", err.Error()))
		return nil, err
	}
	gctrl.AppGroupApi = groupApi

	gctrl.WatchDeployment()
	gctrl.WatchDaemonset()
	gctrl.WatchStatefulSets()
	gctrl.WatchJob()
	gctrl.WatchSecret()
	gctrl.WatchAppGroup()
	// gctrl.WatchEvent()

	manager := NewWorkloadManager(groupApi, helmApi)
	// helmWorkload := NewHelmWorkload(helmApi, groupApi.lister, gctrl.queue)
	manager.DeploymentLister = gctrl.DeploymentLister
	manager.DaemonSetLister = gctrl.DaemonSetLister
	manager.StatefulSetLister = gctrl.StatefulSetLister
	manager.JobLister = gctrl.JobLister
	groupApi.SetLister(gctrl.AppGroupLister)
	gctrl.WorkloadManager = manager
	manager.SetGroupLister(gctrl.AppGroupLister)
	manager.SecretLister = gctrl.SecretLister

	queue := NewDefaultEventQueue(gctrl.WorkloadManager.HandleQueue)
	gctrl.queue = queue
	helmWorkload := NewHelmWorkload(helmApi, gctrl.AppGroupLister, gctrl.queue, gctrl.AppGroupApi)
	gctrl.helmworkload = helmWorkload
	gctrl.WorkloadManager.helmworkload = helmWorkload
	// gctrl.queue = queue

	gctrl.DeploymentInformer.AddEventHandler(gctrl.queue)
	gctrl.DaemonSetInformer.AddEventHandler(gctrl.queue)
	gctrl.StatefulSetInformer.AddEventHandler(gctrl.queue)
	gctrl.JobInformer.AddEventHandler(gctrl.queue)
	gctrl.SecretInformer.AddEventHandler(gctrl.queue)
	gctrl.AppGroupInformer.AddEventHandler(gctrl.queue)
	// gctrl.EventInformer.AddEventHandler(gctrl.queue)

	gctrl.helmworkload = helmWorkload

	return gctrl, nil
}

func (s *AppController) Start() error {
	// informer := s.WatchStatefulSets()
	// dpInformer := s.WatchDeployment()
	// dsInformer := s.WatchDaemonset()
	// jobInformer := s.WatchJob()
	// secretInformer := s.WatchSecret()
	// s.WatchAppGroup()
	// ingressInformer := s.WatchIngressEvents()
	// configMapInformer := s.WatchConfigmap()
	stopCh := make(chan struct{})
	defer close(stopCh)

	s.KubeInformerFactory.Start(stopCh)
	s.ApplicationInformerFactory.Start(stopCh)

	informers := []cache.InformerSynced{
		s.AppGroupInformer.HasSynced,
		s.DeploymentInformer.HasSynced,
		s.DaemonSetInformer.HasSynced,
		s.StatefulSetInformer.HasSynced,
		s.JobInformer.HasSynced,
		s.SecretInformer.HasSynced,
		// s.EventInformer.HasSynced,
		// ingressInformer.HasSynced,
		// configMapInformer.HasSynced,
	}

	if !helper.WaitForNamedCacheSync("groupcontroller", stopCh, informers...) {
		slog.Error("Failed to sync cache")
		return nil
	}
	go s.queue.Run(5, stopCh)
	s.helmworkload.Sync()

	// 启动定时任务

	<-stopCh
	return nil
}

func (s *AppController) WatchDaemonset() {
	informer := s.KubeInformerFactory.Apps().V1().DaemonSets()
	s.DaemonSetInformer = informer.Informer()
	s.DaemonSetLister = appsv1lister.NewDaemonSetLister(s.DaemonSetInformer.GetIndexer())
	// s.DaemonSetInformer.AddEventHandler(s.WorkloadManager)
}

func (s *AppController) WatchDeployment() {
	informer := s.KubeInformerFactory.Apps().V1().Deployments()
	s.DeploymentInformer = informer.Informer()
	s.DeploymentLister = appsv1lister.NewDeploymentLister(s.DeploymentInformer.GetIndexer())
	// s.DeploymentInformer.AddEventHandler(s.WorkloadManager)
	// return informer
}

func (s *AppController) WatchStatefulSets() {
	informer := s.KubeInformerFactory.Apps().V1().StatefulSets()
	s.StatefulSetInformer = informer.Informer()
	s.StatefulSetLister = appsv1lister.NewStatefulSetLister(s.StatefulSetInformer.GetIndexer())
	// s.StatefulSetInformer.AddEventHandler(s.WorkloadManager)

}

func (s *AppController) WatchJob() {
	informer := s.KubeInformerFactory.Batch().V1().Jobs()
	s.JobInformer = informer.Informer()
	s.JobLister = batchv1lister.NewJobLister(s.JobInformer.GetIndexer())
	// s.JobInformer.AddEventHandler(s.WorkloadManager)
}

func (s *AppController) WatchSecret() {
	informer := s.KubeInformerFactory.Core().V1().Secrets()
	s.SecretInformer = informer.Informer()
	s.SecretLister = corev1lister.NewSecretLister(s.SecretInformer.GetIndexer())

}

func (s *AppController) WatchAppGroup() {
	informer := s.ApplicationInformerFactory.Appgroup().V1alpha1().AppGroups()
	s.AppGroupInformer = informer.Informer()
	s.AppGroupLister = v1alpha1Lister.NewAppGroupLister(s.AppGroupInformer.GetIndexer())

}
func (s *AppController) WatchEvent() {
	informer := s.KubeInformerFactory.Core().V1().Events()
	s.EventInformer = informer.Informer()
	s.EventInformer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInit bool) {
			if isInit {
				return
			}
			v1Event, ok := obj.(*v1.Event)
			if !ok {
				slog.Error("Failed to assert object to *v1.Event in AddFunc", "object", obj)
				return
			}
			if v1Event.InvolvedObject.Kind == "Cluster" && v1Event.InvolvedObject.APIVersion == "apps.kubeblocks.io/v1alpha1" &&
				v1Event.Reason == "DeletingCR" {
				err := s.AppGroupApi.DeleteAppGroup(v1Event.InvolvedObject.Namespace, v1Event.InvolvedObject.Name)
				if err != nil {
					slog.Error("Failed to delete app group", "error", err)
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {

		},
		DeleteFunc: func(obj interface{}) {
			// Handle the case where the object is of type cache.DeletedFinalStateUnknown

		},
	})

}
func (w *AppController) changeSslRediret(ingress *networkingv1.Ingress) error {

	// webhook 处理逻辑
	return nil
	// if !helper.IsChildAgent() {
	// 	sslredirect, ok := ingress.Annotations["w7.cc/ssl-redirect"]
	// 	if !ok {
	// 		return nil
	// 	}
	// 	if ingress.Annotations["higress.io/ssl-redirect"] != sslredirect {
	// 		ingress.Annotations["higress.io/ssl-redirect"] = sslredirect
	// 		_, err := w.ClientSet.NetworkingV1().Ingresses(ingress.Namespace).Update(context.TODO(), ingress, metav1.UpdateOptions{})
	// 		if err != nil {
	// 			slog.Error("Failed to update ingress", "error", err)
	// 			return err
	// 		}
	// 	}
	// }
	// return nil
}

func (w *AppController) WatchIngressEvents() cache.SharedIndexInformer {
	// slog.Debug("watch ingress events")
	informer := w.KubeInformerFactory.Networking().V1().Ingresses().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, init bool) {
			if !init {
				ingress, ok := obj.(*networkingv1.Ingress)
				if !ok {
					return
				}
				if helper.IsChildAgent() {
					slog.Debug("Ingress created: %s/%s\n", ingress.Namespace, ingress.Name)
					k3k.SyncIngressHttp(ingress)
				}
				err := w.changeSslRediret(ingress)
				if err != nil {
					slog.Error("Failed to change ssl redirect", "error", err)
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			ingress, ok := newObj.(*networkingv1.Ingress)
			if !ok {
				return
			}
			if helper.IsChildAgent() {
				slog.Debug("Ingress updated: %s/%s\n", ingress.Namespace, ingress.Name)
				k3k.SyncIngressHttp(ingress)
			}
			err := w.changeSslRediret(ingress)
			if err != nil {
				slog.Error("Failed to change ssl redirect", "error", err)
			}
		},
		DeleteFunc: func(obj interface{}) {

			ingress := obj.(*networkingv1.Ingress)
			slog.Debug("Ingress deleted: %s/%s\n", ingress.Namespace, ingress.Name)
			clientset := w.ClientSet
			// 清理与 Ingress 相关的 Secret
			for _, tls := range ingress.Spec.TLS {
				secretName := tls.SecretName
				if secretName != "" {
					// 检查是否有其他 Ingress 引用了该 Secret
					if isSecretReferencedByOtherIngress(clientset, ingress, secretName) {
						fmt.Printf("Secret %s/%s is still referenced by other Ingresses, skipping deletion\n", ingress.Namespace, secretName)
						continue
					}
					// 删除 Secret
					err := clientset.CoreV1().Secrets(ingress.Namespace).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
					if err != nil {
						fmt.Printf("Failed to delete Secret %s/%s: %v\n", ingress.Namespace, secretName, err)
					}
				}
			}
			if helper.IsChildAgent() {
				k3k.SyncIngressHttp(ingress)
			}
		},
	})

	return informer
}

func isSecretReferencedByOtherIngress(clientset *kubernetes.Clientset, deletedIngress *networkingv1.Ingress, secretName string) bool {
	ingresses, err := clientset.NetworkingV1().Ingresses(deletedIngress.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list Ingresses: %v\n", err)
		return false
	}

	for _, ingress := range ingresses.Items {
		// 跳过已删除的 Ingress
		if ingress.Name == deletedIngress.Name && ingress.Namespace == deletedIngress.Namespace {
			continue
		}

		// 检查当前 Ingress 是否引用了指定的 Secret
		for _, tls := range ingress.Spec.TLS {
			if tls.SecretName == secretName {
				return true
			}
		}
	}

	return false
}

func (w *AppController) WatchConfigmap() cache.SharedIndexInformer {
	informer := w.KubeInformerFactory.Core().V1().ConfigMaps().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, init bool) {
			if !init {
				configmap, ok := obj.(*v1.ConfigMap)
				if !ok {
					return
				}
				if helper.IsChildAgent() {
					k3k.SyncConfigmapHttp(configmap)
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			configmap, ok := newObj.(*v1.ConfigMap)
			if !ok {
				return
			}
			if helper.IsChildAgent() {
				k3k.SyncConfigmapHttp(configmap)
			}
		},
		DeleteFunc: func(obj interface{}) {
			configmap, ok := obj.(*v1.ConfigMap)
			if !ok {
				return
			}
			if helper.IsChildAgent() {
				k3k.SyncConfigmapHttp(configmap)
			}
		},
	})
	return informer
}
