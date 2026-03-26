package metrics

import (

	// "sync"

	"github.com/gin-gonic/gin"
	console2 "github.com/w7panel/w7panel/app/metrics/console"
	controller2 "github.com/w7panel/w7panel/app/metrics/http/controller"
	"github.com/w7panel/w7panel/common/middleware"
	"github.com/w7panel/w7panel/common/service/k8s/metrics"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/console"
	httpserver "github.com/we7coreteam/w7-rangine-go/v2/src/http/server"
	_ "modernc.org/sqlite"
)

type Provider struct {
}

func (p Provider) Register(httpServer *httpserver.Server, console console.Console) {
	console.RegisterCommand(new(console2.MetricsCgroup))
	p.RegisterHttpRoutes(httpServer)
	p.startMetricsTicker()

}

func (p Provider) RegisterHttpRoutes(server *httpserver.Server) {
	server.RegisterRouters(func(engine *gin.Engine) {
		// engine.Any("/metrics", controller2.Metrics{}.Promhttp)

		engine.GET("/panel-api/v1/metrics/usage/normal", middleware.Auth{}.Process, controller2.Metrics{}.Usage)
		engine.GET("/panel-api/v1/metrics/usage/disk", middleware.Auth{}.Process, controller2.Metrics{}.UsageDisk)

		engine.GET("/panel-api/v1/metrics/installed", middleware.Auth{}.Process, controller2.Metrics{}.VmOperatorInstalled)
		engine.GET("/panel-api/v1/metrics/state", middleware.Auth{}.Process, controller2.Metrics{}.MetricsState)
	})
}

func (p Provider) startMetricsTicker() {
	// go metrics.Start()
	go metrics.StartNodeMetrics()
	go metrics.StartPodMetrics()
	go metrics.StartCroupMetrics() // 启动cgroup监控
}
