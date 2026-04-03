package containerd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	client "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type manifestEnvelope struct {
	Config    ocispec.Descriptor   `json:"config"`
	Layers    []ocispec.Descriptor `json:"layers"`
	Manifests []ocispec.Descriptor `json:"manifests"`
}

type Manifest struct {
	ctrd     *client.Client
	imageSvc images.Store
	store    content.Store
}

func NewManifest(ctrd *client.Client, imageSvc images.Store) *Manifest {
	return &Manifest{
		ctrd:     ctrd,
		imageSvc: imageSvc,
	}
}

func (r *Manifest) PutManifest(w http.ResponseWriter, req *http.Request, repo, ref string) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mediaType := req.Header.Get("Content-Type")
	if mediaType == "" {
		http.Error(w, "missing content type", http.StatusBadRequest)
		return
	}

	dgst := digest.FromBytes(body)
	go func() {
		if err := r.importToContainerd(context.Background(), repo, ref, mediaType, body); err != nil {
			slog.Error("import to containerd failed", "error", err)
		}
	}()

	w.Header().Set("Location", fmt.Sprintf("/v2/%s/manifests/%s", repo, ref))
	w.Header().Set("Docker-Content-Digest", dgst.String())
	w.WriteHeader(http.StatusCreated)
}

func (r *Manifest) importToContainerd(ctx context.Context, repo, ref, mediaType string, body []byte) error {
	ctx = withNamespace(ctx)
	fullRef := r.imageRef(repo, ref)
	target := ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    digest.FromBytes(body),
		Size:      int64(len(body)),
	}
	//之前ai ingestDescriptor 必须有
	if err := r.ingestDescriptor(ctx, target, body); err != nil {
		return err
	}
	image := images.Image{
		Name:      fullRef,
		Labels:    map[string]string{"registry.repo": repo, "registry.tag": ref},
		Target:    target,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if _, err := r.imageSvc.Create(ctx, image); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			if _, updateErr := r.imageSvc.Update(ctx, image); updateErr != nil {
				return updateErr
			}
		} else if _, updateErr := r.imageSvc.Update(ctx, image); updateErr != nil {
			return updateErr
		}
	}

	img := client.NewImage(r.ctrd, image)
	if err := img.Unpack(ctx, defaults.DefaultSnapshotter); err != nil {
		return fmt.Errorf("unpack image %s with snapshotter %s: %w", fullRef, defaults.DefaultSnapshotter, err)
	}
	return nil
}

func (r *Manifest) imageRef(repo, ref string) string {
	return fmt.Sprintf("%s/%s:%s", REPO, repo, ref)
}

func (r *Manifest) ingestDescriptor(ctx context.Context, desc ocispec.Descriptor, body []byte) error {
	switch desc.MediaType {
	case images.MediaTypeDockerSchema2Manifest, ocispec.MediaTypeImageManifest:
		var manifest manifestEnvelope
		if err := json.Unmarshal(body, &manifest); err != nil {
			return err
		}
		if err := r.ingestChild(ctx, manifest.Config); err != nil {
			return err
		}
		for _, layer := range manifest.Layers {
			if err := r.ingestChild(ctx, layer); err != nil {
				return err
			}
		}
	case images.MediaTypeDockerSchema2ManifestList, ocispec.MediaTypeImageIndex:
		var index manifestEnvelope
		if err := json.Unmarshal(body, &index); err != nil {
			return err
		}
		for _, child := range index.Manifests {
			if err := r.ingestChild(ctx, child); err != nil {
				return err
			}
		}
	}
	if err := writeBlob(ctx, r.store, desc, body); err != nil {
		return err
	}
	return setGCChildLabels(ctx, r.store, desc)
}

func (r *Manifest) ingestChild(ctx context.Context, desc ocispec.Descriptor) error {
	body, err := content.ReadBlob(ctx, r.store, desc)
	if err != nil {
		return fmt.Errorf("read blob %s: %w", desc.Digest, err)
	}
	return r.ingestDescriptor(ctx, desc, body)
}

func writeBlob(ctx context.Context, store content.Store, desc ocispec.Descriptor, body []byte) error {
	if _, err := store.Info(ctx, desc.Digest); err == nil {
		return nil
	}
	return content.WriteBlob(ctx, store, desc.Digest.String(), bytes.NewReader(body), desc)
}

func setGCChildLabels(ctx context.Context, store content.Store, desc ocispec.Descriptor) error {
	children, err := images.Children(ctx, store, desc)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	if len(children) == 0 {
		return nil
	}

	info := content.Info{
		Digest: desc.Digest,
		Labels: map[string]string{},
	}
	fields := make([]string, 0, len(children))
	keys := map[string]uint{}

	for _, child := range children {
		for _, key := range images.ChildGCLabels(child) {
			idx := keys[key]
			keys[key] = idx + 1
			switch {
			case strings.HasSuffix(key, ".sha256."):
				key = fmt.Sprintf("%s%s", key, child.Digest.Hex()[:12])
			case idx > 0 || strings.HasSuffix(key, "."):
				key = fmt.Sprintf("%s%d", key, idx)
			}
			info.Labels[key] = child.Digest.String()
			fields = append(fields, "labels."+key)
		}
	}

	_, err = store.Update(ctx, info, fields...)
	return err
}
