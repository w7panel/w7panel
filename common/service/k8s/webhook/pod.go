package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	k3kTypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/shell"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (m *ResourceMutator) handlePod(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 Pod admission 请求")

	// 解码请求中的 Pod 资源
	pod := &corev1.Pod{}
	if err := (m.decoder).Decode(req, pod); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if pod.Annotations != nil {
		val, ok := pod.Annotations[k3ktypes.W7_CREATE_POD]
		if ok && val == "false" {
			return admission.Denied("不允许创建pod")
		}
	}
	shell.WebHookPid(pod.DeepCopy())
	// // 检查 Pod 是否有 ownerReferences.kind=Cluster
	// _, isClusterNormalPod := pod.Labels["k3k.io/clusterName"]
	// if isClusterNormalPod {
	// 	return m.handleNormalPod(ctx, pod, req)
	// }
	// 纯普通pod
	modified := false
	namespace := pod.Namespace
	if strings.HasPrefix(namespace, "k3k-") && !helper.IsChildAgent() {
		modified = handlePodLimit(m.client, m.sdk, pod, namespace)
	}
	if helper.IsLxcfsEnabled() {
		//https://github.com/ymping/lxcfs-admission-webhook/blob/main/cmd/volume.go
		pod.Spec.Volumes = append(pod.Spec.Volumes, volumesTemplate...)
		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts, volumeMountsTemplate...)
		}
		modified = true
	}
	if !modified {
		return admission.Allowed("Pod cpu memory 无需配置")
	}
	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// 返回修改后的资源
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func handlePodLimit(client client.Client, sdk *k8s.Sdk, pod *corev1.Pod, namespace string) bool {
	modified := false
	if strings.HasPrefix(namespace, "k3k-") {
		k3kName := strings.TrimPrefix(namespace, "k3k-")
		sa, err := getSa(client, sdk, k3kName)
		if err != nil {
			slog.Error("未找到sa")
			return false
		}
		k3kUser := k3kTypes.NewK3kUser(sa)
		if !k3kUser.IsClusterUser() {
			slog.Info("不是集群用户")
			return false
		}
		rang := k3kUser.GetLimitRange()
		if rang == nil {
			slog.Info("未配置limitRange")
			return false
		}
		if k3kUser.IsShared() {
			cpu := rang.Limit.Cpu()
			memory := rang.Limit.Memory()
			if rang.Hard.Cpu().IsZero() && rang.Hard.Memory().IsZero() { // 如果是不限制资源
				cpu := rang.Limit.Cpu()
				memory := rang.Limit.Memory()
				if !cpu.IsZero() && !memory.IsZero() {
					modified = setRequestLimit(pod, *cpu, *memory)
				}
			} else {
				if cpu.IsZero() {
					cpu1 := resource.MustParse("250m")
					cpu = &cpu1
				}
				if memory.IsZero() {
					memory1 := resource.MustParse("500Mi")
					memory = &memory1
				}
				modified = setRequestLimit(pod, *cpu, *memory)
			}
		}

		quantity := k3kUser.GetBandWidth()
		if !quantity.IsZero() {
			if pod.Annotations == nil {
				pod.Annotations = make(map[string]string)
			}
			quantitystr := quantity.String()
			slog.Info("Pod 带宽限制", slog.String("bandwidth", quantitystr))
			// quantitystr = strings.ReplaceAll(quantitystr, "Mi", "Mbps")

			pod.Annotations["kubernetes.io/egress-bandwidth"] = quantitystr
			pod.Annotations["kubernetes.io/ingress-bandwidth"] = quantitystr
			modified = true
		}
	}
	return modified
}

var volumeMountsTemplate = []corev1.VolumeMount{

	{
		Name:      "lxcfs-proc-cpuinfo",
		MountPath: "/proc/cpuinfo",
		ReadOnly:  true,
	},
	{
		Name:      "lxcfs-proc-meminfo",
		MountPath: "/proc/meminfo",
		ReadOnly:  true,
	},
	{
		Name:      "lxcfs-proc-diskstats",
		MountPath: "/proc/diskstats",
		ReadOnly:  true,
	},
	{
		Name:      "lxcfs-proc-stat",
		MountPath: "/proc/stat",
		ReadOnly:  true,
	},
	{
		Name:      "lxcfs-proc-swaps",
		MountPath: "/proc/swaps",
		ReadOnly:  true,
	},
	{
		Name:      "lxcfs-proc-uptime",
		MountPath: "/proc/uptime",
		ReadOnly:  true,
	},
	{
		Name:      "lxcfs-proc-loadavg",
		MountPath: "/proc/loadavg",
		ReadOnly:  true,
	},
	{
		Name:      "lxcfs-sys-devices-system-cpu-online",
		MountPath: "/sys/devices/system/cpu/online",
		ReadOnly:  true,
	},
}
var hostfiletype corev1.HostPathType

func init() {
	hostfiletype = corev1.HostPathFile
}

var volumesTemplate = []corev1.Volume{
	{
		Name: "lxcfs-proc-cpuinfo",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/cpuinfo",
				Type: &hostfiletype,
			},
		},
	},
	{
		Name: "lxcfs-proc-diskstats",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/diskstats",
				Type: &hostfiletype,
			},
		},
	},
	{
		Name: "lxcfs-proc-meminfo",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/meminfo",
				Type: &hostfiletype,
			},
		},
	},
	{
		Name: "lxcfs-proc-stat",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/stat",
				Type: &hostfiletype,
			},
		},
	},
	{
		Name: "lxcfs-proc-swaps",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/swaps",
				Type: &hostfiletype,
			},
		},
	},
	{
		Name: "lxcfs-proc-uptime",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/uptime",
				Type: &hostfiletype,
			},
		},
	},
	{
		Name: "lxcfs-proc-loadavg",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/loadavg",
				Type: &hostfiletype,
			},
		},
	},
	{
		Name: "lxcfs-sys-devices-system-cpu-online",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/sys/devices/system/cpu/online",
				Type: &hostfiletype,
			},
		},
	},
}
