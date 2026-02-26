package auth

import (
	"log/slog"
	"time"

	app "gitee.com/we7coreteam/k8s-offline/app/auth/console"
	controller2 "gitee.com/we7coreteam/k8s-offline/app/auth/http/controller"
	k3kController "gitee.com/we7coreteam/k8s-offline/app/k3k/http/controller"
	"gitee.com/we7coreteam/k8s-offline/common/middleware"
	console2 "gitee.com/we7coreteam/k8s-offline/common/service/console"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/console"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	httpserver "github.com/we7coreteam/w7-rangine-go/v2/src/http/server"
)

type Provider struct {
}

func (p Provider) Register(httpServer *httpserver.Server, console console.Console) {
	console.RegisterCommand(new(app.Register))
	console.RegisterCommand(new(app.Cluster))
	console.RegisterCommand(new(app.Site))
	console.RegisterCommand(new(app.CreateInnerDb))
	console.RegisterCommand(new(app.Unzip))

	p.RegisterHttpRoutes(httpServer)
	if facade.GetConfig().GetBool("site.enabled") {
		// slog.Info("site refresh token timer")
		// go p.RefreshCDToken() // 用户登录时候 触发刷新token
	}
}

func (p Provider) RegisterHttpRoutes(server *httpserver.Server) {
	server.RegisterRouters(func(engine *gin.Engine) {
		engine.POST("/panel-api/v1/login", controller2.Auth{}.Login)

		localApiGroup := engine.Group("/panel-api/v1/auth").Use(middleware.Cors{}.Process)
		{
			localApiGroup.POST("/login", controller2.Auth{}.Login)
			localApiGroup.POST("/register", controller2.Auth{}.Register)
			// localApiGroup.POST("/console/k3k-register", middleware.Auth{}.Process, controller2.Auth{}.RegisterUseUid)
			// localApiGroup.POST("/refresh-token", middleware.Auth{}.Process, controller2.Auth{}.RefreshToken) //废弃
			localApiGroup.POST("/refresh-token2", controller2.Auth{}.RefreshToken2)
			localApiGroup.POST("/init-user", controller2.Auth{}.InitUser)
			localApiGroup.POST("/reset-password", middleware.Auth{}.Process, controller2.Auth{}.ResetPassword)
			localApiGroup.POST("/reset-password-current", middleware.Auth{}.Process, controller2.Auth{}.ResetPasswordCurrent) //设置当前登录用户密码

			localApiGroup.GET("/console/oauth", controller2.Console{}.Redirect)

			localApiGroup.GET("/console/login", controller2.Auth{}.ConsoleLogin)
			localApiGroup.GET("/console/bind", middleware.Auth{}.Process /*middleware.BindConsole{}.Process, middleware.Proxy{}.Process, */, controller2.Console{}.BindConsole)
			localApiGroup.GET("/console/info", middleware.Auth{}.Process, controller2.Console{}.Info)
			localApiGroup.GET("/userinfo", middleware.Auth{}.Process, k3kController.K3k{}.Info)
			// 不需要创始人权限
			localApiGroup.GET("/console/code/:code", middleware.Auth{}.Process, controller2.Console{}.ProxyCouponCode)
			localApiGroup.Any("/console/proxy/*path", middleware.NewAuth("founder").Process, controller2.Console{}.Proxy)

			localApiGroup.POST("/console/register-to-console", middleware.Auth{}.Process, controller2.Console{}.RegisterToConsole) //不能proxy 需要root kubeconfig
			//不能proxy 需要root kubeconfig
			localApiGroup.POST("/console/thirdparty-cd-token", middleware.Auth{}.Process, controller2.Console{}.ThirdPartyCDToken)
			localApiGroup.POST("/console/import-cert", middleware.Auth{}.Process /*middleware.Proxy{}.Process, */, controller2.Console{}.ImportCert)
			localApiGroup.POST("/console/verify-cert", middleware.Auth{}.Process /*middleware.Proxy{}.Process, */, controller2.Console{}.VerifyCert)
			localApiGroup.POST("/console/import-cert-console", middleware.Auth{}.Process /*middleware.Proxy{}.Process, */, controller2.Console{}.ImportCertConsole)
		}
	})
}

func (p Provider) RefreshCDToken() {
	// 一个1分钟的定时器 定时执行console.RefreshToken方法
	tokenResolution := facade.Config.GetDuration("site.token_refresh_resolution")
	time := time.NewTicker(tokenResolution)

	// go console2.RefreshCDToken()

	go func() {
		for range time.C {
			// 刷新token
			slog.Info("刷新token")

			err := console2.VerifyDefaultLicense(true)
			if err != nil {
				slog.Error("刷新license失败", "err", err)
			}
			// err = console2.ReVerifyLicense(sdk)
			// if err != nil {
			// 	slog.Error("刷新license失败", "err", err)
			// }
		}
	}()

}
