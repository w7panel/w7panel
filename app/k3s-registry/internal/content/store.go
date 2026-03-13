package content

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type BlobStore struct {
	baseDir string
	mu      sync.RWMutex
}

var (
	instance *BlobStore
	once     sync.Once
)

// NewBlobStore 创建 Blob 存储实例
func NewBlobStore(baseDir string) (*BlobStore, error) {
	var err error
	once.Do(func() {
		// 确保基础目录存在
		if err = os.MkdirAll(baseDir, 0755); err != nil {
			return
		}

		instance = &BlobStore{
			baseDir: baseDir,
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create blob store: %w", err)
	}

	return instance, nil
}

// PutBlob 存储 blob 数据
func (s *BlobStore) PutBlob(digest string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	blobPath := s.blobPath(digest)
	if err := os.MkdirAll(filepath.Dir(blobPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(blobPath, data, 0644)
}

// GetBlob 获取 blob 数据
func (s *BlobStore) GetBlob(digest string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blobPath := s.blobPath(digest)
	return os.ReadFile(blobPath)
}

// BlobExists 检查 blob 是否存在
func (s *BlobStore) BlobExists(digest string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blobPath := s.blobPath(digest)
	_, err := os.Stat(blobPath)
	return err == nil
}

// GetBlobReader 获取 blob 读取器
func (s *BlobStore) GetBlobReader(digest string) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blobPath := s.blobPath(digest)
	return os.Open(blobPath)
}

// DeleteBlob 删除 blob
func (s *BlobStore) DeleteBlob(digest string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	blobPath := s.blobPath(digest)
	return os.Remove(blobPath)
}

// blobPath 根据 digest 生成 blob 文件路径
func (s *BlobStore) blobPath(digest string) string {
	// 使用 digest 的前两个字符作为子目录
	if len(digest) < 3 {
		return filepath.Join(s.baseDir, digest)
	}
	return filepath.Join(s.baseDir, digest[:2], digest)
}
