package core

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func SigObjectToInfo(obj sigclient.Object, client resource.RESTClient, mapper meta.RESTMapper) (*resource.Info, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	// gvk := deployment.GroupVersionKind()
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping: %v", err)
	}

	// 创建 Info 对象
	info := &resource.Info{
		Client:      client,
		Mapping:     mapping,
		Namespace:   obj.GetNamespace(),
		Name:        obj.GetName(),
		Object:      obj,
		Subresource: "",
	}

	return info, nil
}
