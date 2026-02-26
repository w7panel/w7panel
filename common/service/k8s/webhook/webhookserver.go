package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	k3kTypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 全局缓存，用于存储 VirtualClusterPolicy 的信息
var (
	policyCache     = make(map[string]map[string]string) // 存储 policy name -> annotations
	policyCacheLock sync.RWMutex                         // 用于保护 policyCache 的读写锁
)

func Prepare(sdk *k8s.Sdk) error {
	if err := ensureCertificates(sdk.GetNamespace()); err != nil {
		return err
	}
	caBound, err := os.ReadFile("/tmp/k8s-webhook-server/serving-certs/tls.crt")
	if err != nil {
		return err
	}
	// sdk := k8s.NewK8sClient().Sdk
	mutete := NewWebHookMutate(sdk)
	err = mutete.CreateOrUpdate(caBound, svcName, sdk.GetNamespace(), getHookName(), getOperations())
	if err != nil {
		slog.Error("create or update webhook failed")
		return err
	}
	err = mutete.CreateOrUpdate(caBound, svcName, sdk.GetNamespace(), getHookCrdName(), getCrdOperations())
	if err != nil {
		slog.Error("create or update webhook failed")
		return err
	}
	return nil
}

// func WebHookSetupManager(sdk *k8s.Sdk) (webhook.Server, error) {
// 	if err := ensureCertificates(sdk.GetNamespace()); err != nil {
// 		return nil, err
// 	}
// 	caBound, err := os.ReadFile("/tmp/k8s-webhook-server/serving-certs/tls.crt")
// 	if err != nil {
// 		return nil, err
// 	}
// 	// sdk := k8s.NewK8sClient().Sdk
// 	mutete := NewWebHookMutate(sdk)
// 	err = mutete.CreateOrUpdate(caBound, svcName, sdk.GetNamespace())
// 	if err != nil {
// 		slog.Error("create or update webhook failed")
// 		return nil, err
// 	}
// 	hookServer := webhook.NewServer(webhook.Options{
// 		Host:    "0.0.0.0",
// 		Port:    9443,
// 		CertDir: certDir,
// 		// CertFile: "/tmp/k8s-webhook-server/serving-certs/tls.crt",
// 		// KeyFile:  "/tmp/k8s-webhook-server/serving-certs/tls.key",
// 		//CertFile: "./tmp/k8s-webhook-server/serving-certs/tls.crt",
// 		//KeyFile:  "./tmp/k8s-webhook-server/serving-certs/tls.key",
// 	})

// 	return hookServer, nil
// }

// ResourceMutator 处理各种资源的 webhook
type ResourceMutator struct {
	decoder admission.Decoder
	client  client.Client
	sdk     *k8s.Sdk
}

func NewResourceMutator(client client.Client, sdk *k8s.Sdk) *ResourceMutator {
	scheme := k8s.GetScheme()
	return &ResourceMutator{
		decoder: admission.NewDecoder(scheme),
		client:  client,
		sdk:     sdk,
	}
}

// DomainWhiteListItem 表示白名单中的一个域名项
type DomainWhiteListItem struct {
	Prefix   string `json:"prefix"`
	Domain   string `json:"domain"`
	Disabled bool   `json:"disabled"`
}

/*
*
[2025-08-20 02:36:16.770]       [ERROR] default user info 处理 admission 请求   {"user": {"username":"system:serviceaccount:default:admin","uid":"9a7aedab-a56b-4742-af51-ff22b2dc8d6d","groups":["system:serviceaccounts","system:serviceaccounts:default","system:authenticated"],"extra":{"authentication.kubernetes.io/credential-id":["JTI=543f2783-1bf9-4b3d-b5ed-ced864a24ad2"]}}}
*/
func (m *ResourceMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	slog.Error("处理 admission 请求", slog.String("kind", req.Kind.Kind), slog.String("namespace", req.Namespace), slog.String("name", req.Name),
		slog.String("user", req.UserInfo.Username), slog.String("kindGroup", req.Kind.Group))
	// slog.Error("user info 处理 admission 请求", "user", req.UserInfo)
	// 根据资源类型调用不同的处理函数
	switch req.Kind.Kind {
	case "Service":
		return m.handleService(ctx, req)
	case "StatefulSet":
		return m.handleStatefulSet(ctx, req)
	case "Deployment":
		return m.handleDeployment(ctx, req)
	case "DaemonSet":
		return m.handleDaemonset(ctx, req)
	case "Ingress":
		return m.handleIngress(ctx, req)
	case "VirtualClusterPolicy":
		return m.handleVirtualClusterPolicy(ctx, req)
	case "Pod":
		return m.handlePod(ctx, req)
	case "Secret":
		return m.handleSecret(ctx, req)
	case "ConfigMap":
		return m.handleConfigmap(ctx, req)
	case "McpBridge":
		return m.handleMcpBridge(ctx, req)
	case "Node":
		if req.Kind.Group == "longhorn.io" {
			return m.handleLonghornNode(ctx, req)
		}
		return m.handleNode(ctx, req)
	case "StorageClass":
		return m.handleStorageClass(ctx, req)
	case "Replica":
		if req.Kind.Group == "longhorn.io" {
			return m.handleLonghornReplica(ctx, req)
		}
		return admission.Allowed("不需要修改的资源类型")
	case "Cluster":
		if req.Kind.Group == "apps.kubeblocks.io" {
			return m.handleKubeblocksCluster(ctx, req)
		}
		if req.Kind.Group == "k3k.io" {
			return m.handleK3kCluster(ctx, req)
		}
		return admission.Allowed("不需要修改的资源类型")
	case "ServiceAccount":
		return m.handleServiceAccount(ctx, req)
	default:
		return admission.Allowed("不需要修改的资源类型")
	}
}

