package k8s

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	"github.com/rancher/k3k/pkg/controller/kubeconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type singleton struct {
	*Sdk
	sdks    map[string]*Sdk
	expires map[string]int64 // 缓存过期时间(unix timestamp)
	mu      sync.Mutex
}

// instance 是一个包级别的变量，用于保存唯一的单例对象
var instance *singleton

// once 是一个用于确保初始化操作只执行一次的 sync.Once 对象
var once sync.Once

type Channel interface {
	Channel(token string) (*Sdk, error)
}

// GetInstance 方法返回唯一的单例对象
func NewK8sClient() *singleton {
	// 使用 sync.Once 确保初始化操作只执行一次
	once.Do(func() {
		sdk := NewK8sClientInner()
		instance = &singleton{}
		instance.Sdk = sdk
		instance.sdks = make(map[string]*Sdk)
		instance.expires = make(map[string]int64)
	})
	return instance
}

func (s *singleton) GetSdk() *Sdk {
	return s.Sdk
}

func (s *singleton) ChannelLocal(token string, forceLocal bool) (*Sdk, error) {
	startTime := time.Now()
	
	if forceLocal {
		loadStart := time.Now()
		result, err := s.loadFromCache(token)
		log.Printf("[PERF] ChannelLocal - loadFromCache took %v", time.Since(loadStart))
		log.Printf("[PERF] ChannelLocal total time %v", time.Since(startTime))
		return result, err
	}
	
	channelStart := time.Now()
	result, err := s.Channel(token)
	log.Printf("[PERF] ChannelLocal - Channel took %v", time.Since(channelStart))
	log.Printf("[PERF] ChannelLocal total time %v", time.Since(startTime))
	return result, err
}

func (s *singleton) Channel(token string) (*Sdk, error) {
	startTime := time.Now()
	tokenobj := NewK8sToken(token)
	
	isK3kStart := time.Now()
	isK3k := tokenobj.IsK3kCluster()
	log.Printf("[PERF] Channel - IsK3kCluster took %v", time.Since(isK3kStart))
	
	if isK3k {
		k3kStart := time.Now()
		result, err := s.GetK3kClusterSdk(tokenobj)
		log.Printf("[PERF] Channel - GetK3kClusterSdk took %v", time.Since(k3kStart))
		log.Printf("[PERF] Channel total time %v", time.Since(startTime))
		return result, err
	}
	
	cacheStart := time.Now()
	result, err := s.loadFromCache(token)
	log.Printf("[PERF] Channel - loadFromCache took %v", time.Since(cacheStart))
	log.Printf("[PERF] Channel total time %v", time.Since(startTime))
	return result, err
}
func (s *singleton) loadFromCache(token string) (*Sdk, error) {

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.sdks) > 100 {
		s.sdks = make(map[string]*Sdk)
	}
	sdk, ok := s.sdks[token]
	if !ok {
		sdk2, err := s.Sdk.Channel(token)
		if err != nil {
			return nil, err
		}
		s.sdks[token] = sdk2
		sdk = sdk2

	}
	return sdk, nil
}
func (s *singleton) GetK3kClusterSdkByConfig(k3kconfig *K3kConfig) (*Sdk, error) {
	startTime := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查缓存是否存在且未过期
	cacheCheckStart := time.Now()
	if sdk, ok := s.sdks[k3kconfig.Name]; ok {
		if expireTime, ok := s.expires[k3kconfig.Name]; ok && expireTime > time.Now().Unix() {
			log.Printf("[PERF] GetK3kClusterSdkByConfig - cache hit took %v", time.Since(cacheCheckStart))
			log.Printf("[PERF] GetK3kClusterSdkByConfig total time %v", time.Since(startTime))
			return sdk, nil
		}
	}
	log.Printf("[PERF] GetK3kClusterSdkByConfig - cache miss/check took %v", time.Since(cacheCheckStart))

	toSigStart := time.Now()
	sigClient, err := s.Sdk.ToSigClient()
	log.Printf("[PERF] GetK3kClusterSdkByConfig - ToSigClient took %v", time.Since(toSigStart))
	if err != nil {
		return nil, err
	}
	
	getKubeConfigStart := time.Now()
	kubeconfig, err := GetK3kKubeConfig(sigClient, k3kconfig)
	log.Printf("[PERF] GetK3kClusterSdkByConfig - GetK3kKubeConfig took %v", time.Since(getKubeConfigStart))
	if err != nil {
		return nil, err
	}
	
	clientConfigStart := time.Now()
	clientconfig := clientcmd.NewDefaultClientConfig(*kubeconfig, &clientcmd.ConfigOverrides{})
	restConfig, err := clientconfig.ClientConfig()
	log.Printf("[PERF] GetK3kClusterSdkByConfig - client config setup took %v", time.Since(clientConfigStart))
	if err != nil {
		return nil, err
	}
	
	newSdkStart := time.Now()
	sdk, err := NewForRestConfig(restConfig, "default")
	log.Printf("[PERF] GetK3kClusterSdkByConfig - NewForRestConfig took %v", time.Since(newSdkStart))
	if err != nil {
		return nil, err
	}
	
	createTokenStart := time.Now()
	token, err := sdk.CreateTokenRequest(k3kconfig.Name, 7200, []string{})
	log.Printf("[PERF] GetK3kClusterSdkByConfig - CreateTokenRequest took %v", time.Since(createTokenStart))
	if err != nil {
		return nil, err
	}

	// 缓存结果并设置过期时间(1小时)
	channelStart := time.Now()
	result, err := sdk.Channel(token)
	log.Printf("[PERF] GetK3kClusterSdkByConfig - final Channel took %v", time.Since(channelStart))
	if err == nil {
		s.sdks[k3kconfig.Name] = result
		s.expires[k3kconfig.Name] = time.Now().Add(time.Hour).Unix()
	}
	log.Printf("[PERF] GetK3kClusterSdkByConfig total time %v", time.Since(startTime))
	return result, err
}
func (s *singleton) GetK3kClusterSdk(k8stoken *K8sToken) (*Sdk, error) {
	startTime := time.Now()
	
	getConfigStart := time.Now()
	k3kconfig, err := k8stoken.GetK3kConfig()
	log.Printf("[PERF] GetK3kClusterSdk - GetK3kConfig took %v", time.Since(getConfigStart))
	if err != nil {
		return nil, err
	}
	
	byConfigStart := time.Now()
	result, err := s.GetK3kClusterSdkByConfig(k3kconfig)
	log.Printf("[PERF] GetK3kClusterSdk - GetK3kClusterSdkByConfig took %v", time.Since(byConfigStart))
	log.Printf("[PERF] GetK3kClusterSdk total time %v", time.Since(startTime))
	return result, err
}

