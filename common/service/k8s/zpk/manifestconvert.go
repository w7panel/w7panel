package zpk

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/appgroup"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
	v1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	microapp "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/microapp/v1alpha1"
	"github.com/aws/smithy-go/ptr"
	"github.com/samber/lo"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	// appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
)

const buildimage = "ccr.ccs.tencentyun.com/afan-public/kaniko:w7console-new5-19"

type K8sResourceMetaInterface interface {
	GetTitle() string
	GetRootTitle() string
	GetName() string
	GetIdentifie() string
	GetNamespace() string
	GetReleaseName() string
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetMatchLabels() map[string]string
	GetThirdpartyCDToken() string
}

type K8sResourceInterface interface {
	K8sResourceMetaInterface

	GetImage() string
	GetReplicas() int32
	GetContainerPort() []corev1.ContainerPort
	GetCpu() string
	GetMemory() string
	GetCommand() []string
	GetRuntimeClassName() string
	GetSuffix() string
	GetPodSecurityContext() *corev1.PodSecurityContext
	GetVolumes() []corev1.Volume
	GetVolumeMounts() []corev1.VolumeMount
	GetServicePort() []corev1.ServicePort
	GetServiceLbPort() []corev1.ServicePort
	GetImagePullSecrets() []corev1.LocalObjectReference
	GetEnv() []corev1.EnvVar
	RequireBuildImage() bool
	RequireDomain() bool
	RequireDomainHttps() bool
	GetBuildJobName() string
	GetShellJobName(shellType string) string
	GetHelmInstallJobName(shellType string) string
	GetShellByType(shellType string) types.ShellInterface
	IsUpgrade() bool
	GetServiceAccountName() string
	IsPrivileged() bool
	GetZipUrl() string
	GetZpkUrl() string
	RequireSite() bool
	RequireCreateDb() (bool, string)
	GetLogo() string
	GetVersion() string
	GetHelmConfig() types.HelmConfigInterface
	GetDescription() string
	GetRootDescription() string
	IsHelm() bool
	//microapp
	SupportMicroApp() bool
	GetBackendUrl() string
	GetFrontendUrl() string
	GetMicroAppProps() map[string]string
}

type K8sResourceIngressInterface interface {
	K8sResourceInterface
	GetIngressHost() string
	GetIngressSelectorName() string
	GetIngressClassName() string
	RequireDomainHttps() bool
	GetRoutesByName(name string) []ManifestRouteInterface
	GetFirstPort() int32
	GetIngressSvcName(backendName string) string
	// GetIngressAnnations() map[string]string
}

type ManifestShellInterface interface {
	GetType() string
	GetShell() string
	GetTitle() string
	GetDisployTitle() string
	GetImage() string
}

type ManifestRouteInterface interface {
	GetBackendName() string
	GetIngName() string
	GetBackendPort() int32
	GetPath() string
	GetAnnatations() map[string]string

	GetPathType() string
}

func toPodTemplateSpec(manifest K8sResourceInterface, command []string, restartPolicy corev1.RestartPolicy,
	matchLabels map[string]string, annotations map[string]string) corev1.PodTemplateSpec {

	defaultAnn := manifest.GetAnnotations()
	if annotations != nil && len(annotations) > 0 {
		for k, v := range annotations {
			defaultAnn[k] = v
		}
	}

	isP := manifest.IsPrivileged()
	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        manifest.GetName(),
			Namespace:   manifest.GetNamespace(),
			Labels:      matchLabels,
			Annotations: defaultAnn,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: restartPolicy,
			Containers: []corev1.Container{
				{
					Name:            manifest.GetName(),
					Image:           manifest.GetImage(),
					Ports:           manifest.GetContainerPort(),
					Env:             manifest.GetEnv(),
					Command:         command,
					ImagePullPolicy: corev1.PullIfNotPresent,
					SecurityContext: &corev1.SecurityContext{Privileged: &isP},
				},
			},
			SecurityContext:    manifest.GetPodSecurityContext(),
			ImagePullSecrets:   manifest.GetImagePullSecrets(),
			ServiceAccountName: manifest.GetServiceAccountName(),
		},
	}

	if len(manifest.GetVolumes()) > 0 {
		pod.Spec.Volumes = manifest.GetVolumes()
	}
	if len(manifest.GetVolumeMounts()) > 0 {
		pod.Spec.Containers[0].VolumeMounts = manifest.GetVolumeMounts()
	}
	if (manifest.GetRuntimeClassName() != "") && (manifest.GetRuntimeClassName() != "default") {
		// Bug 修复：改为指针类型
		*(pod.Spec.RuntimeClassName) = manifest.GetRuntimeClassName()
	}
	if (manifest.GetCpu() != "") && (manifest.GetMemory() != "") {
		pod.Spec.Containers[0].Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    helper.ParseResourceLimit(manifest.GetCpu()),
				corev1.ResourceMemory: helper.ParseResourceLimit(manifest.GetMemory()),
			},
		}
	}

	return pod
}

