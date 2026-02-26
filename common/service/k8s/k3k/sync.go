package k3k

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	higressV1 "github.com/alibaba/higress/client/pkg/apis/networking/v1"
	"github.com/rancher/k3k/k3k-kubelet/translate"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type K3kToHostTranslator struct {
	*translate.ToHostTranslator
	maxLength int
}

func NewK3kToHostTranslator(clusterName, clusterNamespace string) *K3kToHostTranslator {
	trans := &translate.ToHostTranslator{
		ClusterName:      clusterName,
		ClusterNamespace: clusterNamespace,
	}
	return &K3kToHostTranslator{
		ToHostTranslator: trans,
		maxLength:        34,
	}
}
func (t *K3kToHostTranslator) TranslateName(namespace string, name string) string {
	var names []string

	if namespace == "" {
		names = []string{name, t.ClusterName}
	} else {
		names = []string{name, namespace, t.ClusterName}
	}

	namePrefix := strings.Join(names, "-")
	// use + as a separator since it can't be in an object name
	nameKey := strings.Join(names, "+")
	// it's possible that the suffix will be in the name, so we use hex to make it valid for k8s
	nameSuffix := hex.EncodeToString([]byte(nameKey))

	return helper.SafeConcatName(t.maxLength, namePrefix, nameSuffix)
}

type SyncObjectInterface interface {
	GetName() string
	GetNamespace() string
}
type SyncObject struct {
	Name      string
	Namespace string
}

type K3kSync struct {
	VirtualName      string `form:"virtualName"`
	VirtualNamespace string `form:"virtualNamespace"`
	K3kName          string `form:"k3kName"`
	K3kNamespace     string `form:"k3kNamespace"`
	K3kMode          string `form:"k3kMode"`
}

func (k *K3kSync) DeepCopy() *K3kSync {
	return &K3kSync{
		VirtualName:      k.VirtualName,
		VirtualNamespace: k.VirtualNamespace,
		K3kName:          k.K3kName,
		K3kNamespace:     k.K3kNamespace,
		K3kMode:          k.K3kMode,
	}
}

func NewSyncObject(name, namespace string) *SyncObject {
	return &SyncObject{Name: name, Namespace: namespace}
}

func (obj *SyncObject) GetName() string {
	return obj.Name
}
func (obj *SyncObject) GetNamespace() string {
	return obj.Namespace
}

// SyncHttp 函数将同步对象数据到 HTTP 服务器
//
// 参数:
//
//	obj: 实现 SyncObjectInterface 接口的对象
//	path: 发送请求的 URL 路径
//
// 返回值:
//
//	如果同步成功，返回 nil
//	如果同步失败，返回错误信息
func SyncHttp(obj SyncObjectInterface, path string) error {
	urlvalues := url.Values{}
	urlvalues.Add("virtualName", obj.GetName())
	urlvalues.Add("virtualNamespace", obj.GetNamespace())
	urlvalues.Add("k3kName", os.Getenv("K3K_NAME"))
	urlvalues.Add("k3kNamespace", os.Getenv("K3K_NAMESPACE"))
	urlvalues.Add("k3kMode", os.Getenv("K3K_MODE"))

	slog.Info("sync start", "urlvalues", urlvalues)

	// postUrl := "http://" + os.Getenv("ROOT_SVCNAME") + "." + os.Getenv("ROOT_NAMESPACE") + ".svc:8000/k8s/k3k/" + path
	postUrl := "http://" + os.Getenv("ROOT_POD_IP") + ":8000/k8s/k3k/" + path
	if helper.IsLocalMock() {
		postUrl = "http://172.16.1.126:9007/k8s/k3k/sync-ingress"
	}
	client := helper.RetryHttpClient()
	resp, err := client.R().SetFormDataFromValues(urlvalues).Post(postUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		slog.Error("sync  error", "errstr", resp.String(), "postUrl", postUrl)
		return fmt.Errorf("sync  error, status code: %d :%s", resp.StatusCode(), resp.String())
	}
	return nil
}

