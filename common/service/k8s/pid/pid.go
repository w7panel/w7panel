package pid

import (
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

type pid struct {
	k8stoken *k8s.K8sToken
}

func NewPid(token string) *pid {
	tokenObj := k8s.NewK8sToken(token)
	return &pid{k8stoken: tokenObj}
}

// 不能在子集群执行
func (p *pid) Handle(param PidParam) (*PidResult, error) {
	root := k8s.NewK8sClient()
	pid := 1 //节点文件管理默认1
	subPid := 0
	if p.k8stoken.IsVirtual() {
		client, err := k8s.NewK8sClient().Channel(p.k8stoken.GetToken())
		if err != nil {
			return nil, err
		}
		podfindApi := newPodFind(root.ClientSet, client.ClientSet)
		clusterPod, err := podfindApi.GetVirtualClusterNodePod(param.Namespace, param.HostIp)
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
			k3kInnerPodPid, err := GetContainerPid(clusterPod, k3kInnerPod, param.ContainerId, false, client)
			if err != nil {
				return nil, err
			}
			subPid = k3kInnerPodPid
		}
	} else {
		podfindApi := newPodFind(root.ClientSet, root.ClientSet)
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
