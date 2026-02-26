package v1alpha1

import (
	"encoding/json"

	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var GROUPTYPE string

const (
	CUSTOM = "custom"
	HELM   = "helm"
	ZPK    = "zpk"
)

var WORKLOADSTATUS string

const (
	// StatusUnknown indicates that a release is in an uncertain state.
	StatusUnknown = "unknown"
	// StatusDeployed indicates that the release has been pushed to Kubernetes.
	StatusDeployed = "deployed"
	// StatusFailed indicates that the release was not successfully deployed.
	StatusFailed = "failed"
	// StatusUninstalling indicates that a uninstall operation is underway.
	StatusPendingInstall = "deploying"

	StatusDeploying = "deploying"
)

type ResourceInfo struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	ApiVersion   string `json:"apiVersion"`
	Kind         string `json:"kind"`
	DeployStatus string `json:"deployStatus"` //是否部署成功 ready
	DeployTitle  string `json:"deployTitle"`  //部署标题
}

func (info ResourceInfo) IsWorkLoad() bool {
	if info.Kind == "Deployment" || info.Kind == "StatefulSet" || info.Kind == "DaemonSet" || info.Kind == "Job" || info.Kind == "CronJob" {
		return true
	}
	return false
}

type AkName struct {
	ApiVersion string `json:"groupVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type DeployItem struct {
	Identifie    string         `json:"identifie"`     //应用标识
	Title        string         `json:"title"`         //应用标题
	ResourceList []ResourceInfo `json:"resourcesList"` //资源列表
	DeployStatus string         `json:"deployStatus"`  //部署状态
}

func (d DeployItem) ComputeIsFailed() bool {
	workList := lo.Filter(d.ResourceList, func(r ResourceInfo, _ int) bool {
		return r.IsWorkLoad()
	})
	return lo.ContainsBy(workList, func(r ResourceInfo) bool {
		return r.DeployStatus == "failed"
	})
}
func (d DeployItem) ComputeIsReady() bool {
	workList := lo.Filter(d.ResourceList, func(r ResourceInfo, _ int) bool {
		return r.IsWorkLoad()
	})
	deployedCount := lo.CountBy(workList, func(r ResourceInfo) bool {
		return r.DeployStatus == StatusDeployed
	})
	return len(workList) == deployedCount
}
func (d DeployItem) ComputeDeployStatus() string {
	if d.ComputeIsFailed() {
		return StatusFailed
	} else if !d.ComputeIsReady() {
		return StatusDeploying
	} else {
		return StatusDeployed
	}
}
func (d DeployItem) ComputeChangeStatus() {
	d.DeployStatus = d.ComputeDeployStatus()
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AppGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppGroupSpec   `json:"spec"`
	Status AppGroupStatus `json:"status"`
}

type ApplicationType string

const (
	ApplicationTypeHelm    ApplicationType = "helm"    //helm chart
	ApplicationTypeZpk     ApplicationType = "zpk"     //zpk 制品库
	ApplicationTypeCustom  ApplicationType = "custom"  //自定义应用
	ApplicationTypeConsole ApplicationType = "console" //控制台应用
)

type HelmConfig struct {
	ChartName  string `json:"chartName"`  //helm chart name
	Repository string `json:"repository"` //helm chart repository
	Version    string `json:"version"`    //版本号
}

type AppGroupSpec struct {
	Identifie        string `json:"identifie"`        //应用标识
	Type             string `json:"type"`             // "helm" or "zpk" or "custom"
	Version          string `json:"version"`          //版本号
	UpgradingVersion string `json:"upgradingVersion"` //更新中的版本号
	Title            string `json:"title"`            //应用标题
	Logo             string `json:"logo"`             //logo地址
	Description      string `json:"description"`      //应用描述
	Suffix           string `json:"suffix"`           //应用名后缀
	// Domains       []string   `json:"domains"`       //域名列表
	// DefaultDomain string     `json:"defaultDomain"` //默认域名
	ZpkUrl     string     `json:"zpkUrl"`     //制品库地址
	HelmConfig HelmConfig `json:"helmConfig"` //helm配置
	// Annotations   map[string]string `json:"annotations"`   //annotations
	IsHelm bool `json:"isHelm"` //是否为helm应用
}

// ApplicationItemStatus 用于记录单个应用的部署状态信息
type AppGroupItemStatus struct {
	Kind              string      `json:"kind"` // "helm" or "deployment" or "statefulset" or "daemonset" or "job"
	ApiVersion        string      `json:"apiVersion"`
	Name              string      `json:"name"`
	Title             string      `json:"title"`
	Ready             bool        `json:"ready"`
	IsHelmWorkLoad    bool        `json:"isHelmWorkLoad"` //是否为helm应用
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty" protobuf:"bytes,8,opt,name=creationTimestamp"`
	DeployStatus      string      `json:"deployStatus"`
	IsZeroReplicas    bool        `json:"isZeroReplicas"` //是否暂停部署 只有一个应用并且replicas为0
}

