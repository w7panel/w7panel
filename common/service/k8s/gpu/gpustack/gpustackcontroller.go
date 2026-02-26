package gpustack

import (
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func Watch() error {
	sdk := k8s.NewK8sClient().Sdk
	controller := NewGpuStackController(sdk)
	err := controller.start()
	if err != nil {
		return err
	}
	return nil
}

type gpuStackController struct {
	sdk     *k8s.Sdk
	factory informers.SharedInformerFactory
}

func NewGpuStackController(sdk *k8s.Sdk) *gpuStackController {
	factory := informers.NewSharedInformerFactory(sdk.ClientSet, 0)
	return &gpuStackController{
		sdk:     sdk,
		factory: factory,
	}
}

func (g *gpuStackController) start() error {
	infomer := g.watchPod()
	stopCh := make(chan struct{})
	defer close(stopCh)
	g.factory.Start(stopCh)
	if !helper.WaitForNamedCacheSync("gpustackcontroller", stopCh, infomer.HasSynced) {
		slog.Debug("Failed to sync cache")
		return nil
	}
	<-stopCh
	return nil
}

func (c *gpuStackController) watchPod() cache.SharedIndexInformer {
	infomer := c.factory.Core().V1().Pods().Informer()
	infomer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
		},
		DeleteFunc: func(obj interface{}) {
			slog.Debug("pod deleted on gpustack controller")
			if deleted, ok := obj.(cache.DeletedFinalStateUnknown); ok {
				// Extract the actual object from the DeletedFinalStateUnknown struct
				obj = deleted.Obj
			}
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}
			isGpuWorker, ok := pod.Labels["w7.cc/groupstack-worker"]
			if !ok || isGpuWorker != "true" {
				return
			}
			groupName, ok := pod.Labels["w7.cc/group-name"]
			if !ok {
				return
			}
			// workerName, ok := pod.Annotations["w7.cc/groupstack-worker-name"]
			// if !ok {
			// 	return
			// }
			gpuStackApiUrl := "http://" + groupName + "." + pod.Namespace + ".svc"
			if helper.IsLocalMock() {
				gpuStackApiUrl = "http://gstabc.b2.sz.w7.com"
			}
			userName := "admin"
			password := ""

			deployment, err := c.sdk.ClientSet.AppsV1().Deployments(pod.Namespace).Get(c.sdk.Ctx, groupName, metav1.GetOptions{})
			if err != nil {
				slog.Error("failed to get deployment", "error", err)
				return
			}
			containerEnvs := deployment.Spec.Template.Spec.Containers[0].Env
			for _, env := range containerEnvs {
				if env.Name == "BOOTSTRAP_PASSWORD" {
					password = env.Value
				}
			}
			if password == "" {
				slog.Debug("password is empty")
				return
			}
			api := NewGpuStackApi(gpuStackApiUrl, userName, password)
			err = api.DeleteGpuStackWorkerByName(pod.Name)
			if err != nil {
				slog.Error("failed to delete worker", "error", err)
				return
			}
		},
	})
	return infomer
}
