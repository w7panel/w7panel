package registry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func PushOciProxy() error {
	kodatapath, ok := os.LookupEnv("KO_DATA_PATH")
	if !ok {
		slog.Error("KO_DATA_PATH not set")
		return fmt.Errorf("KO_DATA_PATH not set")
	}
	dir := fmt.Sprintf("%s/registry/proxy-2.0.5", kodatapath)
	option := crane.WithContext(context.Background())
	options := []crane.Option{option}
	o := crane.GetOptions(options...)
	crane.Insecure(&o)
	return push(dir, "localhost:5000/oci-proxy:2.0.5", false, &o)
}

func push(path, tag string, index bool, o *crane.Options) error {
	imageRefs := ""
	img, err := loadImage(path, index)
	if err != nil {
		return err
	}

	ref, err := name.ParseReference(tag, o.Name...)
	if err != nil {
		return err
	}
	var h v1.Hash
	switch t := img.(type) {
	case v1.Image:
		if err := remote.Write(ref, t, o.Remote...); err != nil {
			return err
		}
		if h, err = t.Digest(); err != nil {
			return err
		}
	case v1.ImageIndex:
		if err := remote.WriteIndex(ref, t, o.Remote...); err != nil {
			return err
		}
		if h, err = t.Digest(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot push type (%T) to registry", img)
	}

	digest := ref.Context().Digest(h.String())
	if imageRefs != "" {
		if err := os.WriteFile(imageRefs, []byte(digest.String()), 0600); err != nil {
			return fmt.Errorf("failed to write image refs to %s: %w", imageRefs, err)
		}
	}
	// panic("unimplemented")
	slog.Info("pushed", "digest", digest.String())
	// Print the digest of the pushed image to stdout to facilitate command composition.
	// fmt.Fprintln(cmd.OutOrStdout(), digest)

	return nil
}

func loadImage(path string, index bool) (partial.WithRawManifest, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		img, err := crane.Load(path)
		if err != nil {
			return nil, fmt.Errorf("loading %s as tarball: %w", path, err)
		}
		return img, nil
	}

	l, err := layout.ImageIndexFromPath(path)
	if err != nil {
		return nil, fmt.Errorf("loading %s as OCI layout: %w", path, err)
	}

	if index {
		return l, nil
	}

	m, err := l.IndexManifest()
	if err != nil {
		return nil, err
	}
	if len(m.Manifests) != 1 {
		return nil, fmt.Errorf("layout contains %d entries, consider --index", len(m.Manifests))
	}

	desc := m.Manifests[0]
	if desc.MediaType.IsImage() {
		return l.Image(desc.Digest)
	} else if desc.MediaType.IsIndex() {
		return l.ImageIndex(desc.Digest)
	}

	return nil, fmt.Errorf("layout contains non-image (mediaType: %q), consider --index", desc.MediaType)
}
