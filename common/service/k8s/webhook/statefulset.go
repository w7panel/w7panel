package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 StatefulSet 资源
func (m *ResourceMutator) handleStatefulSet(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 StatefulSet admission 请求")

	// 解码请求中的 StatefulSet 资源
	statefulset := &appsv1.StatefulSet{}
	if err := (m.decoder).Decode(req, statefulset); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	ResetImage(statefulset.Namespace, statefulset.Name, "StatefulSet", statefulset.Annotations)

	clusterName, ok := statefulset.Labels["cluster"]
	if !ok {
		return admission.Allowed("无需修改 statefulset")
	}
	if statefulset.Annotations == nil {
		statefulset.Annotations = map[string]string{}
	}
	canCreate, okCanCreate := statefulset.Annotations[k3ktypes.W7_CREATE_POD]
	// 检查是否需要修改
	modified := false

	// ConfigMap 名称
	// registriesConfigMapName := "registries"

	// trans := translate.ToHostTranslator{
	// 	ClusterName:      clusterName,
	// 	ClusterNamespace: req.Namespace,
	// }
	// registriesConfigMapName := trans.TranslateName("default", "registries")

	// 检查是否已经挂载了 registries.yaml
	registriesMounted := false

	// 检查所有容器的挂载
	// for _, container := range statefulset.Spec.Template.Spec.Containers {
	// 	for _, volumeMount := range container.VolumeMounts {
	// 		if volumeMount.MountPath == "/etc/rancher/k3s/registries.yaml" {
	// 			registriesMounted = true
	// 			break
	// 		}
	// 	}
	// 	if registriesMounted {
	// 		break
	// 	}
	// }

	// 如果没有挂载 registries.yaml，则添加相应的配置
	if !registriesMounted {
		slog.Info("StatefulSet 没有挂载 registries.yaml，添加挂载配置")

		// 检查是否已经存在 registries-volume
		volumeExists := false
		for _, volume := range statefulset.Spec.Template.Spec.Volumes {
			if volume.Name == "registries-volume" {
				volumeExists = true
				break
			}
		}
		dirType := v1.HostPathFile
		// 如果不存在，添加 volume
		if !volumeExists {
			registriesVolume := v1.Volume{
				Name: "registries-volume",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: "/etc/rancher/k3s/registries.yaml",
						Type: &dirType,
					},
				},
				// ConfigMap: &v1.ConfigMapVolumeSource{
				// 	LocalObjectReference: v1.LocalObjectReference{
				// 		Name: registriesConfigMapName,
				// 	},
				// 	Items: []v1.KeyToPath{
				// 		{
				// 			Key:  "default.cnf",
				// 			Path: "registries.yaml",
				// 		},
				// 	},
				// },
				// },
			}
			statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, registriesVolume)
			modified = true
		}

		// 为每个容器添加 volumeMount
		for i := range statefulset.Spec.Template.Spec.Containers {
			container := &statefulset.Spec.Template.Spec.Containers[i]

			// 检查容器是否已经有了这个挂载
			mountExists := false
			for _, mount := range container.VolumeMounts {
				if mount.Name == "registries-volume" {
					mountExists = true
					break
				}
			}

			// 如果不存在，添加挂载
			if !mountExists {
				registriesMount := v1.VolumeMount{
					Name:      "registries-volume",
					MountPath: "/etc/rancher/k3s/registries.yaml",
					// SubPath:   "registries.yaml",
					ReadOnly: true,
				}
				container.VolumeMounts = append(container.VolumeMounts, registriesMount)
				modified = true
			}
		}

	}

	rs, err := getResourceLimit(m.client, m.sdk, clusterName, statefulset.Labels["role"])
	if err != nil {
		slog.Error("not found resource limit")
	}
	for i := range statefulset.Spec.Template.Spec.Containers {
		container := &statefulset.Spec.Template.Spec.Containers[i]
		if rs != nil {
			container.Resources.Limits = rs
			container.Resources.Requests = v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("0"),
				v1.ResourceMemory: resource.MustParse("0"),
			}
		}
		cmds := container.Command
		if cmds != nil && len(cmds) == 3 {
			cmd3 := cmds[2]
			//wget -P /var/lib/rancher/k3s/server/manifests -O k3k-virtual.yaml http://w7panel.default.svc:8000/ui/yaml/k3k/virtual.yaml
			if !strings.HasPrefix(cmd3, "root_cgroup_raw") {
				append := `
			
root_cgroup_raw=$(cat /proc/1/cgroup)
root_cgroup_stripped="${root_cgroup_raw#0::}"
root_cgroup=$(dirname "$root_cgroup_stripped")
cgroup_root="cgroup-root=$root_cgroup"
mount --make-shared /

				`
				// 				append2 := `
				// wget -P /var/lib/rancher/k3s/server/manifests -O k3k-virtual.yaml http://w7panel.default.svc:8000/ui/yaml/k3k/virtual.yaml
				// `
				cmd3 = append + "" + cmd3
				container.Command = []string{
					cmds[0],
					cmds[1],
					cmd3,
				}
			}

		}
		container.Env = append(container.Env, v1.EnvVar{
			Name: "GOMAXPROCS",
			ValueFrom: &v1.EnvVarSource{
				ResourceFieldRef: &v1.ResourceFieldSelector{
					Divisor:  resource.MustParse("1"),
					Resource: "limits.cpu",
				},
			},
		})
		modified = true
	}
	if statefulset.Spec.Template.Annotations == nil {
		statefulset.Spec.Template.Annotations = map[string]string{}
	}
	if okCanCreate {
		statefulset.Spec.Template.Annotations[k3ktypes.W7_CREATE_POD] = canCreate
		if canCreate == "false" {
			statefulset.Spec.Replicas = ptr.To(int32(0))
		}
		modified = true
	}
	// statefulset.Spec.Template.Spec.HostPID = true
	// modified = true

	// 如果没有修改，直接返回允许
	if !modified {
		return admission.Allowed("StatefulSet 已经挂载了 registries.yaml 或不需要修改")
	}

	// 序列化修改后的 StatefulSet
	marshaledStatefulSet, err := json.Marshal(statefulset)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// 返回修改后的资源
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledStatefulSet)
}