func SyncIngress(params *K3kSync) error {
	k3kConfig := k8s.NewK3kConfig(params.K3kName, params.K3kNamespace, helper.GetApiServerHost(params.K3kNamespace))
	root := k8s.NewK8sClient()
	clientsdk, err := root.GetK3kClusterSdkByConfig(k3kConfig)
	if err != nil {
		return err
	}

	trans := translate.ToHostTranslator{
		ClusterName:      params.K3kName,
		ClusterNamespace: params.K3kNamespace,
	}
	//secret name 过长导致cert-manager 无法生成证书(请求)，所以不需要转换
	// transSecret := NewK3kToHostTranslator(params.K3kName, params.K3kNamespace)
	hostIngressName := trans.TranslateName(params.VirtualNamespace, params.VirtualName)

	ingress, err := clientsdk.ClientSet.NetworkingV1().Ingresses(params.VirtualNamespace).Get(root.Ctx, params.VirtualName, metav1.GetOptions{})
	if err != nil {
		//找不到就删除，找到了就更新
		if errors.IsNotFound(err) {
			err = root.ClientSet.NetworkingV1().Ingresses(params.K3kNamespace).Delete(root.Ctx, hostIngressName, metav1.DeleteOptions{})
			if err != nil {
				slog.Error("delete ingress error", "err", err)
				return err
			}
		}
		slog.Error("get virtual ingress error", "err", err)
		return err
	}
	ingress = ingress.DeepCopy()
	trans.TranslateTo(ingress)
	ingress.Annotations["kubernetes.io/ingress.class"] = "higress"
	if params.K3kMode == "virtual" {
		newAnnations := make(map[string]string)
		for k, v := range ingress.Annotations {
			if k == "kubernetes.io/ingress.class" ||
				// k == "cert-manager.io/cluster-issuer" ||
				k == "higress.io/resource-definer" ||
				// k == "cert-manager.io/renew-before" ||
				k == "higress.io/ssl-redirect" ||
				k == "w7.cc/ssl-redirect" || k == "w7.cc/filecache" || k == "k3k.io/name" || k == "k3k.io/namespace" {
				newAnnations[k] = v
			}
		}
		ingress.Annotations = newAnnations
	}
	rules := ingress.Spec.Rules
	secretNames := []string{}
	// if ingress.Spec.TLS != nil && len(ingress.Spec.TLS) > 0 {
	// 	for k, tls := range ingress.Spec.TLS {
	// 		ingress.Spec.TLS[k].Hosts = nil
	// 		if len(tls.SecretName) > 0 {
	// 			ingress.Spec.TLS[k].SecretName = transSecret.TranslateName(params.VirtualNamespace, tls.SecretName)
	// 			secretNames = append(secretNames, tls.SecretName) //旧的secretName
	// 		}
	// 	}
	// }
	for k, rule := range rules {
		for k1, path := range rule.HTTP.Paths {
			if params.K3kMode == "virtual" { //虚拟集群默认使用80
				rules[k].HTTP.Paths[k1].Backend.Resource = nil
				if rules[k].HTTP.Paths[k1].Backend.Service == nil {
					rules[k].HTTP.Paths[k1].Backend.Service = &networkingv1.IngressServiceBackend{}
				}
				rules[k].HTTP.Paths[k1].Backend.Service.Name = k3kConfig.GetVirtualIngressServiceName()
				rules[k].HTTP.Paths[k1].Backend.Service.Port = networkingv1.ServiceBackendPort{
					Number: 80,
				}
				continue
			}

			service := rules[k].HTTP.Paths[k1].Backend.Service
			if service == nil {
				continue
			}
			rules[k].HTTP.Paths[k1].Backend.Service.Name = trans.TranslateName(params.VirtualNamespace, path.Backend.Service.Name)

		}
	}
	for _, secretName := range secretNames {
		_, err = clientsdk.ClientSet.CoreV1().Secrets(params.VirtualNamespace).Get(root.Ctx, secretName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) { //没有就新建一个，然后同步 保证cert-manager能反向同步到子集群

				defautlData := map[string][]byte{
					"tls.key": []byte(""),
					"tls.crt": []byte(""),
				}
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretName,
						Namespace: params.VirtualNamespace,
					},
					Data: defautlData,
					Type: corev1.SecretTypeTLS,
				}
				_, err = clientsdk.ClientSet.CoreV1().Secrets(params.VirtualNamespace).Create(root.Ctx, secret, metav1.CreateOptions{})
				if err != nil {
					slog.Error("create ingress secret error", "err", err)
					continue
				}
				// if err == nil { //同步secret webhook会自动同步secret，这里就不手动同步了
				// 	secretSync := params.DeepCopy()
				// 	secretSync.VirtualName = secretName
				// 	SyncSecret(secretSync)
				// }
			}
			continue
		}
	}
	// ingress.Spec.Rules
	_, err = root.ClientSet.NetworkingV1().Ingresses(params.K3kNamespace).Get(root.Ctx, hostIngressName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = root.ClientSet.NetworkingV1().Ingresses(params.K3kNamespace).Create(root.Ctx, ingress, metav1.CreateOptions{})
			if err != nil {
				slog.Warn("create ingress error", "err", err)
				return err
			}
			return nil
		}
		return err
	}
	_, err = root.ClientSet.NetworkingV1().Ingresses(params.K3kNamespace).Update(root.Ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		slog.Warn("update ingress error", "err", err)
		return err
	}
	return nil
}

func SyncIngressHttp(ingress *networkingv1.Ingress) error {
	return SyncHttp(ingress, "sync-ingress")
}

