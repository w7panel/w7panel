package k3k

import (
	"context"
	"errors"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	console2 "gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/overselling"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"github.com/rancher/k3k/pkg/apis/k3k.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func LoginByServiceAccount(client *k8s.Sdk, sa *v1.ServiceAccount, seconds int64, updateK3kUser bool) (string, bool, error) {
	k3kUser := types.NewK3kUser(sa)
	isK3kUser := false
	if k3kUser.IsClusterUser() {
		isK3kUser = true
		if sa.Annotations[types.K3K_CLUSTER_POLICY_VERSION] == "" {
			sa.Annotations[types.K3K_CLUSTER_POLICY_VERSION] = "1"
		}
		policyName, ok := sa.Annotations[types.K3K_CLUSTER_POLICY]
		if ok {
			sa.Annotations[types.K3K_CLUSTER_POLICY_VERSION] = types.GetPolicyVersion(policyName)
		}
	}
	// if refreshCdToken {
	err := console.RefreshCDTokenUseOpenid(sa.Name)
	if err != nil {
		slog.Warn("刷新CDToken失败", "err", err)
	}
	// }

	_, err = RefreshK3kUser(k3kUser, client, updateK3kUser)
	if err != nil {
		return "", false, err
	}

	token, err := client.CreateTokenRequest(sa.Name, seconds, k3kUser.GetTokenAud())
	if err != nil {
		return "", false, err
	}
	// 标记最后登录时间 为了触发面板代理重建
	if k3kUser.IsFounder() {
		go console.RegisterLicenseSite(k3kUser.Name)
		go func() {
			// 刷新license
			err = console2.VerifyDefaultLicense(true)
			if err != nil {
				slog.Error("刷新license失败", "err", err)
			}
		}()
	}

	if k3kUser.CanOverSellingCheck() {
		go TryCheckOverSellingResource(client, k3kUser)
	} else {
		go SignLastLoginTime(client, k3kUser)
	}

	return token, isK3kUser, nil
}

func SignLastLoginTime(sdk *k8s.Sdk, user *types.K3kUser) error {
	client, err := sdk.ToSigClient()
	if err != nil {
		return err
	}
	//设置最后登录时间
	_, err = controllerutil.CreateOrPatch(context.TODO(), client, user.ServiceAccount, func() error {
		user.SetLoginTime()
		return nil
	})
	if err != nil {
		return err
	}
	return err
}

func TokenToK3kUser(token string) (*types.K3kUser, error) {
	rootSdk := k8s.NewK8sClient().Sdk
	ktoken := k8s.NewK8sToken(token)
	saName, err := ktoken.GetSaName()
	if err != nil {
		return nil, err
	}
	sa, err := rootSdk.GetServiceAccount(rootSdk.GetNamespace(), saName)
	if err != nil {
		return nil, err
	}
	user := types.NewK3kUser(sa)
	return RefreshK3kUser(user, rootSdk, false)
}

// 登录时候刷新用户权限
func RefreshK3kUser(user *types.K3kUser, rootSdk *k8s.Sdk, update bool) (*types.K3kUser, error) {
	// user := types.NewK3kUser(sa)
	// oldSa := user.ServiceAccount.DeepCopy()
	w7configRepo := config.NewW7ConfigRepository(rootSdk)
	if !user.IsCustomPermission() {
		menuConfig, err := rootSdk.ClientSet.CoreV1().ConfigMaps(user.GetNamespace()).Get(rootSdk.Ctx, user.GetMenuName(), metav1.GetOptions{})
		if err != nil {
			slog.Error("GetMenuConfig error", "error", err)
		}
		if err == nil {
			user.ReplaceMenu(menuConfig)
		}
	}
	// if !user.IsCustomQuota() {
	// 	quotaConfig, err := rootSdk.ClientSet.CoreV1().ConfigMaps(user.GetNamespace()).Get(rootSdk.Ctx, user.GetQuotaName(), metav1.GetOptions{})
	// 	if err != nil {
	// 		slog.Error("GetQuotaConfig error", "error", err)
	// 	}
	// 	if err == nil {
	// 		user.ReplaceQuota(quotaConfig)
	// 	}
	// }
	if !user.IsCustomCost() {
		costConfig, err := rootSdk.ClientSet.CoreV1().ConfigMaps(user.GetNamespace()).Get(rootSdk.Ctx, user.GetCostName(), metav1.GetOptions{})
		if err != nil {
			slog.Error("GetCostConfig error", "error", err)
		}
		if err == nil {
			err := user.ReplaceCost(costConfig)
			if err != nil {
				slog.Error("ReplaceCost error", "error", err)
				return nil, err
			}
		}
	}
	w7config, err := w7configRepo.Get(user.Name)
	if err != nil {
		slog.Error("GetW7Config error", "error", err)
	}
	if w7config != nil {
		user.ReplaceW7Config(w7config)
	}
	// user.SetOverMode(true)
	if update {
		_, err := rootSdk.ClientSet.CoreV1().ServiceAccounts(user.GetNamespace()).Update(rootSdk.Ctx, user.ServiceAccount, metav1.UpdateOptions{})
		if err != nil {
			slog.Error("user update error", "error", err)
			return nil, err
		}
	}
	return user, nil
}

