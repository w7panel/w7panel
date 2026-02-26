package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/shopspring/decimal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type K3kUser struct {
	*k3kUser
}

var once sync.Once

type k3kUser struct {
	*v1.ServiceAccount
	lqr  *LimitRangeQuota
	cost *K3kCost
	// *k3kUserBase
	// *k3kUserOverSelling
	*k3kUserOrder
}

func NewK3kUser(sa *v1.ServiceAccount) *K3kUser {
	jstr, ok := sa.Annotations[W7_QUOTA_LIMIT]
	lqr := &LimitRangeQuota{}
	if ok {
		lqr2, err := NewLimitRangeQuata(jstr)
		if err != nil {
			slog.Error("parse limit range quota error", "error", err)
		} else {
			lqr = lqr2
		}
	}

	u := &K3kUser{k3kUser: &k3kUser{
		ServiceAccount: sa,
		lqr:            lqr,
	}}
	costStr, ok1 := sa.Annotations[W7_COST]

	if ok1 {
		cost2, err := CreateCostFromString(costStr)
		if err != nil {
			slog.Error("parse cost error", "error", err)
		}
		u.cost = cost2
	}
	k3kUserBase := Newk3kUserBase(u.ServiceAccount)
	k3kUserTime := Newk3kUserTime(u.ServiceAccount)
	k3kUserOverSelling := Newk3kUserOverSelling(u.ServiceAccount, k3kUserBase)
	k3kUserOrder := Newk3kUserOrder(u.ServiceAccount, k3kUserOverSelling, k3kUserTime, u.cost)
	u.k3kUserOrder = k3kUserOrder
	return u

}
func (u *k3kUser) IsClusterReady() bool {
	return u.Annotations[K3K_JOB_STATUS] == K3K_STATUS_COMPLETE
}

func (u *k3kUser) IsClusterNew() bool {
	return u.Labels[K3K_CLUSTER_STATUS] == K3K_STATUS_USER_NEW
}

func (u *k3kUser) IsClusterCreating() bool {
	return u.Labels[K3K_CLUSTER_STATUS] == K3K_STATUS_USER_CREATING
}
func (u *k3kUser) IsClusterRecycle() bool {
	return u.Labels[K3K_CLUSTER_STATUS] == K3K_STATUS_USER_RECYCLE
}

func (u *k3kUser) IsClusterLabelReady() bool {
	return u.Labels[K3K_CLUSTER_STATUS] == K3K_STATUS_USER_READY
}
func (u *k3kUser) IsK3kUser() bool {
	return u.Labels[K3K_USER_MODE] == "cluster"
}

func (u *k3kUser) IsClusterUser() bool {
	return u.Labels[K3K_USER_MODE] == "cluster"
}

func (u *k3kUser) IsNormalUser() bool {
	return u.Labels[K3K_USER_MODE] == "normal"
}

func (u *k3kUser) IsInitK3k() bool {
	name, ok := u.Labels[K3K_NAME]
	if ok {
		return name != ""
	}
	return false
}
func (u *k3kUser) GetK3kName() string {
	name, ok := u.Labels[K3K_NAME]
	if ok {
		return name
	}
	return u.Name
}

func (u *k3kUser) GetUserMode() string {
	name, ok := u.Labels[K3K_USER_MODE]
	if ok {
		return name
	}
	return ""
}
func (u *k3kUser) GetClusterMode() string {
	name, ok := u.Annotations[K3K_CLUSTER_MODE]
	if ok {
		return name
	}
	return "unknown"
}

func (u *k3kUser) GetClusterPolicy() string {
	name, ok := u.Annotations[K3K_CLUSTER_POLICY]
	if ok {
		return name
	}
	return ""
}

func (u *k3kUser) GetK3kNamespace() string {
	name, ok := u.Labels[K3K_NAMESPACE]
	if ok {
		return name
	}
	return "k3k-" + u.GetName()
}

func (u *k3kUser) GetK3kJobName() string {
	name, ok := u.Annotations[K3K_JOB_NAME]
	if ok {
		return name
	}
	return ""
}

func (u *k3kUser) GetK3kJobStatus() string {
	name, ok := u.Annotations[K3K_JOB_STATUS]
	if ok {
		return name
	}
	return K3K_STATUS_UNKNOW
}

func (u *k3kUser) GetStorageClass() string {
	if u.lqr != nil {
		return u.lqr.StorageClass
	}
	return ""
}