func ToDeployment(manifest K8sResourceInterface) *v1.Deployment {
	// Bug 修复：引入 appsv1 包
	var Replicas int32 = (manifest.GetReplicas())

	// 创建一个新的 Deployment 对象
	deployment := &v1.Deployment{}
	// 设置 API 版本
	deployment.APIVersion = "apps/v1"
	// 设置资源类型
	deployment.Kind = "Deployment"
	// 设置资源名称
	deployment.SetName(manifest.GetName())
	// 设置命名空间
	deployment.SetNamespace(manifest.GetNamespace())
	// 设置标签
	deployment.SetLabels(manifest.GetLabels())
	// 设置注解
	deployment.SetAnnotations(manifest.GetAnnotations())
	// Bug 修复：改为指针类型
	// 设置副本数量
	deployment.Spec.Replicas = &Replicas
	// 设置选择器
	deployment.Spec.Selector = &metav1.LabelSelector{
		// 设置匹配标签
		MatchLabels: manifest.GetMatchLabels(),
	}
	// 设置 Pod 模板
	deployment.Spec.Template = toPodTemplateSpec(manifest, manifest.GetCommand(), corev1.RestartPolicyAlways, manifest.GetMatchLabels(), map[string]string{})
	return deployment
}

func ToService(manifest K8sResourceInterface) *corev1.Service {
	service := &corev1.Service{}
	service.APIVersion = "v1"
	service.Kind = "Service"
	service.SetName(manifest.GetName())
	service.SetNamespace(manifest.GetNamespace())
	service.SetLabels(manifest.GetLabels())
	service.SetAnnotations(manifest.GetAnnotations())
	service.Spec.Selector = manifest.GetMatchLabels()
	service.Spec.Ports = manifest.GetServicePort()
	return service
}

func ToLoadBalancerService(manifest K8sResourceInterface) *corev1.Service {
	lbClass := "io.cilium/node"
	service := &corev1.Service{}
	service.APIVersion = "v1"
	service.Kind = "Service"
	service.Spec.Type = corev1.ServiceTypeLoadBalancer
	service.SetName(manifest.GetName() + "-lb")
	service.SetNamespace(manifest.GetNamespace())
	service.SetLabels(manifest.GetLabels())
	service.SetAnnotations(manifest.GetAnnotations())
	service.Spec.Selector = manifest.GetMatchLabels()
	service.Spec.Ports = manifest.GetServiceLbPort()
	service.Spec.LoadBalancerClass = &lbClass
	// service.Spec.
	return service
}

func ToBuildJob(p K8sResourceInterface, opt types.BuildImageInterface, shellType string) *batchv1.Job {

	shellTitle := "[应用安装时触发]"
	if shellType == "upgrade" {
		shellTitle = "应用更新时触发"
	}
	option := types.NewBuildImageOption(opt)
	title := shellTitle + p.GetTitle() + "构建镜像任务"
	annotations := map[string]string{
		"title":              title,
		"w7.cc/title":        title,
		"w7.cc/deploy-title": "部署任务",
		"w7.cc/shell-type":   shellType,
		"w7.cc/release-name": p.GetReleaseName(),
		"w7.cc/group-name":   p.GetReleaseName(),
	}
	annotations["w7.cc/job-source"] = "appgroup"
	annotations["w7.cc/job-source-name"] = p.GetName()
	annotations["w7.cc/job-source-title"] = p.GetTitle()
	labels := p.GetLabels()
	labels["w7.cc/job-source"] = "appgroup"
	labels["searchJob"] = p.GetName() + "-build-" + shellType
	labels["w7.cc/shell-type"] = shellType
	backofflimit := int32(3)
	// afterSeconds := int32(3600)
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        p.GetBuildJobName(),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			// TTLSecondsAfterFinished: &afterSeconds,
			BackoffLimit: &backofflimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					//挂载hostPath
					Volumes:       option.GetVolumes(),
					DNSPolicy:     corev1.DNSClusterFirstWithHostNet,
					HostNetwork:   option.GetHostNetwork(),
					HostPID:       option.GetHostNetwork(),
					HostAliases:   option.GetHostAliases(),
					HostIPC:       option.IsInner(),
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:            "docker-build",
							Image:           buildimage,
							Env:             option.ToEnv(),
							WorkingDir:      "/workspace",
							ImagePullPolicy: corev1.PullAlways,
							Command:         []string{"/kaniko/start.sh"},
							VolumeMounts:    option.GetVolumeMounts(),
						},
					},
				},
			},
		},
	}
	return job
}

