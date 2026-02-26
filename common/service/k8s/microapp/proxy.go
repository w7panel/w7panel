package microapp

import (
	"errors"
	"net/http/httputil"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	microapp "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/microapp/v1alpha1"
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
		return nil, errors.New("not found proxy role config")
	}
	proxyServer := roleConfig.ServerUrl
	headers := roleConfig.ProxyRequest.Headers
	query := roleConfig.ProxyRequest.Query
	return helper.ProxyUrl(proxyServer, path, "", headers, query)

}
