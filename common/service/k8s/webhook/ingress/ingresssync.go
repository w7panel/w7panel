package ingress

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ingress struct {
	*networkingv1.Ingress
}

func (ing *ingress) GetChildDomainConfig() string {
	if ing.Annotations == nil {
		return ""
	}
	return ing.Annotations["w7.cc/child-hosts"]
}

func (ing *ingress) IsChild() bool {
	_, ok := ing.Labels["parents"]
	return ok
}

func (ing *ingress) ParentName() string {
	name, ok := ing.Labels["parents"]
	if ok {
		return name
	}
	return ""
}

func (ing *ingress) GetChildDomains() []domain {
	config := ing.GetChildDomainConfig()
	if config == "" {
		return []domain{}
	}
	result := &[]domain{}
	err := json.Unmarshal([]byte(config), result)
	if err != nil {
		return []domain{}
	}
	return *result
}
func (ing *ingress) Has(d domain) bool {
	for _, d2 := range ing.GetChildDomains() {
		if d.Name == d2.Name && d.Host == d2.Host {
			return true
		}
	}
	return false
}

func (ing *ingress) IsEmptyChildDomain() bool {
	return len(ing.GetChildDomains()) == 0
}

type domain struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	AutoSsl     bool   `json:"autoSsl"`     //自动申请证书
	SslRedirect bool   `json:"sslRedirect"` //https强制跳转
}

func newIngressSync(client sigclient.Client, ing *networkingv1.Ingress) ingressSync {
	return ingressSync{
		client:  client,
		ingress: &ingress{ing},
	}
}

type ingressSync struct {
	client sigclient.Client
	*ingress
}

func (ing *ingressSync) Sync() {
	if ing.ingress.IsChild() {
		ing.syncChild()
	} else {
		ing.syncRoot()
	}
}

func (ing *ingressSync) SyncDelete() {
	if ing.ingress.IsChild() {
		ing.syncChildDelete()
	} else {
		ing.syncRootDelete()
	}
}

func (ing *ingressSync) syncRoot() error {
	childs, err := ing.getChildIngress(ing.Name, ing.Namespace)
	if err != nil {
		return err
	}
	childDomains := ing.GetChildDomains()
	for _, domain := range childDomains {
		parentName := ing.Name + "-" + domain.Name
		cloneIng := ing.Ingress.DeepCopy()
		cloneIng.Name = parentName
		syncIngress(ing.client, cloneIng, domain, "", ing.Name)
		for _, child := range childs.Items {
			parentName := ing.Name + "-" + domain.Name
			cloneIng2 := child.DeepCopy()
			cloneIng2.Name = child.Name + "-" + domain.Name
			syncIngress(ing.client, cloneIng2, domain, parentName, ing.Name)
		}
	}

	return nil
}

func (ing *ingressSync) syncChild() error {
	originParentName := ing.ParentName()
	parentIng := &networkingv1.Ingress{}
	err := ing.client.Get(context.TODO(), sigclient.ObjectKey{Namespace: ing.Namespace, Name: originParentName}, parentIng)
	if err != nil {
		return err
	}
	parentWrapperIng := &ingress{parentIng}
	childDomains := parentWrapperIng.GetChildDomains()
	for _, domain := range childDomains {
		parentName := originParentName + "-" + domain.Name
		cloneIng := ing.Ingress.DeepCopy()
		cloneIng.Name = ing.Name + "-" + domain.Name
		syncIngress(ing.client, cloneIng, domain, parentName, originParentName)
	}
	return nil
}

func (ing *ingressSync) syncChildDelete() error {
	childDomains := ing.GetChildDomains()
	for _, domain := range childDomains {
		deleteIngress(ing.client, ing.Name+"-"+domain.Name, ing.Namespace)
	}
	return nil
}

func (ing *ingressSync) syncRootDelete() error {
	req, err := labels.NewRequirement("w7.cc/ingress-rootname", selection.Equals, []string{ing.Name})
	if err != nil {
		return nil
	}
	selector := labels.NewSelector().Add(*req)
	listOptions := sigclient.ListOptions{LabelSelector: selector, Namespace: ing.Namespace}
	err = ing.client.DeleteAllOf(context.TODO(), &networkingv1.Ingress{}, &sigclient.DeleteAllOfOptions{ListOptions: listOptions})
	if err != nil {
		return err
	}
	return err
}

func (ing *ingressSync) getParentIng() (*ingress, error) {
	if !ing.ingress.IsChild() {
		return ing.getIngress(ing.ParentName(), ing.Namespace)
	}
	return ing.ingress, nil
}

func (ing *ingressSync) CurrentIsChild() bool {
	return ing.ingress.IsChild()
}

func (ing *ingressSync) getChildIngress(parentName, namespace string) (*networkingv1.IngressList, error) {
	return getChildIngress(parentName, ing.client, namespace)
}

func (ing *ingressSync) getIngress(name, namespace string) (*ingress, error) {
	rawing := &networkingv1.Ingress{}
	err := ing.client.Get(context.TODO(), sigclient.ObjectKey{Name: name, Namespace: namespace}, rawing)
	if err != nil {
		return nil, err
	}
	return &ingress{Ingress: rawing}, err
}

