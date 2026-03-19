package pid

import (
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

type pid struct {
	rootSdk          *k8s.Sdk
	clientSdk        *k8s.Sdk
	isVirtual        bool
	virtualNamespace string
}

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

// 不能在子集群执行
func (p *pid) Handle(param PidParam) (*PidResult, error) {
	pid := 1 //节点文件管理默认1
	subPid := 0
	if p.isVirtual {

		podfindApi := newPodFind(p.rootSdk.ClientSet, p.clientSdk.ClientSet)
		clusterPod, err := podfindApi.GetVirtualClusterNodePod(param.Namespace, p.virtualNamespace)
		if err != nil {
			return nil, err
		}
		err = checkPodRunning(clusterPod)
		if err != nil {
			return nil, err
		}
		clusterPodPid, err := LoadPid(clusterPod)
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
			k3kInnerPodPid, err := GetContainerPid(clusterPod, k3kInnerPod, param.ContainerId, false, p.clientSdk)
			if err != nil {
				return nil, err
			}
			subPid = k3kInnerPodPid
		}
	} else {
		podfindApi := newPodFind(p.rootSdk.ClientSet, p.rootSdk.ClientSet)
		if param.ContainerId != "" && param.FromPodName != "" {
			//为啥前端传containerId 为了获取pid, 后期因要从annnatation获取pid缓存, 所以需要查询k3kInnerPod
			rootPod, err := podfindApi.GetFromPod(param.FromPodName, param.Namespace, true)
			if err != nil {
				return nil, err
			}
			rootPid, err := LoadPid(rootPod)
			if err != nil {
				return nil, err
			}
			pid = rootPid
		}
	}
	return &PidResult{
		Pid:    pid,
		SubPid: subPid,
	}, nil

}
