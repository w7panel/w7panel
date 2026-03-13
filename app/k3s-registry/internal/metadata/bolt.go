package metadata

import (
	"fmt"
	"sync"

	"go.etcd.io/bbolt"
)

type BoltStore struct {
	db   *bbolt.DB
	path string
}

var (
	defaultBucket = []byte("images")
	once          sync.Once
	store         *BoltStore
)

// NewBoltStore 创建 BoltDB 存储实例
func NewBoltStore(path string) (*BoltStore, error) {
	var err error
	var db *bbolt.DB

	once.Do(func() {
		db, err = bbolt.Open(path, 0644, nil)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open bolt db: %w", err)
	}

	// 创建默认 bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(defaultBucket)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	store = &BoltStore{db: db, path: path}
	return store, nil
}

// GetCatalog 获取镜像目录
func (b *BoltStore) GetCatalog() ([]string, error) {
	var images []string
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		cursor := bucket.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			images = append(images, string(k))
		}
		return nil
	})
	return images, err
}

// GetTags 获取镜像标签
func (b *BoltStore) GetTags(image string) ([]string, error) {
	var tags []string
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		data := bucket.Get([]byte(image))
		if data == nil {
			return nil
		}
		// 解析 tags (简化实现，实际需要更复杂的数据结构)
		// 这里假设数据格式是 JSON 数组
		return nil
	})
	return tags, err
}

// PutManifest 存储镜像 manifest
func (b *BoltStore) PutManifest(image, manifest string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		return bucket.Put([]byte(image), []byte(manifest))
	})
}

// GetManifest 获取镜像 manifest
func (b *BoltStore) GetManifest(image string) (string, error) {
	var manifest string
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		data := bucket.Get([]byte(image))
		if data != nil {
			manifest = string(data)
		}
		return nil
	})
	return manifest, err
}

// Close 关闭数据库连接
func (b *BoltStore) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}
