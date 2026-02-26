package appgroup

import (
	"context"
	"errors"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	v1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// cert-manager cert
type Certificate struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (c *Certificate) DeepCopyObject() runtime.Object {
	return &Certificate{
		TypeMeta:   c.TypeMeta,
		ObjectMeta: c.ObjectMeta,
	}
}

var schemeGroupVersion = schema.GroupVersion{Group: "cert-manager.io", Version: "v1"}

func init() {
	k8s.GetScheme().AddKnownTypes(schemeGroupVersion, &Certificate{})
	// metav1.Reg
}

func getappgroup(client sigclient.Client, ingress *networkingv1.Ingress) (*v1alpha1.AppGroup, error) {
	if ingress.Labels != nil && ingress.Labels["group"] != "" {
		groupName := ingress.Labels["group"]
		appgroup, err := GetAppgroup(groupName, ingress.Namespace, client)
		if err != nil {
			return nil, err
		}
		return appgroup, err
	}
	return nil, errors.New("no group")
}

func getUrl(ingress *networkingv1.Ingress) string {
	scheme := "http://"
	if ingress.Annotations != nil && ingress.Annotations["cert-manager.io/cluster-issuer"] == "w7-letsencrypt-prod" {
		scheme = "https://"
	}
	// if ingress.Spec.TLS != nil && len(ingress.Spec.TLS) > 0 {
	// 	scheme = "https://"
	// }
	if ingress.Spec.Rules[0].HTTP.Paths[0].Path != "/" {
		return scheme + ingress.Spec.Rules[0].Host + "/" + ingress.Spec.Rules[0].HTTP.Paths[0].Path
	}
	return scheme + ingress.Spec.Rules[0].Host
}

//	func patchAppgroup(client sigclient.Client, group *v1alpha1.AppGroup, applyFunc func() error) {
//		controllerutil.CreateOrPatch(client, group, func() error {
//			return applyFunc()
//		})
//	}
func OnAddIngress(client sigclient.Client, ingress *networkingv1.Ingress) {
	group, err := getappgroup(client, ingress)
	if err != nil {
		slog.Error("get appgroup error", "error", err)
		return
	}

	controllerutil.CreateOrPatch(context.Background(), client, group, func() error {
		group.AppendDomain(getUrl(ingress))
		return nil
	})

}

func delCert(client sigclient.Client, namespace, certName string) {
	cert := &Certificate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Certificate",
			APIVersion: "cert-manager.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: namespace,
		},
	}
	err := client.Delete(context.Background(), cert)
	if err != nil {
		slog.Error("delete cert error", "err", err)
	}
}

// 旧绪有Autossl ingress被添加了 new不autossl 就同步删除certmanager
func checkAutossl(client sigclient.Client, old *networkingv1.Ingress, new *networkingv1.Ingress) {
	if old.Annotations == nil || old.Spec.TLS == nil {
		return
	}
	if old.Spec.TLS != nil && len(old.Spec.TLS) == 0 {
		return
	}
	if old.Annotations["cert-manager.io/cluster-issuer"] == "w7-letsencrypt-prod" {
		if new.Annotations == nil {
			new.Annotations = make(map[string]string)
		}
		_, ok := new.Annotations["cert-manager.io/cluster-issuer"]
		if !ok {
			go delCert(client, old.Namespace, old.Spec.TLS[0].SecretName)
		}
	}
}

// 原先有Autossl ingress被删除了 就同步删除certmanager
func checkAutosslDel(client sigclient.Client, old *networkingv1.Ingress) {
	if old.Annotations == nil || old.Spec.TLS == nil {
		return
	}
	if old.Spec.TLS != nil && len(old.Spec.TLS) == 0 {
		return
	}
	_, ok := old.Annotations["cert-manager.io/cluster-issuer"]
	if ok {
		go delCert(client, old.Namespace, old.Spec.TLS[0].SecretName)
	}
}

func OnUpdateIngress(client sigclient.Client, old *networkingv1.Ingress, new *networkingv1.Ingress) {
	checkAutossl(client, old, new)
	group, err := getappgroup(client, new)
	if err != nil {
		slog.Error("get appgroup error", "error", err)
		return
	}
	controllerutil.CreateOrPatch(context.Background(), client, group, func() error {
		group.DeleteDomain(getUrl(old))
		group.AppendDomain(getUrl(new))
		return nil
	})
}

func OnDeleteIngress(client sigclient.Client, ingress *networkingv1.Ingress) {
	checkAutosslDel(client, ingress) //检查
	group, err := getappgroup(client, ingress)
	if err != nil {
		slog.Error("get appgroup error", "error", err)
		return
	}
	controllerutil.CreateOrPatch(context.Background(), client, group, func() error {
		group.DeleteDomain(getUrl(ingress))
		return nil
	})
}