func NeedRelogin(token *k8s.K8sToken) bool {
	saName, err := token.GetSaName()
	if err != nil {
		return false
	}
	// if token.GetLockVersion() != GetSaVersion(saName) || token.GetK3kPolicyVersion() != GetPolicyVersion(token.GetPolicyName()) {
	if token.GetLockVersion() != types.GetSaVersion(saName) {
		return true
	}
	return false
}

func getServiceAccountResource(sa *corev1.ServiceAccount) *overselling.Resource {
	user := k3ktypes.NewK3kUser(sa)
	lqr := user.GetLimitRange()
	if lqr != nil {
		return lqr.GetHardResource()
	}
	return overselling.EmptyResource()
}

func TryCheckOverSellingResource(sdk *k8s.Sdk, k3kuser *types.K3kUser) error {
	if !k3kuser.CanOverSellingCheck() {
		// 不需要资源检测
		return nil
	}
	sigClient, err := sdk.ToSigClient()
	if err != nil {
		return err
	}
	//资源验证有问题
	err = overselling.CanAddResource(k3kuser.GetOverResource(), getServiceAccountResource)
	if err != nil {
		return err
	}
	_, err = controllerutil.CreateOrPatch(context.TODO(), sigClient, k3kuser.ServiceAccount, func() error {
		k3kuser.SetLoginTime()
		k3kuser.SetOverMode(true)
		return nil
	})

	// TODO: 尝试添加资源
	return err
}

func InitCluster(sdk *k8s.Sdk, user *types.K3kUser) error {

	sigClient, err := sdk.ToSigClient()
	if err != nil {
		return err
	}
	k3kClient := k3ktypes.NewK3kClient(sigClient)

	if !user.IsClusterUser() {
		return errors.New("非集群用户,无法操作")
	}
	if user.NeedCreateOrder() {
		return errors.New("未购买集群资源,无法操作")
	}
	if user.IsExpired() {
		return errors.New("用户已过期,请续费")
	}
	if user.GetStorageClass() == "" {
		return errors.New("未设置存储类")
	}
	if user.GetNamespace() == "" {
		return errors.New("未设置命名空间")
	}
	if user.NeedOverSellingCheck() {
		return errors.New("集群资源已售罄,请联系管理员")
	}
	err = k3kClient.Create(user)
	return err
}

func RefreshK3kPolicy(policy *v1alpha1.VirtualClusterPolicy, rootSdk *k8s.Sdk, update bool) error {
	if policy.Annotations == nil {
		return nil
	}
	costName, ok := policy.Annotations["w7.cc/cost-name"]
	if ok && costName != "" {
		costConfig, err := rootSdk.ClientSet.CoreV1().ConfigMaps(policy.Namespace).Get(rootSdk.Ctx, costName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		json, err := types.ConfigMapToCostString(costConfig)
		if err != nil {
			return err
		}
		policy.Annotations["w7.cc/cost"] = json
	}
	return nil
}

func WhMoshiToggle(sdk *k8s.Sdk, k3kuser *types.K3kUser) error {

	sigClient, err := sdk.ToSigClient()
	if err != nil {
		return err
	}

	_, err = controllerutil.CreateOrPatch(context.TODO(), sigClient, k3kuser.ServiceAccount, func() error {
		// k3kuser.SetLoginTime()
		k3kuser.SetWeihu(!k3kuser.IsWeihu())
		return nil
	})

	// TODO: 尝试添加资源
	return err
}