/*
*
执行安装和更新命令的 Job
*/
func ToShellJob(p K8sResourceInterface, shell ManifestShellInterface) *batchv1.Job {
	shellStr := shell.GetShell()
	shellStr = strings.ReplaceAll(shellStr, "%CODE_ZIP_URL%", p.GetZipUrl())
	cmd := []string{"/bin/sh", "-c", shellStr}
	jobName := p.GetShellJobName(shell.GetType())
	matchlabels := map[string]string{
		"job": jobName,
	}
	labels := map[string]string{
		"job":                jobName,
		"app":                p.GetName(),
		"searchJob":          p.GetName() + "-" + shell.GetType(),
		"w7.cc/shell-type":   shell.GetType(),
		"w7.cc/release-name": p.GetReleaseName(),
		"w7.cc/group-name":   p.GetReleaseName(),
	}
	annotations := map[string]string{
		"title":              shell.GetDisployTitle(),
		"w7.cc/title":        shell.GetDisployTitle(),
		"w7.cc/shell-type":   shell.GetType(),
		"w7.cc/release-name": p.GetReleaseName(),
		"w7.cc/group-name":   p.GetReleaseName(),
	}
	annotations["w7.cc/job-source"] = "appgroup"
	annotations["w7.cc/job-source-name"] = p.GetName()
	annotations["w7.cc/job-source-title"] = p.GetTitle()

	labels["w7.cc/job-source"] = "appgroup"

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			Template: toPodTemplateSpec(p, cmd, corev1.RestartPolicyNever, matchlabels, annotations),
		},
	}
	return job
}

// ToIngresses 生成 Ingress
func ToIngresses(app K8sResourceIngressInterface) []networkingv1.Ingress {
	if app.GetIngressHost() == "" {
		return []networkingv1.Ingress{}
	}
	requireHttps := app.RequireDomainHttps()

	if app.GetIngressHost() != "" {
		if app.GetIngressSelectorName() == "" {
			ing := toIngress(app, app.GetName(), app.GetFirstPort(), "/", requireHttps, "", "Prefix")
			ing.Labels["group"] = app.GetReleaseName()
			return []networkingv1.Ingress{ing}
		}
	}
	if app.GetIngressSelectorName() != "" {
		routes := app.GetRoutesByName(app.GetIngressSelectorName())
		if len(routes) == 0 {
			ing := toIngress(app, app.GetName(), app.GetFirstPort(), "/", requireHttps, "", "Prefix")
			ing.Labels["group"] = app.GetReleaseName()
			return []networkingv1.Ingress{ing}
		} else if len(routes) > 0 {
			var ingressList []networkingv1.Ingress
			parentIngName := ""
			for key, route := range routes {
				backendName := route.GetBackendName()
				ing := toIngress(app, app.GetIngressSvcName(backendName), int32(route.GetBackendPort()), route.GetPath(), requireHttps, route.GetIngName(), route.GetPathType())
				for k, v := range route.GetAnnatations() {
					ing.Annotations[k] = v
				}
				if key == 0 {
					parentIngName = ing.GetName()
				}
				if key > 0 {
					ing.Labels["parents"] = parentIngName
				}
				ing.Labels["group"] = app.GetReleaseName()
				ingressList = append(ingressList, ing)
			}
			return ingressList
		}
	}

	return []networkingv1.Ingress{}

}

