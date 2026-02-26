package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	ingApi "gitee.com/we7coreteam/k8s-offline/common/service/k8s/webhook/ingress"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Deployment 资源

// 处理 Ingress 资源
func (m *ResourceMutator) handleIngress(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 Ingress admission 请求")

	defer handlerIngressSync(req, m.decoder, m.client)
	ingress := &networkingv1.Ingress{}
	// 判断是否Delete 请求
	delete := false
	if req.Operation == "DELETE" {
		delete = true
		if err := (m.decoder).DecodeRaw(req.OldObject, ingress); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	} else {
		if err := (m.decoder).Decode(req, ingress); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	}
	// 解码请求中的 Ingress 资源

	if helper.IsChildAgent() {
		defer k3k.SyncHttpAfter(ingress, "sync-ingress")
	}
	if delete {
		defer m.handleIngressDelete(m.client, ingress.DeepCopy())
		return admission.Allowed("删除请求")
	}

	ann := ingress.Annotations
	if ann == nil {
		return admission.Allowed("annotations empty")
	}

	if !helper.IsChildAgent() { //子集群不处理
		sslredirect, ok := ingress.Annotations["w7.cc/ssl-redirect"]
		if !ok {
			return admission.Allowed("no ssl-redirect annotation")
		}
		if ingress.Annotations["higress.io/ssl-redirect"] != sslredirect {
			ingress.Annotations["higress.io/ssl-redirect"] = sslredirect
			marshaledIngress, err := json.Marshal(ingress)
			if err != nil {
				return admission.Errored(http.StatusInternalServerError, err)
			}

			// 返回修改后的资源
			return admission.PatchResponseFromRaw(req.Object.Raw, marshaledIngress)
		}
	}

	return admission.Allowed("所有域名都在白名单中")
}

func (m *ResourceMutator) handleIngressDelete(client client.Client, ingress *networkingv1.Ingress) {

	for _, tls := range ingress.Spec.TLS {
		secretName := tls.SecretName
		if secretName != "" {
			// 检查是否有其他 Ingress 引用了该 Secret
			if isSecretReferencedByOtherIngress(client, ingress, secretName) {
				fmt.Printf("Secret %s/%s is still referenced by other Ingresses, skipping deletion\n", ingress.Namespace, secretName)
				continue
			}
			delSecret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: ingress.Namespace,
				},
			}

			err := client.Delete(context.Background(), delSecret)
			if err != nil {
				slog.Error("ingress webhook delete secret error", "error", err)
				continue
			}
		}
	}

}

func handlerIngressSync(req admission.Request, decoder admission.Decoder, client client.Client) {
	slog.Info("处理 Ingress admission 请求")

	switch req.Operation {
	case "CREATE":
		ing := &networkingv1.Ingress{}
		if err := decoder.Decode(req, ing); err == nil {
			helper.After2SecondRun(func() {
				ingApi.OnAdd(client, ing)
			})
		}
	case "UPDATE":
		olding := &networkingv1.Ingress{}
		ing := &networkingv1.Ingress{}
		err := decoder.Decode(req, ing)
		err2 := decoder.DecodeRaw(req.OldObject, olding)
		if err == nil && err2 == nil {
			helper.After2SecondRun(func() {
				ingApi.OnUpdate(client, olding, ing)
			})
		}
	case "DELETE":
		ing := &networkingv1.Ingress{}
		if err := decoder.DecodeRaw(req.OldObject, ing); err == nil {
			helper.After2SecondRun(func() {
				ingApi.OnDelete(client, ing)
			})
		}
	default:
		slog.Error("不支持的操作类型", "operation", req.Operation)
		return
	}

}

func isSecretReferencedByOtherIngress(clientset client.Client, deletedIngress *networkingv1.Ingress, secretName string) bool {
	// 获取当前命名空间下的所有 Ingress 资源
	ingresses := &networkingv1.IngressList{}
	if err := clientset.List(context.TODO(), ingresses, client.InNamespace(deletedIngress.Namespace)); err != nil {
		slog.Error("Failed to list Ingresses", "error", err)
		return false
	}
	for _, ingress := range ingresses.Items {
		// 跳过已删除的 Ingress
		if ingress.Name == deletedIngress.Name && ingress.Namespace == deletedIngress.Namespace {
			continue
		}

		// 检查当前 Ingress 是否引用了指定的 Secret
		for _, tls := range ingress.Spec.TLS {
			if tls.SecretName == secretName {
				return true
			}
		}
	}

	return false
}

// 解析白名单数据

// 检查域名是否在白名单中
func isDomainInWhiteList(host string, whiteList []DomainWhiteListItem) bool {
	whiteListCount := len(whiteList)
	disableCount := 0
	for _, item := range whiteList {
		// 跳过禁用的项
		if item.Disabled {
			disableCount++
			continue
		}

		// 检查域名是否匹配
		if item.Prefix == "*." {
			// 检查域名是否是白名单域名的子域名
			if strings.HasSuffix(host, "."+item.Domain) || host == item.Domain {
				return true
			}
		} else {
			// 精确匹配
			if host == item.Domain {
				return true
			}
		}
	}
	return disableCount == whiteListCount
}