// 处理 Pod 资源

func setRequestLimit(pod *v1.Pod, cpu resource.Quantity, memory resource.Quantity) bool {

	changed := false
	for j := range pod.Spec.InitContainers {
		if pod.Spec.InitContainers[j].Resources.Limits == nil {
			pod.Spec.InitContainers[j].Resources.Limits = make(v1.ResourceList)
		}
		if pod.Spec.InitContainers[j].Resources.Requests == nil {
			pod.Spec.InitContainers[j].Resources.Requests = make(v1.ResourceList)
		}
		limits := pod.Spec.InitContainers[j].Resources.Limits
		if limits.Cpu().IsZero() || limits.Memory().IsZero() {
			pod.Spec.InitContainers[j].Resources.Limits["cpu"] = cpu
			pod.Spec.InitContainers[j].Resources.Limits["memory"] = memory
			changed = true
		}
		if pod.Spec.InitContainers[j].Resources.Requests.Cpu().IsZero() || pod.Spec.InitContainers[j].Resources.Requests.Memory().IsZero() {
			pod.Spec.InitContainers[j].Resources.Requests["cpu"] = resource.MustParse("0")
			pod.Spec.InitContainers[j].Resources.Requests["memory"] = resource.MustParse("0")
			changed = true
		}
	}
	for i := range pod.Spec.Containers {

		if pod.Spec.Containers[i].Resources.Limits == nil {
			pod.Spec.Containers[i].Resources.Limits = make(v1.ResourceList)
		}
		if pod.Spec.Containers[i].Resources.Requests == nil {
			pod.Spec.Containers[i].Resources.Requests = make(v1.ResourceList)
		}
		limits := pod.Spec.Containers[i].Resources.Limits
		if limits.Cpu().IsZero() || limits.Memory().IsZero() {
			pod.Spec.Containers[i].Resources.Limits["cpu"] = cpu
			pod.Spec.Containers[i].Resources.Limits["memory"] = memory
			changed = true
		}
		if pod.Spec.Containers[i].Resources.Requests.Cpu().IsZero() || pod.Spec.Containers[i].Resources.Requests.Memory().IsZero() {
			pod.Spec.Containers[i].Resources.Requests["cpu"] = resource.MustParse("0")
			pod.Spec.Containers[i].Resources.Requests["memory"] = resource.MustParse("0")
			changed = true
		}
	}
	return changed
}

func (m *ResourceMutator) handleK3sPod(ctx context.Context, pod *v1.Pod, req admission.Request, clusterName string) admission.Response {
	slog.Info("处理 Pod admission handleK3sPod 请求")

	// 检查是否需要修改
	modified := false

	// 检查 Pod 是否有 ownerReferences.kind=Cluster
	role, ok := pod.Labels["role"]
	if !ok {
		return admission.Allowed("Pod 没有 role")
	}
	sa, err := getSa(m.client, m.sdk, clusterName)
	if err != nil {
		return admission.Allowed("未找到sa")
	}
	k3kUser := k3kTypes.NewK3kUser(sa)
	if !k3kUser.IsClusterUser() {
		slog.Info("不是集群用户")
		return admission.Allowed("不是集群用户")
	}
	rang := k3kUser.GetLimitRange()
	if rang == nil {
		slog.Info("未配置limitRange")
		return admission.Allowed("未配置limitRange")
	}
	cpu := rang.Hard.Cpu()
	memory := rang.Hard.Memory()
	if k3kUser.IsVirtual() {
		if role == "server" {
			setRequestLimit(pod, *cpu, *memory)
		}
		modified = true
	}

	if k3kUser.IsShared() {
		if role == "server" {
			cpu2 := resource.MustParse("500m")
			memory2 := resource.MustParse("1Gi")
			setRequestLimit(pod, cpu2, memory2)
			modified = true
		}
		if role == "agent" {
			cpu3 := resource.MustParse("100m")
			memory3 := resource.MustParse("100Mi")
			setRequestLimit(pod, cpu3, memory3)
			modified = true
		}
	}

	quantity := k3kUser.GetBandWidth()
	if !quantity.IsZero() {
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		quantitystr := quantity.String()
		slog.Info("Pod 带宽限制", slog.String("bandwidth", quantitystr))
		// quantitystr = strings.ReplaceAll(quantitystr, "Mi", "Mbps")

		pod.Annotations["kubernetes.io/egress-bandwidth"] = quantitystr
		pod.Annotations["kubernetes.io/ingress-bandwidth"] = quantitystr
		modified = true
	}

	// 如果没有修改，直接返回允许
	if !modified {
		return admission.Allowed("Pod 已有带宽注解或不需要修改")
	}

	// 序列化修改后的 Pod
	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// 返回修改后的资源
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}