func SyncMcpBridge(params *K3kSync) error {
	k3kConfig := k8s.NewK3kConfig(params.K3kName, params.K3kNamespace, helper.GetApiServerHost(params.K3kNamespace))
	root := k8s.NewK8sClient()
	clientsdk, err := root.GetK3kClusterSdkByConfig(k3kConfig)
	if err != nil {
		return err
	}
	rootSigClient, err := root.ToSigClient()
	if err != nil {
		return err
	}
	clientSigClient, err := clientsdk.ToSigClient()
	if err != nil {
		return err
	}

	trans := translate.ToHostTranslator{
		ClusterName:      params.K3kName,
		ClusterNamespace: params.K3kNamespace,
	}

	// hostName := trans.TranslateName(params.VirtualNamespace, params.VirtualName)
	mcpBridge := &higressV1.McpBridge{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.VirtualName,
			Namespace: params.VirtualNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "McpBridge",
			APIVersion: "networking.higress.io/v1",
		},
	}
	nName := types.NamespacedName{Namespace: params.VirtualNamespace, Name: params.VirtualName}
	err = clientSigClient.Get(root.Ctx, nName, mcpBridge)
	if err != nil {
		if errors.IsNotFound(err) {
			err = clientSigClient.Delete(root.Ctx, mcpBridge)
			if err != nil {
				return err
			}
		}
		return err
	}

	newMcpBridge := mcpBridge.DeepCopy()
	trans.TranslateTo(newMcpBridge)

	_, err = controllerutil.CreateOrUpdate(root.Ctx, rootSigClient, newMcpBridge, func() error {
		newMcpBridge.Spec.Registries = mcpBridge.Spec.Registries
		return nil
	})
	if err != nil {
		return err
	}
	return nil

}

func SyncMcpBridgeHttp(bridge SyncObjectInterface) error {
	return SyncHttp(bridge, "sync-mcpbridge")
}

