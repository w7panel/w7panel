package containerd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/v2/core/content"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
)

type containerdBlobHandler struct {
	store   content.Store
	putLock sync.Mutex
}

func NewBlobHandler(store content.Store) *containerdBlobHandler {
	return &containerdBlobHandler{
		store: store,
	}
}
func (handler *containerdBlobHandler) Get(ctx context.Context, repo string, h v1.Hash) (io.ReadCloser, error) {
	return nil, errors.New("not impl")
}

// blobs 是否已存在
func (hd *containerdBlobHandler) Stat(ctx context.Context, repo string, h v1.Hash) (int64, error) {
	dgst, err := digest.Parse(h.String())
	if err != nil {
		return 0, err
	}
	info, err := hd.store.Info(ctx, dgst)
	return info.Size, nil
}

func (hd *containerdBlobHandler) Put(ctx context.Context, repo string, h v1.Hash, rc io.ReadCloser) error {
	hd.putLock.Lock()
	defer hd.putLock.Unlock()
	// v1.Hash 转 digest
	dgst, err := digest.Parse(h.String())
	if err != nil {
		return err
	}
	writer, err := content.OpenWriter(ctx, hd.store, content.WithRef("blob-"+dgst.Encoded()))
	if err != nil {
		if errdefs.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	defer writer.Close()

	// 读取所有数据以便获取大小
	body, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	if _, err := io.Copy(writer, bytes.NewReader(body)); err != nil {
		return err
	}
	if err := writer.Commit(ctx, int64(len(body)), dgst); err != nil && !errdefs.IsAlreadyExists(err) {
		return err
	}
	return nil
}
