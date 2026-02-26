package sa

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"github.com/w7corp/sdk-open-cloud-go/service"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Register struct {
	sdk          *k8s.Sdk
	client       client.Client
	k3kClient    *types.K3kClient
	w7configRepo config.W7ConfigRepositoryInterface
}

func DoRegister(sdk *k8s.Sdk, accessToken *types.ConsoleOAuthAccessToken, userinfo *service.ResultUserinfo, policyName string) (*corev1.ServiceAccount, error) {
	client, err := sdk.ToSigClient()
	if err != nil {
		return nil, err
	}
	k3kClient := types.NewK3kClient(client)
	kconfig, err := k3kClient.GetK3kConfig()
	if err != nil {
		return nil, err
	}
	if policyName != "" {
		register := NewRegister(client, sdk)
		return register.RegisterUseConsole(accessToken, userinfo, kconfig, policyName)
	}
	if kconfig.AllowConsoleRegister && kconfig.DefaultPolicyName != "" {
		register := NewRegister(client, sdk)
		return register.RegisterUseConsole(accessToken, userinfo, kconfig, kconfig.DefaultPolicyName)
	} else {
		return nil, errors.New("不允许控制台注册")
	}
}

func DoRegisterByUid(sdk *k8s.Sdk, uid int) (*corev1.ServiceAccount, error) {
	client, err := sdk.ToSigClient()
	if err != nil {
		return nil, err
	}
	k3kClient := types.NewK3kClient(client)
	kconfig, err := k3kClient.GetK3kConfig()
	if err != nil {
		return nil, err
	}
	if kconfig.AllowConsoleRegister && kconfig.DefaultPolicyName != "" {
		register := NewRegister(client, sdk)
		return register.RegisterUid(uid, kconfig)
	} else {
		return nil, errors.New("不允许控制台注册")
	}
}

// DoRegisterLink 链接注册账号
func DoRegisterLink(sdk *k8s.Sdk, username, password, policyName string) (*corev1.ServiceAccount, error) {
	client, err := sdk.ToSigClient()
	if err != nil {
		return nil, err
	}
	sa, err := sdk.GetServiceAccount("default", username)
	if err != nil {
		//not found 忽略
		if !apierrors.IsNotFound(err) {
			return nil, errors.New("账号已存在..")
		}
	}
	if (err == nil) && sa != nil {
		return nil, errors.New("账号已存在.")
	}

	bpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	register := NewRegister(client, sdk)
	annotations := map[string]string{
		"password": string(bpassword),
	}
	return register.doRegister(policyName, username, "", annotations, true)
}

func NewRegister(client client.Client, sdk *k8s.Sdk) *Register {
	respo := config.NewW7ConfigRepository(sdk)
	return &Register{
		client:       client,
		sdk:          sdk,
		k3kClient:    types.NewK3kClient(client),
		w7configRepo: respo,
	}
}

func (register *Register) RegisterUseConsole(accessToken *types.ConsoleOAuthAccessToken, userinfo *service.ResultUserinfo, k3kConfig *types.K3kConfig, policyName string) (*corev1.ServiceAccount, error) {
	userId := strconv.Itoa(userinfo.UserId)
	anns := map[string]string{
		types.W7_ACCESS_TOKEN: accessToken.ToString(),
	}
	return register.doRegister(policyName, "console-"+userId, userId, anns, false)
}

func (register *Register) RegisterUid(uid int, k3kConfig *types.K3kConfig) (*corev1.ServiceAccount, error) {
	userId := strconv.Itoa(uid)
	return register.doRegister(k3kConfig.DefaultPolicyName, "console-"+userId, userId, nil, false)
}

func (register *Register) doRegister(policyName string, saName string, consoleId string, anns map[string]string, checkAllowRegister bool) (*corev1.ServiceAccount, error) {
	policy, err := register.k3kClient.GetPolicyByName(policyName)
	if err != nil {
		slog.Error("get policy error", "error", err)
		return nil, err
	}
	if checkAllowRegister {
		if policy.Labels["w7.cc/allow-register"] != "true" {
			return nil, errors.New("不允许注册")
		}
	}

	labels := map[string]string{
		"k3k.io/cluster-status": "new",
		"w7.cc/user-mode":       "cluster",
		"k3k.io/policy":         policyName,
		// "w7.cc/console-id":      consoleId,
		"w7.cc/demo-user": policy.Labels["w7.cc/demo-user"],
	}
	if consoleId != "0" && consoleId != "" {
		labels["w7.cc/console-id"] = consoleId
	}
	annotations := policy.Annotations
	annotations["k3k.io/policy"] = policyName
	annotations["k3k.io/policy-title"] = policy.Annotations["title"]
	annotations["k3k.io/cluster-mode"] = string(policy.Spec.AllowedMode)
	// annotations["w7.cc/quota-limit"] = policy.Annotations["w7.cc/quota-limit"]
	costName, ok := policy.Annotations["w7.cc/cost-name"]
	if ok {
		costConfig, err := register.sdk.ClientSet.CoreV1().ConfigMaps("default").Get(register.sdk.Ctx, costName, metav1.GetOptions{})
		if err != nil {
			// return nil, errors.New("获取成本配置失败")
			slog.Error("user register get cost config error", "error", err)
		}
		if err == nil {
			annotations["w7.cc/quota-limit"] = costConfig.Data["quota"]
		}
	}
	if anns != nil {
		for k, v := range anns {
			annotations[k] = v
		}
	}

	sa := &corev1.ServiceAccount{

		ObjectMeta: metav1.ObjectMeta{
			Name:        saName,
			Namespace:   "default",
			Labels:      labels,
			Annotations: annotations,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
	}
	// sa 已存在
	// sa, err = register.sdk.ClientSet.CoreV1().ServiceAccounts("default").Create(register.sdk.Ctx, sa, metav1.CreateOptions{})
	// if err != nil {
	// 	return nil, err
	// }
	_, err = controllerutil.CreateOrPatch(context.Background(), register.client, sa, func() error {
		return nil
	})
	return sa, err

}
