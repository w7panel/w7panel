package longhorn

import (
	"log/slog"
	"time"

	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	longhornV1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var lclient *longhornclient

func init() {
	lclient, _ = NewLonghornClient(k8s.NewK8sClient().Sdk)
}

func OnStart() error {
	nodes, err := lclient.GetNodeList()
	if err != nil {
		slog.Error("longhorn get node list error", "error", err)
		return err
	}
	k8sNodes, err := lclient.GetK8sNodeList()
	if err != nil {
		slog.Error("longhorn get k8s node list error", "error", err)
		return err
	}
	for _, k8sNodes := range k8sNodes.Items {
		knode := &k8sNodes
		ok := WebHookNode(&k8sNodes)
		if ok {
			_, err := lclient.sdk.ClientSet.CoreV1().Nodes().Update(lclient.sdk.Ctx, knode, metav1.UpdateOptions{})
			if err != nil {
				slog.Error("longhorn webhook handler update node error", "error", err)
			}
		}
	}

	// createDefaultVolumeByNodeList(nodes)
	updateAllReplicaCountByNodesList(nodes)
	createDefaultStorageClassByPvc()
	deleteNotSelectorStorageClass(nodes)
	return nil
}

func deleteNotSelectorStorageClass(nodes *longhornV1beta2.NodeList) {
	scs, err := lclient.GetK8sStorageClassList()
	if err != nil {
		slog.Error("longhorn get pvc list error", "error", err)
		return
	}
	scNames := []string{}
	for _, sc := range scs.Items {
		scNames = append(scNames, sc.Name)
	}
	diskSelector := getAllDiskSelector(nodes)
	for _, scName := range scNames {
		if !containsArr(diskSelector, scName) && (strings.HasPrefix(scName, "union") || strings.HasPrefix(scName, "disk")) {
			err := lclient.DeleteK8sStorageClass(scName)
			if err != nil {
				slog.Error("longhorn delete storage class error", "error", err)
			}
		}
	}
}

func WebhookLonghornNode(node *longhornV1beta2.Node) error {
	time.AfterFunc(time.Second*5, func() {
		createStorageClass(node)
		nodes, err := lclient.GetNodeList()
		if err != nil {
			slog.Error("longhorn get node list error", "error", err)
			return
		}
		// createDefaultVolumeByNodeList(nodes)
		updateAllReplicaCountByNodesList(nodes)
		createDefaultStorageClassByPvc()
		deleteNotSelectorStorageClass(nodes)
	})
	return nil
}

func WebHookStorageClass(storageClass *storagev1.StorageClass) bool {
	if storageClass.Name == "longhorn" {
		val, ok := storageClass.Annotations["storageclass.kubernetes.io/is-default-class"]
		if ok && val == "true" {
			storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] = "false"
			return true
		}
	}
	return false
}

func WebHookLonghornReplica() {
	time.AfterFunc(time.Second*5, func() {
		updateNodeLabel()
		updateAllReplicaCount() //更新副本数
	})

}

func WebHookNode(node *v1.Node) bool {
	labels := node.GetLabels()
	isMasterRole, ok := labels["node-role.kubernetes.io/master"]
	if !ok || isMasterRole != "true" {
		return false
	}
	_, ok = labels["node.longhorn.io/create-default-disk"]
	if ok {
		return false
	}
	labels["node.longhorn.io/create-default-disk"] = "config"
	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	configJson := `[ { "path":"/var/lib/longhorn", "allowScheduling":true, "name": "disk-default", "tags":[ "disk-default" ] } ]`
	annotations["node.longhorn.io/default-disks-config"] = configJson
	return true
}

func WebHookDeleteNode(node *v1.Node) {
	time.AfterFunc(time.Second*1, func() {
		lclient.DeleteNode(node.Name)
	})
}

func GetLonghornNodeList() (*longhornV1beta2.NodeList, error) {
	return lclient.GetNodeList()
}
