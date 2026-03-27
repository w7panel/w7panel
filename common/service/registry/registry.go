package registry

import (
	"context"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/spegel-org/spegel/pkg/oci"
	sreg "github.com/spegel-org/spegel/pkg/registry"
	"github.com/spegel-org/spegel/pkg/routing"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	registryNamespace = "k8s.io"
	// containerdRoot    = "/run/k3s/containerd"
	// containerdAddr    = "/run/k3s/containerd/containerd.sock"

	containerdRoot = "/run/containerd"
	containerdAddr = "/run/containerd/containerd.sock"
)

//	type Router interface {
//		// Ready returns true when the router is ready.
//		Ready(ctx context.Context) (bool, error)
//		// Lookup discovers peers with the given key and returns a balancer with the peers.
//		Lookup(ctx context.Context, key string, count int) (routing.Balancer, error)
//		// Advertise broadcasts the availability of the given keys.
//		Advertise(ctx context.Context, keys []string) error
//		// Withdraw stops the broadcasting the availability of the given keys to the network.
//		Withdraw(ctx context.Context, keys []string) error
//	}
type MockRouter struct {
}

var sregRegistry *sreg.Registry
var ContainerDRegistryHandler http.Handler

func init() {
	reg, err := CreateSpegelRegistry(context.Background())
	if err != nil {
		slog.Error("create reg registry err", "err", err)
	}
	sregRegistry = reg
	if sregRegistry != nil {
		logrLogger := log.FromContext(context.Background())
		ContainerDRegistryHandler = sregRegistry.Handler(logrLogger)
	}
}

func NewEmptyRouter() *MockRouter {
	return &MockRouter{}
}
func (m *MockRouter) Ready(ctx context.Context) (bool, error) {
	return true, nil
}

func (m *MockRouter) Lookup(ctx context.Context, key string, count int) (routing.Balancer, error) {
	return nil, errors.Errorf("not implement")
}

func (m *MockRouter) Advertise(ctx context.Context, keys []string) error {
	return nil
}

func (m *MockRouter) Withdraw(ctx context.Context, keys []string) error {
	return nil
}

func newOciStore(ctx context.Context, sock, namespace string, opts ...oci.ContainerdOption) (*oci.Containerd, error) {
	return oci.NewContainerd(ctx, sock, namespace, opts...)
}
func CreateSpegelRegistry(ctx context.Context) (*sreg.Registry, error) {

	ociClient, err := oci.NewClient()
	if err != nil {
		return nil, err
	}
	storeOpts := []oci.ContainerdOption{
		oci.WithContentPath(filepath.Join(containerdRoot, "io.containerd.content.v1.content")),
	}
	ociStore, err := newOciStore(ctx, containerdAddr, registryNamespace, storeOpts...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create OCI store")
	}

	registryOpts := []sreg.RegistryOption{
		// registry.WithRegistryFilters(filters),
		sreg.WithOCIClient(ociClient),
	}
	router := NewEmptyRouter()
	reg, err := sreg.NewRegistry(ociStore, router, registryOpts...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create embedded registry")
	}
	return reg, err
}

func spegelRegistryHandler(ctx context.Context) http.Handler {
	logrLogger := log.FromContext(ctx)
	return sregRegistry.Handler(logrLogger)
}
