package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitee.com/we7coreteam/k8s-offline/app/k3s-registry/internal/content"
	"gitee.com/we7coreteam/k8s-offline/app/k3s-registry/internal/metadata"
)

type RegistryLogic struct {
	metaStore *metadata.BoltStore
	blobStore *content.BlobStore
}

var (
	registryLogicInstance *RegistryLogic
	registryOnce          sync.Once
)

// NewRegistryLogic 创建 Registry 逻辑实例
func NewRegistryLogic() *RegistryLogic {
	registryOnce.Do(func() {
		// 初始化存储
		metaStore, err := metadata.NewBoltStore("/tmp/k3s-registry/metadata.db")
		if err != nil {
			// 记录错误但不中断
			fmt.Printf("Warning: failed to create metadata store: %v\n", err)
		}

		blobStore, err := content.NewBlobStore("/tmp/k3s-registry/blobs")
		if err != nil {
			fmt.Printf("Warning: failed to create blob store: %v\n", err)
		}

		registryLogicInstance = &RegistryLogic{
			metaStore: metaStore,
			blobStore: blobStore,
		}
	})
	return registryLogicInstance
}

// GetCatalog 获取镜像目录
func (l *RegistryLogic) GetCatalog(ctx context.Context) ([]string, error) {
	if l.metaStore == nil {
		return []string{}, fmt.Errorf("metadata store not initialized")
	}
	return l.metaStore.GetCatalog()
}

// GetTags 获取镜像标签
func (l *RegistryLogic) GetTags(ctx context.Context, name string) ([]string, error) {
	if l.metaStore == nil {
		return []string{}, fmt.Errorf("metadata store not initialized")
	}
	return l.metaStore.GetTags(name)
}

// GetManifest 获取镜像 manifest
func (l *RegistryLogic) GetManifest(ctx context.Context, name, reference string) (string, error) {
	if l.metaStore == nil {
		return "", fmt.Errorf("metadata store not initialized")
	}
	return l.metaStore.GetManifest(name + ":" + reference)
}

// PushManifest 推送镜像 manifest
func (l *RegistryLogic) PushManifest(ctx context.Context, name, reference, manifest string) error {
	if l.metaStore == nil {
		return fmt.Errorf("metadata store not initialized")
	}
	return l.metaStore.PutManifest(name+":"+reference, manifest)
}

// GetBlob 获取 blob 数据
func (l *RegistryLogic) GetBlob(ctx context.Context, name, digest string) ([]byte, error) {
	if l.blobStore == nil {
		return nil, fmt.Errorf("blob store not initialized")
	}
	return l.blobStore.GetBlob(digest)
}

// BlobExists 检查 blob 是否存在
func (l *RegistryLogic) BlobExists(ctx context.Context, name, digest string) bool {
	if l.blobStore == nil {
		return false
	}
	return l.blobStore.BlobExists(digest)
}

// InitUpload 初始化 blob 上传
func (l *RegistryLogic) InitUpload(ctx context.Context, name string) (string, error) {
	// 生成唯一的 upload ID
	uuid := generateUUID()
	return uuid, nil
}

// CompleteUpload 完成 blob 上传
func (l *RegistryLogic) CompleteUpload(ctx context.Context, name, uuid, digest string, data []byte) error {
	if l.blobStore == nil {
		return fmt.Errorf("blob store not initialized")
	}
	return l.blobStore.PutBlob(digest, data)
}

// generateUUID 生成 UUID (简化实现)
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
