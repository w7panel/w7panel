package webhook

import (
	"log/slog"
	"strings"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	k8schain "github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "k8s.io/api/core/v1"
)

type image struct {
	Name            string `json:"name"`
	Image           string `json:"image"`
	AnnotationImage string `json:"annotationImage"`
	NewSha256       string `json:"newSha256"`
	Changed         bool   `json:"changed"`
}

func (i *image) ImageIsSha256() bool {
	return strings.Contains(i.Image, "@sha256:")
}
func (i *image) CurrentImageSha256() string {
	if i.ImageIsSha256() {
		return strings.Split(i.Image, "@")[1]
	}
	return ""
}
func (i *image) GetAnnotationImage() string {
	return i.AnnotationImage
}

func (i *image) AnnotationImageIsSha256() bool {
	if i.HasAnnotationImage() {
		return strings.Contains(i.AnnotationImage, "@sha256:")
	}
	return false
}

func (i *image) HasAnnotationImage() bool {
	// 检查Annotations字典中是否存在键为"origin-image-"+i.Name的项
	return i.AnnotationImage != ""
}

func (i *image) GetNativeImage() string {
	if i.ImageIsSha256() {
		return i.GetAnnotationImage()
	}
	return i.Image
}

func (i *image) LoadNewSha256Image(namespace string, imagePullSecrets []string, sdk *k8s.Sdk) {
	if i.ImageIsSha256() && i.AnnotationImageIsSha256() {
		return
	}
	kc, err := k8schain.New(sdk.Ctx, sdk.ClientSet, k8schain.Options{Namespace: namespace, ImagePullSecrets: imagePullSecrets, ServiceAccountName: "no service account"})
	if err != nil {
		slog.Error("parse ref err ", "failed to parse image reference "+i.Image, err)
	}
	option := crane.WithAuthFromKeychain(kc)
	options := []crane.Option{
		option,
	}
	if !i.ImageIsSha256() {
		ref, err := name.ParseReference(i.Image)
		if err != nil {
			slog.Error("parse ref err ", "failed to parse image reference "+i.Image, err)
			// return err
			return
		}
		sha256, err := helper.ImageDigest(i.Image, options...)
		if err != nil {
			slog.Error("parse ref err ", "failed to parse image reference "+i.Image, err)
			// return err
			return
		}
		sha256Image := ref.Context().Digest(sha256).String()
		i.AnnotationImage = i.Image
		i.Image = sha256Image
		i.Changed = true
		return
	}

	if i.HasAnnotationImage() && i.ImageIsSha256() {
		ref, err := name.ParseReference(i.AnnotationImage)
		if err != nil {
			slog.Error("parse ref err ", "failed to parse image reference "+i.AnnotationImage, err)
			// return err
			return
		}
		sha256, err := helper.ImageDigest(i.AnnotationImage, options...)
		if err != nil {
			slog.Error("parse ref err ", "failed to parse image reference "+i.AnnotationImage, err)
			// return err
			return
		}
		currentSha256 := i.CurrentImageSha256()
		if currentSha256 != sha256 {
			i.Image = ref.Context().Digest(sha256).String()
			i.Changed = true
			return
		}
	}
}

func getImagePullSecrets(secrets []v1.LocalObjectReference) []string {
	// secrets := deployment.Spec.Template.Spec.ImagePullSecrets
	var imagePullSecrets []string
	for i := range secrets {
		imagePullSecrets = append(imagePullSecrets, secrets[i].Name)
	}
	return imagePullSecrets
}

