package microapp

import (
	"errors"
	"net/http/httputil"

	"github.com/w7panel/w7panel/common/helper"
	microapp "github.com/w7panel/w7panel/k8s/pkg/apis/microapp/v1alpha1"
)

type MicroAppProxy struct {
	*microapp.MicroApp
	isClusterUser bool
	role          string
}

func NewMicroAppProxy(microapp *microapp.MicroApp, isClusterUser bool, role string) *MicroAppProxy {
	return &MicroAppProxy{
		MicroApp:      microapp,
		isClusterUser: isClusterUser,
		role:          role,
	}
}
func (m *MicroAppProxy) Proxy(path string) (*httputil.ReverseProxy, error) {
	// Check if RoleConfig pointer is nil

	roleConfig, ok := m.Spec.ConfigV2.Props.RoleConfig[m.role]
	if !ok {
		if m.isClusterUser && !m.MicroApp.IsFromRoot() {
			m.role = "founder" // 创始人角色有歧义 站点管理是创始人角色 集群用户是普通用户
		}
		roleConfig, ok = m.Spec.ConfigV2.Props.RoleConfig[m.role]
		if !ok {
			return nil, errors.New("not found proxy role config")
		}
	}
	proxyServer := roleConfig.ServerUrl
	headers := roleConfig.ProxyRequest.Headers
	query := roleConfig.ProxyRequest.Query
	return helper.ProxyUrl(proxyServer, path, "", headers, query)

}
