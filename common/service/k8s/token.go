package k8s

import (
	"sync"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var tokenCache = make(map[string]int64)
var lock sync.Mutex

type K8sToken struct {
	token        string
	mu           sync.RWMutex
	claims       jwtv5.MapClaims
	claimsErr    error
	claimsParsed bool
}

type K3kConfig struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	ApiServer string `json:"apiServer"`
}

func NewK3kConfig(name, namespace, apiServer string) *K3kConfig {
	return &K3kConfig{
		Name:      name,
		Namespace: namespace,
		ApiServer: apiServer,
	}
}

func (t *K3kConfig) GetK3kAgentName() string {
	return helper.GetK3kAgentName(t.Name)
}

// 内部ingress name
func (t *K3kConfig) GetK3kAgentInnerIngressHost() string {
	return helper.GetK3kAgentName(t.Name) + ".w7panel.xyz"
}

func (t *K3kConfig) GetK3kAgentLbHost() string {
	return t.GetVirtualIngressServiceName() + "." + t.Namespace + ".svc:8000"
}

func (t *K3kConfig) GetK3kServer0Name() string {
	return helper.GetK3kServer0Name(t.Name)
}

func (t *K3kConfig) GetK3kServer0ContainerName() string {
	return helper.GetK3kServer0ContainerName(t.Name)
}

func (t *K3kConfig) GetCacheKey() string {
	return t.Name + "-" + t.Namespace
}

func (t *K3kConfig) ToAgentSvc() string {
	if helper.IsLocalMock() {
		return "http://v56.fan.b2.sz.w7.com"
	}
	return "http://" + t.GetK3kAgentName() + ":8000"
}

func (u *K3kConfig) GetVirtualIngressServiceName() string {
	return u.Namespace + "-service-w7"
}

func NewK8sToken(token string) *K8sToken {
	return &K8sToken{
		token: token,
	}
}

// parseClaims 解析并缓存 JWT claims
func (t *K8sToken) parseClaims() (jwtv5.MapClaims, error) {
	t.mu.RLock()
	if t.claimsParsed {
		t.mu.RUnlock()
		return t.claims, t.claimsErr
	}
	t.mu.RUnlock()

	t.mu.Lock()
	defer t.mu.Unlock()

	// 双重检查
	if t.claimsParsed {
		return t.claims, t.claimsErr
	}

	data := jwtv5.MapClaims{}
	jwtToken, _, err := jwtv5.NewParser().ParseUnverified(t.token, data)
	if err != nil {
		t.claimsErr = err
		t.claimsParsed = true
		return nil, err
	}
	t.claims = data
	t.claimsParsed = true
	_ = jwtToken // 仅用于验证，无需使用
	return t.claims, nil
}

func (t *K8sToken) Cache() error {
	expiretime, err := t.GetExpireTime()
	if err != nil {
		return err
	}
	if expiretime == nil {
		return errors.New("expiretime is nil")
	}
	lock.Lock()
	defer lock.Unlock()
	tokenCache[t.token] = expiretime.Unix()
	return nil
}

func (t *K8sToken) IsCacheToken() bool {
	unixtime, ok := tokenCache[t.token]
	if !ok {
		return false
	}
	if unixtime < time.Now().Unix() {
		lock.Lock()
		defer lock.Unlock()
		delete(tokenCache, t.token)
		return false
	}
	return ok
}

func (t *K8sToken) GetSaName() (string, error) {
	sa, _ := getTokenSaName(t.token)
	if sa == "" {
		return "", errors.New("token中没有找到serviceaccount")
	}
	return sa, nil
}

func (t *K8sToken) GetAudience() (jwtv5.ClaimStrings, error) {
	data, err := t.parseClaims()
	if err != nil {
		return nil, err
	}

	aud, ok := data["aud"]
	if !ok {
		return nil, errors.New("token中无audience")
	}
	switch v := aud.(type) {
	case []interface{}:
		result := make(jwtv5.ClaimStrings, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				result[i] = s
			}
		}
		return result, nil
	case string:
		return jwtv5.ClaimStrings{v}, nil
	default:
		return nil, errors.New("audience格式错误")
	}
}

func (t *K8sToken) GetExpireTime() (*jwtv5.NumericDate, error) {
	data, err := t.parseClaims()
	if err != nil {
		return nil, err
	}
	expireData, ok := data["exp"].(float64)
	if !ok || expireData == 0 {
		return nil, nil
	}
	exp := jwtv5.NewNumericDate(time.Unix(int64(expireData), 0))
	return exp, nil
}