func ReSetDeploymentImage(namespace, name string) error {
	sdk := k8s.NewK8sClient()
	deployment, err := sdk.GetDeployment(namespace, name)
	if err != nil {
		return err
	}
	anno := deployment.Annotations
	changed := false
	containers := deployment.Spec.Template.Spec.Containers
	secrets := getImagePullSecrets(deployment.Spec.Template.Spec.ImagePullSecrets)

	for i := range containers {
		if containers[i].ImagePullPolicy != "Always" {
			continue
		}
		oldImage := containers[i].Image
		img := &image{
			Name:            containers[i].Name,
			Image:           oldImage,
			AnnotationImage: anno["origin-image-"+containers[i].Name],
		}
		img.LoadNewSha256Image(deployment.Namespace, secrets, sdk.Sdk)
		if img.Changed {
			changed = true
			containers[i].Image = img.Image
			anno["origin-image-"+containers[i].Name] = img.GetAnnotationImage()
		}
	}
	if changed {
		_, err = sdk.UpdateDeployment(namespace, deployment)
		if err != nil {
			return err
		}
	}
	return nil
	// deployment.Spec.Template.Spec.ImagePullSecrets
}

func ReSetDaemonSetImage(namespace, name string) error {
	sdk := k8s.NewK8sClient()
	ds, err := sdk.GetDaemonset(namespace, name)
	if err != nil {
		return err
	}
	anno := ds.Annotations
	changed := false
	containers := ds.Spec.Template.Spec.Containers
	secrets := getImagePullSecrets(ds.Spec.Template.Spec.ImagePullSecrets)
	for i := range containers {
		if containers[i].ImagePullPolicy != "Always" {
			continue
		}
		oldImage := containers[i].Image
		img := &image{
			Name:            containers[i].Name,
			Image:           oldImage,
			AnnotationImage: anno["origin-image-"+containers[i].Name],
		}
		img.LoadNewSha256Image(ds.Namespace, secrets, sdk.Sdk)
		if img.Changed {
			changed = true
			containers[i].Image = img.Image
			anno["origin-image-"+containers[i].Name] = img.GetAnnotationImage()
		}
	}
	if changed {
		_, err = sdk.UpdateDaemonset(namespace, ds)
		if err != nil {
			return err
		}
	}
	return nil
	// deployment.Spec.Template.Spec.ImagePullSecrets
}

func ReSetStatefulSetImage(namespace, name string) error {
	sdk := k8s.NewK8sClient()
	ds, err := sdk.GetStatefulSet(namespace, name)
	if err != nil {
		return err
	}
	anno := ds.Annotations
	changed := false
	containers := ds.Spec.Template.Spec.Containers
	secrets := getImagePullSecrets(ds.Spec.Template.Spec.ImagePullSecrets)
	for i := range containers {
		if containers[i].ImagePullPolicy != "Always" {
			continue
		}
		oldImage := containers[i].Image
		img := &image{
			Name:            containers[i].Name,
			Image:           oldImage,
			AnnotationImage: anno["origin-image-"+containers[i].Name],
		}
		img.LoadNewSha256Image(ds.Namespace, secrets, sdk.Sdk)
		if img.Changed {
			changed = true
			containers[i].Image = img.Image
			anno["origin-image-"+containers[i].Name] = img.GetAnnotationImage()
		}
	}
	if changed {
		_, err = sdk.UpdateStatefulSet(namespace, ds)
		if err != nil {
			return err
		}
	}
	return nil
	// deployment.Spec.Template.Spec.ImagePullSecrets
}

func ResetImage(namespace, name, workloadtype string, anno map[string]string) error {
	slog.Info("reset image", "workload type", workloadtype)
	// sha256Image, ok := os.LookupEnv("SHA256_IMAGE")
	// if !ok || sha256Image == "false" {
	// 	return nil
	// }
	if anno["w7.cc/image-to-sha256"] != "true" {
		return nil
	}
	time.AfterFunc(time.Second*5, func() {
		ResetImageNow(namespace, name, workloadtype, anno)
	})
	return nil

}

func ResetImageNow(namespace, name, workloadtype string, anno map[string]string) error {
	switch workloadtype {
	case "Deployment":
		ReSetDeploymentImage(namespace, name)
		break
	case "DaemonSet":
		ReSetDaemonSetImage(namespace, name)
		break
	case "StatefulSet":
		ReSetStatefulSetImage(namespace, name)
		break
	default:
		slog.Info("reset image err", "workload type", workloadtype, "name", name, "namespace", namespace)

	}
	return nil
}
