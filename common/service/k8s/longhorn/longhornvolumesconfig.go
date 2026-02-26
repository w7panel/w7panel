package longhorn

import (
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ConfigName = "longhorn-volumes-config"

func initLonghornVolumesConfig(sdk *k8s.Sdk) error {
	configmap, err := sdk.ClientSet.CoreV1().ConfigMaps("default").Get(sdk.Ctx, ConfigName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			configmap = &v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: ConfigName,
				},
				Data: map[string]string{
					"customs": defaultVolumeName,
					"default": defaultVolumeName,
				},
			}
			_, err = sdk.ClientSet.CoreV1().ConfigMaps("default").Create(sdk.Ctx, configmap, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		}
		return err
	}
	cliststr, ok := configmap.Data["customs"]
	if !ok {
		return errors.NewBadRequest("customs not found")
	}
	customs := strings.Split(cliststr, ",")
	//check has defaultVolumeName
	for _, v := range customs {
		if strings.TrimSpace(v) == defaultVolumeName {
			return nil
		}
	}
	customs = append(customs, defaultVolumeName)
	configmap.Data["customs"] = strings.Join(customs, ",")
	_, err = sdk.ClientSet.CoreV1().ConfigMaps("default").Update(sdk.Ctx, configmap, metav1.UpdateOptions{})
	return err

}
