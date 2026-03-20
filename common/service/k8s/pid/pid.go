package pid

import (
	"log/slog"

	"github.com/w7panel/w7panel/common/service/k8s"
	"github.com/w7panel/w7panel/common/service/k8s/terminal"
	corev1 "k8s.io/api/core/v1"
)

type pid struct {
	rootSdk          *k8s.Sdk
	clientSdk        *k8s.Sdk
	isVirtual        bool
	virtualNamespace string
}

// var pidCache *helper.MemoryCache

//	func init() {
//		pidCache = helper.NewMemoryCache()
//	}
func NewPid(token string) (*pid, error) {
	tokenObj := k8s.NewK8sToken(token)
	var client *k8s.Sdk
	ns := "default"
	if tokenObj.IsK3kCluster() {
		clientSdk, err := k8s.NewK8sClient().Channel(token)
		if err != nil {
			return nil, err
		}
		client = clientSdk
		ns = tokenObj.GetNamespace()
	}
	root := k8s.NewK8sClient()

	return &pid{rootSdk: root.Sdk, clientSdk: client, isVirtual: tokenObj.IsK3kCluster(), virtualNamespace: ns}, nil
}

func NewPidTest(saName string) (*pid, error) {
	root := k8s.NewK8sClient()
	var client *k8s.Sdk
	isVirtual := false
	if saName != "" {
		k3kConfig := k8s.NewK3kConfig(saName, "k3k-"+saName, "http://test.cc")
		clientSdk, err := root.GetK3kClusterSdkByConfig(k3kConfig)
		if err != nil {
			return nil, err
		}
		client = clientSdk
		isVirtual = true
	}
	return &pid{rootSdk: root.Sdk, clientSdk: client, isVirtual: isVirtual, virtualNamespace: "k3k-" + saName}, nil
}

// func cacheKey(pod *corev1.Pod) string {
// 	return helper.Set(cacheKey(pod), pod.Status.PodIP)
// }

// 不能在子集群执行
func (p *pid) Handle(param PidParam) (*PidResult, error) {
	pid := 1 //节点文件管理默认1
	subPid := 0
	proxyIp := ""
	var agentPod *corev1.Pod
	if p.isVirtual {

		podfindApi := newPodFind(p.rootSdk.ClientSet, p.clientSdk.ClientSet)
		clusterPod, err := podfindApi.GetVirtualClusterNodePod(p.virtualNamespace, param.HostIp)
		if err != nil {
			return nil, err
		}
		err = checkPodRunning(clusterPod)
		if err != nil {
			return nil, err
		}
		// clusterPod.Status.HostIP 就是主集群pod ip
		daemonsetPod, err := p.rootSdk.GetDaemonsetAgentPod(p.rootSdk.GetNamespace(), clusterPod.Status.HostIP)
		if err != nil {
			slog.Error("get  daemonsetPod err", "err", err)
			return nil, err
		}
		err = checkPodRunning(daemonsetPod)
		if err != nil {
			return nil, err
		}
		// clusterPodKey := cacheKey(clusterPod)
		// val, ok := pidCache.Get(clusterPodKey)
		clusterPodPid, err := GetContainerPid(daemonsetPod, clusterPod, param.ContainerId, false, p.rootSdk)
		if err != nil {
			return nil, err
		}

		pid = clusterPodPid
		if param.ContainerId != "" && param.FromPodName != "" {
			//为啥前端传containerId 为了获取pid, 后期因要从annnatation获取pid缓存, 所以需要查询k3kInnerPod
			k3kInnerPod, err := podfindApi.GetFromPod(param.FromPodName, param.Namespace, true)
			if err != nil {
				return nil, err
			}
			k3kInnerPodPid, err := GetContainerPid(clusterPod, k3kInnerPod, param.ContainerId, false, p.rootSdk) //必须rootsdk
			if err != nil {
				return nil, err
			}
			subPid = k3kInnerPodPid
		}
		agentPod = daemonsetPod
		proxyIp = daemonsetPod.Status.PodIP
	} else {
		podfindApi := newPodFind(p.rootSdk.ClientSet, p.rootSdk.ClientSet)
		daemonsetPod, err := p.rootSdk.GetDaemonsetAgentPod(p.rootSdk.GetNamespace(), param.HostIp)
		if err != nil {
			slog.Error("get  daemonsetPod err", "err", err)
			return nil, err
		}
		if param.ContainerId != "" && param.FromPodName != "" {

			//为啥前端传containerId 为了获取pid, 后期因要从annnatation获取pid缓存, 所以需要查询k3kInnerPod
			rootPod, err := podfindApi.GetFromPod(param.FromPodName, param.Namespace, true)
			if err != nil {
				return nil, err
			}
			rootPid, err := GetContainerPid(daemonsetPod, rootPod, param.ContainerId, true, p.rootSdk)
			if err != nil {
				return nil, err
			}
			pid = rootPid

		}
		agentPod = daemonsetPod
		proxyIp = daemonsetPod.Status.PodIP

	}
	pwd := "/"
	if param.FromPodName != "" && param.FromPodContainerName != "" {
		pwd1, err := p.GetPwd(param)
		if err != nil {
			pwd = "/"
		} else {
			pwd = pwd1
		}
	}
	return &PidResult{
		Pid:      pid,
		SubPid:   subPid,
		ProxyIp:  proxyIp,
		AgentPod: agentPod,
		Pwd:      pwd,
	}, nil

}

func (self *pid) GetPwd(params PidParam) (string, error) {
	session := terminal.NewTerminalSession(nil)
	defer session.Close()
	sdk := self.rootSdk
	if self.isVirtual {
		sdk = self.clientSdk
	}
	err := sdk.RunExec(session, params.Namespace, params.FromPodName, params.FromPodContainerName, []string{"pwd"}, false)
	if err != nil {
		return "", err
	}
	return string(session.GetWriterBytes()), nil
}
