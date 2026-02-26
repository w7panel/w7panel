package webhook

import (
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
)

const WebHookName = "webhook-w7panel"

type WebHookMutate struct {
	sdk *k8s.Sdk
}

func NewWebHookMutate(sdk *k8s.Sdk) *WebHookMutate {
	return &WebHookMutate{
		sdk: sdk,
	}
}
func getOperations() []admissionregistrationv1.RuleWithOperations {
	if helper.IsChildAgent() {
		return getAgentOperations()
	}
	return getDefaultOperations()
}

func getCrdOperations() []admissionregistrationv1.RuleWithOperations {
	if helper.IsChildAgent() {
		return getAgentCrdOperations()
	}
	return getDefaultCrdOperations()
}
func getAgentOperations() []admissionregistrationv1.RuleWithOperations {
	return []admissionregistrationv1.RuleWithOperations{
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE", "DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"networking.k8s.io"},
				APIVersions: []string{"v1"},
				Resources:   []string{"ingresses"},
			},
		},

		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"configmaps"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE", "DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"secrets"},
			},
		},
	}
}

func getAgentCrdOperations() []admissionregistrationv1.RuleWithOperations {
	return []admissionregistrationv1.RuleWithOperations{
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE", "DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"networking.higress.io"},
				APIVersions: []string{"v1"},
				Resources:   []string{"mcpbridges"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"apps.kubeblocks.io"},
				APIVersions: []string{"v1alpha1"},
				Resources:   []string{"clusters"},
			},
		},
	}
}

func getDefaultOperations() []admissionregistrationv1.RuleWithOperations {
	return []admissionregistrationv1.RuleWithOperations{
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"services"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"serviceaccounts"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"apps"},
				APIVersions: []string{"v1"},
				Resources:   []string{"statefulsets"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"apps"},
				APIVersions: []string{"v1"},
				Resources:   []string{"deployments"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"apps"},
				APIVersions: []string{"v1"},
				Resources:   []string{"daemonsets"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE", "DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"networking.k8s.io"},
				APIVersions: []string{"v1"},
				Resources:   []string{"ingresses"},
			},
		},

		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"pods", "pods/status"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE", "DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"secrets"},
			},
		},

		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE", "DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"nodes"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"storage.k8s.io"},
				APIVersions: []string{"v1"},
				Resources:   []string{"storageclasses"},
			},
		},
	}
}

func getDefaultCrdOperations() []admissionregistrationv1.RuleWithOperations {
	return []admissionregistrationv1.RuleWithOperations{
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"k3k.io"},
				APIVersions: []string{"v1alpha1"},
				Resources:   []string{"virtualclusterpolicies", "clusters"},
			},
		},

		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE", "DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"networking.higress.io"},
				APIVersions: []string{"v1"},
				Resources:   []string{"mcpbridges"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE", "DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"longhorn.io"},
				APIVersions: []string{"v1beta2"},
				Resources:   []string{"replicas"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"CREATE", "UPDATE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"longhorn.io"},
				APIVersions: []string{"v1beta2"},
				Resources:   []string{"nodes"},
			},
		},
		{
			Operations: []admissionregistrationv1.OperationType{"DELETE"},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{"apps.kubeblocks.io"},
				APIVersions: []string{"v1alpha1"},
				Resources:   []string{"clusters"},
			},
		},
	}
}

func (w *WebHookMutate) CreateOrUpdate(caBound []byte, svcName string, namespace string, hookName string, rules []admissionregistrationv1.RuleWithOperations) error {
	path := "/mutate"
	port := int32(9443)
	sideEffects := admissionregistrationv1.SideEffectClassNoneOnDryRun
	policy := admissionregistrationv1.Ignore

	// 创建 webhook 配置对象
	webhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MutatingWebhookConfiguration",
			APIVersion: "admissionregistration.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: hookName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: getSvcHost(namespace),
				// 添加命名空间选择器，处理所有命名空间
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "kubernetes.io/metadata.name",
							Operator: metav1.LabelSelectorOpExists,
						},
					},
				},
				Rules: rules,
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      svcName,
						Namespace: namespace,
						Path:      &path,
						Port:      &port,
					},
					CABundle: caBound,
				},
				AdmissionReviewVersions: []string{"v1"},
				SideEffects:             &sideEffects,
				FailurePolicy:           &policy,
			},
		},
	}

	// 尝试获取现有的 webhook 配置
	existingConfig, err := w.sdk.ClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(w.sdk.Ctx, hookName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// 如果不存在，创建新的配置
			_, err := w.sdk.ClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(w.sdk.Ctx, webhookConfig, metav1.CreateOptions{})
			return err
		}
		return err
	}

	// 如果已存在，更新配置
	// 保留现有的 ResourceVersion 以避免冲突
	webhookConfig.ObjectMeta.ResourceVersion = existingConfig.ObjectMeta.ResourceVersion

	_, err = w.sdk.ClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(w.sdk.Ctx, webhookConfig, metav1.UpdateOptions{})
	return err
}
