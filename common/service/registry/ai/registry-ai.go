package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/errdefs"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	DefaultListenAddr       = "127.0.0.1:5000"
	DefaultContainerdSocket = "/run/containerd/containerd.sock"
	DefaultContainerdNS     = "k8s.io"
	DefaultImportRegistry   = "registry.local.w7.cc"
)

type Config struct {
	ListenAddr       string
	ContainerdSocket string
	ContainerdNS     string
	ImportRegistry   string
}

type Registry struct {
	cfg          Config
	ctrd         *client.Client
	store        content.Store
	imageSvc     images.Store
	uploadLock   sync.Mutex
	uploadMetaMu sync.RWMutex
	uploadRepos  map[string]string
}

type tagMetadata struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

type manifestEnvelope struct {
	Config    ocispec.Descriptor   `json:"config"`
	Layers    []ocispec.Descriptor `json:"layers"`
	Manifests []ocispec.Descriptor `json:"manifests"`
}

func (c Config) withDefaults() Config {
	if c.ListenAddr == "" {
		c.ListenAddr = DefaultListenAddr
	}
	if c.ContainerdSocket == "" {
		c.ContainerdSocket = DefaultContainerdSocket
	}
	if c.ContainerdNS == "" {
		c.ContainerdNS = DefaultContainerdNS
	}
	if c.ImportRegistry == "" {
		c.ImportRegistry = DefaultImportRegistry
	}
	return c
}

func New(cfg Config) (*Registry, error) {
	cfg = cfg.withDefaults()

	ctrd, err := client.New(cfg.ContainerdSocket)
	if err != nil {
		return nil, fmt.Errorf("connect containerd: %w", err)
	}

	return &Registry{
		cfg:         cfg,
		ctrd:        ctrd,
		store:       ctrd.ContentStore(),
		imageSvc:    ctrd.ImageService(),
		uploadRepos: map[string]string{},
	}, nil
}

func (r *Registry) Close() error {
	if r.ctrd == nil {
		return nil
	}
	return r.ctrd.Close()
}

func (r *Registry) NewServer() *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", r)
	return &http.Server{
		Addr:              r.cfg.ListenAddr,
		Handler:           withRegistryHeaders(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func withRegistryHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
		next.ServeHTTP(w, req)
	})
}