// u.Name,
//
//	u.GetRole(),
//	u.Annotations[W7_CONSOLE_ID],
//	u.GetK3kName(),
//	u.GetK3kNamespace(),
//	u.GetApiServerHost(),
//	u.GetClusterMode(),
//	u.GetClusterPolicy(),
//	u.GetLockVersion(),
//	u.GetClusterPolicyVersion(),
//	"https://kubernetes.default.svc.cluster.local",
//	"k3s",
//
// 判断是不是虚拟集群
func (t *K8sToken) IsK3kCluster() bool {
	s, err := t.GetAudience()
	if err != nil {
		return false
	}
	return len(s) >= 6
}
func (t *K8sToken) IsVirtual() bool {
	return t.K3kMode() == "virtual"
}

func (t *K8sToken) K3kMode() string {
	if t.IsK3kCluster() {
		v, err := t.GetAudience()
		if err == nil {
			return v[6]
		}
	}
	return ""
}

// k3kuser.go 如果是集群用户返回founder 为了显示菜单
// k8stoken.go 如果是集群用户返回普通用户角色，因为auth.go要根据role限制请求
func (t *K8sToken) Role() string {
	if t.IsK3kCluster() {
		return "normal"
	}
	s, err := t.GetAudience()
	if err != nil {
		return "normal"
	}
	return s[1]
}

func (t *K8sToken) IsShared() bool {
	return t.K3kMode() == "shared"
}

// u.Name,  0
//
//	u.GetRole(), 1
//	u.Annotations[W7_CONSOLE_ID], 2
//	u.GetK3kName(), 3
//	u.GetK3kNamespace(), 4
//	u.GetApiServerHost(), 5
//	u.GetClusterMode(), 6
//	u.GetClusterPolicy(), 7
//	u.GetLockVersion(), 8
//	u.GetClusterPolicyVersion(), 9
//	"https://kubernetes.default.svc.cluster.local",
//	"k3s",
func (t *K8sToken) GetPolicyName() string {
	if t.IsK3kCluster() {
		v, err := t.GetAudience()
		if err == nil && len(v) > 7 {
			return v[7]
		}
	}
	return ""
}

func (t *K8sToken) GetLockVersion() string {
	if t.IsK3kCluster() {
		v, err := t.GetAudience()
		if err == nil && len(v) > 8 {
			return v[8]
		}
	}
	return "1"
}

func (t *K8sToken) GetK3kPolicyVersion() string {
	if t.IsK3kCluster() {
		v, err := t.GetAudience()
		if err == nil && len(v) > 9 {
			return v[9]
		}
	}
	return "1"
}

func (t *K8sToken) GetK3kConfig() (*K3kConfig, error) {
	if !t.IsK3kCluster() {
		return nil, errors.New("不是虚拟集群")
	}

	aud, err := t.GetAudience()
	if err != nil {
		return nil, err
	}

	return &K3kConfig{
		Name:      aud[3],
		Namespace: aud[4],
		ApiServer: aud[5],
	}, nil
}
func (t *K8sToken) GetRole() string {
	aud, err := t.GetAudience()
	if err != nil {
		return "normal"
	}
	if len(aud) > 1 {
		return aud[1]
	}
	return "normal"
}

func (t *K8sToken) IsFounder() bool {
	return t.GetRole() == "founder"
}

func (t *K8sToken) GetNamespace() string {
	aud, err := t.GetAudience()
	if err != nil {
		return ""
	}
	return aud[4]
}

func getTokenSaName(token string) (string, *jwtv5.NumericDate) {
	data := jwtv5.MapClaims{}
	jwtToken, _, err := jwtv5.NewParser().ParseUnverified(token, data)
	if err != nil {
		return "", nil
	}

	expireData, _ := jwtToken.Claims.GetExpirationTime()

	saName, ok := data["kubernetes.io/serviceaccount/service-account.name"]
	if ok {
		//获取过期时间

		return saName.(string), expireData
	}

	kubernetesIO, ok := data["kubernetes.io"].(map[string]interface{})
	if !ok {
		return "", expireData
	}

	serviceaccount, ok := kubernetesIO["serviceaccount"].(map[string]interface{})
	if !ok {
		return "", expireData
	}

	serviceaccountName, ok := serviceaccount["name"].(string)
	if !ok {
		return "", expireData
	}
	return serviceaccountName, expireData
}
