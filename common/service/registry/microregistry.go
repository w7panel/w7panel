package registry

import (
	"net/http"

	micro "github.com/google/go-containerregistry/pkg/registry"
)

func CreateMicroRegistry() http.Handler {
	return micro.New(micro.WithBlobHandler(micro.NewInMemoryBlobHandler()))
}
