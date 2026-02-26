package shell

import (
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/gpu"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/informers"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var mu sync.Mutex

var oslock sync.Mutex

const (
	SERVRE_ENV_FILE = "/host/etc/systemd/system/k3s.service.env"
	AGENT_ENV_FILE  = "/host/etc/systemd/system/k3s-agent.service.env"
)

func initRegistries(k8sClient *k8s.Sdk) error {

	// 初始化k8s客户端
	ryaml, err := helper.ReadRegistries()
	if err != nil {
		ryaml = []byte("")
	}
	clabels := map[string]string{
		"data-hash": "init",
	}
	//有的话 不处理
	// configmap, err := k8sClient.ClientSet.CoreV1().ConfigMaps(k8sClient.GetNamespace()).Get(k8sClient.Ctx, "registries", metav1.GetOptions{})
	// if err != nil {
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "registries",
			Namespace: k8sClient.GetNamespace(),
		},
	}
	// }

	data := map[string]string{
		"default.cnf": string(ryaml),
	}
	applyconfig, err := applyv1.ExtractConfigMap(configmap, "k8s-offline")
	if err != nil {
		slog.Error("extract configmap error", "err", err)
		return err
	}
	applyconfig = applyconfig.WithData(data).WithLabels(clabels)
	_, err = k8sClient.ClientSet.CoreV1().ConfigMaps(k8sClient.GetNamespace()).Apply(k8sClient.Ctx, applyconfig, metav1.ApplyOptions{FieldManager: "k8s-offline"})
	if err != nil {
		slog.Error("create configmap error", "err", err)
		return err
	}
	return nil

}

func isControlNode(node *v1.Node) bool {
	nodeNameEnv, ok := os.LookupEnv("NODE_NAME")
	if !ok {
		slog.Error("get node name env error")
		return false
	}

	labels := node.Labels
	if labels["node-role.kubernetes.io/control-plane"] == "true" && nodeNameEnv == node.Name {
		return true
	}
	return false
}

func isCurrentDaemonsetNode(node *v1.Node) bool {
	nodeNameEnv, ok := os.LookupEnv("NODE_NAME")
	if !ok {
		slog.Error("get node name env error")
		return false
	}
	if nodeNameEnv == node.Name {
		// slog.Debug("isCurrentDaemonsetNode", "nodeNameEnv", nodeNameEnv, "node.Name", node.Name)
		return true
	}
	return false
}

func ShellWatch() {
	// 初始化k8s客户端

	k8sClient := k8s.NewK8sClient()
	gpu.InitK3sGpu(k8sClient.GetSdk()) //初始化gpu configmap
	controller := NewK3sConfigController(k8sClient.GetSdk())
	controller.Start()
}

type k3sConfigController struct {
	*k8s.Sdk
	KubeInformerFactory   informers.SharedInformerFactory
	gpuOperatorIsDeployed bool
	nodeLister            listerv1.NodeLister
}

func NewK3sConfigController(sdk *k8s.Sdk) *k3sConfigController {
	kubeInformerFactory := informers.NewSharedInformerFactory(sdk.ClientSet, 0)
	// nodeLister := kubeInformerFactory.Core().V1().Nodes().Lister()
	return &k3sConfigController{
		Sdk:                 sdk,
		KubeInformerFactory: kubeInformerFactory,
		// nodeLister:            nodeLister,
		gpuOperatorIsDeployed: false,
	}
}

// 关闭输出
func WaitForNamedCacheSync(controllerName string, stopCh <-chan struct{}, cacheSyncs ...cache.InformerSynced) bool {
	// klog.Infof("Waiting for caches to sync for %s", controllerName)

	if !cache.WaitForCacheSync(stopCh, cacheSyncs...) {
		// utilruntime.HandleError(fmt.Errorf("unable to sync caches for %s", controllerName))
		return false
	}

	// klog.Infof("Caches are synced for %s", controllerName)
	return true
}

func (s *k3sConfigController) Start() error {
	informer := s.WatchK3sRegistry()
	nodeInformer := s.WatchNodeInformer()
	secretInformer := s.WatchSecretInformer()
	nodeInformer.GetIndexer().List()
	// dsInformer := s.WatchDaemonset()
	stopCh := make(chan struct{})
	defer close(stopCh)
	s.KubeInformerFactory.Start(stopCh)
	if !WaitForNamedCacheSync("shellcontroller", stopCh, informer.HasSynced, nodeInformer.HasSynced, secretInformer.HasSynced) {
		slog.Debug("Failed to sync cache")
		return nil
	}
	// 启动定时任务
	s.StartGpuTimer()

	<-stopCh
	return nil
}

func (s *k3sConfigController) WatchSecretInformer() cache.SharedIndexInformer {
	informer := s.KubeInformerFactory.Core().V1().Secrets().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			// slog.Debug("update secret")
			slog.Debug("update secret")
			s.handleK3sEnvSecret(newObj.(*v1.Secret))
		},
	})
	return informer
}