func (u *k3kUser) GetStorageRequestSize() string {
	if u.lqr != nil {
		result := u.lqr.GetHardRequestStorage()
		return result.String()
	}
	return "5Gi"
}

func (u *k3kUser) GetClusterSysStorageRequestSize() string {

	defaultSize := u.GetStorageRequestSize()
	// return defaultSize
	if u.GetClusterMode() == "virtual" {
		return defaultSize
	}
	return "5Gi" //shared 模式默认给5Gi
	// if u.lqr != nil {
	// 	result, ok := u.lqr.Hard[SysStorageSize]
	// 	if ok {
	// 		return result.String()
	// 	}
	// }
	// return "1Gi"
}

func (u *k3kUser) GetClusterDataStorageRequestSize() string {

	defaultSize := u.GetStorageRequestSize()
	quantity, err := resource.ParseQuantity(defaultSize)
	if err != nil {
		slog.Error("parse quantity error", "error", err)
	}
	shareSize := resource.MustParse("5Gi")
	if (quantity.Cmp(shareSize)) > 0 {
		quantity.Sub(shareSize)
		return quantity.String()
	}
	//
	return "5Gi"
}

func (u *k3kUser) GetAgentName() string {
	return helper.GetK3kAgentName(u.GetK3kName())
}

func (u *k3kUser) GetVirtualIngressServiceName() string {
	return u.GetK3kNamespace() + "-service-w7"
}

func (u *k3kUser) GetApiServerHost() string {
	return helper.GetApiServerHost(u.GetK3kNamespace())
}

func (u *k3kUser) GetKubeconfigMapName() string {
	return "k3k-kubeconfig-" + u.GetK3kName()
}

func (u *k3kUser) GetDefaultVolumeName() string {
	// if u.IsVirtual() {
	// 	return "default-volume"
	// }
	// return u.GetClusterServer0PvcName()
	return "default-volume"
}

// 是否维护模式
func (u *k3kUser) IsWeihu() bool {
	return u.Labels[W7_WH_MODE] == "true"
}

func (u *k3kUser) SetWeihu(ok bool) {
	val := "false"
	if ok {
		val = "true"
	}
	u.Labels[W7_WH_MODE] = val
}

func (u *k3kUser) GetMenu() string {

	if u.IsWeihu() { //维护模式菜单
		whMenu := []string{"cluster", "cluster-panel", "cluster-resource", "app", "app-apps", "app-apps-delete"}
		json, _ := json.Marshal(whMenu)
		return string(json)
	}
	name, ok := u.Annotations[W7_MENU]
	if ok {
		if console.IsFree() {
			name = strings.ReplaceAll(name, "system-manage", "system-manage-free") //替换为不存在的多租户管理菜单
		}
		return name
	}
	return ""
}

func (u *k3kUser) GetMenuName() string {
	name, ok := u.Annotations[W7_MENU_NAME]
	if ok {
		return name
	}
	return ""
}

func (u *k3kUser) GetQuotaName() string {
	name, ok := u.Annotations[W7_QUOTA_LIMIT_NAME]
	if ok {
		return name
	}
	return ""
}

func (u *k3kUser) GetCostName() string {
	name, ok := u.Annotations[W7_COST_NAME]
	if ok {
		return name
	}
	return ""
}

func (u *k3kUser) GetDebugMode() string {
	if !u.IsClusterUser() {
		return "true"
	}
	name, ok := u.Annotations[K3K_DEBUG]
	if ok {
		return name
	}
	return "false"
}

