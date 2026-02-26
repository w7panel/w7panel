package types

import (
	"os"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ToK3kJob(k3kUser *K3kUser) *batchv1.Job {

	// 生成随机的Job名称
	jobName := "k3k-create-" + strings.ToLower(helper.RandomString(10))
	ingressClass := "higress"
	mode := "shared"
	k3smode := "4"
	// 如果模式为虚拟，则更新IngressClass和k3s模式
	if mode == "virtual" {
		ingressClass = "traefik"
		k3smode = "5"
	}

	// 设置环境变量
	//ToK3kJob
	envs := []corev1.EnvVar{
		{
			Name:  "K3K_NAME",
			Value: k3kUser.GetK3kName(),
		},
		{
			Name:  "K3K_NAMESPACE",
			Value: k3kUser.GetK3kNamespace(),
		},
		{
			Name:  "KUBECONFIG_PATH", //k3k-k7-k7-kubeconfig.yaml
			Value: "/tmp/" + k3kUser.GetK3kNamespace() + "-" + k3kUser.GetK3kName() + "-kubeconfig.yaml",
		},
		{
			Name:  "KUBECONFIG_SERVER", // k3k-cccc-service.k3k-cccc
			Value: k3kUser.GetK3kNamespace() + "-service." + k3kUser.GetK3kNamespace() + "",
		},
		{
			Name:  "STORAGE_CLASS_NAME",
			Value: k3kUser.GetStorageClass(),
		},
		{
			Name:  "K3K_MODE",
			Value: k3kUser.GetClusterMode(),
		},
		{
			Name:  "INGRESS_CLASS",
			Value: ingressClass,
		},
		{
			Name:  "K3S_MODE",
			Value: k3smode,
		},
		{
			Name:  "K3K_POLICY",
			Value: k3kUser.GetClusterPolicy(),
		},
		{
			Name:  "K3K_STORAGE_REQUEST_SIZE",
			Value: k3kUser.GetClusterSysStorageRequestSize(),
		},
		{
			Name:  "K3K_PVC_STORAGE_REQUEST_SIZE",
			Value: k3kUser.GetClusterDataStorageRequestSize(),
		},
		{
			Name:  "DEFAULT_VOLUME_NAME",
			Value: k3kUser.GetDefaultVolumeName(),
		},

		// 设置Job的标题
	}
	// 设置Job的注解
	title := "初始化虚拟集群"
	annotations := map[string]string{
		"title":              title,
		"w7.cc/title":        title,
		"w7.cc/deploy-title": "初始化虚拟集群",

		// 设置Job的标签
	}
	labels := map[string]string{
		"k3k-job":      "true",
		"k3k-sa":       k3kUser.Name,
		"job-name":     jobName,
		"w7.cc/suffix": k3kUser.Name,
	}
	// labels["w7.cc/job-source"] = "appgroup"
	// labels["searchJob"] = p.GetName() + "-build-" + shellType

	// 设置Job的重试次数和超时时间
	// labels["w7.cc/shell-type"] = shellType
	backofflimit := int32(1)

	// 创建Job对象
	afterSeconds := int32(600)
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      labels,
			Annotations: annotations,
			Namespace:   "default",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &afterSeconds,
			BackoffLimit:            &backofflimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: helper.ServiceAccountName(),
					//挂载hostPath
					RestartPolicy: corev1.RestartPolicyNever,

					Containers: []corev1.Container{
						{
							Name:            "create-cluster",
							Image:           helper.SelfImage(),
							Env:             envs,
							WorkingDir:      "/tmp",
							ImagePullPolicy: corev1.PullAlways,
							Command:         []string{"sh", "-c", "${KO_DATA_PATH}/shell/k3k-create.sh"},

							// SecurityContext: &corev1.SecurityContext{,
							// Args:            []string{shellkubectl.kubernetes.io/restartedAt
						},
					},
				},
			},
		},
	}
	return job
}

