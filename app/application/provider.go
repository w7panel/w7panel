package application

import (
	"log/slog"
	"os"
	"time"

	consoleShell "gitee.com/we7coreteam/k8s-offline/app/application/console"
	controller2 "gitee.com/we7coreteam/k8s-offline/app/application/http/controller"
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/middleware"
	console2 "gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	appctl "gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/core"
	gpustack "gitee.com/we7coreteam/k8s-offline/common/service/k8s/gpu/gpustack"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/higress"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/longhorn"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/mcp"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/shell"
	"gitee.com/we7coreteam/k8s-offline/common/service/registry"
	"gitee.com/we7coreteam/k8s-offline/common/service/s3"
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

	console.RegisterCommand(new(consoleShell.Goshell))
	console.RegisterCommand(new(consoleShell.K8sCheckResource))
	console.RegisterCommand(new(consoleShell.IngressUpgrade))
	console.RegisterCommand(new(consoleShell.MetricsInstall))
	console.RegisterCommand(new(consoleShell.UninstallStorePanel)) //删除商店安装的面板
	console.RegisterCommand(new(consoleShell.DomainParseConfig))   //删除商店安装的面板
	console.RegisterCommand(new(consoleShell.BeianCheck))          //删除商店安装的面板
	p.RegisterValidateRule()
	p.RegisterHttpRoutes(httpServer)
	console2.SetConsoleApi(facade.GetConfig().GetString("app.console_base_url"))
	if helper.IsLocalMock() {
		// console2.SetConsoleApi("http://172.16.1.116:9004")
	}

	// p.CRD() //upgrade.sh 中处理
	if facade.GetConfig().GetBool("longhorn.watch") {

		go longhorn.OnStart()
	}
	if facade.GetConfig().GetBool("registry.watch") {

		go shell.ShellWatch()
	}

	// if facade.GetConfig().GetBool("higress.watch") {

	// 	go higress.Watch()
	// }
	// go converter.ConvertOpenApiToSchema()
	if facade.GetConfig().GetBool("clean.enabled") {

		go p.cleanS3()
	}

	if facade.GetConfig().GetBool("app.watch") {
		slog.Info("开始监听AppGroup资源变更事件")
		go appctl.Watch()
	}
	if facade.GetConfig().GetBool("gpustack.watch") {

		go gpustack.Watch()
	}

	if facade.GetConfig().GetBool("mcp.watch") {

		go mcp.Watch()
	}
	if facade.GetConfig().GetBool("k8s.watch") {
		go core.StartControlManager()
	}

	go k8s.CheckLogo()
	// go k3k.SyncAgentIngress()
	go higress.LoadBkConfig()

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
	webdavMethods := []string{"PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "LINK", "UNLINK", "GET", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "POST"}
	server.RegisterRouters(func(engine *gin.Engine) {
		apiGroup := engine.Group("/panel-api/v1") //.Use(middleware.Cors{}.Process)
		{
			apiGroup.GET("/namespaces", middleware.Auth{}.Process, controller2.Namespaces{}.GetList)
			apiGroup.GET("/helm/releases", middleware.Auth{}.Process, controller2.Helm{}.List)
			apiGroup.GET("/helm/releases/:name", middleware.Auth{}.Process, controller2.Helm{}.Info)
			apiGroup.POST("/helm/releases/:name", middleware.Auth{}.Process, controller2.Helm{}.InstallUseRepo)
			apiGroup.DELETE("/helm/releases/:name", middleware.Auth{}.Process, controller2.Helm{}.UnInstall)
			apiGroup.PUT("/helm/releases/:name/reuse", middleware.Auth{}.Process, controller2.Helm{}.ReUseValues)
			apiGroup.GET("/app-info", controller2.Helm{}.AppInfo)

		}

		localApiGroup := engine.Group("/panel-api/v1") //.Use(middleware.Cors{}.Process)
		{
			localApiGroup.GET("/tty", middleware.Auth{}.Process, controller2.PodExec{}.Tty)
			localApiGroup.GET("/nodetty", middleware.Auth{}.Process, controller2.PodExec{}.NodeTty)
			localApiGroup.GET("/download/*path", controller2.File{}.Download)
			localApiGroup.POST("/cp", middleware.Auth{}.Process, controller2.PodExec{}.KubectlCp) //kubectl cp文件
			localApiGroup.POST("/cppid", middleware.Auth{}.Process, controller2.File{}.CpPidFile) //pid文件移动
			localApiGroup.POST("/mvpid", middleware.Auth{}.Process, controller2.File{}.CpPidFile) //pid文件移动
			localApiGroup.GET("/exec", middleware.Auth{}.Process, controller2.PodExec{}.Exec)
			localApiGroup.POST("/exec2", middleware.Auth{}.Process, controller2.PodExec{}.Exec)
			localApiGroup.GET("/pid", middleware.Auth{}.Process, controller2.PodExec{}.GetAgentPodAndPid) //获取所在pod和pid
			localApiGroup.GET("/nodepid", middleware.Auth{}.Process, controller2.PodExec{}.GetNodePid)    //获取所在pod和pid

			localApiGroup.POST("/yaml", middleware.Auth{}.Process, controller2.Yaml{}.ApplyYamlOld) // 直接提交yaml
			localApiGroup.PUT("/rollback", middleware.Auth{}.Process, controller2.Yaml{}.Rollback)  // 回滚资源
			// localApiGroup.POST("/kcompose", middleware.Auth{}.Process, controller2.Yaml{}.ApplyDockerCompose)   // 直接提交yaml
			localApiGroup.POST("/kcompose", middleware.Auth{}.Process, controller2.Yaml{}.ConvertDockerComposeOld) // 转化kompose
			localApiGroup.POST("/pinyin", middleware.Auth{}.Process, controller2.Util{}.Pinyin)                    // pinyin
			localApiGroup.GET("/dnsip", middleware.Auth{}.Process, controller2.Util{}.DnsIp)
			localApiGroup.GET("/dns-cname", middleware.Auth{}.Process, controller2.Util{}.DnsCName)
			localApiGroup.GET("/myip", middleware.Auth{}.Process, controller2.Util{}.MyIp)
			localApiGroup.POST("/db-conn-test", middleware.Auth{}.Process, controller2.Util{}.DbConnTest) // 数据库连接测试
			localApiGroup.POST("/ping-etcd", middleware.Auth{}.Process, controller2.Util{}.PintEtcd)      // etcd连接测试
			localApiGroup.GET("/captcha", controller2.Util{}.Captcha)
			localApiGroup.POST("/verify-captcha", controller2.Util{}.VerifyCaptcha)
			for _, method := range webdavMethods {
				// 不转发到子pod
				localApiGroup.Handle(method, "/namespaces/:namespace/services/:name/proxy-root/*path", middleware.Auth{}.Process, controller2.Proxy{}.ProxyService)

				localApiGroup.Handle(method, "/namespaces/:namespace/services/:name/proxy/*path", middleware.Auth{}.Process, middleware.Proxy{}.Process, controller2.Proxy{}.ProxyService)
				localApiGroup.Handle(method, "/namespaces/:namespace/pods/:name/proxy/*path", middleware.Auth{}.Process, middleware.Proxy{}.Process, controller2.Proxy{}.ProxyPod)
				//代理转发
				localApiGroup.Handle(method, "/:name/proxy/*path", middleware.Auth{}.Process, controller2.Proxy{}.ProxyCommon) // 转发到子pod 文件管理需要访问agent 不能走proxy middleware
			}

			// localApiGroup.Any("/v1/:name/proxy/*path", controller2.Proxy{}.ProxyCommon)

			localApiGroup.Any("/proxy-url/", controller2.Proxy{}.ProxyAddr)

			//获取需要删除的副本
			localApiGroup.GET("/longhorn/need-delete-replica", middleware.Auth{}.Process, controller2.Longhorn{}.GetNeedDeleteReplicas)
			localApiGroup.GET("/longhorn/volumes/status", middleware.Auth{}.Process, controller2.Longhorn{}.GetVolumesStatus)

			localApiGroup.GET("/k3s/env/gogc", middleware.Auth{}.Process, controller2.K3s{}.GoGc)
			localApiGroup.POST("/k3s/env/gogc", middleware.Auth{}.Process, controller2.K3s{}.GoGcToggle)
			localApiGroup.GET("/kubeblocks/installjobyaml", middleware.Auth{}.Process, controller2.KubeBlocks{}.InstallJobYaml)
			localApiGroup.POST("/kubeblocks/install", middleware.Auth{}.Process, controller2.KubeBlocks{}.Install)

		}
		gpuGroup := engine.Group("/panel-api/v1/gpu").Use(middleware.Auth{}.Process, middleware.Proxy{}.Process)
		{
			gpuGroup.POST("/enabled-gpu", controller2.Gpu{}.EnableGpu)                   // 开启关闭gpu
			gpuGroup.POST("/install-hami", controller2.Gpu{}.InstallHami)                // 安装hami
			gpuGroup.POST("/install-gpu-operator", controller2.Gpu{}.InstallGpuOperator) // 安装gpu-op
			gpuGroup.GET("/config", controller2.Gpu{}.GetGpuConfig)                      // 获取配置
			gpuGroup.GET("/hami/metrics/real", controller2.Gpu{}.HamiMetricsReal)        // hami实时监控利用率百分比
			gpuGroup.GET("/summary", controller2.Gpu{}.GpuSummary)
			gpuGroup.GET("/node/devices", controller2.Gpu{}.NodesDevices)
			gpuGroup.POST("/gpustack/worker", controller2.Gpu{}.CreateGpuStackWorker)
		}

		for _, method := range webdavMethods {
			engine.Handle(method, "/panel-api/v1/files/webdav/*path", middleware.Auth{}.Process, controller2.Webdav{}.Handle)
			engine.Handle(method, "/panel-api/v1/files/webdav-agent/:pid/subagent/:subpid/agent/*path", middleware.Auth{}.Process, controller2.Webdav{}.HandlePidSubPid)

			engine.Handle(method, "/panel-api/v1/files/webdav-agent/:pid/agent/*path", middleware.Auth{}.Process, controller2.Webdav{}.HandlePid)
		}
		//临时兼容旧版api
		engine.Any("/k8s/v1/namespaces/:namespace/services/:name/proxy-no/*path", middleware.ProxyNoAuth{}.Process, controller2.Proxy{}.ProxyNoAuthService)

		engine.POST("/panel-api/v1/files/compress-agent/:pid/compress", middleware.Auth{}.Process, controller2.CompressAgent{}.Compress)
		engine.POST("/panel-api/v1/files/compress-agent/:pid/extract", middleware.Auth{}.Process, controller2.CompressAgent{}.Extract)
		engine.POST("/panel-api/v1/files/compress-agent/:pid/subagent/:subpid/compress", middleware.Auth{}.Process, controller2.CompressAgent{}.Compress)
		engine.POST("/panel-api/v1/files/compress-agent/:pid/subagent/:subpid/extract", middleware.Auth{}.Process, controller2.CompressAgent{}.Extract)

		engine.POST("/panel-api/v1/files/permission-agent/:pid/chmod", middleware.Auth{}.Process, controller2.PermissionAgent{}.Chmod)
		engine.POST("/panel-api/v1/files/permission-agent/:pid/chown", middleware.Auth{}.Process, controller2.PermissionAgent{}.Chown)
		engine.POST("/panel-api/v1/files/permission-agent/:pid/subagent/:subpid/chmod", middleware.Auth{}.Process, controller2.PermissionAgent{}.Chmod)
		engine.POST("/panel-api/v1/files/permission-agent/:pid/subagent/:subpid/chown", middleware.Auth{}.Process, controller2.PermissionAgent{}.Chown)

		engine.GET("/panel-api/v1/kubeconfig", middleware.Auth{}.Process, middleware.Proxy{}.Process, controller2.Proxy{}.Kubeconfig)
		engine.Any("/panel-api/v1/s3bucket", middleware.Auth{}.Process, controller2.File{}.Upload).Use(middleware.Cors{}.Process)

		// 安全的未授权接口 - 只返回必要的公开字段
		engine.GET("/panel-api/v1/noauth/site/beian", controller2.Site{}.Beian)
		engine.GET("/panel-api/v1/noauth/site/k3k-config", controller2.Site{}.K3kConfig)
		engine.GET("/panel-api/v1/noauth/site/init-user", controller2.Site{}.InitUser)

		engine.GET("/panel-api/v1/microapp/top", middleware.Auth{}.Process, controller2.MicroApp{}.List)                     //获取microapp列表
		engine.Any("/panel-api/v1/microapp/:name/proxy/*path", middleware.Auth{}.Process, controller2.Proxy{}.ProxyMicroApp) //microapp proxy

	})
}

func (p Provider) RegisterS3Server(server *httpserver.Server) {
	s3.Init(facade.Config.GetString("s3.base_dir"))
}

func (p Provider) cleanS3() {
	sen := facade.Config.GetDuration("clean.interval")
	ticker := time.NewTicker(sen)
	quit := make(chan struct{})

	for {
		select {
		// default:
		// 	slog.Info("Mrunning", runtime.NumGoroutine())
		// 	time.Sleep(1 * time.Second)
		case <-quit:
			ticker.Stop()
			return

		case <-ticker.C:
			s3dir := facade.Config.GetString("s3.base_dir")
			err := os.RemoveAll(s3dir + "/upload")
			if err != nil {
				slog.Error("clean s3 error", "err", err)
			}
		}

	}
}

func (p Provider) pushProxyOci() {
	time.AfterFunc(30*time.Second, func() {
		err := registry.PushOciProxy()
		if err != nil {
			slog.Error("push oci proxy error", "err", err)
		}
	})
}

// 提前缓存helm升级信息
