package controller

import (
	"context"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type Upgrade struct {
	controller.Abstract
}

func (self Upgrade) UpgradeImage(http *gin.Context) {
	type ParamsValidate struct {
		Kind          string `form:"kind" validate:"required"`
		ApiVersion    string `form:"apiVersion" validate:"required"`
		Name          string `form:"name" validate:"required"`
		Namespace     string `form:"namespace" `
		ContainerName string `form:"containerName" validate:"required"`
		Image         string `form:"image" validate:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	sdk := k8s.NewK8sClient().Sdk
	clientset := sdk.ClientSet

	switch params.Kind {
	case "Deployment":
		err := self.updateDeploymentImage(params.Namespace, params.Name, params.ContainerName, params.Image, clientset)
		if err != nil {
			http.JSON(500, gin.H{"error": err.Error()})
			return
		}
	case "DaemonSet":
		err := self.updateDaemonSetImage(params.Namespace, params.Name, params.ContainerName, params.Image, clientset)
		if err != nil {
			http.JSON(500, gin.H{"error": err.Error()})
			return
		}
	case "StatefulSet":
		err := self.updateStatefulSetImage(params.Namespace, params.Name, params.ContainerName, params.Image, clientset)
		if err != nil {
			http.JSON(500, gin.H{"error": err.Error()})
			return
		}
	default:
		http.JSON(400, gin.H{"error": "Unsupported kind: " + params.Kind})
		return
	}

	http.JSON(200, gin.H{"message": "Image updated successfully"})
}

func (self Upgrade) updateDeploymentImage(namespace, name, containerName, image string, clientset *kubernetes.Clientset) error {

	patch := []byte(`{"spec":{"template":{"spec":{"containers":[{"name":"` + containerName + `","image":"` + image + `"}]}}}}`)
	_, err := clientset.AppsV1().Deployments(namespace).Patch(
		context.TODO(),
		name,
		types.StrategicMergePatchType,
		patch,
		metav1.PatchOptions{},
	)
	return err
}

func (self Upgrade) updateDaemonSetImage(namespace, name, containerName, image string, clientset *kubernetes.Clientset) error {

	patch := []byte(`{"spec":{"template":{"spec":{"containers":[{"name":"` + containerName + `","image":"` + image + `"}]}}}}`)
	_, err := clientset.AppsV1().DaemonSets(namespace).Patch(
		context.TODO(),
		name,
		types.StrategicMergePatchType,
		patch,
		metav1.PatchOptions{},
	)
	return err
}

func (self Upgrade) updateStatefulSetImage(namespace, name, containerName, image string, clientset *kubernetes.Clientset) error {

	patch := []byte(`{"spec":{"template":{"spec":{"containers":[{"name":"` + containerName + `","image":"` + image + `"}]}}}}`)
	_, err := clientset.AppsV1().StatefulSets(namespace).Patch(
		context.TODO(),
		name,
		types.StrategicMergePatchType,
		patch,
		metav1.PatchOptions{},
	)
	return err
}
