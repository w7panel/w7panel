package longhorn

import longhornV1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"

type longNodeVolumes struct {
	volume *longhornV1beta2.Volume
	nodes  *longhornV1beta2.NodeList
}

func NewlongNodeVolumes(nodes *longhornV1beta2.NodeList, volumes *longhornV1beta2.Volume) *longNodeVolumes {
	return &longNodeVolumes{volume: volumes, nodes: nodes}
}

func (l *longNodeVolumes) NeedReplicaCount() (int, error) {
	result := 0
	nodes := l.nodes
	selector := l.volume.Spec.DiskSelector
	if selector == nil {
		selector = []string{}
	}
	for _, node := range nodes.Items {
		for _, disk := range node.Spec.Disks {
			// if disk.AllowScheduling {
			tags := disk.Tags
			if containsAll(tags, selector) {
				result++
			}
			// }
		}
	}
	return result, nil
}

type VolumeReplica struct {
	volume   *longhornV1beta2.Volume
	replicas *longhornV1beta2.ReplicaList
}

type volumeReplicaList []VolumeReplica

// contains checks if a slice contains a specific element
func containsArr(slice []string, element string) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func (l volumeReplicaList) GetNeedDeleteReplicas(diskSelector []string, nodeIds []string) *longhornV1beta2.ReplicaList {
	rlist := &longhornV1beta2.ReplicaList{}
	for _, vr := range l {
		if vr.volume.Spec.DiskSelector != nil && len(vr.replicas.Items) > 0 {
			if containsAll(vr.volume.Spec.DiskSelector, diskSelector) && vr.volume.Spec.NumberOfReplicas > 1 {
				for _, replica := range vr.replicas.Items {
					//判断nodeId是否在nodeIds中
					if containsArr(nodeIds, replica.Spec.NodeID) {
						rlist.Items = append(rlist.Items, replica)
					}

				}
			}
		}
	}
	return rlist
}
