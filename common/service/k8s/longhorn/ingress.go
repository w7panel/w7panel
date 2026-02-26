package longhorn

import (
	"context"
	"fmt"
	"log/slog"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (w *longhorncontroller) WatchIngressEvents() cache.SharedIndexInformer {
	// slog.Debug("watch ingress events")
	informer := w.factory.KubeInformerFactory.Networking().V1().Ingresses().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {

			ingress := obj.(*networkingv1.Ingress)
			slog.Debug("Ingress deleted: %s/%s\n", ingress.Namespace, ingress.Name)
			clientset := w.sdk.ClientSet
			// 清理与 Ingress 相关的 Secret
			for _, tls := range ingress.Spec.TLS {
				secretName := tls.SecretName
				if secretName != "" {
					// 检查是否有其他 Ingress 引用了该 Secret
					if isSecretReferencedByOtherIngress(clientset, ingress, secretName) {
						fmt.Printf("Secret %s/%s is still referenced by other Ingresses, skipping deletion\n", ingress.Namespace, secretName)
						continue
					}
					// 删除 Secret
					err := clientset.CoreV1().Secrets(ingress.Namespace).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
					if err != nil {
						fmt.Printf("Failed to delete Secret %s/%s: %v\n", ingress.Namespace, secretName, err)
					}
				}
			}
		},
	})

	return informer
}

func isSecretReferencedByOtherIngress(clientset *kubernetes.Clientset, deletedIngress *networkingv1.Ingress, secretName string) bool {
	ingresses, err := clientset.NetworkingV1().Ingresses(deletedIngress.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list Ingresses: %v\n", err)
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
