package gpustack

import (
	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/**
backendUrl: >-
               /k8s/v1/namespaces/default/services/gpustack-backend-vlbxxuha/proxy-no
           group: gpustack-backend-vlbxxuha
           image: >-
               swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/gpustack/gpustack:v0.5.1
           password: mcibezmroj
           pvcName: gpustack-backend-vlbxxuha
           serverUrl: http://gpustack-backend-vlbxxuha.default.svc
           storageClass: disk1
           username: admin
           workerToken: gpustack-backend-vlbxxuha

*/

type WorkerConfig struct {
	Image   string `form:"image" binding:"required"`
	PvcName string `form:"pvcName" binding:"required"`
	// StorageClass string `form:"storageClass" binding:"required"`
	ServerUrl        string `form:"serverUrl" binding:"required"`
	Group            string `form:"group" binding:"required"`
	Password         string `form:"password"`
	WorkerToken      string `form:"workerToken" binding:"required"`
	Namespace        string `form:"namespace" binding:"required"`
	GpuCores         string `form:"gpucores" binding:"required"`
	Gpu              string `form:"gpu" binding:"required"`
	GpuMem           string `form:"gpumem" binding:"required"`
	Cpu              string `form:"cpu"`
	Memory           string `form:"memory"`
	RuntimeClassName string `form:"runtimeClassName"`
}

type GpuStackWorker struct {
	sdk *k8s.Sdk
}

func (w *GpuStackWorker) podContainer(config *WorkerConfig) corev1.Container {
	if config.Cpu != "" {
		config.Cpu = "0"
	}
	if config.Memory != "" {
		config.Memory = "0"
	}
	return corev1.Container{
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "gpustack-backend",
				MountPath: "/var/lib/gpustack/cache",
				SubPath:   "gpustack/cache",
			},
		},
		Resources: corev1.ResourceRequirements{
			Requests: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    resource.MustParse(config.Cpu),
				corev1.ResourceMemory: resource.MustParse(config.Memory),
				"nvidia.com/gpu":      resource.MustParse(config.Gpu),
				"nvidia.com/gpucores": resource.MustParse(config.GpuCores),
				"nvidia.com/gpumem":   resource.MustParse(config.GpuMem),
			},
			Limits: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    resource.MustParse(config.Cpu),
				corev1.ResourceMemory: resource.MustParse(config.Memory),
				"nvidia.com/gpu":      resource.MustParse(config.Gpu),
				"nvidia.com/gpucores": resource.MustParse(config.GpuCores),
				"nvidia.com/gpumem":   resource.MustParse(config.GpuMem),
			},
		},
		Name:  "gpustack-backend",
		Image: config.Image,
		Command: []string{
			"/bin/sh",
		},
		Args: []string{
			"-c",
			"gpustack start --server-url $SERVER_URL --token $TOKEN --worker-ip $POD_IP --enable-ray",
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: 8080,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "RELEASE_NAME_SUFFIX",
				Value: config.Group,
			},
			{
				Name:  "SERVER_URL",
				Value: config.ServerUrl,
			},
			{
				Name:  "TOKEN",
				Value: config.WorkerToken,
			},
			{
				Name:  "HF_ENDPOINT",
				Value: "https://hf-mirror.com",
			},

			// {
			// 	Name:  "BOOTSTRAP_PASSWORD",
			// 	Value: config.Password,
			// },
			{
				Name: "POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				},
			},
		},
		ImagePullPolicy: corev1.PullIfNotPresent,
	}
}

func NewGpuStackWorker(sdk *k8s.Sdk) *GpuStackWorker {
	return &GpuStackWorker{
		sdk: sdk,
	}
}
func (w *GpuStackWorker) CreateWorker(config *WorkerConfig) (*appsv1.StatefulSet, error) {
	statefulSet := w.CreateStatefulset(config)
	return w.sdk.ClientSet.AppsV1().StatefulSets(config.Namespace).Create(w.sdk.Ctx, statefulSet, metav1.CreateOptions{})
}

func (w *GpuStackWorker) CreateStatefulset(config *WorkerConfig) *appsv1.StatefulSet {
	// name := config.Group + "-" + helper.RandomString(10)
	name := "gpustack-worker" + "-" + helper.RandomString(10)
	runtimeClassName := "nvidia"
	if config.RuntimeClassName != "" {
		runtimeClassName = config.RuntimeClassName
	}
	labels := map[string]string{
		"app":                     config.Group,
		"w7.cc/group-name":        config.Group,
		"w7.cc/groupstack-worker": "true",
	}
	matchlabels := map[string]string{
		"app":                          config.Group,
		"w7.cc/group-name":             config.Group,
		"w7.cc/groupstack-worker":      "true",
		"w7.cc/groupstack-worker-name": name,
	}
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: config.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchlabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchlabels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "gpustack-backend",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: config.PvcName,
								},
							},
						},
					},
					RuntimeClassName: &runtimeClassName,
					Containers:       []corev1.Container{w.podContainer(config)},
				},
			},
		},
	}
	return statefulset
}
