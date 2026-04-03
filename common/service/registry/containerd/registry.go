package containerd

import (
	"context"
	"net/http"
	"strings"

	micro "github.com/google/go-containerregistry/pkg/registry"
)

type Registry struct {
	google   http.Handler
	manifest *Manifest
}

func InitReigstry(ctx context.Context) (http.Handler, error) {
	return NewRegistry()
}

func NewRegistry() (*Registry, error) {

	c, err := CreateClient()
	if err != nil {
		return nil, err
	}
	contentStore := c.ContentStore()
	handler := NewBlobHandler(contentStore)
	manifest := NewManifest(c, c.ImageService())
	google := micro.New(micro.WithBlobHandler(handler))
	return &Registry{
		google:   google,
		manifest: manifest,
	}, nil
}

func (r *Registry) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if r.isPutManifestRequest(req, http.StatusCreated) {
		name, reference := r.extractNameAndReference(req.URL.Path)
		if name != "" && reference != "" {
			r.manifest.PutManifest(rw, req, name, reference)
			return
		}
	} else {
		r.google.ServeHTTP(rw, req)
	}
}

func (r *Registry) isPutManifestRequest(req *http.Request, status int) bool {
	// Docker Registry API: PUT /v2/{name}/manifests/{reference}
	return req.Method == http.MethodPut &&
		strings.Contains(req.URL.Path, "/manifests/") &&
		status == http.StatusCreated
}

// extractNameAndReference 从路径中提取仓库名称和 reference (digest/tag)
func (r *Registry) extractNameAndReference(path string) (string, string) {
	// 路径格式：/v2/{name}/manifests/{reference}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "manifests" && i+1 < len(parts) {
			// 提取 name: /v2/{name}/manifests/...
			name := strings.Join(parts[2:i], "/")
			reference := parts[i+1]
			return name, reference
		}
	}
	return "", ""
}
