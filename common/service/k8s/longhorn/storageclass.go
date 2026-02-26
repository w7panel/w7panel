package longhorn

import (
	"log/slog"

	longhornV1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

/*
  - kind: StorageClass
    apiVersion: storage.k8s.io/v1
    metadata:
    name: longhorn
    annotations:
    storageclass.kubernetes.io/is-default-class: "true"
    provisioner: driver.longhorn.io
    allowVolumeExpansion: true
    reclaimPolicy: "Delete"
    volumeBindingMode: Immediate
    parameters:
    numberOfReplicas: "3"
    staleReplicaTimeout: "30"
    fromBackup: ""
    fsType: "ext4"
    dataLocality: "disabled"
    unmapMarkSnapChainRemoved: "ignored"
    disableRevisionCounter: "true"
    dataEngine: "v1"
*/
func (w *longhorncontroller) WatchLonghornStorageClass() cache.SharedIndexInformer {
	// slog.Debug("WatchLonghornVolumesEvents")
	informer := w.factory.KubeInformerFactory.Storage().V1().StorageClasses().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInit bool) {
			sc, ok := obj.(*storagev1.StorageClass)
			if ok {
				w.HandleStorageClass(sc)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			sc, ok := newObj.(*storagev1.StorageClass)
			if ok {
				w.HandleStorageClass(sc)
			}
		},
		DeleteFunc: func(obj interface{}) {

		},
	})
	// informer.Run(make(chan struct{}))
	return informer
}

func (w *longhorncontroller) HandleStorageClass(storageClass *storagev1.StorageClass) {
	if storageClass.Name == "longhorn" {
		val, ok := storageClass.Annotations["storageclass.kubernetes.io/is-default-class"]
		if ok && val == "true" {
			storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] = "false"
			_, err := w.sdk.ClientSet.StorageV1().StorageClasses().Update(w.sdk.Ctx, storageClass, metav1.UpdateOptions{})
			if err != nil {
				slog.Error("Failed to update storage class", "err", err)
			}
		}
	}
}

func (w *longhorncontroller) createStorageClass(node *longhornV1beta2.Node) error {
	return createStorageClass(node)
}

func createDefaultStorageClass(defaultDiskClassName string) error {
	sc, err := lclient.GetK8sStorageClass(defaultDiskClassName)
	if err != nil {
		slog.Error("GetK8sStorageClass error", "err", err)
		return err
	}
	if sc.Annotations == nil {
		sc.Annotations = make(map[string]string)
	}
	val, ok := sc.Annotations["storageclass.kubernetes.io/is-default-class"]
	if !ok || val != "true" {
		sc.Annotations["storageclass.kubernetes.io/is-default-class"] = "true"
		_, err := lclient.UpdateK8sStorageClass(sc)
		if err != nil {
			return err
		}
	}
	return nil
}

func createDefaultStorageClassByPvc() error {
	pvc, err := lclient.sdk.ClientSet.CoreV1().PersistentVolumeClaims("default").Get(lclient.sdk.Ctx, defaultVolumeName, metav1.GetOptions{})
	if err != nil {
		slog.Error("Get ConfigMap error", "err", err)
		return err
	}
	scName := pvc.Spec.StorageClassName
	if scName != nil {
		return createDefaultStorageClass(*scName)
	}
	return nil

}

func webhookLonghornStorageClass(storageClass *storagev1.StorageClass) {
	if storageClass.Name == "longhorn" {
		val, ok := storageClass.Annotations["storageclass.kubernetes.io/is-default-class"]
		if ok && val == "true" {
			storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] = "false"
			_, err := lclient.sdk.ClientSet.StorageV1().StorageClasses().Update(lclient.sdk.Ctx, storageClass, metav1.UpdateOptions{})
			if err != nil {
				slog.Error("Failed to update storage class", "err", err)
			}
		}
	}
}

func createStorageClass(node *longhornV1beta2.Node) error {
	storageClassList, err := lclient.GetK8sStorageClassList()
	if err != nil {
		return err
	}
	var storagemap map[string]bool = make(map[string]bool)
	for _, storageClass := range storageClassList.Items {
		storagemap[storageClass.Name] = true
	}
	for _, disk := range node.Spec.Disks {
		for _, tag := range disk.Tags {
			// 检查storageClassList 是否有tag的storageClass
			ok := storagemap[tag]
			if !ok {
				expansition := true
				sc := &storagev1.StorageClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: tag,
						// Annotations: map[string]string{
						// 	"storageclass.kubernetes.io/is-default-class": "true",
						// },
					},
					Provisioner: "driver.longhorn.io",
					Parameters: map[string]string{
						"numberOfReplicas":          "1",
						"staleReplicaTimeout":       "300",
						"fromBackup":                "",
						"fsType":                    "ext4",
						"dataLocality":              "disabled",
						"diskSelector":              tag,
						"unmapMarkSnapChainRemoved": "ignored",
						"disableRevisionCounter":    "true",
						"dataEngine":                "v1",
					},

					AllowVolumeExpansion: &expansition,
				}
				_, err := lclient.CreateK8sStorageClass(sc)
				if err != nil {
					slog.Error("create storage class error", "err", err)
					continue
				}
			}
		}

	}

	return nil
}