func SyncConfigmap(params *K3kSync) error {
	k3kConfig := k8s.NewK3kConfig(params.K3kName, params.K3kNamespace, helper.GetApiServerHost(params.K3kNamespace))
	root := k8s.NewK8sClient()
	clientsdk, err := root.GetK3kClusterSdkByConfig(k3kConfig)
	if err != nil {
		return err
	}

	trans := translate.ToHostTranslator{
		ClusterName:      params.K3kName,
		ClusterNamespace: params.K3kNamespace,
	}
	hostName := trans.TranslateName(params.VirtualNamespace, params.VirtualName)

	configMap, err := clientsdk.ClientSet.CoreV1().ConfigMaps(params.VirtualNamespace).Get(root.Ctx, params.VirtualName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			err = root.ClientSet.CoreV1().ConfigMaps(params.K3kNamespace).Delete(root.Ctx, hostName, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
		return err
	}
	rootSigClient, err := root.ToSigClient()
	if err != nil {
		return err
	}
	newConfigmap := configMap.DeepCopy()
	trans.TranslateTo(newConfigmap)

	_, err = controllerutil.CreateOrUpdate(root.Ctx, rootSigClient, newConfigmap, func() error {
		newConfigmap.Data = configMap.Data
		newConfigmap.BinaryData = configMap.BinaryData
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func SyncConfigmapHttp(configmap *corev1.ConfigMap) error {
	return SyncHttp(configmap, "sync-configmap")
}

func SyncSecret(params *K3kSync) error {
	k3kConfig := k8s.NewK3kConfig(params.K3kName, params.K3kNamespace, helper.GetApiServerHost(params.K3kNamespace))
	root := k8s.NewK8sClient()
	clientsdk, err := root.GetK3kClusterSdkByConfig(k3kConfig)
	if err != nil {
		return err
	}

	// trans := translate.ToHostTranslator{
	// 	ClusterName:      params.K3kName,
	// 	ClusterNamespace: params.K3kNamespace,
	// }

	//secret name 过长导致cert-manager 无法生成证书(请求)，所以需要转换
	// trans := NewK3kToHostTranslator(params.K3kName, params.K3kNamespace) //certmanager 证书无法生成 先不转化secret name
	hostName := params.VirtualName //trans.TranslateName(params.VirtualNamespace, params.VirtualName)

	secret, err := clientsdk.ClientSet.CoreV1().Secrets(params.VirtualNamespace).Get(root.Ctx, params.VirtualName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			err = root.ClientSet.CoreV1().Secrets(params.K3kNamespace).Delete(root.Ctx, hostName, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
		return err
	}
	if secret.Type != corev1.SecretTypeTLS {
		return nil
	}
	rootSigClient, err := root.ToSigClient()
	if err != nil {
		return err
	}
	newSecret := secret.DeepCopy()
	newSecret.SetResourceVersion("")
	newSecret.SetUID("")
	newSecret.SetNamespace(params.K3kNamespace)
	// trans.TranslateTo(newSecret)
	if newSecret.Labels == nil {
		newSecret.Labels = make(map[string]string)
	}

	_, err = controllerutil.CreateOrUpdate(root.Ctx, rootSigClient, newSecret, func() error {
		newSecret.Data = secret.Data
		newSecret.Labels["w7.cc/sync"] = "true"
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// cert-manager 集群同步到子集群的secret
func SyncToChildSecret(secret *corev1.Secret) error {
	if secret.Annotations == nil {
		slog.Error("SyncToChildSecret", "err", "secret annotations is nil")
		return nil
	}
	// if secret.Type != corev1.SecretTypeTLS {
	// 	slog.Error("SyncToChildSecret", "err", "secret type is not TLS", "secretName", secret.Name)
	// 	return nil
	// }
	//只有cert-manager.io/common-name的secret才会同步到子集群
	_, ok := secret.Annotations["cert-manager.io/common-name"]
	if !ok {
		slog.Error("SyncToChildSecret", "err", "secret annotations cert-manager.io/common-name is nil")
		return nil
	}
	K3kNamespace := secret.Namespace
	if !strings.HasPrefix(K3kNamespace, "k3k-") {
		slog.Error("SyncToChildSecret", "err", "K3kNamespace is not k3k-xxx")
		return nil
	}
	K3kName := strings.ReplaceAll(K3kNamespace, "k3k-", "")

	if K3kName == "" {
		slog.Error("SyncToChildSecret", "err", "K3kName is empty")
		return nil
	}
	params := &K3kSync{
		K3kName:          K3kName,
		K3kNamespace:     secret.Namespace,
		VirtualNamespace: "default",
		VirtualName:      secret.Name,
	}

	k3kConfig := k8s.NewK3kConfig(params.K3kName, secret.Namespace, helper.GetApiServerHost(secret.Namespace))
	root := k8s.NewK8sClient()
	clientsdk, err := root.GetK3kClusterSdkByConfig(k3kConfig)
	if err != nil {
		slog.Error("SyncToChildSecret", "err", err)
		return err
	}
	clientSigClient, err := clientsdk.ToSigClient()
	if err != nil {
		slog.Error("SyncToChildSecret", "err", err)
		return err
	}
	newSecret := secret.DeepCopy()
	_, ok2 := newSecret.Annotations["k3k.io/name"] //如果有k3k.io/name 则认为是子集群同步过来的，不需要再翻译回来
	if ok2 {
		trans := translate.ToHostTranslator{
			ClusterName:      params.K3kName,
			ClusterNamespace: params.K3kNamespace,
		}
		trans.TranslateFrom(newSecret)
	}

	newSecret.Namespace = "default"
	if newSecret.Labels == nil {
		newSecret.Labels = make(map[string]string)
	}
	slog.Error("SyncToChildSecret", "secretName", secret.Name, "K3kNamespace", K3kNamespace)
	_, err = controllerutil.CreateOrPatch(root.Ctx, clientSigClient, newSecret, func() error {
		newSecret.Annotations = secret.Annotations
		newSecret.Labels = secret.Labels
		// newSecret.Namespace = params.VirtualNamespace
		newSecret.Data = secret.Data
		return nil
	})
	if err != nil {
		slog.Error("SyncToChildSecret", "err", err)
		return err
	}
	return nil
}
func SyncSecretHttp(secret *corev1.Secret) error {
	if secret.Type != corev1.SecretTypeTLS {
		return nil
	}
	if secret.Annotations != nil {
		_, ok := secret.Annotations["cert-manager.io/common-name"] //cert-manager 集群同步到子集群的secret
		if ok {
			//SyncToChildSecret 同步过来的不来回同步
			return nil
		}
	}
	return SyncHttp(secret, "sync-secret")
}

func SyncHttpAfter(object SyncObjectInterface, path string) error {
	time.AfterFunc(time.Second*5, func() {
		SyncHttp(object, path)
	})
	return nil
}

func SyncAgentIngress() error {
	if helper.IsK3kVirtual() {
		//面板代理 ingress 同步到主集群
		return SyncHttp(NewSyncObject("ing-k3k-agent", "default"), "sync-ingress")
	}
	return nil
}

func SyncDownStatic(name, zpkurl string) error {
	if helper.IsK3kVirtual() {
		//面板代理 ingress 同步到主集群
		return SyncHttp(NewSyncObject(name, zpkurl), "sync-down-static")
	}
	return nil
}
func SyncMicroApp() error {
	if helper.IsK3kVirtual() {
		//面板代理 ingress 同步到主集群
		return SyncHttp(NewSyncObject("default", "default"), "sync-microapp")
	}
	return nil
}
