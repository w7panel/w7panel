package types

import (
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/patrickmn/go-cache"
	"github.com/w7corp/sdk-open-cloud-go/service"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

const K3K_MENU_FOUNDER_NAME = "k3k.permission.founder"
const W7_CREATE_POD = "w7.cc/create-pod"

var K3kInitStatus string

const W7_USER_MODE_FOUNDER = "founder"
const W7_USER_MODE_CLUSTER = "cluster"
const W7_USER_MODE_NORMAL = "normal"

const W7_ORDER_PAID = "paid"
const W7_ORDER_RETURN = "return"
const W7_ORDER_WAIT = "wait"

const W7_BUY_MODE_PAID = "paid"         //付费模式
const W7_BUY_MODE_GIVE = "give"         //赠送模式
const W7_BUY_MODE_WAIT_PAY = "wait-pay" //待支付模式

// 初始化顺序
const K3K_STATUS_UNKNOW = "unknow"
const K3K_STATUS_RUNNING = "running"
const K3K_STATUS_COMPLETE = "complete"
const K3K_STATUS_FAILED = "failed"

const K3K_STATUS_USER_NEW = "new"           //无资源
const K3K_STATUS_USER_CREATING = "creating" //创建中
const K3K_STATUS_USER_READY = "ready"       //有资源
const K3K_STATUS_USER_WAIT = "wait"         //待回收
const K3K_STATUS_USER_RECYCLE = "recycle"   //回收中

// 待回收时间注解键
const K3K_PENDING_RECYCLE_TIME = "w7.cc/pending-recycle-time"
const K3K_EXPIRE_TIME = "w7.cc/expiretime"

const K3kFinalizerName = "k3k.sa/finalizer"

const (
	K3K_CLUSTER_MODE_VIRTUAL = "virtual"
	K3K_CLUSTER_MODE_SHARED  = "shared"
	K3K_USER_MODE            = "w7.cc/user-mode" //用户模式
	W7_WH_MODE               = "w7.cc/weihu"     //维护模式

	K3K_NAME       = "w7.cc/k3k-name"       //集群名称
	K3K_NAMESPACE  = "w7.cc/k3k-namespace"  //命名空间
	K3K_JOB_NAME   = "w7.cc/k3k-job-name"   //执行安装的job名称
	K3K_JOB_STATUS = "w7.cc/k3k-job-status" //执行安装的job状态
	// K3K_STORAGE_CLASS          = "k3k.io/storageclass"
	// K3K_STORAGE_REQUEST_SIZE   = "k3k.io/storage-request-size"
	K3K_CLUSTER_MODE           = "k3k.io/cluster-mode"
	K3K_CLUSTER_STATUS         = "k3k.io/cluster-status"
	K3K_CLUSTER_POLICY         = "k3k.io/policy"
	K3K_DEBUG                  = "w7.cc/debug" //调试模式
	K3K_LOCK_VERSION           = "w7.cc/version"
	W7_CONSOLE_ID              = "w7.cc/console-id" //控制台id
	K3K_CLUSTER_POLICY_VERSION = "w7.cc/policy-version"
	W7_PAUSE                   = "w7.cc/pause"               //回收中设置暂停
	W7_MENU                    = "w7.cc/menu"                //菜单
	W7_ROLE                    = "w7.cc/role"                //角色
	W7_MENU_NAME               = "w7.cc/menu-name"           //菜单menu名称
	W7_QUOTA_LIMIT             = "w7.cc/quota-limit"         //配额限制
	W7_QUOTA_LIMIT_LOCK        = "w7.cc/quota-limit-lock"    //配额锁定
	W7_QUOTA_LIMIT_NAME        = "w7.cc/quota-limit-name"    //配额限制
	W7_BANDWIDTH               = "w7.cc/bandwidth"           //带宽限制
	W7_WEB_SHELL               = "w7.cc/web-shell"           //web终端
	W7_FILE_EDITTOR            = "w7.cc/file-editor"         //文件编辑器
	W7_ACCESS_TOKEN            = "w7.cc/access-token"        //访问令牌
	W7_DOMAIN_WHITE_LIST       = "w7.cc/domain-white-list"   //域名白名单
	W7_DEMO_USER               = "w7.cc/demo-user"           //演示用户
	W7_SYS_STORAGE_PVC_NAME    = "w7.cc/sys-pvc-name"        //系统存储PVC名称
	W7_RETURN_ORDER_INFO       = "w7.cc/return-order-info"   //需要退款处理的订单信息
	W7_BASE_ORDER_SN           = "w7.cc/base-order-sn"       //基础订单号
	W7_BASE_ORDER_PASS         = "w7.cc/base-order-pass"     //跳过基础订单购买 true or false
	W7_RENEW_ORDER_SN          = "w7.cc/renew-order-sn"      //续费订单号
	W7_EXPAND_ORDER_SN         = "w7.cc/expand-order-sn"     //扩容订单号
	W7_BUY_MODE                = "w7.cc/buy-mode"            //付费模式
	W7_BASE_ORDER_STATUS       = "w7.cc/base-order-status"   //基础订单状态 paid return 支付or 退回
	W7_RENEW_ORDER_STATUS      = "w7.cc/renew-order-status"  //续费订单状态 paid return 支付or 退回
	W7_EXPAND_ORDER_STATUS     = "w7.cc/expand-order-status" //扩容订单状态 paid return 支付or 退回
	W7_COST                    = "w7.cc/cost"                //费用
	// W7_COST_PACKAGE            = "w7.cc/cost-package"        //套餐
	W7_COST_NAME          = "w7.cc/cost-name"          //费用套餐
	W7_USER_MODE          = "w7.cc/user-mode"          //用户模式 cluster founder normal
	W7_OVER_MODE          = "w7.cc/over-mode"          //资源状态 wait(等待检测) no-resource(无资源) success(检测通过)
	W7_OVER_RESOURCE      = "w7.cc/over-resource"      //资源状态 wait(等待检测) no-resource(无资源) success(检测通过)
	W7_OVER_BASE_RESOURCE = "w7.cc/over-base-resource" //首次购买资源
	W7_LOGIN_TIME         = "w7.cc/login-time"
)

const Bandwidth v1.ResourceName = "bandwidth"
const DataStorageSize v1.ResourceName = "data.storage"     // 数据存储大小
const SysStorageSize v1.ResourceName = "sys.storage"       // 系统存储大小
const ExpandStorageSize v1.ResourceName = "expand.storage" // 系统存储大小

type ConsoleOAuthAccessToken struct {
	*service.ResultAccessToken
}

func NewConsoleOAuthAccessToken(jsonStr string) (*ConsoleOAuthAccessToken, error) {
	var result ConsoleOAuthAccessToken
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func NewConsoleOAuthAccessToken2(rt *service.ResultAccessToken) *ConsoleOAuthAccessToken {
	return &ConsoleOAuthAccessToken{ResultAccessToken: rt}
}
func (c *ConsoleOAuthAccessToken) IsExpired() bool {
	return c.ExpireTime < time.Now().Second()
}
func (c *ConsoleOAuthAccessToken) ToString() string {
	result, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(result)
}

var saVersions = make(map[string]string)
var policyVersions = make(map[string]string)

func DelSaVersion(name string) {
	delete(saVersions, name)
}
func SetSaVersion(name string, version string) {
	if version == "" {
		version = "1"
	}
	saVersions[name] = version
}

func GetSaVersion(name string) string {
	if v, ok := saVersions[name]; ok {
		return v
	}
	return "1"
}
func SetPolicyVersion(name string, version string) {
	if version == "" {
		version = "1"
	}
	policyVersions[name] = version
}
func GetPolicyVersion(name string) string {
	if v, ok := policyVersions[name]; ok {
		return v
	}
	return "1"
}
func DelPolicyVersion(name string) {
	delete(policyVersions, name)
}

func NeedRelogin(token *k8s.K8sToken) bool {
	saName, err := token.GetSaName()
	if err != nil {
		return false
	}
	// if token.GetLockVersion() != GetSaVersion(saName) || token.GetK3kPolicyVersion() != GetPolicyVersion(token.GetPolicyName()) {
	if token.GetLockVersion() != GetSaVersion(saName) {
		return true
	}
	return false
}

var k3kcache *cache.Cache

var K3kregCnf = `
mirrors:
  registry.local.w7.cc:
  docker.io:
    endpoint:
      - "https://mirror.ccs.tencentyun.com"
      - "https://registry.cn-hangzhou.aliyuncs.com"
      - "https://docker.m.daocloud.io"
      - "https://docker.1panel.live"
  quay.io:
    endpoint:
      - "https://quay.m.daocloud.io"
      - "https://quay.dockerproxy.com"
  gcr.io:
    endpoint:
      - "https://gcr.m.daocloud.io"
      - "https://gcr.dockerproxy.com"
  ghcr.io:
    endpoint:
      - "https://ghcr.m.daocloud.io"
      - "https://ghcr.dockerproxy.com"
  k8s.gcr.io:
    endpoint:
      - "https://k8s-gcr.m.daocloud.io"
      - "https://k8s.dockerproxy.com"
  registry.k8s.io:
    endpoint:
      - "https://k8s.m.daocloud.io"
      - "https://k8s.dockerproxy.com"
  mcr.microsoft.com:
    endpoint:
      - "https://mcr.m.daocloud.io"
      - "https://mcr.dockerproxy.com"
  nvcr.io:
    endpoint:
      - "https://nvcr.m.daocloud.io"
  "*":

`

type VirtualClusterPolicy struct {
	metav1.ObjectMeta `json:"metadata"`
	metav1.TypeMeta   `json:",inline"`
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