func (s *singleton) Clear(k3kName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sdks, k3kName)
	delete(s.expires, k3kName)
}

func extract(ctx context.Context, client client.Client, cluster *v1alpha1.Cluster, hostServerIP string) (*clientcmdapi.Config, error) {
	cfg := kubeconfig.New()
	kubeconfig, err := cfg.Generate(ctx, client, cluster, hostServerIP, 443)
	return kubeconfig, err
}

// GetK3kKubeConfig 从Kubernetes集群中获取k3k的kubeconfig配置
//
// 参数：
//
//	sigClient: Kubernetes的client，用于与Kubernetes集群进行交互
//	k3kconfig: K3kConfig结构体，包含了k3k集群的配置信息
//
// 返回值：
//
//	*clientcmdapi.Config: 解析后的kubeconfig配置
//	error: 如果获取kubeconfig失败，则返回错误信息
func GetK3kKubeConfig(sigClient client.Client, k3kconfig *K3kConfig) (*clientcmdapi.Config, error) {
	startTime := time.Now()

	secret := &corev1.Secret{}
	kubeConfigName := "k3k-" + k3kconfig.Name + "-kubeconfig"
	
	getSecretStart := time.Now()
	err := sigClient.Get(context.TODO(), types.NamespacedName{Name: kubeConfigName, Namespace: k3kconfig.Namespace}, secret)
	log.Printf("[PERF] GetK3kKubeConfig - Get Secret took %v", time.Since(getSecretStart))
	if err != nil {
		return nil, err
	}

	kubeconfigYaml := secret.Data["kubeconfig.yaml"]
	if len(kubeconfigYaml) == 0 {
		return nil, errors.New("kubeconfig.yaml is empty")
	}
	
	loadStart := time.Now()
	kubeconfig, err := clientcmd.Load(kubeconfigYaml)
	log.Printf("[PERF] GetK3kKubeConfig - Load kubeconfig took %v", time.Since(loadStart))
	log.Printf("[PERF] GetK3kKubeConfig total time %v", time.Since(startTime))
	return kubeconfig, err
}

func NewForCmdConfig(kubeconfig *clientcmdapi.Config) (*Sdk, error) {
	clientconfig := clientcmd.NewDefaultClientConfig(*kubeconfig, &clientcmd.ConfigOverrides{})
	restConfig, err := clientconfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	sdk, err := NewForRestConfig(restConfig, "default")
	if err != nil {
		return nil, err
	}
	return sdk, nil
}
