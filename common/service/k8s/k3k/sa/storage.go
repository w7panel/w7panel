package sa

import (
	"context"

	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Storage struct {
	client.Client
}

func NewStorage(client client.Client) *Storage {
	return &Storage{
		Client: client,
	}
}

/*
*

	扩容集群存储
*/
func (s *Storage) Handle(ctx context.Context, user *k3ktypes.K3kUser) error {
	pvc := &v1.PersistentVolumeClaim{}
	err := s.Client.Get(ctx, types.NamespacedName{Namespace: user.GetK3kNamespace(), Name: user.GetClusterServer0PvcName()}, pvc)
	if err != nil {
		return err
	}
	st := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	stNewStr := user.GetClusterSysStorageRequestSize()
	stNew, err := resource.ParseQuantity(stNewStr)
	if err != nil {
		return err
	}
	if st.Cmp(stNew) < 0 {
		pvc.Spec.Resources.Requests[v1.ResourceStorage] = stNew
		err := s.Client.Update(ctx, pvc)
		if err != nil {
			return err
		}
	}
	return nil
}
