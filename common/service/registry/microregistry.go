package registry

import (
	"net/http"
	"os"

	micro "github.com/google/go-containerregistry/pkg/registry"
)

func CreateMicroRegistry() http.Handler {
	blobsDir := os.Getenv("REGISTRY_BLOBS_DIR")
	if blobsDir == "" {
		blobsDir = os.TempDir() + "/blobs"
	}
	info, err := os.Stat(blobsDir)
	if err != nil || !info.IsDir() {
		os.MkdirAll(blobsDir, 0755)
	}
	blobsHandler := micro.NewDiskBlobHandler(blobsDir)
	return micro.New(micro.WithBlobHandler(blobsHandler))
}
