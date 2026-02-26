package k3k

import (
	consoleShell "gitee.com/we7coreteam/k8s-offline/app/k3k/console"
	controller2 "gitee.com/we7coreteam/k8s-offline/app/k3k/http/controller"
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/middleware"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/console"
	httpserver "github.com/we7coreteam/w7-rangine-go/v2/src/http/server"
)

type Provider struct {
}

func (p Provider) Register(httpServer *httpserver.Server, console console.Console) {

	console.RegisterCommand(new(consoleShell.ClusterUpgrade))
	console.RegisterCommand(new(consoleShell.QxUpgrade))
	console.RegisterCommand(new(consoleShell.K3kOrderReturnCheck))    //处理有退款记录的用户
	console.RegisterCommand(new(consoleShell.K3kOrderReturnCheckOne)) //处理有退款记录的用户one
	p.RegisterHttpRoutes(httpServer)
	// if facade.Config.GetBool("k3k.watch") {
	// 	go k3kapi.Watch()
	// }
	if helper.IsChildAgent() {
		go k3k.SyncMicroApp()
	}
}

func (p Provider) RegisterHttpRoutes(server *httpserver.Server) {

	server.RegisterRouters(func(engine *gin.Engine) {
		k3kGroup := engine.Group("/panel-api/v1/k3k") //.Use(middleware.Cors{}.Process)
		{
			k3kGroup.GET("/info", middleware.Auth{}.Process, controller2.K3k{}.Info)                        // 登录信息
			k3kGroup.POST("/init", middleware.Auth{}.Process, controller2.K3k{}.ReInitCluster)              // 初始化集群
			k3kGroup.POST("/init-cluster", middleware.Auth{}.Process, controller2.K3k{}.ReInitClusterSuper) // 创始人初始化集群
			k3kGroup.POST("/sync-ingress", controller2.K3k{}.SyncIngress)                                   //
			k3kGroup.POST("/sync-configmap", controller2.K3k{}.SyncConfigmap)                               //
			k3kGroup.POST("/sync-mcpbridge", controller2.K3k{}.SyncMcpBridge)                               //
			k3kGroup.POST("/sync-secret", controller2.K3k{}.SyncSecret)                                     //
			k3kGroup.POST("/sync-down-static", controller2.K3k{}.SyncDownStatic)                            //
			k3kGroup.POST("/sync-microapp", controller2.K3k{}.SyncMicroApp)                                 //microapp同步到子集群
			k3kGroup.POST("/login", middleware.Auth{}.Process, controller2.K3k{}.Login)                     //
			k3kGroup.POST("/wh", middleware.Auth{}.Process, controller2.K3k{}.WhMoshi)                      // 维护模式 切换

			k3kGroup.GET("/order/config", middleware.Auth{}.Process, controller2.Order{}.GetConfig) // 获取配置
			k3kGroup.GET("/order/price", middleware.Auth{}.Process, controller2.Order{}.GetPrice)   // 获取当前价格

			k3kGroup.POST("/order/base", middleware.Auth{}.Process, controller2.Order{}.CreateBaseResourceOrder) // 创建基础资源订单

			k3kGroup.POST("/order/renew", middleware.Auth{}.Process, controller2.Order{}.CreateRenewOrder)   // 创建续费订单
			k3kGroup.POST("/order/expand", middleware.Auth{}.Process, controller2.Order{}.CreateExpandOrder) // 创建续费订单
			k3kGroup.POST("/order/notify", controller2.Order{}.OrderNotify)                                  // 支付回调 不需要登录
			k3kGroup.POST("/order/refresh", middleware.Auth{}.Process, controller2.Order{}.Refresh)          // 当前用户拉取支付状态

			k3kGroup.POST("/order/license", middleware.Auth{}.Process, controller2.Order{}.CreateLicenseOrder) // 面板授权购买

			k3kGroup.POST("/storage/resize", middleware.Auth{}.Process, controller2.K3k{}.ResizeSysStorage) // 扩容系统存储

		}

		k3kGroup1 := engine.Group("/panel-api/v1/k3k/overselling") //.Use(middleware.Cors{}.Process)
		{
			k3kGroup1.GET("/config", middleware.Auth{}.Process, controller2.OverSelling{}.OverSellingConfig)         // 获取超卖配置
			k3kGroup1.GET("/current-resource", middleware.Auth{}.Process, controller2.OverSelling{}.CurrentResource) // 获取超卖百分比*当前集群资源
			k3kGroup1.POST("/check", middleware.Auth{}.Process, controller2.OverSelling{}.CheckResource)             // 检查是否超出集群配置
		}

		k8kGroup := engine.Group("/panel-api/v1") //.Use(middleware.Cors{}.Process)
		{
			k8kGroup.GET("/userinfo", middleware.Auth{}.Process, controller2.K3k{}.Info) // 登录信息
			k8kGroup.GET("/idc-list", controller2.K3k{}.IdcResource)                     // IDC资源列表
		}
	})
}
