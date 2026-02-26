package webhook

import (
	"fmt"
	"log/slog"
	"os"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	k3kTypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var svcName = ""
var svcHost = ""

type WebhookConfig struct {
	Name        string
	Description string
	URL         string
	Secret      string
}

func init() {
	name, ok := os.LookupEnv("SVC_NAME")
	if !ok {
		// slog.Error("SVC_NAME not set")
		return
	}
	svcName = name
}
func getSvcHost(namespace string) string {
	return svcName + "." + namespace + ".svc"
}

func getSecret() string {
	return svcName + "-webhook-tls"
}

func getHookName() string {
	return svcName + "-webhook"
}

func getHookCrdName() string {
	return svcName + "-crd-webhook"
}

type void struct{}

func getSa(client client.Client, sdk *k8s.Sdk, saName string) (*v1.ServiceAccount, error) {
	sa := &v1.ServiceAccount{}
	err := client.Get(sdk.Ctx, types.NamespacedName{Name: saName, Namespace: "default"}, sa) //获取不到 未找到原因
	if err != nil {
		if errors.IsNotFound(err) {
			sa2, err := sdk.ClientSet.CoreV1().ServiceAccounts("default").Get(sdk.Ctx, saName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			return sa2, nil
		}
		return nil, err
	}
	return sa, err
}

func getResourceLimit(client client.Client, sdk *k8s.Sdk, clusterName string, role string) (v1.ResourceList, error) {
	sa, err := getSa(client, sdk, clusterName)
	if err != nil {
		return nil, err
	}
	k3kUser := k3kTypes.NewK3kUser(sa)
	if !k3kUser.IsClusterUser() {
		slog.Info("不是集群用户")
		return nil, err
	}
	rang := k3kUser.GetLimitRange()
	if rang == nil {
		slog.Info("未配置limitRange")
		return nil, err
	}
	cpu := rang.Hard.Cpu()
	memory := rang.Hard.Memory()
	if k3kUser.IsVirtual() {
		rs := v1.ResourceList{}
		cpuCopy := cpu.DeepCopy()
		memoryCopy := memory.DeepCopy()
		cpuPoint := *cpu
		memoryPoint := *memory
		if k3kUser.IsWeihu() {
			cpuPoint.Add(cpuCopy)
			memoryPoint.Add(memoryCopy)
		}
		rs["cpu"] = cpuPoint
		rs["memory"] = memoryPoint

		return rs, err
	}

	if k3kUser.IsShared() {
		if role == "server" {
			cpu2 := resource.MustParse("500m")
			memory2 := resource.MustParse("1Gi")
			rs := v1.ResourceList{}
			rs["cpu"] = cpu2
			rs["memory"] = memory2
			return rs, err

		}
		if role == "agent" {
			cpu3 := resource.MustParse("100m")
			memory3 := resource.MustParse("100Mi")
			rs := v1.ResourceList{}
			rs["cpu"] = cpu3
			rs["memory"] = memory3
			return rs, err
		}
	}
	return nil, fmt.Errorf("err not found")
}