func ToK3kDaemonSet(k3kUser *K3kUser) *appsv1.DaemonSet {
	labels := map[string]string{
		"k3k-agent-pod": "true",
		"k3k-sa":        k3kUser.Name,
		"k3k-name":      k3kUser.GetK3kName(),
		"k3k-namespace": k3kUser.GetK3kNamespace(),
	}
	pod := ToK3kPod(k3kUser)
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   k3kUser.GetAgentName(),
			Labels: labels,
			Annotations: map[string]string{
				"helm-version":     os.Getenv("HELM_VERSION"),
				"root-pod-ip":      os.Getenv("POD_IP"),
				"title":            "面板代理",
				"w7.cc/create-svc": "true",
				"w7.cc.app/ports":  "{\"8000\":8000}",
			},
			Namespace: "default",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: pod.Spec,
			},
		},
	}
	return ds
}

func ToK3kPod(k3kUser *K3kUser) *corev1.Pod {
	// shell := "cd /tmp && k3kcli kubeconfig generate --namespace $K3K_NAMESPACE --name $K3K_NAME && k8s-offline"
	clusterMode := k3kUser.GetClusterMode()
	// k3smode := "4"
	// if mode == "virtual" {
	// 	k3smode = "5"
	// }
	// fileMode := int32(0600)
	envs := []corev1.EnvVar{
		{
			Name:  "K3K_NAME",
			Value: k3kUser.GetK3kName(),
		},
		{
			Name:  "K3K_NAMESPACE",
			Value: k3kUser.GetK3kNamespace(),
		},
		// {
		// 	Name:  "KUBECONFIG", //k3k-k7-k7-kubeconfig.yaml
		// 	Value: "/w7-config/kubeconfig.yaml",
		// },
		{
			Name:  "K3K_CLUSTER_HOST",
			Value: k3kUser.GetK3kNamespace() + "-service." + k3kUser.GetK3kNamespace() + "",
		},
		{
			Name:  "K3K_MODE",
			Value: k3kUser.GetClusterMode(),
		},
		{
			Name:  "SERVICE_ACCOUNT_NAME",
			Value: k3kUser.GetK3kName(),
		},
		{
			Name:  "STORAGE_CLASS_NAME",
			Value: k3kUser.GetStorageClass(),
		},
		{
			Name:  "WEBHOOK_ENABLED",
			Value: "true",
		},
		{
			Name:  "APP_WATCH",
			Value: "true",
		},
		{
			Name:  "MICROAPP_PATH",
			Value: "/data/microapp",
		},
		{
			Name:  "HELM_VERSION",
			Value: os.Getenv("HELM_VERSION"),
		},
		{
			Name:  "CLUSTER_MODE",
			Value: clusterMode,
		},
		{
			Name:  "IS_CHILD",
			Value: "true",
		},
		{
			Name:  "ROOT_SVCNAME",
			Value: os.Getenv("RELEASE_NAME_SUFFIX"),
		},
		{
			Name:  "ROOT_NAMESPACE",
			Value: os.Getenv("HELM_NAMESPACE"),
		},
		{
			Name:  "SVC_NAME",
			Value: k3kUser.GetAgentName(),
		},
		{
			Name:  "SVC_LB_CLASS",
			Value: os.Getenv("SVC_LB_CLASS"),
		},
		{
			Name:  "K8S_WATCH",
			Value: "true",
		},
		{
			Name:  "HIGRESS_WATCH",
			Value: "true",
		},
		{
			Name:  "SITE_ENABLED",
			Value: "true",
		},
		{
			Name:  "ROOT_POD_IP",
			Value: os.Getenv("POD_IP"),
		},
		{
			Name:  "STATIC_DOWN_ENABLED",
			Value: "true",
		},
		{
			Name:  "PANEL_TOKEN",
			Value: helper.RandomString(32),
		},
	}
	root := true
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k3kUser.GetAgentName(),
			Namespace: k3kUser.Namespace,
			Labels: map[string]string{
				"k3k-agent-pod": "true",
				"k3k-sa":        k3kUser.Name,
				"k3k-name":      k3kUser.GetK3kName(),
				"k3k-namespace": k3kUser.GetK3kNamespace(),
			},
			Annotations: map[string]string{
				"helm-version": os.Getenv("HELM_VERSION"),
				"root-pod-ip":  os.Getenv("POD_IP"),
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			// Volumes: []corev1.Volume{
			// 	{
			// 		Name: "kubeconfig",
			// 		VolumeSource: corev1.VolumeSource{
			// 			ConfigMap: &corev1.ConfigMapVolumeSource{
			// 				LocalObjectReference: corev1.LocalObjectReference{
			// 					Name: k3kUser.GetKubeconfigMapName(),
			// 				},
			// 				Items: []corev1.KeyToPath{
			// 					{
			// 						Key:  "kubeconfig",
			// 						Path: "kubeconfig.yaml",
			// 						Mode: &fileMode,
			// 					},
			// 				},
			// 			},
			// 		},
			// 	},
			// },
			ServiceAccountName: k3kUser.Name,
			HostPID:            true,
			// AutomountServiceAccountToken: true,
			InitContainers: []corev1.Container{
				{
					Name:  "w7panel-agent-init",
					Image: helper.SelfImage(),
					Env:   envs,

					ImagePullPolicy: corev1.PullAlways,
					Command:         []string{"sh", "-c", "${KO_DATA_PATH}/shell/k3k-agent-upgrade.sh"},
					// Args: []string{"server:start"},
				},
			},
			Containers: []corev1.Container{
				{

					Name:  "w7panel-agent",
					Image: helper.SelfImage(),
					Env:   envs,
					// WorkingDir:      "/tmp",
					SecurityContext: &corev1.SecurityContext{
						Privileged: &root,
					},
					ImagePullPolicy: corev1.PullAlways,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("0"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("250m"),
							corev1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 8000,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							ContainerPort: 9443,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					// Command:         []string{"sh", "-c", shell},
					// VolumeMounts: []corev1.VolumeMount{
					// 	{
					// 		Name:      "kubeconfig",
					// 		MountPath: "/w7-config/",
					// 		ReadOnly:  false,
					// 	},
					// },
					Args: []string{"server:start"},

					// Args:            []string{shell
				},
			},
		},
	}

	return pod
}

