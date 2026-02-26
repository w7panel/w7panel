package longhorn

import (
	"log/slog"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func (w *longhorncontroller) WatchNodeEvents() cache.SharedIndexInformer {
	// slog.Debug("WatchLonghornVolumesEvents")
	informer := w.factory.KubeInformerFactory.Core().V1().Nodes().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInit bool) {
			node, ok := obj.(*v1.Node)
			if ok {
				handleNodeEvent(node)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			node, ok := newObj.(*v1.Node)
			if ok {
				handleNodeEvent(node)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// node, ok := obj.(*v1.Node)
			// if ok {
			// 	// w.HandleNodeEvent(node)
			// }
		},
	})
	// informer.Run(make(chan struct{}))
	return informer
}

func handleNodeEvent(node *v1.Node) bool {
	//node is master

	labels := node.GetLabels()
	isMasterRole, ok := labels["node-role.kubernetes.io/master"]
	if !ok || isMasterRole != "true" {
		return false
	}
	defaultDisk, ok := labels["node.longhorn.io/create-default-disk"]
	if ok && defaultDisk == "true" {
		return false
	}
	labels["node.longhorn.io/create-default-disk"] = "true"
	// return true
	_, err := lclient.sdk.ClientSet.CoreV1().Nodes().Update(lclient.sdk.Ctx, node, metav1.UpdateOptions{})
	if err != nil {
		slog.Error("Failed to update node", "err", err)
	}
	return true
}
