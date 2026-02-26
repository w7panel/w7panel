package main

import (
	"bytes"
	_ "embed"
	"io"
	"log"
	"log/slog"
	"net"
	http2 "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/we7coreteam/k8s-offline/app/application"
	"gitee.com/we7coreteam/k8s-offline/app/application/http/controller"
	"gitee.com/we7coreteam/k8s-offline/app/auth"
	"gitee.com/we7coreteam/k8s-offline/app/k3k"
	metrics2 "gitee.com/we7coreteam/k8s-offline/app/metrics"
	"gitee.com/we7coreteam/k8s-offline/app/zpk"
	helper2 "gitee.com/we7coreteam/k8s-offline/common/helper"
	middleware2 "gitee.com/we7coreteam/k8s-offline/common/middleware"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	registryClient "gitee.com/we7coreteam/k8s-offline/common/service/registry"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/grafana/pyroscope-go"
	"github.com/spf13/viper"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	app "github.com/we7coreteam/w7-rangine-go/v2/src"
	"github.com/we7coreteam/w7-rangine-go/v2/src/core/helper"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/response"
	cachecontrol "go.eigsys.de/gin-cachecontrol/v2"
	"go.uber.org/automaxprocs/maxprocs"
)

//go:embed config.yaml
var ConfigFileContent []byte

var Asset = (os.DirFS(os.Getenv("KO_DATA_PATH")))

func pyroscope2() {
	serverAddress := facade.Config.GetString("pyroscope.server_address")
	helmreleaseName := facade.Config.GetString("app.helm_release_name")
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "w7panel-offline",
		ServerAddress:   serverAddress,
		Logger:          pyroscope.StandardLogger,
		Tags:            map[string]string{"releasename": helmreleaseName},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		log.Fatalf("error starting pyroscope profiler: %v", err)
	}
}

func registryServer() error {
	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalln(err)
	}
	s := &http2.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           registry.New(registry.WithBlobHandler(registry.NewInMemoryBlobHandler())),
	}

	errCh := make(chan error)
	go func() { errCh <- s.Serve(listener) }()
	time.AfterFunc(5*time.Second, func() {
		err := registryClient.PushOciProxy()
		if err != nil {
			slog.Error("main PushOciProxy", "error", err)
		}
	})
	<-errCh
	return err
}

func init() {
	const PR_SET_CHILD_SUBREAPER = 36
	_, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0, 0)
	if errno != 0 {
		slog.Warn("Failed to set child subreaper", "error", errno)
	} else {
		slog.Info("set child subreaper successfully")
	}

	signal.Ignore(syscall.SIGCHLD)
	slog.Info("SIGCHLD ignored for auto child process reaping")
}

func main() {
	os.Setenv("APP_DIR", helper2.GetAppHomeDir())
	maxprocs.Set(maxprocs.Logger(nil))

	newApp := app.NewApp(app.Option{
		DefaultConfigLoader: func(config *viper.Viper) {
			config.SetConfigType("yaml")
			err := config.MergeConfig(bytes.NewReader(helper.ParseConfigContentEnv(ConfigFileContent)))
			if err != nil {
				panic(err)
			}
		},
	})

	if os.Getenv("DISABLE_LOG") == "true" {
		log.SetOutput(io.Discard)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		slog.SetDefault(logger)
	}

	response.SetSuccessResponseHandler(func(ctx *gin.Context, data any, statusCode int) {
		ctx.JSON(statusCode, data)
	})

	if facade.Config.GetBool("pyroscope.enabled") {
		go pyroscope2()
	}

	single := k8s.NewK8sClient()
	go single.GetSdk().CreateServiceAccountSecret(single.GetSdk().GetServiceAccountName())

	httpServer := new(http.Provider).Register(newApp.GetConfig(), newApp.GetConsole(), newApp.GetServerManager()).Export()
	httpServer.Use(middleware.GetPanicHandlerMiddleware()).Use(middleware2.HostCheck{}.Process)

	httpServer.RegisterRouters(func(engine *gin.Engine) {
		engine.Use(middleware2.Cors{}.Process)
		microappPath := facade.Config.GetString("static.microapp_path")
		os.MkdirAll(microappPath, 0755)

		router := engine.Group("").
			Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".pdf", ".mp4"}))).
			Use(cachecontrol.New(cachecontrol.CacheAssetsForeverPreset))
		routerNocache := engine.Group("").
			Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".pdf", ".mp4"})))

		staticPath := facade.Config.GetString("app.static_path")
		router.Static("/assets", staticPath+"/assets")
		router.Static("/longhorn", staticPath+"/longhorn")
		router.Static("/charts", staticPath+"/charts")
		router.Static("/schema", staticPath+"/schema")
		routerNocache.Static("/ui/microapp", microappPath)
		router.Static("/ui/plugin", staticPath+"/plugin")
		router.Static("/ui/wasm", staticPath+"/wasm")
		router.Static("/ui/yaml", staticPath+"/yaml")

		router.StaticFileFS("/index.html", "index.html", http2.FS(Asset))
		router.StaticFileFS("/k3s-agent.sh", "k3s-agent.sh", http2.FS(Asset))
		router.StaticFileFS("/k3s-server.sh", "k3s-server.sh", http2.FS(Asset))
		router.StaticFileFS("/favicon.ico", "icon.jpg", http2.FS(Asset))
		router.StaticFileFS("/micro.html", "micro.html", http2.FS(Asset))
		router.StaticFileFS("/logo.png", "logo.png", http2.FS(Asset))
	})

	httpServer.RegisterRouters(
		func(engine *gin.Engine) {
			engine.Any("/k8s-proxy/*path",
				middleware2.Auth{}.Process,
				middleware2.K8sFilter{}.Process,
				controller.Proxy{}.ProxyK8s)
		},
	)

	new(application.Provider).Register(httpServer, newApp.GetConsole())
	new(auth.Provider).Register(httpServer, newApp.GetConsole())
	new(metrics2.Provider).Register(httpServer, newApp.GetConsole())
	new(zpk.Provider).Register(httpServer, newApp.GetConsole())
	new(k3k.Provider).Register(httpServer, newApp.GetConsole())

	// NoRoute 必须在所有 Provider 注册之后
	httpServer.RegisterRouters(
		func(engine *gin.Engine) {
			engine.NoRoute(middleware2.Html{}.Process)
		},
	)

	newApp.RunConsole()
}
