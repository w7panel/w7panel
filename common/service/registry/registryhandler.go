package registry

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RegistryHandler struct {
	memoryHandler http.Handler
	spegelHandler http.Handler
	uploadStatus  sync.Map // map[string]bool - tracks upload completion by digest
}

func InitReigstry(ctx context.Context) (*RegistryHandler, error) {
	memory := CreateMicroRegistry()
	reg, err := CreateSpegelRegistry(context.Background())
	if err != nil {
		slog.Error("create reg registry err", "err", err)
		return nil, err
	}
	logrLogger := log.FromContext(ctx)
	return NewRegistryHandler(memory, reg.Handler(logrLogger)), nil
}

func NewRegistryHandler(pre, next http.Handler) *RegistryHandler {
	return &RegistryHandler{memoryHandler: pre, spegelHandler: next}
}

// bufferResponseWriter 用于缓冲响应，避免污染原始 ResponseWriter
type bufferResponseWriter struct {
	http.ResponseWriter
	buffer  *bytes.Buffer
	status  int
	written bool
	headers http.Header
}

func newBufferResponseWriter(rw http.ResponseWriter) *bufferResponseWriter {
	return &bufferResponseWriter{
		ResponseWriter: rw,
		buffer:         &bytes.Buffer{},
		headers:        make(http.Header),
		status:         http.StatusOK,
	}
}

func (brw *bufferResponseWriter) Header() http.Header {
	return brw.headers
}

func (brw *bufferResponseWriter) Write(data []byte) (int, error) {
	brw.written = true
	return brw.buffer.Write(data)
}

func (brw *bufferResponseWriter) WriteHeader(statusCode int) {
	brw.status = statusCode
}

func (brw *bufferResponseWriter) commit() {
	// 将缓冲的 header 复制到真实 ResponseWriter
	for k, vv := range brw.headers {
		for _, v := range vv {
			brw.ResponseWriter.Header().Add(k, v)
		}
	}
	brw.ResponseWriter.WriteHeader(brw.status)
	brw.ResponseWriter.Write(brw.buffer.Bytes())
}

func (r *RegistryHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	values.Set("ns", defaultDamain) //默认仓库域名
	req.URL.RawQuery = values.Encode()
	if isReadOnly(req.Method) {
		// 只读请求：先尝试 memoryHandler，如果返回非 200 则用 spegelHandler
		brw := newBufferResponseWriter(rw)
		r.memoryHandler.ServeHTTP(brw, req)

		if brw.status != http.StatusOK {
			// memoryHandler 未成功处理，使用 spegelHandler
			r.spegelHandler.ServeHTTP(rw, req)
			return
		}

		// memoryHandler 成功处理，提交响应
		brw.commit()
		return
	}

	// 修改请求：只用 memoryHandler 处理
	brw := newBufferResponseWriter(rw)
	r.memoryHandler.ServeHTTP(brw, req)

	brw.commit()
	// 检查是否是 PUT /v2/{name}/manifests/{reference} 请求，判断上传是否完成
	if r.isPutManifestRequest(req, brw.status) {
		name, reference := r.extractNameAndReference(req.URL.Path)
		if name != "" && reference != "" {
			// 上传完成，ctr import
			// ref, err := parseRef(reference)
			// if err != nil {
			// 	slog.Warn("parse ref err", "err", err)
			// 	return
			// }
			// if ref.
			sourceUrl := "127.0.0.1:8000/" + name + ":" + reference
			targeUrl := defaultDamain + "/" + name + ":" + reference
			go func() {
				err := PullToContainerD(context.TODO(), sourceUrl, targeUrl)
				if err != nil {
					slog.Warn("pull to containerd err", "err", err)
				}
			}()
		}
	}

}

// isPutManifestRequest 检查是否是 PUT manifest 请求且成功
func (r *RegistryHandler) isPutManifestRequest(req *http.Request, status int) bool {
	// Docker Registry API: PUT /v2/{name}/manifests/{reference}
	return req.Method == http.MethodPut &&
		strings.Contains(req.URL.Path, "/manifests/") &&
		status == http.StatusCreated
}

// extractNameAndReference 从路径中提取仓库名称和 reference (digest/tag)
func (r *RegistryHandler) extractNameAndReference(path string) (string, string) {
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

// getManifestURL 获取上传镜像 manifest 地址
func (r *RegistryHandler) getManifestURL(req *http.Request, name, reference string) string {
	baseURL := ""
	if req.Host != "" {
		baseURL = fmt.Sprintf("http://%s", req.Host)
	}
	return fmt.Sprintf("%s/v2/%s/manifests/%s", baseURL, name, reference)
}

// IsUploadComplete 检查指定 reference 的上传是否完成
func (r *RegistryHandler) IsUploadComplete(reference string) bool {
	completed, ok := r.uploadStatus.Load(reference)
	if !ok {
		return false
	}
	return completed.(bool)
}

// GetManifestURL 获取已上传镜像的 manifest 地址
func (r *RegistryHandler) GetManifestURL(name, reference string) string {
	if completed, ok := r.uploadStatus.Load(reference); !ok || !completed.(bool) {
		return ""
	}
	// 返回完整的 manifest URL
	return fmt.Sprintf("/v2/%s/manifests/%s", name, reference)
}

func isReadOnly(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
