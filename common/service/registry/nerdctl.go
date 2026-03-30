package registry

import (
	"context"
	"os"

	// "github.com/containerd/containerd"
	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/cmd/container"
	"github.com/containerd/nerdctl/v2/pkg/cmd/image"
	"github.com/containerd/nerdctl/v2/pkg/imgutil/commit"
	"github.com/containerd/nerdctl/v2/pkg/referenceutil"
	"github.com/opencontainers/go-digest"
)

// commit a image
func Commit(ctx context.Context, client *containerd.Client, rawRef string, req string, options types.ContainerCommitOptions) error {
	return container.Commit(ctx, client, rawRef, req, options)
}

func CommitOne(ctx context.Context, client *containerd.Client, rawRef string, req string, options types.ContainerCommitOptions) (digest.Digest, error) {
	parsedReference, err := referenceutil.Parse(rawRef)
	if err != nil {
		return "", err
	}

	// changes, err := parseChanges(options.Change)
	// if err != nil {
	// 	return err
	// }

	opts := &commit.Opts{
		Author:             options.Author,
		Message:            options.Message,
		Ref:                parsedReference.String(),
		Pause:              options.Pause,
		Changes:            commit.Changes{},
		Compression:        options.Compression,
		Format:             options.Format,
		EstargzOptions:     options.EstargzOptions,
		ZstdChunkedOptions: options.ZstdChunkedOptions,
	}

	ctn, err := client.LoadContainer(ctx, req)
	if err != nil {
		return "", err
	}

	imageID, err := commit.Commit(ctx, client, ctn, opts, options.GOptions)
	if err != nil {
		return "", err
	}
	return imageID, nil
}

func Pull(ctx context.Context, client *containerd.Client, rawRef string, options types.ImagePullOptions) error {
	return image.Pull(ctx, client, rawRef, options)
}

func CommitToContainerD(ctx context.Context, rawRef, containerId string) (digest.Digest, error) {
	client, err := CreateClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	return CommitOne(ctx, client, rawRef, containerId, types.ContainerCommitOptions{
		GOptions: types.GlobalCommandOptions{DataRoot: "/tmp", Address: containerAddr()},
	})
}
func PullToContainerD(ctx context.Context, rawRef string) error {
	client, err := containerd.New(containerAddr(), containerd.WithDefaultNamespace(registryNamespace))
	if err != nil {
		return err
	}
	defer client.Close()
	return Pull(ctx, client, rawRef, types.ImagePullOptions{
		// Std
		Stdout:                 os.Stdout,
		Stderr:                 os.Stderr,
		ProgressOutputToStdout: true,
		Mode:                   "always",
		GOptions: types.GlobalCommandOptions{
			Namespace: registryNamespace,
		},
	})
}