type AppGroupStatus struct {
	Items          []AppGroupItemStatus `json:"items"`
	DeployItems    []DeployItem         `json:"deployInfo"`
	DeployStatus   string               `json:"deployStatus"`
	Ready          bool                 `json:"ready"`
	IsZeroReplicas bool                 `json:"isZeroReplicas"` //是否暂停部署 只有一个应用并且replicas为0
}

func (s AppGroupStatus) ComputeDeployIsFailed() bool {
	return lo.ContainsBy(s.DeployItems, func(r DeployItem) bool {
		return r.ComputeIsFailed()
	})
}
func (s AppGroupStatus) ComputeDeployIsReady() bool {
	deployedCount := lo.CountBy(s.DeployItems, func(r DeployItem) bool {
		return r.ComputeIsReady()
	})
	return len(s.DeployItems) == deployedCount
}
func (s AppGroupStatus) ComputeDeployStatus() string {
	if s.ComputeDeployIsFailed() {
		return StatusFailed
	} else if !s.ComputeDeployIsReady() {
		return StatusDeploying
	} else {
		return StatusDeployed
	}
}

func (g *AppGroup) ComputeStatus() {
	for in, v1 := range g.Status.DeployItems {
		if v1.DeployStatus != StatusDeployed {
			g.Status.DeployItems[in].DeployStatus = g.Status.DeployItems[in].ComputeDeployStatus()
		}
	}
	if g.Status.DeployStatus != StatusDeployed {
		g.Status.DeployStatus = g.Status.ComputeDeployStatus()
	}
	if g.Status.DeployStatus == StatusDeployed {
		if g.Spec.UpgradingVersion != "" && g.Spec.Version != g.Spec.UpgradingVersion {
			g.Spec.Version = g.Spec.UpgradingVersion
		}
	}
	g.Status.IsZeroReplicas = false
	if len(g.Status.Items) == 1 {
		g.Status.IsZeroReplicas = g.Status.Items[0].IsZeroReplicas
	}

	if len(g.Status.Items) > 0 {
		ready := lo.EveryBy(g.Status.Items, func(v AppGroupItemStatus) bool {
			return v.Ready || v.IsZeroReplicas
		})
		g.Status.Ready = ready
	}
	if len(g.Status.Items) == 0 {
		// g.Status.IsZeroReplicas = true
		g.Status.Ready = true
	}
}

func (g *AppGroup) GetDomains() []string {
	if (g.Annotations != nil) && (g.Annotations["w7.cc/domains"] != "") {
		result := []string{}
		err := json.Unmarshal([]byte(g.Annotations["w7.cc/domains"]), &result)
		if err != nil {
			return []string{}
		}
		return result
	}
	return []string{}
}

func (g *AppGroup) GetDefaultDomain() string {
	if (g.Annotations != nil) && (g.Annotations["w7.cc/default-domain"] != "") {
		return g.Annotations["w7.cc/default-domain"]
	}
	return ""
}

func (g *AppGroup) SetDomain(domains []string) {
	if g.Annotations == nil {
		g.Annotations = make(map[string]string)
	}
	data, err := json.Marshal(domains)
	if err == nil {
		g.Annotations["w7.cc/domains"] = string(data)
	}
}
func (g *AppGroup) AppendDomain(url string) {
	domains := g.GetDomains()
	domains = append(domains, url)
	g.SetDomain(lo.Uniq(domains))
}

func (g *AppGroup) DeleteDomain(url string) {
	domains := g.GetDomains()
	domains = lo.Filter(domains, func(item string, index int) bool {
		return item != url
	})
	g.SetDomain(domains)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AppGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AppGroup `json:"items"`
}