func toIngress(app K8sResourceIngressInterface, svcName string, port int32, path string, requireHttps bool, ingName string, pathTypeStr string) networkingv1.Ingress {
	pathType := networkingv1.PathType(pathTypeStr)
	annotations := map[string]string{}
	tls := []networkingv1.IngressTLS{}
	if requireHttps {
		annotations = map[string]string{
			"cert-manager.io/cluster-issuer": "w7-letsencrypt-prod",
			"cert-manager.io/renew-before":   "30m",
		}
		tls = []networkingv1.IngressTLS{
			{
				Hosts:      []string{app.GetIngressHost()},
				SecretName: strings.ToLower(strings.ReplaceAll(app.GetIngressHost(), ".", "-")) + "-tls-secret",
			},
		}
	}
	ingressClassName := app.GetIngressClassName()
	if ingressClassName == "" {
		ingressClassName = "higress"
		// 兼容k3s虚拟集群模式
		// k3kMode, ok := os.LookupEnv("K3K_MODE")
		// if ok && k3kMode == "virtual" {
		// 	ingressClassName = "traefik"
		// }
	}
	annotations["kubernetes.io/ingress.class"] = ingressClassName
	annotations["w7.cc/ingress-selector-name"] = app.GetIngressSelectorName()
	// annotations["w7.cc/ingress-selector-path"] =
	if ingName == "" {
		ingName = "ing-" + helper.RandomString(10)
	}

	ingress := networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingName,
			Namespace:   app.GetNamespace(),
			Annotations: annotations,
			Labels:      app.GetLabels(),
		},
		Spec: networkingv1.IngressSpec{
			TLS: tls,
			Rules: []networkingv1.IngressRule{
				{
					Host: app.GetIngressHost(), //TODO
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     path,
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: svcName,
											Port: networkingv1.ServiceBackendPort{
												Number: port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ingress
}

func ToMysqlSecret(p K8sResourceIngressInterface) *corev1.Secret {
	envs := p.GetEnv()
	username := ""
	password := ""
	for _, env := range envs {
		if env.Name == "MYSQL_ROOT_USERNAME" {
			username = env.Value
		}
		if env.Name == "MYSQL_ROOT_PASSWORD" {
			password = env.Value
		}
	}
	name := strings.ReplaceAll(p.GetIdentifie(), "-", "_")
	labels := map[string]string{
		"app.kubernetes.io/component": name,
	}
	return toSecret(p, username, password, labels, false)
}

func ToRedisSecret(p K8sResourceIngressInterface) *corev1.Secret {
	envs := p.GetEnv()
	username := ""
	password := ""
	for _, env := range envs {
		if env.Name == "REDIS_PASSWORD" {
			password = env.Value
		}
	}
	labels := map[string]string{
		"app.kubernetes.io/component": "w7_redis",
	}
	return toSecret(p, username, password, labels, true)
}

func toSecret(p K8sResourceIngressInterface, username, password string, l map[string]string, isRedis bool) *corev1.Secret {
	labels := p.GetLabels()
	for k, v := range l {
		labels[k] = v
	}
	port := p.GetFirstPort()
	portstr := strconv.Itoa(int(port))
	host := helper.ClusterDomain(p.GetName(), p.GetNamespace()) //p.GetName() + "." + p.GetNamespace() + ".svc.cluster.local"

	data := map[string][]byte{
		"HOST":                []byte(host),
		"PORT":                []byte(portstr),
		"USERNAME":            []byte(username),
		"PASSWORD":            []byte(password),
		"MYSQL_HOST":          []byte(host),
		"MYSQL_PORT":          []byte(portstr),
		"MYSQL_ROOT_USERNAME": []byte(username),
		"MYSQL_ROOT_PASSWORD": []byte(password),
	}
	if isRedis {
		data = map[string][]byte{
			"HOST":           []byte(host),
			"PORT":           []byte(portstr),
			"PASSWORD":       []byte(password),
			"REDIS_HOST":     []byte(host),
			"REDIS_PORT":     []byte(portstr),
			"REDIS_PASSWORD": []byte(password),
		}
	}
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.GetName(),
			Namespace: p.GetNamespace(),
			Labels:    labels,
		},
		Data: data,
		Type: corev1.SecretTypeOpaque,
	}
}

func ToShellJob2(manifest K8sResourceInterface, ingress K8sResourceIngressInterface, shellType string) *batchv1.Job {

	shell := manifest.GetShellByType(shellType)
	initContainers := []corev1.Container{}
	containers := []corev1.Container{}
	c := []corev1.Container{}
	host := ingress.GetIngressHost()
	releaseName := manifest.GetReleaseName()
	deploymentName := manifest.GetName()
	namespace := manifest.GetNamespace()
	cdToken := manifest.GetThirdpartyCDToken()
	// afterSeconds := int32(3600)
	shellTitle := "[应用安装时触发]"
	deployTitle := "安装脚本"
	if shellType == "upgrade" {
		shellTitle = "[应用更新时触发]"
		deployTitle = "更新脚本"
	}
	recordApp := []string{}
	// if manifest.RequireSite() && shellType == "install" { //创建站点会更新版本号
	if manifest.RequireSite() {
		cmd := "ko-app/k8s-offline site:register --thirdPartyCDToken=" + cdToken + " --host=" + host + " --releaseName=" + releaseName + " --deploymentName=" + deploymentName + " --namespace=" + namespace
		siteC := corev1.Container{
			Name:  "check-site",
			Image: helper.SelfImage(),
			Env: []corev1.EnvVar{
				{
					Name:  "LONGHORN_WATCH",
					Value: "false",
				},
				{
					Name:  "METRIC_ENABLED",
					Value: "false",
				},
			},
			Command:         []string{"/bin/sh", "-c", cmd},
			ImagePullPolicy: corev1.PullIfNotPresent,
		}
		useragent, ok := os.LookupEnv("USER_AGENT")
		if ok {
			siteC.Env = append(siteC.Env, corev1.EnvVar{
				Name:  "USER_AGENT",
				Value: useragent,
			})
		}
		c = append(c, siteC)
		recordApp = append(recordApp, "")
	}
	if shell != nil {
		shellStr := shell.GetShell()
		shellStr = strings.ReplaceAll(shellStr, "%CODE_ZIP_URL%", manifest.GetZipUrl())
		shellcmd := []string{"/bin/sh", "-c", shellStr}
		isP := manifest.IsPrivileged()
		shellImage := shell.GetImage()
		if manifest.IsHelm() {
			shellImage = helper.SelfImage()
		}
		if shellImage == "" {
			shellImage = manifest.GetImage()
		}
		envs := manifest.GetEnv()
		envs = lo.Filter[corev1.EnvVar](envs, func(env corev1.EnvVar, _ int) bool {
			return helper.IsValidEnvVarName(env.Name)
		})
		envs = append(envs, corev1.EnvVar{
			Name:  "RELEASE_NAME_SUFFIX",
			Value: manifest.GetReleaseName(),
		})

		shellContainer := corev1.Container{
			Name:            manifest.GetName(),
			Image:           shellImage,
			Ports:           manifest.GetContainerPort(),
			Env:             envs,
			Command:         shellcmd,
			ImagePullPolicy: corev1.PullIfNotPresent,
			SecurityContext: &corev1.SecurityContext{Privileged: &isP},
		}
		c = append(c, shellContainer)
		recordApp = append(recordApp, manifest.GetName())
	}
	ok, dbName := manifest.RequireCreateDb()
	if ok {
		cmd := "/ko-app/k8s-offline db:create-inner --database=" + dbName + " --namespace=" + namespace
		createDbC := corev1.Container{
			Name:  "create-db",
			Image: helper.SelfImage(),
			Env: []corev1.EnvVar{
				{
					Name:  "LONGHORN_WATCH",
					Value: "false",
				},
				{
					Name:  "METRIC_ENABLED",
					Value: "false",
				},
				{
					Name:  "HIGRESS_WATCH",
					Value: "false",
				},
				{
					Name:  "REGISTRY_WATCH",
					Value: "false",
				},
			},
			Command:         []string{"/bin/sh", "-c", cmd},
			ImagePullPolicy: corev1.PullIfNotPresent,
		}
		c = append(c, createDbC)
		recordApp = append(recordApp, "")
	}

	if len(c) == 0 {
		return nil
	}
	if len(c) == 1 {
		containers = c
	}
	if len(c) > 1 {
		lastIndex := len(c) - 1
		initContainers = c[0:lastIndex]
		containers = c[lastIndex:]
	}

	matchlabels := map[string]string{
		"job": manifest.GetShellJobName(shellType),
	}
	labels := map[string]string{
		"job":                manifest.GetShellJobName(shellType),
		"app":                manifest.GetName(),
		"searchJob":          manifest.GetName() + "-" + shellType,
		"w7.cc/shell-type":   shellType,
		"w7.cc/release-name": manifest.GetReleaseName(),
		"w7.cc/group-name":   manifest.GetReleaseName(),
	}
	// jobTitle := shellTitle + "初始化" + manifest.GetTitle()
	aTitle := manifest.GetTitle()
	if aTitle == "" {
		aTitle = "默认任务"
	}
	jobTitle := shellTitle + aTitle
	recordAppJson, err := json.Marshal(recordApp)
	if err != nil {
		recordAppJson = []byte("")
	}
	annotations := map[string]string{
		"title":              jobTitle,
		"w7.cc/title":        jobTitle,
		"w7.cc/deploy-title": deployTitle,
		"w7.cc/shell-type":   shellType,
		"recordApp":          string(recordAppJson),
		"w7.cc/release-name": manifest.GetReleaseName(),
		"w7.cc/group-name":   manifest.GetReleaseName(),
	}

	annotations["w7.cc/job-source"] = "appgroup"
	annotations["w7.cc/job-source-name"] = manifest.GetName()
	annotations["w7.cc/job-source-title"] = manifest.GetTitle()

	labels["w7.cc/job-source"] = "appgroup"

	defaultAnn := manifest.GetAnnotations()
	if annotations != nil && len(annotations) > 0 {
		for k, v := range annotations {
			defaultAnn[k] = v
		}
	}

	volumes := []corev1.Volume{}
	volumes = append(volumes, corev1.Volume{
		Name: "tmp",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
	volumes = append(volumes, manifest.GetVolumes()...)

	volumesMounts := []corev1.VolumeMount{}
	volumesMounts = append(volumesMounts, corev1.VolumeMount{
		Name:      "tmp",
		MountPath: "/tmp-w7-data",
	})
	volumesMounts = append(volumesMounts, manifest.GetVolumeMounts()...)

	if len(volumesMounts) > 0 {
		for i := 0; i < len(containers); i++ {
			containers[i].VolumeMounts = volumesMounts
		}
		for i := 0; i < len(initContainers); i++ {
			initContainers[i].VolumeMounts = volumesMounts
		}
		// pod.Spec.Containers[0].VolumeMounts = manifest.GetVolumeMounts()
	}

	pod := corev1.PodTemplateSpec{

		ObjectMeta: metav1.ObjectMeta{
			Name:        manifest.GetName(),
			Namespace:   manifest.GetNamespace(),
			Labels:      matchlabels,
			Annotations: defaultAnn,
		},
		Spec: corev1.PodSpec{

			RestartPolicy:      corev1.RestartPolicyNever,
			Volumes:            volumes,
			InitContainers:     initContainers,
			Containers:         containers,
			SecurityContext:    manifest.GetPodSecurityContext(),
			ImagePullSecrets:   manifest.GetImagePullSecrets(),
			ServiceAccountName: manifest.GetServiceAccountName(),
		},
	}

	if (manifest.GetRuntimeClassName() != "") && (manifest.GetRuntimeClassName() != "default") {
		// Bug 修复：改为指针类型
		*(pod.Spec.RuntimeClassName) = manifest.GetRuntimeClassName()
	}
	backofflimit := int32(3)
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        manifest.GetShellJobName(shellType),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			// TTLSecondsAfterFinished: &afterSeconds,
			BackoffLimit: &backofflimit,
			// Selector: &metav1.LabelSelector{
			// 	MatchLabels: matchlabels,
			// },
			Template: pod,
		},
	}
	return job
}

// 安装helm 应用的job
func ToHelmShellJob(p K8sResourceInterface, shell ManifestShellInterface) *batchv1.Job {
	shellStr := shell.GetShell()
	shellStr = strings.ReplaceAll(shellStr, "%CODE_ZIP_URL%", p.GetZipUrl())
	cmd := []string{"/bin/sh", "-c", shellStr}
	jobName := p.GetHelmInstallJobName(shell.GetType())
	var backofflimit int32
	backofflimit = 1
	matchlabels := map[string]string{
		"job": jobName,
		// "job-random": helper.RandomString(10),
	}
	labels := map[string]string{
		"job":                jobName,
		"app":                p.GetName(),
		"w7.cc/shell-type":   shell.GetType(),
		"searchJob":          p.GetName() + "-" + shell.GetType(),
		"w7.cc/release-name": p.GetReleaseName(),
		"w7.cc/group-name":   p.GetReleaseName(),
	}
	annotations := map[string]string{
		"title":                   shell.GetDisployTitle(),
		"w7.cc/title":             shell.GetDisployTitle(),
		"w7.cc/deploy-title":      "helm安装任务",
		"w7.cc/shell-type":        shell.GetType(),
		"w7.cc/release-name":      p.GetReleaseName(),
		"w7.cc/group-name":        p.GetReleaseName(),
		"helm.sh/resource-policy": "keep",
		// "helm.sh/hook": "pre-install, pre-upgrade",
		// "helm.sh/hook-weight": "-5",
		// "helm.sh/hook-delete-policy": "before-hook-creation, hook-succeeded",
	}
	annotations["w7.cc/job-source"] = "appgroup"
	annotations["w7.cc/job-source-name"] = p.GetReleaseName()
	annotations["w7.cc/job-source-title"] = p.GetTitle()
	annotations["w7.cc/identifie"] = p.GetIdentifie()
	annotations["w7.cc/helm-install"] = "true"

	labels["w7.cc/job-source"] = "appgroup"
	labels["w7.cc/identifie"] = p.GetIdentifie()
	labels["w7.cc/helm-install"] = "true"

	container := corev1.Container{
		Name:  "helm-go",
		Image: helper.SelfImage(),
		Env: []corev1.EnvVar{
			{
				Name:  "LONGHORN_WATCH",
				Value: "false",
			},
			{
				Name:  "METRIC_ENABLED",
				Value: "false",
			},
			{
				Name:  "REGISTRY_WATCH",
				Value: "false",
			},
			{
				Name:  "HIGRESS_WATCH",
				Value: "false",
			},
			{
				Name:  "APP_WATCH",
				Value: "false",
			},
		},
		Command:         cmd,
		ImagePullPolicy: corev1.PullAlways,
	}

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backofflimit,
			// TTLSecondsAfterFinished: &afterSeconds,
			// BackoffLimit: &backofflimit,
			// Selector: &metav1.LabelSelector{
			// 	MatchLabels: matchlabels,
			// },

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      p.GetName(),
					Namespace: p.GetNamespace(),
					Labels:    matchlabels,
				},
				Spec: corev1.PodSpec{
					//挂载hostPath
					DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers:         []corev1.Container{container},
					ServiceAccountName: p.GetServiceAccountName(),
				},
			},
		},
	}
	return job
}