func (u *k3kUser) CanInitCluster() bool {
	if !u.IsOverSellingSuccess() {
		return false
	}
	if u.IsClusterUser() {
		return u.IsClusterCreating() || u.IsClusterNew()
	}
	return false
}
func (u *k3kUser) ToArray() map[string]string {
	needCreateOrder := "true"
	if !u.NeedCreateOrder() {
		needCreateOrder = "false"
	}
	canReNew := "false"
	if err := u.CanRenewError(); err == nil {
		canReNew = "true"
	}
	needRenew := "false"
	if u.NeedRenew() {
		needRenew = "true"
	}
	canExpand := "false"
	if err := u.CanExpandError(); err == nil {
		canExpand = "true"
	}
	expiretime, err := u.GetExpireTime()
	expiretimeStr := "" //expiretime.Format("2006-01-02 15:04:05")
	if err == nil {
		expiretimeStr = expiretime.Format("2006-01-02 15:04:05")
	}
	hasPassword := false
	passval, passok := u.Annotations["password"]
	if passok && passval != "" {
		hasPassword = true
	}
	// price, err := u.GetBasePrice()
	// if err != nil {
	// 	price = decimal.Zero
	// }
	uprice, err := u.GetUnitPrice()
	if err != nil {
		uprice = decimal.Zero
	}
	// if u.cost != nil && u.cost.IsGive() {
	// 	price = decimal.Zero
	// }
	result := map[string]string{
		K3K_USER_MODE: u.GetUserMode(),
		K3K_NAME:      u.GetK3kName(),
		K3K_NAMESPACE: u.GetK3kNamespace(),
		// K3K_STORAGE_CLASS:        u.GetStorageClass(),
		// K3K_STORAGE_REQUEST_SIZE: u.GetStorageRequestSize(),
		K3K_JOB_NAME:              u.GetK3kJobName(),
		K3K_JOB_STATUS:            u.GetK3kJobStatus(),
		K3K_CLUSTER_MODE:          u.GetClusterMode(),
		K3K_CLUSTER_POLICY:        u.GetClusterPolicy(),
		"w7.cc/username":          u.Name,
		K3K_DEBUG:                 u.GetDebugMode(),
		W7_MENU:                   u.GetMenu(),
		W7_QUOTA_LIMIT:            u.Annotations[W7_QUOTA_LIMIT],
		W7_FILE_EDITTOR:           u.Annotations[W7_FILE_EDITTOR],
		W7_WEB_SHELL:              u.Annotations[W7_WEB_SHELL],
		W7_DOMAIN_WHITE_LIST:      u.Annotations[W7_DOMAIN_WHITE_LIST], // 白名单域名
		W7_DEMO_USER:              u.Labels[W7_DEMO_USER],
		W7_SYS_STORAGE_PVC_NAME:   u.GetClusterServer0PvcName(), // 系统存储PVC名称
		W7_COST:                   u.Annotations[W7_COST],
		"w7.cc/can-init-cluster":  boolToString(u.CanInitCluster()),       //是否可以初始化集群
		"w7.cc/need-create-order": needCreateOrder,                        //需要创建初始订单
		"w7.cc/need-over-check":   boolToString(u.NeedOverSellingCheck()), //是否必须超额检查 首次购买
		"w7.cc/can-over-check":    boolToString(u.CanOverSellingCheck()),  //是否可以超额检查 扩容不强制检查
		"w7.cc/has-over-resource": boolToString(!u.CanOverSellingCheck()), //资源是否超额
		"w7.cc/has-password":      boolToString(hasPassword),              //是否设置了密码
		// "w7.cc/base-price-total":  price.String(),                         //初始订单价格
		"w7.cc/unit-price-total": uprice.String(),        //续费订单单价
		"w7.cc/can-renew":        canReNew,               //是否可以续费
		"w7.cc/need-renew":       needRenew,              //必须续费
		"w7.cc/can-expand":       canExpand,              //是否可以扩容
		"w7.cc/over-mode":        u.Labels[W7_OVER_MODE], //检测资源是否足够
		K3K_EXPIRE_TIME:          expiretimeStr,
		K3K_CLUSTER_STATUS:       u.Labels[K3K_CLUSTER_STATUS],
		// "w7.cc/diff-day":          strconv.FormatInt(int64(u.GetDiffDays()), //废弃
		"w7.cc/diff-month": u.GetDiffMonths().String(),
		W7_ROLE:            u.GetRole(),
		W7_WH_MODE:         u.Labels[W7_WH_MODE],
	}
	if !u.IsClusterUser() {
		// result[W7_FILE_EDITTOR] = "true"
		// result[W7_WEB_SHELL] = "true"
	}
	return result

}

func (u *k3kUser) GetRole() string {
	// if u.IsClusterUser() {
	// 	return "founder"
	// }
	role, ok := u.Annotations[W7_ROLE]
	if ok {
		return role
	}
	if u.Labels[W7_USER_MODE] == "founder" {
		return "founder"
	}
	if u.Labels[W7_USER_MODE] == "normal" {
		return "normal"
	}
	return "normal"
}