func deleteIngress(client sigclient.Client, name, namespace string) error {
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	return client.Delete(context.TODO(), ing, &sigclient.DeleteOptions{})
}
func getChildIngress(parentName string, client sigclient.Client, namespace string) (*networkingv1.IngressList, error) {
	ings := &networkingv1.IngressList{}
	req, err := labels.NewRequirement("parents", selection.Equals, []string{parentName})
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*req)
	err = client.List(context.TODO(), ings, &sigclient.ListOptions{Namespace: namespace, LabelSelector: selector})
	return ings, err
}

func syncIngress(client sigclient.Client, ingress *networkingv1.Ingress, domain domain, parentName, fromName string) (*networkingv1.Ingress, error) {
	newIngress := ingress.DeepCopy()
	newIngress.SetResourceVersion("")
	newIngress.SetUID("")

	if newIngress.Annotations == nil {
		newIngress.Annotations = make(map[string]string)
	}
	_, err := controllerutil.CreateOrUpdate(context.TODO(), client, newIngress, func() error {
		tls := []networkingv1.IngressTLS{
			{
				Hosts:      []string{domain.Host},
				SecretName: strings.ToLower(strings.ReplaceAll(domain.Host, ".", "-")) + "-tls-secret",
			},
		}
		newIngress.Spec.TLS = tls
		newIngress.Spec.Rules = ingress.Spec.Rules
		newIngress.Spec.Rules[0].Host = domain.Host
		// newIngress.Spec.Rules =
		// newIngress.Spec.Rules[0].IngressRuleValue
		if parentName == "" {
			delete(newIngress.Labels, "parents")
		} else {
			newIngress.Labels["parents"] = parentName
		}
		delete(newIngress.Annotations, "w7.cc/child-hosts")
		// newIngress.Labels["parents"] = rootName
		newIngress.Labels["w7.cc/ingress-rootname"] = fromName
		newIngress.Labels["w7.cc/hide"] = "true"
		if domain.AutoSsl {
			newIngress.Annotations["cert-manager.io/cluster-issuer"] = "w7-letsencrypt-prod"
			newIngress.Annotations["cert-manager.io/renew-before"] = "30m"
		} else {
			delete(newIngress.Annotations, "cert-manager.io/cluster-issuer")
			delete(newIngress.Annotations, "cert-manager.io/renew-before")
		}
		if domain.SslRedirect {
			newIngress.Annotations["w7.cc/ssl-redirect"] = "true"
		} else {
			delete(newIngress.Annotations, "w7.cc/ssl-redirect")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return newIngress, err
}

func doSync(client sigclient.Client, ing *networkingv1.Ingress, delete bool) {
	sync := newIngressSync(client, ing)
	if delete {
		sync.SyncDelete()
	} else {
		sync.Sync()
	}
}

func OnAdd(client sigclient.Client, ingress *networkingv1.Ingress) {
	appgroup.OnAddIngress(client, ingress)
	if !isMain(ingress) {
		return
	}
	doSync(client, ingress, false)
}

func OnUpdate(client sigclient.Client, old *networkingv1.Ingress, newObj *networkingv1.Ingress) {
	appgroup.OnUpdateIngress(client, old, newObj)
	olding := ingress{old}
	newing := ingress{newObj}
	if !isMain(newObj) {
		return
	}
	// path变更后 获取负载均衡的子域名
	if olding.Spec.Rules[0].IngressRuleValue.HTTP != nil && newing.Spec.Rules[0].IngressRuleValue.HTTP != nil {
		oldPath := olding.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path
		newPath := newing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path
		oldPathMd5 := strings.ToLower(helper.StringToMD5(oldPath))
		if oldPath != newPath {
			parentName := newObj.Name
			// if newObj.Labels["parents"] != "" {
			// 	parentName = newObj.Labels["parents"]
			// }
			childs, err := getChildIngress(parentName, client, newObj.Namespace)
			if err != nil {
				slog.Error("get child ingress error", "error", err)
			}
			if err == nil {
				for _, child := range childs.Items {
					if child.Labels != nil && child.Labels["parentsPath"] == (oldPathMd5) {
						childPointer := &child
						_, err = controllerutil.CreateOrPatch(context.TODO(), client, childPointer, func() error {
							childPointer.Labels["parentsPath"] = strings.ToLower(helper.StringToMD5(newing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path))
							childPointer.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path = newing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path
							return nil
						})
						if err != nil {
							slog.Error("patch child ingress parentsPath error", "error", err)
						}
					}
				}
			}
		}
	}
	// 删除不存在的子域名
	childDomains := olding.GetChildDomains()
	for _, domain := range childDomains {
		if !newing.Has(domain) {
			deleteIngress(client, newObj.Name+"-"+domain.Name, newObj.Namespace) //删除不存在的子域名
		}
	}
	// 同步
	doSync(client, newObj, false)

}

func OnDelete(client sigclient.Client, newObj *networkingv1.Ingress) {
	appgroup.OnDeleteIngress(client, newObj)
	if !isMain(newObj) {
		return
	}
	doSync(client, newObj, true)
}

// 是否非复制出来的域名
func isMain(obj *networkingv1.Ingress) bool {
	if obj.Labels == nil {
		return true
	}
	if obj.Labels["w7.cc/ingress-rootname"] != "" {
		return false
	}
	return true
}