// func (s *k3sConfigController) WatchConfigmapInformer() cache.SharedIndexInformer {
// informer := s.KubeInformerFactory.Core().V1().ConfigMaps().Informer()
// informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
// 	AddFunc: func(obj interface{}, isInit bool) {
// 		s.handleK3kConfig(obj.(*v1.ConfigMap))
// 	},
// 	UpdateFunc: func(oldObj, newObj interface{}) {
// 		// slog.Debug("update secret")
// 		slog.Debug("update secret")
// 		s.handleK3kConfig(newObj.(*v1.ConfigMap))
// 	},
// 	DeleteFunc: func(obj interface{}) {
// 		s.deleteK3kConfig(obj.(*v1.ConfigMap))
// 	},
// })
// return informer
// }

func (s *k3sConfigController) WatchDaemonset() cache.SharedIndexInformer {
	informer := s.KubeInformerFactory.Apps().V1().DaemonSets().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInit bool) {
			s.HandleGpuDaemonset(obj.(*appsv1.DaemonSet), false)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			s.HandleGpuDaemonset(newObj.(*appsv1.DaemonSet), false)
		},
		DeleteFunc: func(obj interface{}) {
			s.HandleGpuDaemonset(obj.(*appsv1.DaemonSet), true)
		},
	})
	return informer
}

func (s *k3sConfigController) WatchNodeInformer() cache.SharedIndexInformer {
	nodeInformer := s.KubeInformerFactory.Core().V1().Nodes()
	informer := nodeInformer.Informer()
	s.nodeLister = nodeInformer.Lister()
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInit bool) {
			node := obj.(*v1.Node)
			s.initK3sEnvSecret(node)
			s.initK3sNodeSwap(node)
			// slog.Debug("add node", "node", node.Name)
			if isControlNode(node) {
				// slog.Debug("control node", "node", node.Name)
				// 初始化k8s客户端
				k8sClient := k8s.NewK8sClient()
				initRegistries(k8sClient.Sdk)
				s.initK3sConfig(node)
			} else {
				slog.Error("not control node", "node", node.Name)
				// if !isInit {
				// 	slog.Debug("not init node", "node", node.Name)
				// 	return
				// }
				nodes, err := s.nodeLister.List(labels.Everything())
				if err != nil {
					slog.Error("list nodes error", "err", err)
					return
				}
				for _, controlNode := range nodes {
					if controlNode.Labels["node-role.kubernetes.io/control-plane"] == "true" { //如果是主节点
						// 如果不一致，更新节点的配置
						if controlNode.Annotations[k3sSwapAnnotation] != node.Annotations[k3sSwapAnnotation] {
							node.Annotations[k3sSwapAnnotation] = controlNode.Annotations[k3sSwapAnnotation]
							node.Annotations[k3sSwapLastModifyAnnotation] = "controller"
							_, err := s.Sdk.ClientSet.CoreV1().Nodes().Update(s.Ctx, node, metav1.UpdateOptions{})
							if err != nil {
								slog.Error("update node error", "error", err)
								return
							}
							slog.Debug("update node", "node", node.Name)
						}
					}
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			new := newObj.(*v1.Node)
			old := oldObj.(*v1.Node)
			newHash, ok := new.Labels["k3s.io/node-config-hash"]
			oldHash, ok1 := old.Labels["k3s.io/node-config-hash"]
			if newHash != oldHash && ok && ok1 {
				s.initK3sConfig(new)
			}
			s.updateK3sNodeSwap(new)
			s.loadPublicIp(new)
		},
	})
	return informer
}

func (s *k3sConfigController) initK3sConfig(node *v1.Node) error {
	if !isControlNode(node) {
		return nil
	}
	k8sClient := s.Sdk
	k3sConfig := NewK3sConfigByNode(node)
	//多个server 只有一个节点会执行初始化操作
	if k3sConfig.IsOutDB() && k3sConfig.DbUrl() == "" {
		return nil
	}
	if k3sConfig.IsNotFirstNode() {
		return nil
	}
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k3s.config",
			Namespace: "kube-system",
		},
	}
	data := map[string]string{
		"k3s.cluster-init":       k3sConfig.IsClusterInitString(),
		"k3s.datastore-endpoint": k3sConfig.DbUrl(),
		"k3s.config":             (k3sConfig.k3sRawConfig),
		"k3s.mode":               k3sConfig.GetMode(),
		"k3s.default-tls-san":    strings.Join(k3sConfig.GetDefaultTlsSanIp(), ","),
	}
	labels := map[string]string{
		"data-hash": "init",
	}
	slog.Debug("k3s config", "data", data)
	applyconfig, err := applyv1.ExtractConfigMap(configmap, "k8s-offline")
	if err != nil {
		slog.Error("extract configmap k3s.config  error", "err", err)
		return err
	}
	applyconfig = applyconfig.WithData(data).WithLabels(labels)
	_, err = k8sClient.ClientSet.CoreV1().ConfigMaps("kube-system").Apply(k8sClient.Ctx, applyconfig, metav1.ApplyOptions{FieldManager: "k8s-offline", Force: true})
	if err != nil {
		slog.Error("create configmap k3s.config error", "err", err)
		return err
	}
	return nil

}

