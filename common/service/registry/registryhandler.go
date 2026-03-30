package registry

import (
	"bytes"
	"net/http"
)

type RegistryHandler struct {
	memoryHandler http.Handler
	spegelHandler http.Handler
}

func NewRegistry() *RegistryHandler {
	return &RegistryHandler{}
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
	r.memoryHandler.ServeHTTP(rw, req)
}

func isReadOnly(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
