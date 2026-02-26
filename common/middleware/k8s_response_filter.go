package middleware

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type K8sResponseFilter struct {
	middleware.Abstract
	RemoveFields []string
}

func NewK8sResponseFilter() K8sResponseFilter {
	return K8sResponseFilter{
		RemoveFields: []string{
			"managedFields",
		},
	}
}

func (f K8sResponseFilter) Process(ctx *gin.Context) {
	if ctx.Request.Method != "GET" {
		ctx.Next()
		return
	}

	if !strings.Contains(ctx.GetHeader("Accept"), "application/json") {
		ctx.Next()
		return
	}

	blw := &bodyLogWriter{body: bytes.Buffer{}, ResponseWriter: ctx.Writer}
	ctx.Writer = blw
	ctx.Next()

	if blw.body.Len() > 0 {
		content := blw.body.Bytes()
		filtered := f.filterManagedFields(content)
		ctx.Data(blw.statusCode, "application/json", filtered)
	}
}

func (f K8sResponseFilter) filterManagedFields(content []byte) []byte {
	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return content
	}

	f.removeFieldRecursive(data, "managedFields")

	filtered, err := json.Marshal(data)
	if err != nil {
		return content
	}

	return filtered
}

func (f K8sResponseFilter) removeFieldRecursive(data interface{}, fieldName string) {
	switch v := data.(type) {
	case map[string]interface{}:
		delete(v, fieldName)
		for key := range v {
			f.removeFieldRecursive(v[key], fieldName)
		}
	case []interface{}:
		for i := range v {
			f.removeFieldRecursive(v[i], fieldName)
		}
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body       bytes.Buffer
	statusCode int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