func (u *k3kUser) GetTokenAud() []string {
	if !u.IsK3kUser() {
		return []string{
			u.Name,
			u.GetRole(),
			u.Labels[W7_CONSOLE_ID],
			"https://kubernetes.default.svc.cluster.local",
			"k3s",
		}
	}
	// return []string{}
	return []string{
		u.Name,
		u.GetRole(),
		u.Labels[W7_CONSOLE_ID],
		u.GetK3kName(),
		u.GetK3kNamespace(),
		u.GetApiServerHost(),
		u.GetClusterMode(),
		u.GetClusterPolicy(),
		u.GetLockVersion(),
		u.GetClusterPolicyVersion(),
		"https://kubernetes.default.svc.cluster.local",
		"k3s",
	}
}

func (u *k3kUser) GetLockVersion() string {
	version, ok := u.Annotations[K3K_LOCK_VERSION]
	if ok {
		return version
	}
	return "1"
}

// 这是一个内存值，用于标记集群策略版本 auth.go 登录时候 实时查询
func (u *k3kUser) GetClusterPolicyVersion() string {
	version, ok := u.Annotations[K3K_CLUSTER_POLICY_VERSION]
	if ok {
		return version
	}
	return "1"
}

func (u *k3kUser) Running(jobName string) {
	u.Annotations[K3K_JOB_STATUS] = K3K_STATUS_RUNNING
	u.Annotations[K3K_JOB_NAME] = jobName
	u.Labels[K3K_CLUSTER_STATUS] = K3K_STATUS_USER_CREATING //创建中
}

func (u *k3kUser) ReNew() {
	u.Annotations[K3K_JOB_STATUS] = K3K_STATUS_UNKNOW
	u.Annotations[K3K_JOB_NAME] = ""
	u.DelPendingRecycleTime()
	delete(u.Annotations, K3K_EXPIRE_TIME)
	delete(u.Labels, W7_BASE_ORDER_SN)
	delete(u.Labels, W7_BASE_ORDER_STATUS)
}

func (u *k3kUser) ToK3kConfig() *k8s.K3kConfig {
	return &k8s.K3kConfig{
		Name:      u.GetK3kName(),
		Namespace: u.GetK3kNamespace(),
		ApiServer: u.GetApiServerHost(),
	}
}

// 获取资源回收阶段
func (u *k3kUser) GetResourceStatus() string {
	status, ok := u.Labels[K3K_CLUSTER_STATUS]
	if !ok {
		return K3K_STATUS_USER_NEW // 默认为有资源状态
	}
	return status
}

// 设置资源回收阶段
func (u *k3kUser) SetResourceStatus(status string) {
	if u.Labels == nil {
		u.Labels = make(map[string]string)
	}
	u.Labels[K3K_CLUSTER_STATUS] = status
}

// 检查是否过期

func (u *k3kUser) IsVirtual() bool {
	return u.GetClusterMode() == K3K_CLUSTER_MODE_VIRTUAL
}

func (u *k3kUser) IsShared() bool {
	return u.GetClusterMode() == K3K_CLUSTER_MODE_SHARED
}

func (u *k3kUser) GetClusterServer0PvcName() string {
	return "varlibrancherk3s-" + u.GetK3kNamespace() + "-server-0"
}

func (u *k3kUser) GetBandWidth() resource.Quantity {
	if u.lqr != nil {
		return u.lqr.GetBandWidth()
	}
	return resource.MustParse("0M")
}

func (u *k3kUser) GetLimitRange() *LimitRangeQuota {
	return u.lqr
}

func (u *k3kUser) GetConsoleId() string {
	return u.Labels["w7.cc/console-id"]
}

// 自定义权限菜单
func (u *k3kUser) IsCustomPermission() bool {
	return u.Annotations["w7.cc/menu-name"] == ""
}

// 自定义配额
func (u *k3kUser) IsCustomQuota() bool {
	return u.Annotations["w7.cc/quota-limit-name"] == ""
}

func (u *k3kUser) IsCustomCost() bool {
	return u.Annotations["w7.cc/cost-name"] == ""
}

func (u *k3kUser) ReplaceMenu(menu *v1.ConfigMap) {
	// u.Annotations[W7_MENU_NAME] = menu.Name
	u.Annotations[K3K_DEBUG] = menu.Data["debug"]
	u.Annotations[W7_MENU] = menu.Data["menu"]
	u.Annotations[W7_WEB_SHELL] = menu.Data["webshell"]
	u.Annotations[W7_FILE_EDITTOR] = menu.Data["fileeditor"]
	if menu.Labels[W7_ROLE] != "" {
		u.Labels[W7_ROLE] = menu.Labels[W7_ROLE]
	}
}

