package k8s

import (
	"context"
	"log/slog"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const LOGO_NAME = "k3k.logo.config"

func CheckLogo() error {
	sdk := NewK8sClient()
	configMap, err := sdk.ClientSet.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), LOGO_NAME, metav1.GetOptions{})
	if err != nil {
		slog.Error("Failed to get logo config", "error", err)
		return err
	}
	return WriteLogo(configMap)
}

func WriteLogo(configMap *corev1.ConfigMap) error {
	if configMap.Namespace == "kube-system" && configMap.Name == LOGO_NAME {
		kodata, ok := os.LookupEnv("KO_DATA_PATH")
		if ok {
			logoData, ok := configMap.BinaryData["default-cnf"]
			if ok {
				// save configMap
				err := os.WriteFile(kodata+"/assets/logo.png", logoData, 0644)
				if err != nil {
					slog.Error("Failed to write logo", "error", err)
					return err
				}
			} // save configMap
		}
	}
	return nil
}