/**
type AppGroupSpec struct {
	Identifie     string            `json:"identifie"`     //应用标识
	Type          string            `json:"type"`          // "helm" or "zpk" or "custom"
	Version       string            `json:"version"`       //版本号
	Title         string            `json:"title"`         //应用标题
	Logo          string            `json:"logo"`          //logo地址
	Description   string            `json:"description"`   //应用描述
	Domains       []string          `json:"domains"`       //域名列表
	DefaultDomain string            `json:"defaultDomain"` //默认域名
	ZpkUrl        string            `json:"zpkUrl"`        //制品库地址
	HelmConfig    HelmConfig        `json:"helmConfig"`    //helm配置
	Annotations   map[string]string `json:"annotations"`   //annotations
	IsHelm        bool              `json:"isHelm"`        //是否为helm应用
}

*/

func ToAppGroup(p K8sResourceInterface, installResult []v1alpha1.DeployItem) *v1alpha1.AppGroup {

	aType := "zpk"
	isHelm := false
	if p.IsHelm() {
		aType = "helm"
		isHelm = true
	}
	helmConfig := p.GetHelmConfig()
	helmSpec := v1alpha1.HelmConfig{
		Repository: helmConfig.GetRepository(),
		Version:    helmConfig.GetVersion(),
		ChartName:  helmConfig.GetChartName(),
	}

	obj := appgroup.CreateAppGroup(p.GetReleaseName(), p.GetNamespace())
	obj.Labels = p.GetLabels()
	obj.Annotations = p.GetAnnotations()
	obj.Spec = v1alpha1.AppGroupSpec{
		Type:        aType,
		Title:       p.GetRootTitle(),
		Logo:        p.GetLogo(),
		Identifie:   p.GetIdentifie(),
		Description: p.GetRootDescription(),
		Suffix:      p.GetSuffix(),
		ZpkUrl:      p.GetZpkUrl(),
		Version:     p.GetVersion(),
		HelmConfig:  helmSpec,
		IsHelm:      isHelm,
	}
	obj.Status = v1alpha1.AppGroupStatus{
		DeployItems:  installResult,
		DeployStatus: v1alpha1.StatusPendingInstall,
		Items:        []v1alpha1.AppGroupItemStatus{},
		// InstallResult: v1alpha1.InstallResult{
		// 	Resource: p.GetResourceName(),
		// },
	}
	return obj

}