func ToK3kAgentService(k3kUser *K3kUser) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        k3kUser.GetAgentName(),
			Namespace:   k3kUser.Namespace,
			Annotations: map[string]string{"w7.cc/title": "agent服务"},
			Labels: map[string]string{
				"k3k-sa":        k3kUser.Name,
				"k3k-name":      k3kUser.GetK3kName(),
				"k3k-namespace": k3kUser.GetK3kNamespace(),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8000,
					TargetPort: intstr.FromInt(8000),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       9443,
					TargetPort: intstr.FromInt(9443),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"k3k-agent-pod": "true",
				"k3k-name":      k3kUser.GetK3kName(),
				"k3k-namespace": k3kUser.GetK3kNamespace(),
			},
		},
	}
}

func ToVirtualIngressService(k3kUser *K3kUser) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        k3kUser.GetVirtualIngressServiceName(),
			Namespace:   k3kUser.GetK3kNamespace(),
			Annotations: map[string]string{"w7.cc/title": "k3k服务w7"},
			Labels: map[string]string{
				"k3k-sa":        k3kUser.Name,
				"k3k-name":      k3kUser.GetK3kName(),
				"k3k-namespace": k3kUser.GetK3kNamespace(),
				"cluster":       k3kUser.GetK3kName(),
				"role":          "server",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(443),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "https-hook",
					Port:       9443,
					TargetPort: intstr.FromInt(9443),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "http-panel",
					Port:       8000,
					TargetPort: intstr.FromInt(8000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				// "cluster": "true",
				"cluster": k3kUser.GetK3kName(),
				"role":    "server",
			},
		},
	}
}
