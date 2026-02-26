package core

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var cacheobj *k8sCache = &k8sCache{}

func InitCache(cache cache.Cache) {
	cacheobj = NewK8sCache(cache)
}

type k8sCache struct {
	client.Client
	cache.Cache
	context.Context
}

func NewK8sCache(cache cache.Cache) *k8sCache {
	return &k8sCache{
		Cache: cache,
	}
}

func (c *k8sCache) GetServiceAccount(saName, namespace string) (*v1.ServiceAccount, error) {
	sa := &v1.ServiceAccount{}
	err := c.Cache.Get(c.Context, client.ObjectKey{
		Namespace: namespace, Name: saName,
	}, sa)
	return sa, err
}