func ToMicroApp(p K8sResourceInterface) *microapp.MicroApp {

	if !p.SupportMicroApp() {
		return nil
	}
	obj := microapp.MicroApp{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "microapp.w7.cc/v1alpha1",
			Kind:       "MicroApp",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        p.GetName(),
			Namespace:   p.GetNamespace(),
			Labels:      p.GetLabels(),
			Annotations: p.GetAnnotations(),
		},
		Spec: microapp.MicroAppSpec{
			BackendUrl:  p.GetBackendUrl(),
			Description: p.GetDescription(),
			FrontendUrl: p.GetFrontendUrl(),
			Framework:   "wujie",
			Title:       p.GetTitle(),
			Logo:        p.GetLogo(),
			Config: microapp.MicroAppConfig{
				Props: p.GetMicroAppProps(),
			},
		},
	}
	//
	return &obj
}

func toBuildPodSpec(option types.BuildImageOption) corev1.PodSpec {
	return corev1.PodSpec{
		//挂载hostPath
		Volumes:       option.GetVolumes(),
		DNSPolicy:     corev1.DNSClusterFirstWithHostNet,
		HostNetwork:   option.GetHostNetwork(),
		HostPID:       option.GetHostNetwork(),
		HostAliases:   option.GetHostAliases(),
		HostIPC:       option.IsInner(),
		RestartPolicy: corev1.RestartPolicyNever,
		Containers: []corev1.Container{
			{
				Name:            "docker-build",
				Image:           buildimage,
				Env:             option.ToEnv(),
				WorkingDir:      "/workspace",
				ImagePullPolicy: corev1.PullAlways,
				Command:         []string{"/kaniko/start.sh"},
				VolumeMounts:    option.GetVolumeMounts(),
			},
		},
	}
}

