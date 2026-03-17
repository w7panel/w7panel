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

// func (p *pid) Handle(param PidParam) (*PidResult, error) {
// 	root := k8s.NewK8sClient()
// 	if p.k8stoken.IsVirtual() {
// 		client, err := k8s.NewK8sClient().Channel(p.k8stoken.GetToken())
// 		if err != nil {
// 			return nil, err
// 		}
// 		podfindApi := newPodFind(root.ClientSet, client.ClientSet)
// 		clusterPod, err := podfindApi.GetVirtualClusterNodePod(param.Namespace, param.HostIp)
// 		if err != nil {
// 			return nil, err
// 		}
// 		err = checkPodRunning(clusterPod)
// 		if err != nil {
// 			return nil, err
// 		}
// 		agentPod, err = podfindApi.GetPanelAgentPod(clusterPod.Status.HostIP)
// 		if err != nil {
// 			return nil, err
// 		}
// 		clusterPodPid, err := GetContainerPid(clusterPod, "", true, root.Sdk)
// 		if err != nil {
// 			return nil, err
// 		}

// 	} else {

// 	}

// }