func (u *k3kUser) ReplaceW7Config(config *config.W7Config) {
	if config != nil && config.UserInfo != nil {
		u.Labels[W7_CONSOLE_ID] = strconv.Itoa(config.UserInfo.UserId)
		u.Annotations["w7.cc/console-nickname"] = config.UserInfo.Nickname
		// u.Annotations[W7_USER_MODE] = config.UserInfo.UserMode
		// return 0, fmt.Errorf("user cost is not empty")
	}
}

func (u *k3kUser) ReplaceQuota(config *v1.ConfigMap) error {
	if u.IsClusterReady() || u.IsClusterLabelReady() {
		return nil
	}
	if config.Annotations == nil {
		config.Annotations = make(map[string]string)
	}
	if u.Annotations[W7_QUOTA_LIMIT_LOCK] == "true" {
		return nil
	}
	u.Annotations[W7_QUOTA_LIMIT] = config.Data["quota"]
	// lqr := &LimitRangeQuota{}
	lqr2, err := NewLimitRangeQuata(config.Data["quota"])
	if err != nil {
		slog.Error("parse limit range error", "error", err)
		return err
	}
	u.lqr = lqr2
	return nil
	// u.Annotations[W7_QUATA_LIMIT_NAME] = config.Name
}

func (u *k3kUser) ReplaceCost(config *v1.ConfigMap) error {

	cost, err := ConfigMapToCost(config)
	if err != nil {
		slog.Error("parse cost error", "error", err)
		return err
	}
	u.ReplaceQuota(config)
	if config.Annotations == nil {
		config.Annotations = make(map[string]string)
	}
	jsonCost, err := cost.ToJsonString()
	if err != nil {
		slog.Error("parse cost error", "error", err)
		return err
	}
	u.Annotations[W7_COST] = (jsonCost)
	// u.Annotations[W7_COST_PACKAGE] = config.Data["packageConfig"]

	// cost2, err := CreateCostFromString(string(data))
	// if err != nil {
	// 	slog.Error("parse cost error", "error", err)
	// 	return err
	// }
	u.cost = cost
	u.k3kUserOrder.cost = cost
	return nil
	// u.Annotations[W7_QUATA_LIMIT_NAME] = config.Name
}

func (u *k3kUser) GetBasePrice() (decimal.Decimal, error) {

	unit, err := u.GetUnitPrice()
	if err != nil {
		return decimal.Decimal{}, err
	}
	days, err := u.GetBaseDay()
	if err != nil {
		return decimal.Decimal{}, err
	}
	return unit.Mul(decimal.NewFromFloat(days)), nil

}

func (u *k3kUser) GetUnitPrice() (decimal.Decimal, error) {
	compute, err := u.GetOrderCompute()
	if err != nil {
		return decimal.Decimal{}, err
	}
	return compute.GetUnitPrice(), nil
}

//	func (u *k3kUser) GetBaseHour() (int64, error) {
//		if u.lqr == nil {
//			return 0, fmt.Errorf("limit range not set")
//		}
//		return u.lqr.GetHour(), nil
//	}
//
// 首次购买默认赠送天数
func (u *k3kUser) GetBaseDay() (float64, error) {
	if u.lqr == nil {
		return 0, fmt.Errorf("limit range not set")
	}
	return u.lqr.GetDays(), nil
}

func (u *k3kUser) GetDefaultUnitQuantity() UnitQuantity {
	if u.lqr != nil {
		return u.lqr.GetDefaultUnitQuantity()
	}
	return UnitQuantity{Quantity: 0, Unit: "month"}
}

func (u *k3kUser) CanResizeSysStorage(unUsed resource.Quantity, resizeTo resource.Quantity) bool {
	if u.lqr != nil {
		return u.lqr.CanResizeSysStorage(unUsed, resizeTo)
	}
	return false
}

func (u *k3kUser) ResizeSysStorage(storageSize resource.Quantity) {
	if u.lqr != nil {
		u.lqr.ResizeSysStorage(storageSize)
		json := u.GetLimitRange().ToString()
		u.Annotations[W7_QUOTA_LIMIT] = json
		u.Annotations[W7_QUOTA_LIMIT_NAME] = ""
	}
}

func (u *k3kUser) GetCost() *K3kCost {
	return u.cost
}

// 是否需要购买基础资源，

func (u *k3kUser) IsFounder() bool {
	return u.Labels[W7_USER_MODE] == W7_USER_MODE_FOUNDER
}

func (u *k3kUser) IsNormal() bool {
	return u.Labels[W7_USER_MODE] == W7_USER_MODE_NORMAL
}

