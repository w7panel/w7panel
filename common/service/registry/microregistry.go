package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	micro "github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
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

// ExportToTar 将 registry 中的镜像导出为 tar 文件
// repo: 仓库名称，例如 "localhost:5000/myimage"
// reference: 镜像引用 (tag 或 digest)
// tarPath: 输出的 tar 文件路径
func ExportToTar(ctx context.Context, blobsDir, repo, reference, tarPath string) error {
	img, err := loadImageFromRegistry(ctx, blobsDir, repo, reference)
	if err != nil {
		return fmt.Errorf("failed to load image from registry: %w", err)
	}

	ref, err := name.ParseReference(repo + ":" + reference)
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}

	if err := crane.Save(img, ref.Name(), tarPath); err != nil {
		return fmt.Errorf("failed to save image to tar: %w", err)
	}

	return nil
}

// ExportToTarWithDigest 使用 digest 导出镜像到 tar 文件
func ExportToTarWithDigest(ctx context.Context, blobsDir, repo, digest, tarPath string) error {
	img, err := loadImageFromRegistry(ctx, blobsDir, repo, digest)
	if err != nil {
		return fmt.Errorf("failed to load image from registry: %w", err)
	}

	ref, err := name.ParseReference(repo + "@" + digest)
	if err != nil {
		return fmt.Errorf("failed to parse digest reference: %w", err)
	}

	if err := crane.Save(img, ref.Name(), tarPath); err != nil {
		return fmt.Errorf("failed to save image to tar: %w", err)
	}

	return nil
}

// loadImageFromRegistry 从 registry 磁盘存储中加载镜像
func loadImageFromRegistry(ctx context.Context, blobsDir, repo, reference string) (v1.Image, error) {
	// 从 manifest 中获取镜像信息
	manifestPath := filepath.Join(blobsDir, "sha256")

	// 如果 reference 是 digest 格式 (sha256:xxx)
	var digestHex string
	if strings.HasPrefix(reference, "sha256:") {
		digestHex = strings.TrimPrefix(reference, "sha256:")
	} else {
		// 如果是 tag，需要查找对应的 digest
		// 这里简化处理，假设 tag 文件存在于 manifests 目录
		tagManifestPath := filepath.Join(blobsDir, "manifests", repo, reference)
		if _, err := os.Stat(tagManifestPath); err == nil {
			// 从 manifest 内容中提取 digest
			// 简化实现：直接返回错误，需要使用完整的 manifest 解析
			return nil, fmt.Errorf("tag resolution not implemented, please use digest")
		}
		return nil, fmt.Errorf("tag not found: %s", reference)
	}

	// 读取 manifest blob
	manifestBlobPath := filepath.Join(manifestPath, digestHex)
	if _, err := os.Stat(manifestBlobPath); err != nil {
		return nil, fmt.Errorf("manifest not found: %s", digestHex)
	}

	// 尝试作为 OCI layout 加载
	return loadImageFromOCILayout(blobsDir, digestHex)
}

// loadImageFromOCILayout 从 OCI layout 格式加载镜像
func loadImageFromOCILayout(blobsDir, digestHex string) (v1.Image, error) {
	// 尝试从 blobs 目录构建 OCI layout 结构
	layoutPath := filepath.Join(blobsDir, "oci-layout")
	if _, err := os.Stat(layoutPath); err == nil {
		// 存在 OCI layout，直接加载
		l, err := layout.ImageIndexFromPath(blobsDir)
		if err != nil {
			return nil, err
		}
		h := v1.Hash{
			Algorithm: "sha256",
			Hex:       digestHex,
		}
		return l.Image(h)
	}

	// 创建临时的 OCI layout 结构
	tmpDir, err := os.MkdirTemp("", "oci-layout-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 oci-layout 文件
	ociLayoutContent := `{"imageLayoutVersion":"1.0.0"}`
	if err := os.WriteFile(filepath.Join(tmpDir, "oci-layout"), []byte(ociLayoutContent), 0644); err != nil {
		return nil, err
	}

	// 创建 index.json
	indexJSON := fmt.Sprintf(`{
		"schemaVersion": 2,
		"manifests": [{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"digest": "sha256:%s",
			"size": 0
		}]
	}`, digestHex)

	indexPath := filepath.Join(tmpDir, "index.json")
	if err := os.WriteFile(indexPath, []byte(indexJSON), 0644); err != nil {
		return nil, err
	}

	// 复制 blobs
	blobsSource := filepath.Join(blobsDir, "sha256")
	blobsDest := filepath.Join(tmpDir, "blobs", "sha256")
	if err := os.MkdirAll(blobsDest, 0755); err != nil {
		return nil, err
	}

	// 复制所有需要的 blobs
	entries, err := os.ReadDir(blobsSource)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			src := filepath.Join(blobsSource, entry.Name())
			dst := filepath.Join(blobsDest, entry.Name())
			copyFile(src, dst)
		}
	}

	l, err := layout.ImageIndexFromPath(tmpDir)
	if err != nil {
		return nil, err
	}
	h := v1.Hash{
		Algorithm: "sha256",
		Hex:       digestHex,
	}
	return l.Image(h)
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// GetImageFromRegistry 从 registry 获取镜像对象（用于进一步处理）
func GetImageFromRegistry(ctx context.Context, blobsDir, repo, reference string) (v1.Image, error) {
	return loadImageFromRegistry(ctx, blobsDir, repo, reference)
}

// ExportToWriter 将镜像导出到 io.Writer（可用于 HTTP 响应）
func ExportToWriter(ctx context.Context, blobsDir, repo, reference string, w io.Writer) error {
	img, err := loadImageFromRegistry(ctx, blobsDir, repo, reference)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	// 使用 crane.Export 写入 tar 流
	if err := crane.Export(img, w); err != nil {
		return fmt.Errorf("failed to export image: %w", err)
	}

	return nil
}

// ListImages 列出 registry 中的所有镜像
func ListImages(blobsDir string) ([]string, error) {
	var images []string

	// 遍历 manifests 目录（如果存在）
	manifestsDir := filepath.Join(blobsDir, "manifests")
	if info, err := os.Stat(manifestsDir); err == nil && info.IsDir() {
		filepath.Walk(manifestsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				rel, err := filepath.Rel(manifestsDir, path)
				if err != nil {
					return nil
				}
				images = append(images, rel)
			}
			return nil
		})
	}

	// 或者遍历 blobs 目录查找 manifest
	shaDir := filepath.Join(blobsDir, "sha256")
	if info, err := os.Stat(shaDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(shaDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					images = append(images, "sha256:"+entry.Name())
				}
			}
		}
	}

	return images, nil
}
