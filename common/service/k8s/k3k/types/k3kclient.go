package types

import (
	"context"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"

	// _ "github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K3kClient struct {
	k3kClient client.Client
}

func NewK3kClient(client client.Client) *K3kClient {
	return &K3kClient{
		k3kClient: client,
	}
}

func (k *K3kClient) Create(k3kUser *K3kUser) error {

	job := ToK3kJob(k3kUser)
	err := k.k3kClient.Create(context.Background(), job)
	if err != nil {
		return err
	}
	k3kUser.Running(job.Name)
	err = k.k3kClient.Update(context.Background(), k3kUser.ServiceAccount)
	if err != nil {
		return err
	}
	return nil

}

func (k *K3kClient) Delete(user *K3kUser) error {
	namespace := user.GetK3kNamespace()
	clusterName := user.GetK3kName()
	// cluster :=
	cluster := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "k3k.io/v1alpha1",
		},
	}

	err := k.k3kClient.Delete(context.Background(), cluster)
	if err != nil {
		return err
	}
	return err
}

func (k *K3kClient) DeleteNamespace(user *K3kUser) error {
	namespace := user.GetK3kNamespace()
	// cluster :=
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	err := k.k3kClient.Delete(context.Background(), ns)
	if err != nil {
		return err
	}
	return err
}

func (k *K3kClient) GetPolicy(user *K3kUser) (*v1alpha1.VirtualClusterPolicy, error) {
	policy := &v1alpha1.VirtualClusterPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      user.GetClusterPolicy(),
			Namespace: user.GetK3kNamespace(),
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualClusterPolicy",
			APIVersion: "k3k.io/v1alpha1",
		},
	}
	err := k.k3kClient.Get(context.Background(), types.NamespacedName{Namespace: user.Namespace, Name: user.GetClusterPolicy()}, policy)
	if err != nil {
		return nil, err
	}
	return policy, err
}

func (k *K3kClient) GetPolicyByName(name string) (*v1alpha1.VirtualClusterPolicy, error) {
	policy := &v1alpha1.VirtualClusterPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualClusterPolicy",
			APIVersion: "k3k.io/v1alpha1",
		},
	}
	err := k.k3kClient.Get(context.Background(), types.NamespacedName{Name: name}, policy)
	if err != nil {
		return nil, err
	}
	return policy, err
}

func (k *K3kClient) GetKubeConfig(user *K3kUser) (*clientcmdapi.Config, error) {
	kubeconfig, err := k8s.GetK3kKubeConfig(k.k3kClient, user.ToK3kConfig())
	return kubeconfig, err
}

func (k *K3kClient) GetKubeConfigYaml(user *K3kUser) ([]byte, error) {
	kubeconfig, err := k.GetKubeConfig(user)
	if err != nil {
		return nil, err
	}
	kubeconfigData, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubeconfigData, err
}

func (k *K3kClient) GetCluster(user *K3kUser) (*v1alpha1.Cluster, error) {
	cluster := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      user.GetK3kName(),
			Namespace: user.GetK3kNamespace(),
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "k3k.io/v1alpha1",
		},
	}
	err := k.k3kClient.Get(context.Background(), types.NamespacedName{Namespace: user.Namespace, Name: user.GetK3kName()}, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, err
}
func (k *K3kClient) GetK3kConfig() (*K3kConfig, error) {

	configmap := &corev1.ConfigMap{}
	err := k.k3kClient.Get(context.Background(), types.NamespacedName{Namespace: "kube-system", Name: "k3k.config"}, configmap)
	if err != nil {
		return nil, err
	}
	secretConfig := NewK3kConfigByConfigmap(configmap)
	return secretConfig, err
}

/**
kind: ConfigMap
apiVersion: v1
metadata:
    name: k3k.config
    namespace: kube-system
type: Opaque
data:
    allowConsoleRegister: "true"
    defaultPolicyName: "yibqvzoz"

*/
//	func (k *K3kClient) TokenToK3kUser(token string) (*k3kUser, error) {
//		k8stoken := k8s.NewK8sToken(token)
//		saName, err := k8stoken.GetSaName()
//		if err != nil {
//			return nil, err
//		}
//		sa, err := k.sdk.GetServiceAccount(k.sdk.GetNamespace(), saName)
//		if err != nil {
//			return nil, err
//		}
//		kuser := NewK3kUser(sa)
//		return kuser, nil
//	}
// func Extract(ctx context.Context, client client.Client, cluster *v1alpha1.Cluster, hostServerIP string) (*clientcmdapi.Config, error) {
// 	cfg := kubeconfig.New()
// 	kubeconfig, err := cfg.Generate(ctx, client, cluster, hostServerIP)
// 	return kubeconfig, err
// }
