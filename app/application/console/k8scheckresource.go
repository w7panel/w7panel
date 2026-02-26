package console

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/spf13/cobra"
	console2 "github.com/we7coreteam/w7-rangine-go/v2/src/console"
)

const (
	checkInterval = 10 * time.Second
	timeout       = 11 * time.Second // 应该根据实际需求设置合理的超时时间
)

type K8sCheckResource struct {
	console2.Abstract
}

type resourceOption struct {
	resources []string
	timeout   int
}

var resourceOp = resourceOption{}

func (c K8sCheckResource) GetName() string {
	return "k8s-check-resource"
}

func (c K8sCheckResource) Configure(cmd *cobra.Command) {
	// username password register
	cmd.Flags().StringArrayVar(&resourceOp.resources, "resources", []string{}, "resources")
	cmd.Flags().IntVar(&resourceOp.timeout, "timeout", 300, "默认300秒")
}

func (c K8sCheckResource) GetDescription() string {
	return "检查k8s资源是否存在,并且运行状态是否正常"
}

// 每10秒调用一次check 方法
// 超出timeout 时间后退出

func (c K8sCheckResource) Handle(cmd *cobra.Command, args []string) {
	sdk := k8s.NewK8sClientInner()
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	timeoutChan := time.After(time.Duration(resourceOp.timeout) * time.Second)

	for {
		select {
		case <-ticker.C:
			slog.Info("检查资源 step")
			err := check(resourceOp.resources, sdk)
			if err != nil {
				slog.Error("资源检查失败", "error", err)
			}
			if err == nil {
				slog.Info("资源检查成功")
				ticker.Stop()
				os.Exit(0)
			}
		case <-timeoutChan:
			slog.Info("检查超时")
			os.Exit(1)
		}
	}
}

func check(resources []string, sdk *k8s.Sdk) error {
	ok := []bool{}
	slog.Info("检查资源", "resources", resources)
	for _, resource := range resources {
		gvkn := strings.Split(resource, ".")
		if len(gvkn) != 4 {
			slog.Error("资源格式不正确, 应为: group/version.kind.name, 实际: ", "resource", resource)
			os.Exit(1)
		}
		apiVersion := gvkn[0]
		kind := gvkn[1]
		name := gvkn[2]
		namespace := gvkn[3]
		check := false
		switch apiVersion {
		case "v1":
			// 处理 core 资源
			pod, err := sdk.ClientSet.CoreV1().Pods(namespace).Get(sdk.Ctx, name, metav1.GetOptions{})
			if err != nil {
				check = false
				return err
			}
			if pod.Status.Phase == "Running" {
				check = true
			}
		case "apps/v1":
			switch kind {
			case "Deployment":
				dt, err := sdk.ClientSet.AppsV1().Deployments(namespace).Get(sdk.Ctx, name, metav1.GetOptions{})
				if err != nil {
					check = false
					return err
				}
				if dt.Status.ReadyReplicas == dt.Status.Replicas {
					check = true
				}
			case "StatefulSet":
				// 处理 StatefulSet 资源
				sf, err := sdk.ClientSet.AppsV1().StatefulSets(namespace).Get(sdk.Ctx, name, metav1.GetOptions{})
				if err != nil {
					check = false
					return err
				}
				if sf.Status.ReadyReplicas == sf.Status.Replicas {
					check = true
				}
			case "DaemonSet":
				// 处理 DaemonSet 资源
				ds, err := sdk.ClientSet.AppsV1().DaemonSets(namespace).Get(sdk.Ctx, name, metav1.GetOptions{})
				if err != nil {
					check = false
					return err
				}
				// slog("sssssssss"+ds.Status.NumberAvailable == ds.Status.DesiredNumberScheduled)
				// if ds.Status.NumberAvailable == ds.Status.DesiredNumberScheduled && ds.Status.NumberReady > 0 {
				if ds.Status.NumberReady > 0 {
					check = true
				}
				// 处理 Deployment 资源
			}
		case "batch/v1":
			switch kind {
			case "Job":
				// 处理 Job 资源
				slog.Info("检查 Job 资源")
				job, err := sdk.ClientSet.BatchV1().Jobs(namespace).Get(sdk.Ctx, name, metav1.GetOptions{})
				if err != nil {
					check = false
					return err
				}
				if job.Status.Succeeded >= 1 {
					check = true
				}
			}

		default:
			slog.Error("不支持的资源类型", "resourceType", gvkn[0])
			return fmt.Errorf("不支持的资源类型: %s", gvkn[0])
		}
		if check {
			ok = append(ok, check)
		}

	}
	if (len(resources) - len(ok)) > 0 {
		err := fmt.Errorf("资源检查失败 成功个数: %d, 失败个数: %d", len(ok), (len(resources) - len(ok)))
		return err
	}

	return nil
}
