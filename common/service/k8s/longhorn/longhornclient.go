package longhorn

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/go-resty/resty/v2"
	longhornV1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	v1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	longhornClient "github.com/longhorn/longhorn-manager/k8s/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

const (
	namespace         = "longhorn-system"
	defaultVolumeName = "default-volume"
	baseUrl           = "http://longhorn-backend.longhorn-system.svc:9500"
	// baseUrl = "http://longhorn.fan.b2.sz.w7.com"
)

type longhornclient struct {
	sdk       *k8s.Sdk
	client    *longhornClient.Clientset
	namespace string
}

func NewLonghornClient(sdk *k8s.Sdk) (*longhornclient, error) {
	result := &longhornclient{sdk: sdk}
	restConfig, err := sdk.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	client, err := longhornClient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	result.client = client
	result.namespace = namespace
	return result, nil
}

func (c *longhornclient) GetVolume(name string) (*longhornV1beta2.Volume, error) {
	return c.client.LonghornV1beta2().Volumes(c.namespace).Get(c.sdk.Ctx, name, v1.GetOptions{})
}

func (c *longhornclient) UpdateVolume(volume *longhornV1beta2.Volume) (*longhornV1beta2.Volume, error) {
	return c.client.LonghornV1beta2().Volumes(c.namespace).Update(c.sdk.Ctx, volume, v1.UpdateOptions{})
}

func (c *longhornclient) GetVolumeList() (*longhornV1beta2.VolumeList, error) {
	return c.client.LonghornV1beta2().Volumes(c.namespace).List(c.sdk.Ctx, v1.ListOptions{})
}

func (c *longhornclient) GetNodeList() (*longhornV1beta2.NodeList, error) {
	return c.client.LonghornV1beta2().Nodes(c.namespace).List(c.sdk.Ctx, v1.ListOptions{})
}

func (c *longhornclient) DeleteNode(name string) error {
	return c.client.LonghornV1beta2().Nodes(c.namespace).Delete(c.sdk.Ctx, name, v1.DeleteOptions{})
}

func (c *longhornclient) GetK8sNodeList() (*corev1.NodeList, error) {
	return c.sdk.ClientSet.CoreV1().Nodes().List(c.sdk.Ctx, v1.ListOptions{})
}

func (c *longhornclient) GetK8sStorageClassList() (*storagev1.StorageClassList, error) {
	return c.sdk.ClientSet.StorageV1().StorageClasses().List(c.sdk.Ctx, v1.ListOptions{})
}

func (c *longhornclient) DeleteK8sStorageClass(name string) error {
	return c.sdk.ClientSet.StorageV1().StorageClasses().Delete(c.sdk.Ctx, name, v1.DeleteOptions{})
}

func (c *longhornclient) GetK8sStorageClass(name string) (*storagev1.StorageClass, error) {
	return c.sdk.ClientSet.StorageV1().StorageClasses().Get(c.sdk.Ctx, name, v1.GetOptions{})
}

func (c *longhornclient) UpdateK8sStorageClass(sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	return c.sdk.ClientSet.StorageV1().StorageClasses().Update(c.sdk.Ctx, sc, v1.UpdateOptions{})
}

func (c *longhornclient) CreateK8sStorageClass(sc *storagev1.StorageClass) (*storagev1.StorageClass, error) {
	return c.sdk.ClientSet.StorageV1().StorageClasses().Create(c.sdk.Ctx, sc, v1.CreateOptions{})
}

func (c *longhornclient) UpdateNode(node *longhornV1beta2.Node) (*longhornV1beta2.Node, error) {
	return c.client.LonghornV1beta2().Nodes(c.namespace).Update(c.sdk.Ctx, node, v1.UpdateOptions{})
}

func (c *longhornclient) GetReplicaList() (*longhornV1beta2.ReplicaList, error) {
	return c.client.LonghornV1beta2().Replicas(c.namespace).List(c.sdk.Ctx, v1.ListOptions{})
}

func (c *longhornclient) WatchNodeList() (watch.Interface, error) {
	return c.client.LonghornV1beta2().Nodes(c.namespace).Watch(c.sdk.Ctx, v1.ListOptions{})
}

func (c *longhornclient) GetLonghornReplicaList() (*v1beta2.ReplicaList, error) {
	return c.client.LonghornV1beta2().Replicas(c.namespace).List(c.sdk.Ctx, v1.ListOptions{})
}

func (c *longhornclient) UpdateK8sNodeLabel(node *corev1.Node, labels map[string]string) (*corev1.Node, error) {
	apply, err := applyv1.ExtractNode(node, "k8s-offline")
	if err != nil {
		return nil, err
	}
	apply.WithLabels(labels)
	return c.sdk.ClientSet.CoreV1().Nodes().Apply(c.sdk.Ctx, apply, metav1.ApplyOptions{FieldManager: "k8s-offline"})
}

