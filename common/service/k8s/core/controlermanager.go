package core

import (
	"log/slog"
	"os"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/service"
	webhooklocal "gitee.com/we7coreteam/k8s-offline/common/service/k8s/webhook"
	"github.com/go-logr/logr"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// slogr implements logr.LogSink interface

func webHookSetupManager(sdk *k8s.Sdk) (webhook.Server, error) {
	err := webhooklocal.Prepare(sdk)
	if err != nil {
		slog.Error("prepare webhook failed", "err", err)
		return nil, err
	}
	hookServer := webhook.NewServer(webhook.Options{
		Host:    "0.0.0.0",
		Port:    9443,
		CertDir: webhooklocal.CertDir,
	})

	return hookServer, nil
}

func StartControlManager() error {
	isAgent := os.Getenv("IS_AGENT") == "true"
	if isAgent {
		// return nil //higress 需要监听
	}
	// slog.Info("start control manager")
	// 设置controller-runtime的日志记录器
	// ctrl.SetLogger(logr.New(&slogr{logger: slog.Default()}))
	ctrl.SetLogger(logr.FromSlogHandler(slog.Default().Handler()))

	sdk := k8s.NewK8sClient().Sdk
	config, err := sdk.ToRESTConfig()
	if err != nil {
		return err
	}
	options := manager.Options{Scheme: k8s.GetScheme()}
	if facade.GetConfig().GetBool("webhook.enabled") {
		slog.Info("webhook enabled")
		webhookserver, err := webHookSetupManager(sdk)
		if err != nil {
			slog.Error("webhook setup manager failed", "err", err)
			return err
		}
		options.WebhookServer = webhookserver
	}
	// options.BaseContext = sdk.Ctx
	mgr, err := manager.New(config, options)
	if err != nil {
		return err
	}
	//必须调用，否则会导致webhook server无法启动
	if facade.GetConfig().GetBool("webhook.enabled") {
		hookServer := mgr.GetWebhookServer()

		hookServer.Register("/mutate", &webhook.Admission{
			Handler: webhooklocal.NewResourceMutator(mgr.GetClient(), sdk),
		})
	}

	err = service.SvcSetupManager(mgr)
	if err != nil {
		return err
	}
	if facade.GetConfig().GetBool("higress.watch") {
		// slog.Info("higress watch enabled") //走webhook
		// err := higress.SetupManager(mgr, sdk)
		// if err != nil {
		// 	slog.Error("setup higress controllers failed", "err", err)
		// 	return err
		// }
	}

	if facade.GetConfig().GetBool("k3k.watch") {
		slog.Info("k3k watch enabled")
		err = k3k.SetupK3kControllers(mgr)
		if err != nil {
			slog.Error("setup k3k controllers failed", "err", err)
			return err
		}
	}
	// cache.Init

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		slog.Error("manager start failed", "err", err)
		return err
	}
	return nil
}
