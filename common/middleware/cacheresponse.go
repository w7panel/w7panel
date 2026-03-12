package middleware

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type CacheResponse struct {
	middleware.Abstract
}

// responseWriter 包装 gin.ResponseWriter 以捕获响应数据
type responseWriter2 struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func newResponseWriter(w gin.ResponseWriter) *responseWriter2 {
	return &responseWriter2{
		ResponseWriter: w,
		body:           bytes.NewBuffer(nil),
		status:         http.StatusOK,
	}
}

func (w *responseWriter2) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter2) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter2) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// CacheEntry 缓存条目
type CacheEntry struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	ContentType string
	CreatedAt   time.Time
	ExpireAt    time.Time
}

// IsExpired 检查缓存是否过期
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpireAt)
}

// GetCacheKey 生成缓存键
func GetCacheKey(c *gin.Context) string {
	return c.Request.URL.String()
}

// cacheStore 简单的内存缓存存储
var cacheStore = make(map[string]*CacheEntry)

// CacheResponseWithExpire 带过期时间的响应缓存中间件
// duration: 缓存过期时间
func CacheResponseWithExpire(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅缓存 GET 请求
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		cacheKey := GetCacheKey(c)

		// 检查缓存
		if entry, exists := cacheStore[cacheKey]; exists && !entry.IsExpired() {
			// 从缓存恢复响应
			for key, values := range entry.Headers {
				for _, value := range values {
					c.Writer.Header().Add(key, value)
				}
			}
			c.Writer.Header().Set("X-Cache-Status", "HIT")
			c.Writer.WriteHeader(entry.StatusCode)
			c.Writer.Write(entry.Body)
			c.Abort()
			return
		}

		// 创建响应包装器以捕获响应
		writer := newResponseWriter(c.Writer)
		c.Writer = writer

		// 执行后续处理
		c.Next()

		// 缓存成功的响应 (2xx 状态码)
		if writer.status >= http.StatusOK && writer.status < http.StatusMultipleChoices {
			body := writer.body.Bytes()
			cacheStore[cacheKey] = &CacheEntry{
				StatusCode:  writer.status,
				Headers:     c.Writer.Header().Clone(),
				Body:        body,
				ContentType: c.Writer.Header().Get("Content-Type"),
				CreatedAt:   time.Now(),
				ExpireAt:    time.Now().Add(duration),
			}
			c.Writer.Header().Set("X-Cache-Status", "MISS")
		}
	}
}

// CacheResponse 默认 5 分钟过期时间的响应缓存中间件
func (self CacheResponse) Process(c *gin.Context) {
	CacheResponseWithExpire(5 * time.Minute)(c)
}

// ClearCache 清除指定键的缓存
func ClearCache(key string) {
	delete(cacheStore, key)
}

// ClearAllCache 清除所有缓存
func ClearAllCache() {
	cacheStore = make(map[string]*CacheEntry)
}

// GetCacheKeys 获取所有缓存键
func GetCacheKeys() []string {
	keys := make([]string, 0, len(cacheStore))
	for k := range cacheStore {
		keys = append(keys, k)
	}
	return keys
}

// 确保 responseWriter 实现 gin.ResponseWriter 的所有必需方法
var _ gin.ResponseWriter = (*responseWriter2)(nil)

// 实现 Size 方法
func (w *responseWriter2) Size() int {
	return w.body.Len()
}

// 实现 Written 方法
func (w *responseWriter2) Written() bool {
	return w.body.Len() > 0
}

// 实现 WriteNow 方法
func (w *responseWriter2) WriteNow() {
}

// 实现 Close 方法
func (w *responseWriter2) Close() error {
	if closer, ok := w.ResponseWriter.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// 实现 Hijack 方法
func (w *responseWriter2) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(interface {
		Hijack() (net.Conn, *bufio.ReadWriter, error)
	}); ok {
		return hijacker.Hijack()
	}
	return nil, nil, nil
}
