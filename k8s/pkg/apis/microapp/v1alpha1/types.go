package v1alpha1

import (
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MicroApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MicroAppSpec `json:"spec"`
}

type MicroAppConfig struct {
	Props map[string]string `json:"props,omitempty"`
}

// type Headers map[string]string
// type Query map[string]string

type ProxyRequest struct {
	Headers map[string]string `json:"headers,omitempty"`
	Query   map[string]string `json:"query,omitempty"`
}

type Role struct {
	// +k8s:optional
	// +optional
	// +nullable
	LoadMode      string            `json:"load_mode,omitempty"`
	ProxyRequest  ProxyRequest      `json:"proxy_request,omitempty"`
	FrontendProps map[string]string `json:"frontend_props,omitempty"`
	ServerUrl     string            `json:"serverUrl,omitempty"`
}

type RoleConfig map[string]Role

// type RoleConfig struct {
// 	// +k8s:optional
// 	// +optional
// 	// +nullable
// 	Founder *Role `json:"founder,omitempty"`
// 	// +k8s:optional
// 	// +optional
// 	// +nullable
// 	Super *Role `json:"super,omitempty"`
// 	// +k8s:optional
// 	// +optional
// 	// +nullable
// 	Normal *Role `json:"normal,omitempty"`
// 	// +k8s:optional
// 	// +optional
// 	// +nullable
// 	Tech *Role `json:"tech,omitempty"`
// }

type Props struct {
	// +k8s:optional
	// +optional
	// +nullable
	RoleConfig RoleConfig `json:"roleConfig,omitempty"`
}

type MicroAppConfig2 struct {
	// +k8s:optional
	// +optional
	// +nullable
	Props Props `json:"props,omitempty"`
}

type MicroAppSpec struct {
	Framework   string         `json:"framework,omitempty"`
	BackendUrl  string         `json:"backendUrl,omitempty"`
	FrontendUrl string         `json:"frontendUrl"`
	ProxyUrl    string         `json:"proxyUrl,omitempty"`
	Title       string         `json:"title"`
	Logo        string         `json:"logo,omitempty"`
	Config      MicroAppConfig `json:"config,omitempty"`
	// +k8s:optional
	// +optional
	// +nullable
	ConfigV2    MicroAppConfig2 `json:"config-v2,omitempty"`
	Description string          `json:"description,omitempty"`
	// +k8s:optional
	// +optional
	// +nullable
	Bindings []Bindings `json:"bindings,omitempty"`
}

type Menu struct {
	Displayorder int    `json:"displayorder,omitempty"`
	Do           string `json:"do"`
	Icon         string `json:"icon,omitempty"`
	IsDefault    int    `json:"is_default,omitempty"`
	Location     string `json:"location,omitempty"`
	Title        string `json:"title"`
}

type Bindings struct {
	Framework         string `json:"framework,omitempty"`
	IsDefaultRegister int    `json:"is_default_register,omitempty"`
	Location          string `json:"location,omitempty"`
	Support           string `json:"support,omitempty"`
	Menu              []Menu `json:"menu,omitempty"`
	Name              string `json:"name"`
	Status            int    `json:"status,omitempty"`
	Title             string `json:"title"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MicroAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MicroApp `json:"items"`
}

func (c *MicroApp) IsFromRoot() bool {
	if c.Labels != nil && c.Labels["microapp.w7.cc/from"] == "root" {
		return true
	}
	return false
}

// 角色数量
func (c *MicroApp) RoleCount() int {
	return lo.CountBy(c.Spec.Bindings, func(item Bindings) bool {
		return item.Support == "thirdparty_cd"
	})
}
