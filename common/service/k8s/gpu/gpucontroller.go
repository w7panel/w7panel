package gpu

import (
	"context"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	gpuclassv1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/gpuclass/v1alpha1"
	gpuclassclientset "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/gpuclass/clientset/versioned"
	gpuinformer "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/gpuclass/informers/externalversions"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types2 "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type GpuController struct {
	*k8s.Sdk
	factory         informers.SharedInformerFactory
	gpuClassApi     gpuclassclientset.Interface
	helmApi         *k8s.Helm
	gpuclassFactory gpuinformer.SharedInformerFactory
}

func Watch() {
	gpuController, err := NewGpuController(k8s.NewK8sClient().Sdk)
	if err != nil {
		return
	}
	err = gpuController.Start()
	if err != nil {
		slog.Error("gpu controller start error", "error", err)
	}
}

func NewGpuController(sdk *k8s.Sdk) (*GpuController, error) {
	config, err := sdk.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	gpuClassApi, err := gpuclassclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	helmApi := k8s.NewHelm(sdk)
	return &GpuController{
		Sdk:             sdk,
		factory:         informers.NewSharedInformerFactory(sdk.ClientSet, 0),
		gpuClassApi:     gpuClassApi,
		helmApi:         helmApi,
		gpuclassFactory: gpuinformer.NewSharedInformerFactory(gpuClassApi, 0),
		// gpuClassApi:     gpuClassApi,
		// helmApi:         &k8s.Helm{Sdk: sdk},
	}, nil
}

// func (g *GpuController) Init() error {

// }

func (g *GpuController) Start() error {
	job := g.WatchJob()
	gpuclass := g.WatchGpuClass()
	stopCh := make(chan struct{})
	defer close(stopCh)
	g.factory.Start(stopCh)
	g.gpuclassFactory.Start(stopCh)
	if !helper.WaitForNamedCacheSync("gpucontroller", stopCh, job.HasSynced, gpuclass.HasSynced) {
		slog.Debug("Failed to sync cache")
		return nil
	}
	<-stopCh
	return nil
}

func (g *GpuController) WatchJob() cache.SharedIndexInformer {
	infomer := g.factory.Batch().V1().Jobs().Informer()
	infomer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			job, ok := obj.(*batchv1.Job)
			if ok {
				g.HandleJob(job)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			job, ok := newObj.(*batchv1.Job)
			if ok {
				g.HandleJob(job)
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	})
	return infomer
}
func (g *GpuController) WatchGpuClass() cache.SharedIndexInformer {
	informer := g.gpuclassFactory.Gpuclass().V1alpha1().GpuClasses().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			gpuclass := obj.(*gpuclassv1alpha1.GpuClass)
			g.HandleGpuClass(gpuclass)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			gpuclass := newObj.(*gpuclassv1alpha1.GpuClass)
			g.HandleGpuClass(gpuclass)
		},
		DeleteFunc: func(obj interface{}) {

		},
	})
	return informer
}
func (g *GpuController) HandleGpuClass(gpuclass *gpuclassv1alpha1.GpuClass) error {

	if g.helmApi.Exists(gpuclass.Namespace, "w7.cc/identifie="+NVIDIA_IDENTIFIE) {
		gpuMode := INSTALLED
		return g.PatchMode(gpuclass.Namespace, string(gpuMode), true)
	}
	if g.helmApi.Exists(gpuclass.Namespace, "w7.cc/identifie="+HAMI_IDENTIFIE) {
		hamiMode := INSTALLED
		return g.PatchMode(gpuclass.Namespace, string(hamiMode), false)
	}
	return nil
}

func (g *GpuController) PatchMode(namespace, mode string, isGpuOperator bool) error {
	patchData := []byte(`{"spec":{"hamiMode": "` + mode + `"}}`)
	if isGpuOperator {
		patchData = []byte(`{"spec":{"gpuOperatorMode": "` + mode + `"}}`)

	}
	_, err := g.gpuClassApi.GpuclassV1alpha1().GpuClasses(namespace).Patch(context.TODO(), NVIDIA_CLASS_NAME,
		types2.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		slog.Error("patch nvidia-gpuoperator error", "error", err)
	}
	return err
}
func (g *GpuController) HandleJob(job *batchv1.Job) error {

	if job.Annotations["w7.cc/helm-install"] != "true" {
		return nil
	}

	if job.Annotations["w7.cc/identifie"] == NVIDIA_IDENTIFIE {
		gpuMode := UNINSTALL
		if g.helmApi.Exists(job.Namespace, "w7.cc/identifie="+NVIDIA_IDENTIFIE) {
			gpuMode = INSTALLED
			return g.PatchMode(job.Namespace, string(gpuMode), true)
		}
		gpuMode = INSTALLING
		if job.Status.Succeeded > 0 {
			gpuMode = INSTALLED
		}
		if job.Status.Failed > 0 {
			gpuMode = UNINSTALL
		}
		return g.PatchMode(job.Namespace, string(gpuMode), true)
	}
	if job.Annotations["w7.cc/identifie"] == HAMI_IDENTIFIE {
		hamiMode := UNINSTALL
		if g.helmApi.Exists(job.Namespace, "w7.cc/identifie="+HAMI_IDENTIFIE) {
			hamiMode = INSTALLED
			return g.PatchMode(job.Namespace, string(hamiMode), false)
		}
		hamiMode = INSTALLING
		if job.Status.Succeeded > 0 {
			hamiMode = INSTALLED
		}
		if job.Status.Failed > 0 {
			hamiMode = UNINSTALL
		}
		return g.PatchMode(job.Namespace, string(hamiMode), false)

	}
	return nil

}
