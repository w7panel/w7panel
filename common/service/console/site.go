package console

import (
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

// 如果是创始人 登录时候 自动注册license site
func RegisterLicenseSite(saName string) error {
	licenseClient, err := NewDefaultLicenseClient()
	if err != nil {
		slog.Error("创建license client失败", "err", err)
		return err
	}
	_, err = licenseClient.CreateLicenseSite(saName, false)
	if err != nil {
		slog.Error("创建license site失败", "err", err)
		return err
	}
	return nil
	// err = VerifyLicense(license)
	// if err != nil {
	// 	slog.Error("验证license失败", "err", err)
	// 	return err
	// }
	// return nil

}
func RegisterSite(token, releaseName, host string) (appSecret *AppSecret, err error) {
	cdClient := NewConsoleCdClient(token)
	return cdClient.CreateSite(host, releaseName)
}

func PatchAppId(client *k8s.Sdk, appSecret *AppSecret, deploymentName string, namespace string) (err error) {
	patchData := `{
		"spec": {
			"template": {
				"spec": {
					"containers": [
						{
							"name": "%s",
							"env": [
								{
									"name": "APP_ID",
									"value": "%s"
								},
								{
									"name": "APP_SECRET",
									"value": "%s"
								}
							]
						}
					]
				}
			}
		}
	}`
	patchData = fmt.Sprintf(patchData, deploymentName, appSecret.AppId, appSecret.AppSecret)
	//deployment 修改env
	//patch deployment

	_, err = client.ClientSet.
		AppsV1().
		Deployments(namespace).
		Patch(client.Ctx, deploymentName, k8stypes.StrategicMergePatchType, []byte(patchData), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}