func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch {
	case req.URL.Path == "/v2/" || req.URL.Path == "/v2":
		w.WriteHeader(http.StatusOK)
	case strings.HasPrefix(req.URL.Path, "/v2/"):
		r.handleV2(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (r *Registry) handleV2(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/v2/")

	switch {
	case strings.HasSuffix(path, "/blobs/uploads/"):
		repo := strings.TrimSuffix(path, "/blobs/uploads/")
		r.handleBlobUploadStart(w, req, repo)
	case strings.Contains(path, "/blobs/uploads/"):
		idx := strings.Index(path, "/blobs/uploads/")
		repo := path[:idx]
		uploadID := path[idx+len("/blobs/uploads/"):]
		r.handleBlobUpload(w, req, repo, uploadID)
	case strings.Contains(path, "/blobs/"):
		idx := strings.Index(path, "/blobs/")
		repo := path[:idx]
		dgst := path[idx+len("/blobs/"):]
		r.handleBlob(w, req, repo, dgst)
	case strings.Contains(path, "/manifests/"):
		idx := strings.Index(path, "/manifests/")
		repo := path[:idx]
		ref := path[idx+len("/manifests/"):]
		r.handleManifest(w, req, repo, ref)
	default:
		http.NotFound(w, req)
	}
}

func (r *Registry) handleBlobUploadStart(w http.ResponseWriter, req *http.Request, repo string) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := req.ParseForm(); err == nil && req.Form.Get("digest") != "" {
		r.handleMonolithicBlobUpload(w, req, repo)
		return
	}

	uploadID := fmt.Sprintf("%d", time.Now().UnixNano())
	r.setUploadRepo(uploadID, repo)
	ctx := withNamespace(req.Context(), r.cfg.ContainerdNS)
	writer, err := content.OpenWriter(ctx, r.store, content.WithRef(r.uploadRef(uploadID)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := writer.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	location := fmt.Sprintf("/v2/%s/blobs/uploads/%s", repo, uploadID)
	w.Header().Set("Location", location)
	w.Header().Set("Range", "0-0")
	w.Header().Set("Docker-Upload-UUID", uploadID)
	w.WriteHeader(http.StatusAccepted)
}

func (r *Registry) handleMonolithicBlobUpload(w http.ResponseWriter, req *http.Request, repo string) {
	dgst, err := digest.Parse(req.Form.Get("digest"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if digest.FromBytes(body) != dgst {
		http.Error(w, "digest mismatch", http.StatusBadRequest)
		return
	}
	if err := r.putBlob(withNamespace(req.Context(), r.cfg.ContainerdNS), dgst, body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/%s", repo, dgst))
	w.Header().Set("Docker-Content-Digest", dgst.String())
	w.WriteHeader(http.StatusCreated)
}

func (r *Registry) handleBlobUpload(w http.ResponseWriter, req *http.Request, repo, uploadID string) {
	uploadRepo, ok := r.getUploadRepo(uploadID)
	if !ok {
		http.Error(w, "unknown upload", http.StatusNotFound)
		return
	}
	if uploadRepo != repo {
		http.Error(w, "upload repository mismatch", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodPatch:
		size, err := r.appendUpload(withNamespace(req.Context(), r.cfg.ContainerdNS), uploadID, req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads/%s", repo, uploadID))
		w.Header().Set("Docker-Upload-UUID", uploadID)
		w.Header().Set("Range", fmt.Sprintf("0-%d", max(size-1, 0)))
		w.WriteHeader(http.StatusAccepted)
	case http.MethodPut:
		ctx := withNamespace(req.Context(), r.cfg.ContainerdNS)
		dgst, err := digest.Parse(req.URL.Query().Get("digest"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if _, err := r.appendUpload(ctx, uploadID, req.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		size, err := r.commitUpload(ctx, uploadID, dgst)
		if err != nil {
			status := http.StatusInternalServerError
			if strings.Contains(err.Error(), "unexpected commit digest") {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}
		r.deleteUploadRepo(uploadID)
		w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/%s", repo, dgst))
		w.Header().Set("Docker-Content-Digest", dgst.String())
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.WriteHeader(http.StatusCreated)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Registry) handleBlob(w http.ResponseWriter, req *http.Request, _ string, dgstText string) {
	if req.Method != http.MethodHead && req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	dgst, err := digest.Parse(dgstText)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := withNamespace(req.Context(), r.cfg.ContainerdNS)
	info, err := r.store.Info(ctx, dgst)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Docker-Content-Digest", dgst.String())
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))
	if req.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	body, err := content.ReadBlob(ctx, r.store, ocispec.Descriptor{Digest: dgst})
	if err != nil {
		http.NotFound(w, req)
		return
	}
	w.Write(body)
}

func (r *Registry) handleManifest(w http.ResponseWriter, req *http.Request, repo, ref string) {
	switch req.Method {
	case http.MethodPut:
		r.handleManifestPut(w, req, repo, ref)
	case http.MethodHead, http.MethodGet:
		r.handleManifestRead(w, req, repo, ref)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Registry) handleManifestPut(w http.ResponseWriter, req *http.Request, repo, ref string) {
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
	if err := r.importToContainerd(req.Context(), repo, ref, mediaType, body); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/v2/%s/manifests/%s", repo, ref))
	w.Header().Set("Docker-Content-Digest", dgst.String())
	w.WriteHeader(http.StatusCreated)
}

func (r *Registry) handleManifestRead(w http.ResponseWriter, req *http.Request, repo, ref string) {
	meta, err := r.resolveManifest(repo, ref)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	manifestDigest, err := digest.Parse(meta.Digest)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	ctx := withNamespace(req.Context(), r.cfg.ContainerdNS)
	body, err := content.ReadBlob(ctx, r.store, ocispec.Descriptor{Digest: manifestDigest})
	if err != nil {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", meta.MediaType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.Header().Set("Docker-Content-Digest", meta.Digest)
	if req.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Write(body)
}

func (r *Registry) resolveManifest(repo, ref string) (*tagMetadata, error) {
	if strings.HasPrefix(ref, "sha256:") {
		dgst, err := digest.Parse(ref)
		if err != nil {
			return nil, err
		}
		ctx := withNamespace(context.Background(), r.cfg.ContainerdNS)
		info, err := r.store.Info(ctx, dgst)
		if err != nil {
			return nil, err
		}
		return &tagMetadata{
			MediaType: inferManifestMediaType(ctx, r.store, dgst),
			Digest:    dgst.String(),
			Size:      info.Size,
		}, nil
	}

	ctx := withNamespace(context.Background(), r.cfg.ContainerdNS)
	image, err := r.imageSvc.Get(ctx, r.imageRef(repo, ref))
	if err != nil {
		return nil, err
	}
	return &tagMetadata{
		MediaType: image.Target.MediaType,
		Digest:    image.Target.Digest.String(),
		Size:      image.Target.Size,
	}, nil
}

func (r *Registry) importToContainerd(ctx context.Context, repo, ref, mediaType string, body []byte) error {
	ctx = withNamespace(ctx, r.cfg.ContainerdNS)
	fullRef := r.imageRef(repo, ref)
	target := ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    digest.FromBytes(body),
		Size:      int64(len(body)),
	}

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

func (r *Registry) ingestDescriptor(ctx context.Context, desc ocispec.Descriptor, body []byte) error {
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

func (r *Registry) ingestChild(ctx context.Context, desc ocispec.Descriptor) error {
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

func (r *Registry) appendUpload(ctx context.Context, uploadID string, src io.Reader) (int64, error) {
	r.uploadLock.Lock()
	defer r.uploadLock.Unlock()

	writer, err := content.OpenWriter(ctx, r.store, content.WithRef(r.uploadRef(uploadID)))
	if err != nil {
		return 0, err
	}
	defer writer.Close()

	if _, err := io.Copy(writer, src); err != nil {
		return 0, err
	}
	status, err := writer.Status()
	if err != nil {
		return 0, err
	}
	return status.Offset, nil
}

func (r *Registry) commitUpload(ctx context.Context, uploadID string, expected digest.Digest) (int64, error) {
	r.uploadLock.Lock()
	defer r.uploadLock.Unlock()

	writer, err := content.OpenWriter(ctx, r.store, content.WithRef(r.uploadRef(uploadID)))
	if err != nil {
		return 0, err
	}
	defer writer.Close()

	status, err := writer.Status()
	if err != nil {
		return 0, err
	}
	if err := writer.Commit(ctx, status.Offset, expected); err != nil && !errdefs.IsAlreadyExists(err) {
		return 0, err
	}
	return status.Offset, nil
}

func (r *Registry) putBlob(ctx context.Context, dgst digest.Digest, body []byte) error {
	writer, err := content.OpenWriter(ctx, r.store, content.WithRef("blob-"+dgst.Encoded()))
	if err != nil {
		if errdefs.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	defer writer.Close()

	if _, err := io.Copy(writer, bytes.NewReader(body)); err != nil {
		return err
	}
	if err := writer.Commit(ctx, int64(len(body)), dgst); err != nil && !errdefs.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *Registry) uploadRef(uploadID string) string { return "upload-" + uploadID }

func (r *Registry) imageRef(repo, ref string) string {
	return fmt.Sprintf("%s/%s:%s", r.cfg.ImportRegistry, repo, ref)
}

func (r *Registry) setUploadRepo(uploadID, repo string) {
	r.uploadMetaMu.Lock()
	defer r.uploadMetaMu.Unlock()
	r.uploadRepos[uploadID] = repo
}

func (r *Registry) getUploadRepo(uploadID string) (string, bool) {
	r.uploadMetaMu.RLock()
	defer r.uploadMetaMu.RUnlock()
	repo, ok := r.uploadRepos[uploadID]
	return repo, ok
}

func (r *Registry) deleteUploadRepo(uploadID string) {
	r.uploadMetaMu.Lock()
	defer r.uploadMetaMu.Unlock()
	delete(r.uploadRepos, uploadID)
}

func inferManifestMediaType(ctx context.Context, store content.Store, dgst digest.Digest) string {
	body, err := content.ReadBlob(ctx, store, ocispec.Descriptor{Digest: dgst})
	if err != nil {
		return ocispec.MediaTypeImageManifest
	}
	var probe struct {
		MediaType string `json:"mediaType"`
	}
	if err := json.Unmarshal(body, &probe); err != nil || probe.MediaType == "" {
		return ocispec.MediaTypeImageManifest
	}
	return probe.MediaType
}

func withNamespace(ctx context.Context, ns string) context.Context {
	return namespaces.WithNamespace(ctx, ns)
}

func max[T ~int64 | ~int](a, b T) T {
	if a > b {
		return a
	}
	return b
}
