package logic

import (
	"context"
	"sync"

	"github.com/w7panel/w7panel/app/k3s-registry/model"
)

type ContainersLogic struct {
	// 这里可以注入 containerd 客户端等依赖
}

var (
	containersLogicInstance *ContainersLogic
	containersOnce          sync.Once
)

// NewContainersLogic 创建 Containers 逻辑实例
func NewContainersLogic() *ContainersLogic {
	containersOnce.Do(func() {
		containersLogicInstance = &ContainersLogic{}
	})
	return containersLogicInstance
}

// List 获取容器列表
func (l *ContainersLogic) List(ctx context.Context) ([]model.ContainerInfo, error) {
	// 简化实现，返回空列表
	// 实际实现需要从 containerd 获取容器列表
	return []model.ContainerInfo{}, nil
}

// Get 获取容器详情
func (l *ContainersLogic) Get(ctx context.Context, id string) (*model.ContainerInfo, error) {
	// 简化实现
	// 实际实现需要从 containerd 获取容器详情
	return &model.ContainerInfo{
		ID:      id,
		Status:  "running",
		Image:   "example:latest",
		ImageID: "sha256:123456",
		PID:     12345,
	}, nil
}

// GetLayers 获取容器镜像层信息
func (l *ContainersLogic) GetLayers(ctx context.Context, id string) (*model.ContainerLayers, error) {
	// 简化实现
	// 实际实现需要从 containerd 获取镜像层信息
	return &model.ContainerLayers{
		ContainerID: id,
		Image:       "example:latest",
		Layers: []model.LayerInfo{
			{
				Digest: "sha256:abc123",
				Size:   1024,
			},
			{
				Digest: "sha256:def456",
				Size:   2048,
			},
		},
		TotalSize:  3072,
		LayerCount: 2,
	}, nil
}