func (c *longhornclient) CreateVolume(name string, size int64, diskSelector []string) (*longhornV1beta2.Volume, error) {

	volume := &longhornV1beta2.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Volume",
			APIVersion: longhornV1beta2.SchemeGroupVersion.String(),
		},
		Spec: longhornV1beta2.VolumeSpec{
			AccessMode:       "rwo",
			BackingImage:     "",
			DataEngine:       "v1",
			NumberOfReplicas: 1,
			DataLocality:     "disabled",
			Frontend:         "blockdev",
			DiskSelector:     diskSelector,
			NodeSelector:     []string{},
			Size:             10737418240,
		},
	}
	return c.client.LonghornV1beta2().Volumes(c.namespace).Create(c.sdk.Ctx, volume, v1.CreateOptions{})
}

func (c *longhornclient) GetDisksCount(diskSelector []string) (int, error) {

	result := 0
	nodes, err := c.GetNodeList()
	if err != nil {
		return 0, err
	}
	for _, node := range nodes.Items {
		for _, disk := range node.Spec.Disks {
			if disk.AllowScheduling {
				tags := disk.Tags
				if containsAll(tags, diskSelector) {
					result++
				}
			}
		}
	}
	return result, nil
}
func (c *longhornclient) GetVolumeReplicaCompose() (*volumeReplicaCompose, error) {
	replicas, err := c.GetReplicaList()
	if err != nil {
		return nil, err
	}
	volumes, err := c.GetVolumeList()
	if err != nil {
		return nil, err
	}
	return &volumeReplicaCompose{
		replicas: replicas,
		volumes:  volumes,
	}, nil
}

type volumeReplicaCompose struct {
	replicas *longhornV1beta2.ReplicaList
	volumes  *longhornV1beta2.VolumeList
}

func (l *volumeReplicaCompose) GetVolumeReplicasByVolumeName(volumeName string) *longhornV1beta2.ReplicaList {
	rlist := &longhornV1beta2.ReplicaList{}
	for _, replica := range l.replicas.Items {
		if replica.Labels["longhornvolume"] == volumeName {
			rlist.Items = append(rlist.Items, replica)
		}
	}
	return rlist
}

func (l *volumeReplicaCompose) GetVolumeReplicas() volumeReplicaList {
	result := volumeReplicaList{}
	for _, volume := range l.volumes.Items {
		// if volume.Status.State == longhornV1beta2.VolumeStateAttached {
		result = append(result, VolumeReplica{
			volume:   &volume,
			replicas: l.GetVolumeReplicasByVolumeName(volume.Name),
		})
		// }
	}
	return result
}

func createPv(volumeName string, pvName string, scName string) error {
	if scName == "" {
		scName = "longhorn"
	}
	json := fmt.Sprintf(`{"fsType": "ext4", "pvName": "%s", "storageClassName": "`+scName+`"}`, pvName)
	return longhornVolumeApiAction(volumeName, "pvCreate", json)
}

func createPvc(volumeName string, pvcName string, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}
	json := fmt.Sprintf(`{"pvcName": "%s", "namespace": "%s"}`, pvcName, namespace)
	return longhornVolumeApiAction(volumeName, "pvcCreate", json)
}

// containsAll checks if slice a contains all elements of slice b

func updateVolumeReplicaCountApi(name string, count int) error {
	// json := `{"replicaCount": 6}`
	json := fmt.Sprintf(`{"replicaCount": %d}`, count)
	return longhornVolumeApiAction(name, "updateReplicaCount", json)
}

func longhornVolumeApiAction(volumeName string, action string, json string) error {
	slog.Info("longhornclient longhornVolumeApiAction: ", "volumeName", volumeName, "action", action)
	postUrl := baseUrl + "/v1/volumes/" + volumeName + "?action=" + action
	response, err := resty.New().R().SetBody(json).SetHeader("content-type", "application/json").SetHeader("Accept", "application/json").Post(postUrl)
	if err != nil {
		slog.Error("longhornclient longhornVolumeApiAction error: ", "err", err)
	}

	if response.StatusCode() != http.StatusOK {
		slog.Error("longhornclient longhornVolumeApiAction error response: %s", "err", response.String())
		return errors.New("longhornclient UpdateVolumeReplicaCount error: " + response.Status() + ": content: " + response.String())
	}
	return nil
}

func LonghorStoragePercentage(value string) error {
	postUrl := baseUrl + "/v1/settings/storage-over-provisioning-percentage"
	json := `{"actions":{},"applied":true,"definition":{"category":"scheduling","default":"100","description":"The over-provisioning percentage defines how much storage can be allocated relative to the hard drive's capacity","displayName":"Storage Over Provisioning Percentage","range":{"minimum":0},"readOnly":false,"required":true,"type":"int"},"id":"storage-over-provisioning-percentage","links":{"self":"http://longhorn.fan.b2.sz.w7.com/v1/settings/storage-over-provisioning-percentage"},"name":"storage-over-provisioning-percentage","type":"setting","value":"%s"}`
	json = fmt.Sprintf(json, value)
	response, err := resty.New().R().SetBody(json).SetHeader("content-type", "application/json").SetHeader("Accept", "application/json").Put(postUrl)
	if err != nil {
		slog.Error("longhornclient longhornVolumeApiAction error: ", "err", err)
	}

	if response.StatusCode() != http.StatusOK {
		slog.Error("longhornclient LonghorStoragePercentage error response: %s", "err", response.String())
		return errors.New("longhornclient LonghorStoragePercentage error: " + response.Status() + ": content: " + response.String())
	}
	return nil
}
