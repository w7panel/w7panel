package longhorn

import (
	"fmt"
	"log/slog"

	"github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type defaultvolume struct {
	longhornclient *longhornclient
}

func NewDefaultVolume(longhornclient *longhornclient) *defaultvolume {
	return &defaultvolume{
		longhornclient: longhornclient,
	}
}

func createDefaultVolume() error {
	nodes, err := lclient.GetNodeList()
	if err != nil {
		return err
	}
	if len(nodes.Items) == 0 {
		return fmt.Errorf("no nodes found")
	}
	return createDefaultVolumeByNodeList(nodes)
}
func createDefaultVolumeByNodeList(nodes *v1beta2.NodeList) error {

	if len(nodes.Items) == 0 {
		return fmt.Errorf("no nodes found")
	}
	diskSelector := []string{}
	for _, node := range nodes.Items {
		if len(node.Spec.Disks) > 0 {
			for name, disk := range node.Spec.Disks {
				if len(disk.Tags) > 0 {
					diskSelector = disk.Tags
					break
				} else {
					diskSelector = []string{name}
					break
				}
			}
		}
	}
	if len(diskSelector) == 0 {
		return fmt.Errorf("no disk found")
	}
	if len(diskSelector) == 0 {
		slog.Debug("no diskSelector found")
		return nil
	}
	defaultScName := diskSelector[0]
	//check default volume pvc exists
	_, err := lclient.sdk.ClientSet.CoreV1().PersistentVolumeClaims("default").Get(lclient.sdk.Ctx, defaultVolumeName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// create default volume pvc if not exists
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: defaultVolumeName,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
					StorageClassName: func() *string { s := defaultScName; return &s }(),
				},
			}
			_, err = lclient.sdk.ClientSet.CoreV1().PersistentVolumeClaims("default").Create(lclient.sdk.Ctx, pvc, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create default volume pvc: %v", err)
			}
		}
		return err
	}
	return nil

}