func (u *k3kUser) Pause() {
	u.Annotations[W7_PAUSE] = "true"
}

func (u *k3kUser) UnPause() {
	u.Annotations[W7_PAUSE] = "false"
}

func (u *k3kUser) IsPause() bool {
	return u.Annotations[W7_PAUSE] == "true"
}

// 是否必须超额检查，扩容不需要超额检查
func (u *k3kUser) SetOverMode(ok bool) error {
	if ok && u.Labels[W7_OVER_MODE] != "success" {
		isExpand := u.IsExpand()
		u.Labels[W7_OVER_MODE] = "success"
		if isExpand { // 如果是扩容
			// u.lqr.ResetHard(rlist)
			// bs := NewBaseResource(&K3kUser{u})
			// rlist := bs.GetExpand(u.GetOverResource())
			u.lqr.Expand(u.GetOverResource())
			// u.lqr.ResetHard(rlist)
			delete(u.Annotations, W7_OVER_RESOURCE) //重复执行 会导致扩容的资源累加多次
		} else {
			u.lqr.ResetHard(u.GetOverResource())
		}
		u.Annotations[W7_QUOTA_LIMIT] = u.GetLimitRange().ToString()
		u.Annotations[W7_QUOTA_LIMIT_NAME] = ""
	} else {
		u.Labels[W7_OVER_MODE] = "no-resource" //资源不足
		// delete(u.Annotations, W7_OVER_BASE_RESOURCE)
	}
	return nil
}

func (u *k3kUser) SetLoginTime() {
	u.Annotations[W7_LOGIN_TIME] = time.Now().Format(time.DateTime)
}

func (u *k3kUser) GetOrderCompute() (*K3kOrderCompute, error) {
	if u.cost == nil {
		return nil, errors.New("当前用户未配置费用套餐，无法生成计算器")
	}
	if u.lqr == nil {
		return nil, errors.New("未配置资源限额, 无法生成计算器")
	}
	return NewK3kOrderComputeWithCostLimitRange(u.lqr, u.cost), nil
}

// 锁定退款订单
func (u *k3kUser) LockReturnK3kOrder(order *console.K3kOrder) error {
	if u.lqr != nil {
		currentResource := u.lqr.GetHardBuyResource()
		orderRes := K3kOrderToBuyResource(order)
		currentTime, err := u.GetExpireTime()

		if order.BuyMode == "base" {
			currentResource = currentResource.Sub(orderRes)
			if err == nil {
				currentTime = time.Now()
			}
		}
		if order.BuyMode == "expand" {
			currentResource = currentResource.Sub(orderRes)
		}
		if order.BuyMode == "renew" {
			currentResource = ZeroBuyResource
			hour, err := decimal.NewFromString(order.Hour)
			if err == nil {
				sec := hour.Mul(decimal.NewFromInt(3600))
				if err == nil {
					durations := time.Second * time.Duration(-sec.IntPart())
					currentTime.Add(durations)
				}
			}
		}
		lockOrder := LockReturnK3kOrder{
			CurrentTime: currentTime.Unix(),
			BuyResource: currentResource,
			OrderSn:     order.OrderSn,
			BuyMode:     order.BuyMode,
		}
		data, err := json.Marshal(lockOrder)
		if err != nil {
			return err
		}
		u.Annotations[W7_RETURN_ORDER_INFO] = string(data)
	}
	return nil
}

func (u *k3kUser) GetLockReturnK3kOrder() (*LockReturnK3kOrder, error) {
	data, ok := u.Annotations[W7_RETURN_ORDER_INFO]
	if ok {
		lockOrder := &LockReturnK3kOrder{}
		err := json.Unmarshal([]byte(data), lockOrder)
		if err != nil {
			return nil, err
		}
		return lockOrder, nil
	}
	return nil, errors.New("not found")
}

func (u *k3kUser) ProcessReturnK3kOrder() error {
	rorder, err := u.GetLockReturnK3kOrder()
	if err != nil {
		return err
	}
	u.SetOverMode(true) //等待中的资源 让生效
	u.lqr.ResetHard(rorder.ToOverSellingResource())
	if rorder.CurrentTime > 0 {
		u.Annotations[K3K_EXPIRE_TIME] = time.Unix(rorder.CurrentTime, 9).Format("2006-01-02 15:04:05")
	}
	delete(u.Annotations, W7_RETURN_ORDER_INFO)
	return nil
}
