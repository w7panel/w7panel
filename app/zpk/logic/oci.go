package logic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type oci struct {
	*remote.Repository
	ociPath string
	sdk     *k8s.Sdk
}

const (
	MediaTypeIcon       = "application/vnd.w7.formula.icon+png"
	MediaTypeFilesJson  = "application/vnd.w7.formula.files.json+json"
	MediaTypeCodeZip    = "application/vnd.w7.formula.code.zip+zip"
	MediaTypeWebCodeZip = "application/vnd.w7.formula.code.web.zip+zip"
)

func MediaToRealType(mtype string) string {
	switch mtype {
	case "icon":
		return MediaTypeIcon
	case "files":
		return MediaTypeFilesJson
	case "codezip":
		return MediaTypeCodeZip
	case "webzip":
		return MediaTypeWebCodeZip
	default:
		return mtype
	}
}

func NewOCI(sdk *k8s.Sdk, ociPath string) (*oci, error) {
	originPath := ociPath
	ociPath = strings.Replace(ociPath, "/oci://", "oci://", 1)
	ociPath = strings.Replace(ociPath, "oci://", "", 1)
	repo, err := remote.NewRepository(ociPath)
	if err != nil {
		return nil, err
	}
	return &oci{
		ociPath:    originPath,
		Repository: repo,
		sdk:        sdk,
	}, nil
}
func (o *oci) GetCodeZipUrl() string {
	return helper.SelfReqUrl() + "api/v1/zpk/oci/down/" + o.ociPath + "?mediaType=codezip"
}

func (o *oci) GetWebCodeZipUrl() string {
	return helper.SelfReqUrl() + "api/v1/zpk/oci/down/" + o.ociPath + "?mediaType=webzip"
}

func (o *oci) GetIconUrl() string {
	return helper.SelfReqUrl() + "api/v1/zpk/oci/down/" + o.ociPath + "?mediaType=" + MediaTypeIcon
}

// DockerConfigJSON represents the structure of a .dockerconfigjson file
type DockerConfigJSON struct {
	Auths map[string]DockerAuthConfig `json:"auths"`
}

// DockerAuthConfig contains authorization information for connecting to a registry
type DockerAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth,omitempty"`
}

func (o *oci) GetCredential(registry string, namespace string) (*auth.Credential, error) {
	if namespace == "" {
		namespace = o.sdk.GetNamespace()
	}

	// 列出指定命名空间中的所有 Secret
	secretList, err := o.sdk.ClientSet.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Secret 列表失败: %w", err)
	}

	// 遍历所有 Secret
	for _, secret := range secretList.Items {
		// 只处理 docker-registry 类型的 Secret
		if secret.Type != "kubernetes.io/dockerconfigjson" {
			continue
		}

		// 获取 .dockerconfigjson 数据
		dockerConfigJSON, exists := secret.Data[".dockerconfigjson"]
		if !exists || dockerConfigJSON == nil {
			continue
		}

		// 解析 docker 配置
		var config DockerConfigJSON
		if err := json.Unmarshal(dockerConfigJSON, &config); err != nil {
			slog.Warn("解析 docker 配置失败", "secret", secret.Name, "error", err)
			continue
		}

		// 查找指定 registry 的认证信息
		if authConfig, exists := config.Auths[registry]; exists {
			if authConfig.Username == "" || authConfig.Password == "" {
				// 如果没有明确的用户名和密码，但有 auth 字段，尝试解码
				if authConfig.Auth != "" {
					username, password, err := decodeDockerAuth(authConfig.Auth)
					if err != nil {
						slog.Warn("解码 auth 字段失败", "error", err)
						continue
					}
					authConfig.Username = username
					authConfig.Password = password
				} else {
					continue
				}
			}

			return &auth.Credential{
				Username: authConfig.Username,
				Password: authConfig.Password,
			}, nil
		}
	}

	// 如果没有找到匹配的认证信息，返回特定错误
	return nil, fmt.Errorf("未找到 registry %s 的认证信息", registry)
}

// decodeDockerAuth 解码 base64 编码的 auth 字段
func decodeDockerAuth(auth string) (username, password string, err error) {

	decoded, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", "", fmt.Errorf("base64 解码失败: %w", err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("无效的 auth 格式")
	}

	return parts[0], parts[1], nil
}

// FetchArtifact 从远程OCI注册表获取指定类型的文件
func (o *oci) Download(ctx context.Context, mediaType string, outputPath string) error {
	// 创建远程仓库客户端
	repo := o.Repository
	mediaType = MediaToRealType(mediaType)

	creden, err := o.GetCredential(repo.Reference.Registry, "")
	if err != nil {
		slog.Error("获取认证失败", "err", err)
	}
	if creden != nil {
		repo.Client = &auth.Client{
			Cache:      auth.NewCache(),
			Credential: auth.StaticCredential(repo.Reference.Registry, *creden),
		}
	}
	// 获取清单
	_, rc, err := repo.FetchReference(ctx, repo.Reference.Reference)
	if err != nil {
		return fmt.Errorf("获取引用失败: %w", err)
	}
	defer rc.Close()

	// 读取清单内容
	manifestBytes, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("读取清单内容失败: %w", err)
	}

	// 解析清单
	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return fmt.Errorf("解析清单失败: %w", err)
	}

	// 查找指定媒体类型的层
	var layerDesc *ocispec.Descriptor
	for _, layer := range manifest.Layers {
		if layer.MediaType == mediaType {
			layerDesc = &layer
			break
		}
	}

	if layerDesc == nil {
		return fmt.Errorf("未找到媒体类型为 %s 的层", mediaType)
	}

	// 获取层内容
	layerRC, err := repo.Fetch(ctx, *layerDesc)
	if err != nil {
		return fmt.Errorf("获取层内容失败: %w", err)
	}
	defer layerRC.Close()

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outFile.Close()

	// 将层内容写入文件
	if _, err := io.Copy(outFile, layerRC); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
