package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/rancher/k3k/k3k-kubelet/translate"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (m *ResourceMutator) handleDeployment(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 Deployment admission 请求")

	// 解码请求中的 Deployment 资源
	deployment := &appsv1.Deployment{}
	if err := (m.decoder).Decode(req, deployment); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	ResetImage(deployment.Namespace, deployment.Name, "Deployment", deployment.Annotations)
	clusterName, ok := deployment.Labels["cluster"]
	if !ok {
		return admission.Allowed("无需修改 deployment")
	}

	// 检查是否需要修改
	modified := false

	trans := translate.ToHostTranslator{
		ClusterName:      clusterName,
		ClusterNamespace: req.Namespace,
	}
	registriesConfigMapName := trans.TranslateName("default", "registries")

	// 检查是否已经挂载了 registries.yaml
	registriesMounted := false

	// 检查所有容器的挂载
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			if volumeMount.MountPath == "/etc/rancher/k3s/registries.yaml" {
				registriesMounted = true
				break
			}
		}
		if registriesMounted {
			break
		}
	}

	// 如果没有挂载 registries.yaml，则添加相应的配置
	if !registriesMounted {
		slog.Info("Deployment 没有挂载 registries.yaml，添加挂载配置")

		// 检查是否已经存在 registries-volume
		volumeExists := false
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.Name == "registries-volume" {
				volumeExists = true
				break
			}
		}

		// 如果不存在，添加 volume
		if !volumeExists {
			registriesVolume := v1.Volume{
				Name: "registries-volume",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: registriesConfigMapName,
						},
						Items: []v1.KeyToPath{
							{
								Key:  "default.cnf",
								Path: "registries.yaml",
							},
						},
					},
				},
			}
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, registriesVolume)
			modified = true
		}

		// 为每个容器添加 volumeMount
		for i := range deployment.Spec.Template.Spec.Containers {
			container := &deployment.Spec.Template.Spec.Containers[i]

			// 检查容器是否已经有了这个挂载
			mountExists := false
			for _, mount := range container.VolumeMounts {
				if mount.MountPath == "/etc/rancher/k3s/registries.yaml" {
					mountExists = true
					break
				}
			}

			// 如果不存在，添加挂载
			if !mountExists {
				registriesMount := v1.VolumeMount{
					Name:      "registries-volume",
					MountPath: "/etc/rancher/k3s/registries.yaml",
					SubPath:   "registries.yaml",
					ReadOnly:  true,
				}
				container.VolumeMounts = append(container.VolumeMounts, registriesMount)
				modified = true
			}
		}
	}

	// 检查 pullPolicy 是否为 Always，并修改镜像地址为 sha256 格式
	// for i := range deployment.Spec.Template.Spec.Containers {
	// 	container := &deployment.Spec.Template.Spec.Containers[i]
	// 	if container.ImagePullPolicy == v1.PullAlways {
	// 		// 记录原始镜像地址和标签到 Annotations
	// 		if deployment.Annotations == nil {
	// 			deployment.Annotations = make(map[string]string)
	// 		}
	// 		deployment.Annotations["original-image-"+container.Name] = container.Image

	// 		// 修改镜像地址为 sha256 格式
	// 		sha256Image, err := getImageSHA256(container.Image)
	// 		if err != nil {
	// 			slog.Error("获取镜像 sha256 失败", slog.String("error", err.Error()))
	// 			continue
	// 		}
	// 		container.Image = sha256Image
	// 		modified = true
	// 	}
	// }

	// 如果没有修改，直接返回允许
	if !modified {
		return admission.Allowed("Deployment 已经挂载了 registries.yaml 或不需要修改")
	}

	// 序列化修改后的 Deployment
	marshaledDeployment, err := json.Marshal(deployment)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// 返回修改后的资源
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledDeployment)
}
