package longhorn

import (
	"fmt"
	"log/slog"

	longhornV1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
)

func checkDefaultDiskTag(node *longhornV1beta2.Node) error {
	if (node).Spec.Disks == nil {
		return fmt.Errorf("node %s has no disks", (node).Name)
	}
	for name, disk := range node.Spec.Disks {
		if disk.Tags == nil {
			disk.Tags = []string{}
		}
		if len(disk.Tags) == 0 {
			disk.Tags = append(disk.Tags, name)
			node.Spec.Disks[name] = disk
			slog.Debug("disk %s tags: %v", name, disk.Tags)
			_, err := lclient.UpdateNode(node)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
