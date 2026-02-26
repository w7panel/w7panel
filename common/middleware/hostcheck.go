package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type HostCheck struct {
	middleware.Abstract
}

func (self HostCheck) Process(c *gin.Context) {
	// 移除域名限制，允许任意域名/IP访问（用于K8s容器部署）
	c.Next()
}
