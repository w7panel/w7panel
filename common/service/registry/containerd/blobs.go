package containerd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/v2/core/content"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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
	// dgst, err := digest.Parse(h.String())
	// if err != nil {
	// 	return nil, err
	// }
	// body, err := content.ReadBlob(withNamespace(ctx), handler.store, ocispec.Descriptor{Digest: dgst})
	// if err != nil {
	// 	return nil, err
	// }
	// return io.NopCloser(bytes.NewReader(body)), nil
	return nil, errors.New("not impl")
}

// blobs 是否已存在
func (hd *containerdBlobHandler) Stat(ctx context.Context, repo string, h v1.Hash) (int64, error) {
	dgst, err := digest.Parse(h.String())
	if err != nil {
		return 0, err
	}
	info, err := hd.store.Info(withNamespace(ctx), dgst)
	if err != nil {
		return 0, registry.ErrNotFound()
	}
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
	ctx = withNamespace(ctx)

	// 读取所有数据以便获取大小
	body, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	desc := ocispec.Descriptor{
		Digest: dgst,
		Size:   int64(len(body)),
	}
	if err := content.WriteBlob(ctx, hd.store, dgst.String(), bytes.NewReader(body), desc); err != nil && !errdefs.IsAlreadyExists(err) {
		return err
	}
	return nil
}
