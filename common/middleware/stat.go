package middleware

import (
	"github.com/gin-gonic/gin"

	// "github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"

	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type Stat struct {
	middleware.Abstract
}

func (self Stat) Process(c *gin.Context) {
	// 打印请求前的内存和请求后的内存

	// requestPath := c.Request.URL.Path

	// 请求前：获取当前的 Goroutine 数量和内存占用
	// beforeGoroutines := runtime.NumGoroutine()
	// var beforeMemStats runtime.MemStats
	// runtime.ReadMemStats(&beforeMemStats)

	// 打印请求前的信息
	// log.Printf("Before Request: Path: %s, Goroutines: %d, Memory Alloc: %d MB\n", requestPath, beforeGoroutines, beforeMemStats.Alloc/1024/1024)

	// 处理请求
	c.Next()

	// 请求后：获取当前的 Goroutine 数量和内存占用
	// afterGoroutines := runtime.NumGoroutine()
	// var afterMemStats runtime.MemStats
	// runtime.ReadMemStats(&afterMemStats)

	// 打印请求后的信息
	// log.Printf("After Request: Path: %s, Goroutines: %d, Memory Alloc: %d MB\n", requestPath, afterGoroutines, afterMemStats.Alloc/1024/1024)
	// log.Printf("After Request: Path: %s, Goroutines: %d, Memory Sys: %d MB\n", requestPath, afterGoroutines, afterMemStats.Sys/1024/1024)
	// log.Printf("After Request: Path: %s, Goroutines: %d, Memory HeapSys: %d MB\n", requestPath, afterGoroutines, afterMemStats.HeapSys/1024/1024)
	// log.Printf("After Request: Path: %s, Goroutines: %d, Memory HeapAlloc: %d MB\n", requestPath, afterGoroutines, afterMemStats.HeapAlloc/1024/1024)
	// log.Printf("After Request: Path: %s, Goroutines: %d, Memory HeapReleased: %d MB\n", requestPath, afterGoroutines, afterMemStats.HeapReleased/1024/1024)
	// log.Printf("After Request: Path: %s, Goroutines: %d, Memory HeapIdle: %d MB\n", requestPath, afterGoroutines, afterMemStats.HeapIdle/1024/1024)
	// log.Printf("After Request: Path: %s, Goroutines: %d, Memory HeapInuse: %d MB\n", requestPath, afterGoroutines, afterMemStats.HeapInuse/1024/1024)
	// 打印内存变化
	// memDiff := afterMemStats.Alloc - beforeMemStats.Alloc
	// log.Printf("Memory Change: Path: %s, Change: %d MB\n", requestPath, memDiff/1024/1024)
}
