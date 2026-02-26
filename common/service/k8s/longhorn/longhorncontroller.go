package longhorn

import (
	"log/slog"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	longhornV1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	longhornutil "github.com/longhorn/longhorn-manager/util"
	"k8s.io/client-go/tools/cache"
)

// longhornbeta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"

func WatchLonghorn() error {
	sdk := k8s.NewK8sClientInner()
	initLonghornVolumesConfig(sdk)
	controller, err := Newlonghorncontroller(sdk)
	if err != nil {
		slog.Error("WatchLonghorn error", "err", err)
		return err
	}
	return controller.Start()
}

type longhorncontroller struct {
	sdk            *k8s.Sdk
	factory        *longhornutil.InformerFactories
	longhornclient *longhornclient
	defaultvolume  *defaultvolume
}

func Newlonghorncontroller(sdk *k8s.Sdk) (*longhorncontroller, error) {
	longhornclient, err := NewLonghornClient(sdk)
	if err != nil {
		return nil, err
	}
	// volumes := w.factory.LonghornInformerFactory.Longhorn().V1beta2().Volumes().Informer()
	// longhornutil.New
	defaultvolume := NewDefaultVolume(longhornclient)
	factory := longhornutil.NewInformerFactories("longhorn-system", sdk.ClientSet, longhornclient.client, time.Minute*10)
	return &longhorncontroller{sdk: sdk, longhornclient: longhornclient, factory: factory, defaultvolume: defaultvolume}, nil
}

func (w *longhorncontroller) Start() error {
	longhornNode := w.WatchLonghornNodesEvents()
	// longhornIngress := w.WatchIngressEvents() // 放到appgroupcontroller 里
	replica := w.WatchLonghornReplicaEvents()
	sc := w.WatchLonghornStorageClass()
	nodewc := w.WatchNodeEvents()
	w.UpdateNodeLabel()

	stopCh := make(chan struct{})
	defer close(stopCh)
	w.factory.Start(stopCh)
	if !cache.WaitForNamedCacheSync("longhornController", stopCh,
		longhornNode.HasSynced, replica.HasSynced, sc.HasSynced, nodewc.HasSynced) {
		slog.Debug("Failed to sync cache")
		return nil
	}
	<-stopCh
	return nil
}

func (w *longhorncontroller) WatchLonghornReplicaEvents() cache.SharedIndexInformer {
	// slog.Debug("WatchLonghornVolumesEvents")
	informer := w.factory.LhInformerFactory.Longhorn().V1beta2().Replicas().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInit bool) {
			w.UpdateNodeLabel()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			w.UpdateNodeLabel()
		},
		DeleteFunc: func(obj interface{}) {
			w.UpdateNodeLabel()
		},
	})
	// informer.Run(make(chan struct{}))
	return informer
}

func (w *longhorncontroller) CheckHasDefaultVolume() error {
	return createDefaultVolume()
}

func (w *longhorncontroller) WatchLonghornNodesEvents() cache.SharedIndexInformer {
	// slog.Debug("WatchLonghornNodesEvents")
	informer := w.factory.LhInformerFactory.Longhorn().V1beta2().Nodes().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			slog.Debug("Node added")
			node := obj.(*longhornV1beta2.Node)
			updateAllReplicaCount()
			w.checkDefault(node)
			w.createStorageClass(node)
			createDefaultStorageClassByPvc()

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			slog.Debug("Node updated")
			node := newObj.(*longhornV1beta2.Node)
			updateAllReplicaCount()
			w.checkDefault(node)
			w.createStorageClass(node)
			createDefaultStorageClassByPvc()

		},
		DeleteFunc: func(obj interface{}) {
			slog.Debug("Node deleted")
			updateAllReplicaCount()
		},
	})
	// informer.Run(make(chan struct{}))
	return informer
}

func (w *longhorncontroller) checkDefault(node *longhornV1beta2.Node) error {
	slog.Debug("checkDefault")
	err := checkDefaultDiskTag(node)
	if err != nil {
		slog.Error("checkDefaultDiskTag error", "err", err)
		return err
	}
	err = w.CheckHasDefaultVolume()
	if err != nil {
		slog.Error("CheckHasDefaultVolume error", "err", err)
	}
	return err
}

// 给node 添加storage 标签
func (w *longhorncontroller) UpdateNodeLabel() error {
	return updateNodeLabel()
}

func (w *longhorncontroller) GetLonghornReplicaByNodeName(list *longhornV1beta2.ReplicaList, nodeName string) *longhornV1beta2.Replica {
	for _, node := range list.Items {
		if node.Spec.NodeID == nodeName {
			return &node
		}
	}
	return nil
}
