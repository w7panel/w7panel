package zpk

import (
	consolezpk "gitee.com/we7coreteam/k8s-offline/app/zpk/console"
	controller "gitee.com/we7coreteam/k8s-offline/app/zpk/http"
	"gitee.com/we7coreteam/k8s-offline/common/middleware"
	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/console"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	httpserver "github.com/we7coreteam/w7-rangine-go/v2/src/http/server"
)

type Provider struct {
}

func (p Provider) Register(httpServer *httpserver.Server, console console.Console) {

	p.RegisterValidateRule()
	p.RegisterHttpRoutes(httpServer)

	console.RegisterCommand(new(consolezpk.HelmCmd))
	console.RegisterCommand(new(consolezpk.MetricsUpgrade))

}

func (p Provider) RegisterValidateRule() {
	if v, ok := facade.GetValidator().Engine().(*validator.Validate); ok {
		v.RegisterValidation("id", func(fl validator.FieldLevel) bool {
			if id, ok := fl.Field().Interface().(uint); ok {
				if id > 0 {
					return true
				}
			}

			return false
		})

		v.RegisterValidation("page", func(fl validator.FieldLevel) bool {
			if page, ok := fl.Field().Interface().(uint); ok {
				if page > 0 {
					return true
				}
			}

			return false
		})
		v.RegisterTranslation("page", facade.GetTranslator(), func(ut ut.Translator) error {
			return ut.Add("page", "{0} 格式错误", true) // see universal-translator for details
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("page", fe.Field())
			return t
		})
	}
}

func (p Provider) RegisterHttpRoutes(server *httpserver.Server) {
	server.RegisterRouters(func(engine *gin.Engine) {

		localApiGroup := engine.Group("/panel-api/v1/zpk").Use(middleware.Cors{}.Process)
		{
			localApiGroup.GET("/config", middleware.Auth{}.Process, controller.Zpk{}.GetConfig)                      //ManifestPackge RequireLimit 判断是否共享环境
			localApiGroup.GET("/", middleware.Auth{}.Process, controller.Zpk{}.List)                                 //安装列表
			localApiGroup.PUT("/install", middleware.Auth{}.Process, controller.Zpk{}.Install)                       // 安装或更新
			localApiGroup.GET("/upgrade-info", middleware.Auth{}.Process, controller.Zpk{}.UpgradeInfo)              // 更新信息
			localApiGroup.Any("/build-image-success", middleware.Auth{}.Process, controller.Zpk{}.BuildImageSuccess) // 卸载插件
			localApiGroup.GET("/trandition/env", middleware.Auth{}.Process, controller.Zpk{}.TranditionList)         // 传统应用环境
			localApiGroup.POST("/trandition/install", middleware.Auth{}.Process, controller.Zpk{}.InstallTrandition) // 传统应用安装
			localApiGroup.GET("/out-depends/env", middleware.Auth{}.Process, controller.Zpk{}.OutDependEnv)          // 外部依赖环境变量
			localApiGroup.POST("/helm/memory", middleware.Auth{}.Process, controller.Zpk{}.GenHelmMemory)            // 外部依赖环境变量
			localApiGroup.GET("/helm/chart-yaml", middleware.Auth{}.Process, controller.Zpk{}.ChartYaml)             // 获取chart.yaml
			localApiGroup.GET("/last-version/env", middleware.Auth{}.Process, controller.Zpk{}.LastVersionEnv)       // 更新时候 需要获取上次配置的环境变量
			localApiGroup.GET("/oci/down/*oci", controller.Zpk{}.OciDown)
			// OCI下载
			localApiGroup.POST("/buildimage/job", middleware.Auth{}.Process, controller.Zpk{}.BuildImageJob)         // 构建镜像job
			localApiGroup.POST("/buildimage/cronjob", middleware.Auth{}.Process, controller.Zpk{}.BuildImageCronJob) // 构建镜像定时job
			localApiGroup.GET("/local-url", middleware.Auth{}.Process, controller.Zpk{}.LocalZpkUrl)                 // 构建镜像定时job
		}

	})
}
