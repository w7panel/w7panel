package registry

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
)

func Import(ctx context.Context, client *containerd.Client, options types.ImageImportOptions) error {
	// cmd.Import()
	// image.Import()
	// container.Commit()
	return nil
}