func (s *k3sConfigController) WatchK3sRegistry() cache.SharedIndexInformer {
	informer := s.KubeInformerFactory.Core().V1().ConfigMaps().Informer()
	nodes, err := s.Sdk.ClientSet.CoreV1().Nodes().List(s.Ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("get nodes error", "err", err)
	}
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			configmap := newObj.(*v1.ConfigMap)
			oldconfigmap := oldObj.(*v1.ConfigMap)
			if configmap.Labels["data-hash"] != oldconfigmap.Labels["data-hash"] && configmap.Labels["data-hash"] != "init" {
				if configmap.Name == "registries" {
					slog.Debug("update registries")
					s.handleRegistry(configmap, nodes)
				}
				if configmap.Name == "k3s.config" && configmap.Namespace == "kube-system" {
					slog.Debug("update k3s config")
					s.handleK3sConfig(configmap, nodes)
				}

			}
		},
	})
	return informer
}

func (s *k3sConfigController) handleRegistry(config *v1.ConfigMap, nodes *v1.NodeList) error {
	mu.Lock()
	defer mu.Unlock()
	slog.Info("handle registry")
	data, ok := config.Data["default.cnf"]
	if !ok {
		return nil
	}
	slog.Debug("reg.yaml", "data", data)
	_, err := helper.YamlParse([]byte(data))
	if err != nil {
		slog.Error("parse yaml error", "err", err)
		return err
	}

	err = helper.WriterRegistries([]byte(data))
	if err != nil {
		slog.Error("run shell error", "err", err)
		return err
	}
	s.restart(nodes)
	return err
}

func (s *k3sConfigController) handleK3sConfig(config *v1.ConfigMap, nodes *v1.NodeList) error {

	for _, node := range nodes.Items {
		if isControlNode(&node) {
			k3sConfig := NewK3sConfigByNode(&node)
			change := false
			if config.Data["k3s.cluster-init"] == "true" && !k3sConfig.IsClusterInit() {
				k3sConfig.k3sConfigYaml["cluster-init"] = "true"
				change = true
			}
			if config.Data["k3s.datastore-endpoint"] != "" && !k3sConfig.IsOutDB() && !k3sConfig.IsClusterInit() {
				k3sConfig.k3sConfigYaml["datastore-endpoint"] = config.Data["k3s.datastore-endpoint"]
				change = true
			}
			// if config.Data["k3s.tls-san"] != "" { //可能会覆盖之前的配置，暂时不启用
			ips := strings.Split(config.Data["k3s.tls-san"], ",")
			if len(ips) > 0 {
				k3sConfig.k3sConfigYaml["tls-san"] = ips
			} else {
				delete(k3sConfig.k3sConfigYaml, "tls-san")
			}
			change = true

			// }
			if change {
				data, err := helper.YamlToBytes(k3sConfig.k3sConfigYaml)
				if err != nil {
					slog.Error("k3s.config yaml to bytes error", "err", err)
					return err
				}
				err = helper.WriteK3sConfig(data)
				if err != nil {
					slog.Error("write /etc/rancher/k3s/config.yaml error", "err", err)
					return err
				}

				s.restartNode(&node)
			}
		}
	}

	return nil
}

func (s *k3sConfigController) restart(nodes *v1.NodeList) error {
	time.AfterFunc(time.Second*15, func() {
		oslock.Lock()
		s.restartNodes(nodes)
		defer oslock.Unlock()
	})
	return nil
}

func (s *k3sConfigController) restartNodes(nodes *v1.NodeList) error {
	nodeNameEnv, ok := os.LookupEnv("NODE_NAME")
	if !ok {
		slog.Error("get node name error")
		return nil
	}
	for _, node := range nodes.Items {
		if node.Name == nodeNameEnv {
			s.restartNode(&node)
		}
	}
	return nil
}

func (s *k3sConfigController) restartNode(nodes *v1.Node) error {
	time.AfterFunc(time.Second*15, func() {
		slog.Debug("restart k3s")
		oslock.Lock()
		shell := "systemctl restart k3s"
		if !isControlNode(nodes) {
			shell = "systemctl restart k3s-agent"
		}
		helper.RunNcenterBinsh(shell)
		defer oslock.Unlock()
		os.Exit(0)
	})
	return nil

}
