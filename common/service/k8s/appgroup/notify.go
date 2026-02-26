package appgroup

import (
	"errors"
	"net/url"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	v1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
)

func NeedNotifyInstalled(group *v1alpha1.AppGroup) bool {
	if group.Annotations == nil {
		return false
	}
	if val, ok := group.Annotations["w7.cc/ticket"]; ok && val == "" {
		return false
	}
	_, ok := group.Annotations["w7.cc/notify-installed"]
	return !ok
}

func NotifyInstalled(group *v1alpha1.AppGroup) error {
	if group.Annotations == nil {
		return nil
	}
	if val, ok := group.Annotations["w7.cc/ticket"]; ok && val == "" {
		return nil
	}
	_, ok := group.Annotations["w7.cc/notify-installed"]
	if ok {
		return nil
	}
	group.Annotations["w7.cc/notify-installed"] = "true"

	zpkUrl := group.Spec.ZpkUrl
	uri, err := url.Parse(zpkUrl)
	if err != nil {
		return err
	}
	uri.Path = "/zpk/respo/install/complete-notify"
	err = notify(uri.String(), group.Annotations["w7.cc/ticket"])
	if err != nil {
		return err
	}
	return err
}

func NotifyDeleted(group *v1alpha1.AppGroup) error {
	if group.Annotations == nil {
		return nil
	}
	if val, ok := group.Annotations["w7.cc/ticket"]; ok && val == "" {
		return nil
	}
	zpkUrl := group.Spec.ZpkUrl
	uri, err := url.Parse(zpkUrl)
	if err != nil {
		return err
	}
	uri.Path = "/zpk/respo/uninstall/complete-notify"
	err = notify(uri.String(), group.Annotations["w7.cc/ticket"])
	if err != nil {

		return err
	}
	return err
}

func notify(url string, ticket string) error {

	data := map[string]string{
		"ticket": ticket,
	}
	res, err := helper.RetryHttpClient().R().SetFormData(data).Post(url)
	if err != nil {
		return err
	}
	if res.StatusCode() != 200 {
		return errors.New("appgroup notify error")
	}
	return err
}