func ToZpkBuildJob(opt types.BuildImageInterface) *batchv1.Job {

	option := types.NewBuildImageOption(opt)
	title := opt.GetTitle()
	annotations := map[string]string{
		"title":       title,
		"w7.cc/title": title,
	}
	labels := opt.GetLabels()

	backofflimit := int32(1)
	afterSeconds := int32(3600)
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opt.GetBuildJobName(),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &afterSeconds,
			BackoffLimit:            &backofflimit,
			Template: corev1.PodTemplateSpec{
				Spec: toBuildPodSpec(option),
			},
		},
	}
	return job
}

func ToZpkBuildCronJob(opt types.BuildImageInterface, schedule string) *batchv1.CronJob {

	option := types.NewBuildImageOption(opt)
	title := opt.GetTitle()
	annotations := map[string]string{
		"title":       title,
		"w7.cc/title": title,
	}
	labels := opt.GetLabels()

	backofflimit := int32(3)
	afterSeconds := int32(3600)
	job := &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opt.GetBuildJobName(),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.CronJobSpec{
			// TTLSecondsAfterFinished: &afterSeconds,
			Schedule: schedule, //"*/1 * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					TTLSecondsAfterFinished: &afterSeconds,
					BackoffLimit:            &backofflimit,
					Template: corev1.PodTemplateSpec{
						Spec: toBuildPodSpec(option),
					},
				},
			},
		},
	}
	return job
}
func ToBeianCheckJob(info K8sResourceInterface, host string) *batchv1.Job {
	shellStr := "ko-app/k8s-offline beian-check --host=" + host
	cmd := []string{"/bin/sh", "-c", shellStr}
	jobName := "beian-job-" + helper.RandomString(8)

	labels := map[string]string{
		"job": jobName,
	}
	annotations := map[string]string{}
	annotations["w7.cc/deploy-title"] = "域名备案检测"
	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "OnFailure",
			Containers: []corev1.Container{
				{
					Name:  "beian-check",
					Image: helper.SelfImage(),
					// Env:             ,
					Command:         cmd,
					ImagePullPolicy: corev1.PullIfNotPresent,
				},
			},
			ServiceAccountName: info.GetServiceAccountName(),
		},
	}
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			Template: pod,
		},
	}
	return job
}

func ToKubeblockInstallJob(saName string) *batchv1.Job {
	shellStr := `
kubectl apply -f $KO_DATA_PATH/crds-kubeblocks/1.0.1/ --server-side
kubectl apply -f $KO_DATA_PATH/crds-kubeblocks/snapshot/ --server-side
helm upgrade kubeblocks $KO_DATA_PATH/charts/kubeblocks-1.0.1.tgz -n kb-system --create-namespace --install
`
	cmd := []string{"/bin/sh", "-c", shellStr}
	jobName := "kubeblock-job-" + helper.RandomString(8)

	labels := map[string]string{
		"job":       jobName,
		"kubeblock": "install",
	}
	annotations := map[string]string{}
	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{
				{
					Name:  "kubeblocks-install",
					Image: helper.SelfImage(),
					// Env:             ,
					Command:         cmd,
					ImagePullPolicy: corev1.PullIfNotPresent,
				},
			},
			ServiceAccountName: saName,
		},
	}
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: ptr.Int32(300),
			Template:                pod,
		},
	}
	return job
}
