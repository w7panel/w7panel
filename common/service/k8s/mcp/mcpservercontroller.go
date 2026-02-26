package mcp

import (
	"encoding/json"
	"log/slog"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	mcpserverv1alpha1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/mcpserver/v1alpha1"
	mcpserverversioned "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/mcpserver/clientset/versioned"
	mcpserverinformers "gitee.com/we7coreteam/k8s-offline/k8s/pkg/client/mcpserver/informers/externalversions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/tools/cache"
)

func Watch() {
	sdk := k8s.NewK8sClient()
	controller, err := NewMcpServerController(sdk.Sdk)
	if err != nil {
		slog.Error("failed to create mcpserver controller", "error", err)
		return
	}
	if err := controller.Start(); err != nil {
		slog.Error("failed to start mcpserver controller", "error", err)
		return
	}
}

type mcpservercontroller struct {
	sdk       *k8s.Sdk
	factory   mcpserverinformers.SharedInformerFactory
	clientSet *mcpserverversioned.Clientset
}

func NewMcpServerController(sdk *k8s.Sdk) (*mcpservercontroller, error) {
	restConfig, err := sdk.ToRESTConfig()
	if err != nil {
		slog.Error("failed to create rest config", "error", err)
		return nil, err
	}

	clientset, err := mcpserverversioned.NewForConfig(restConfig)
	if err != nil {
		slog.Error("failed to create clientset", "error", err)
		return nil, err
	}

	factory := mcpserverinformers.NewSharedInformerFactory(clientset, 0)
	return &mcpservercontroller{
		sdk:       sdk,
		factory:   factory,
		clientSet: clientset,
	}, nil
}

func (s *mcpservercontroller) Start() error {
	informer := s.WatchMcpServerInformer()
	stopCh := make(chan struct{})
	defer close(stopCh)

	s.factory.Start(stopCh)
	if !cache.WaitForNamedCacheSync("mcpservercontroller", stopCh, informer.HasSynced) {
		slog.Debug("Failed to sync cache")
		return nil
	}

	<-stopCh
	return nil
}

func (s *mcpservercontroller) WatchMcpServerInformer() cache.SharedIndexInformer {
	informer := s.factory.Mcpserver().V1alpha1().McpServers().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, isInit bool) {
			server := obj.(*mcpserverv1alpha1.McpServer)
			s.handleMcpServer(server)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			server := newObj.(*mcpserverv1alpha1.McpServer)
			s.handleMcpServer(server)
		},
		DeleteFunc: func(obj interface{}) {
			server := obj.(*mcpserverv1alpha1.McpServer)
			s.handleMcpServerDelete(server)
		},
	})
	return informer
}

func (s *mcpservercontroller) handleMcpServer(server *mcpserverv1alpha1.McpServer) {
	slog.Info("Handling McpServer", "name", server.Name, "url", server.Spec.Url)

	// 根据command类型处理
	switch server.Spec.Command {
	case "uv", "uvx", "npx":
		err := s.createDeploymentForSSE(server)
		if err != nil {
			slog.Error("Failed to create deployment for SSE",
				"name", server.Name, "error", err)
		}
	default:
		slog.Warn("Unsupported command type", "command", server.Spec.Command)
	}
}

func (s *mcpservercontroller) createDeploymentForSSE(server *mcpserverv1alpha1.McpServer) error {
	// 创建deployment配置
	image := "supercorp/supergateway:3.0.1-uvx"
	if server.Spec.Command == "npx" {
		image = "supercorp/supergateway:3.0.1"
	}
	var proxyArgs = []string{server.Spec.Command}
	proxyArgs = append(proxyArgs, server.Spec.Args...)
	proxyArgsStr := strings.Join(proxyArgs, " ")
	args := []string{"--stdio", proxyArgsStr, "--port", "8000"}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
			Annotations: map[string]string{
				"w7.cc/create-svc": "true",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": server.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": server.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  server.Name,
							Image: image,
							Args:  args,
							Env:   convertToEnvVars(server.Spec.Env),
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8000,
								},
							},
						},
					},
				},
			},
		},
	}

	// 创建或更新deployment
	_, err := s.sdk.ClientSet.AppsV1().Deployments(s.sdk.GetNamespace()).Create(
		s.sdk.Ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			patchData := map[string]interface{}{
				"spec": deployment.Spec,
			}
			specBytes, err := json.Marshal(patchData["spec"])
			if err != nil {
				return err
			}
			_, err = s.sdk.ClientSet.AppsV1().Deployments(s.sdk.GetNamespace()).Patch(
				s.sdk.Ctx, deployment.Name, types.StrategicMergePatchType,
				[]byte(`{"spec":`+string(specBytes)+`}`),
				metav1.PatchOptions{})
		}
		return err
	}
	return nil
}

func convertToEnvVars(envs map[string]string) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	for k, v := range envs {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return envVars
}

func int32Ptr(i int32) *int32 { return &i }

func (s *mcpservercontroller) handleMcpServerDelete(server *mcpserverv1alpha1.McpServer) {
	// 实现McpServer删除处理逻辑
	slog.Info("Handling McpServer deletion", "name", server.Name)
	err := s.sdk.ClientSet.AppsV1().Deployments(s.sdk.GetNamespace()).Delete(s.sdk.Ctx, server.Name, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete deployment", "name", server.Name, "error", err)
	}
}
