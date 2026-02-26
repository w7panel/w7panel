package webhook

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	k3ktypes "gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// 处理 Service 资源
func (m *ResourceMutator) handleServiceAccount(ctx context.Context, req admission.Request) admission.Response {
	slog.Info("处理 ServiceAccount admission 请求")

	if req.Operation != "UPDATE" {
		return admission.Allowed("无需修改 ServiceAccount")
	}

	sa := &v1.ServiceAccount{}
	if err := (m.decoder).Decode(req, sa); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	k3kUsr := k3ktypes.NewK3kUser(sa)
	if !k3kUsr.IsVirtual() {
		return admission.Allowed("无需修改 ServiceAccount")
	}

	oldSa := &v1.ServiceAccount{}
	if err := (m.decoder).DecodeRaw(req.OldObject, oldSa); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	oldK3kUser := k3ktypes.NewK3kUser(oldSa)
	if !oldK3kUser.IsVirtual() {
		return admission.Allowed("无需修改 ServiceAccount")
	}
	defer m.DoPause(ctx, oldK3kUser, k3kUsr)
	oldL := oldK3kUser.GetLimitRange()
	newL := k3kUsr.GetLimitRange()
	if oldL.IsCpuMemoryBandWidthChange(newL) || oldK3kUser.IsWeihu() != k3kUsr.IsWeihu() {
		defer m.restartClusterServer(ctx, k3kUsr.GetK3kNamespace()+"-server", k3kUsr.GetK3kNamespace(), k3kUsr.GetBandWidth())
		return admission.Allowed("修改 ServiceAccount 的 CPU 和内存限制")
	}

	return admission.Allowed("无需修改 ServiceAccount")
}

func (m *ResourceMutator) DoPause(ctx context.Context, oldK3kUser *k3ktypes.K3kUser, currentUser *k3ktypes.K3kUser) error {
	if currentUser.IsPause() != oldK3kUser.IsPause() {
		statefulSetName := currentUser.GetK3kNamespace() + "-server"
		namespace := currentUser.GetK3kNamespace()
		statuset := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      statefulSetName,
				Namespace: namespace,
			},
		}
		err := m.client.Get(context.Background(), types.NamespacedName{Name: statefulSetName, Namespace: namespace}, statuset)
		if err != nil {
			slog.Error("serviceaccount get statefulset 失败", "err", err)
			return err
		}
		if statuset == nil {
			return nil
		}
		copy := statuset.DeepCopy()
		_, err = controllerutil.CreateOrPatch(context.Background(), m.client, copy, func() error {
			if copy.Annotations == nil {
				copy.Annotations = make(map[string]string)
			}
			if copy.Spec.Template.Annotations == nil {
				copy.Spec.Template.Annotations = make(map[string]string)
			}
			if currentUser.IsPause() {
				copy.Spec.Template.Annotations[k3ktypes.W7_CREATE_POD] = "false"
				copy.Annotations[k3ktypes.W7_CREATE_POD] = "false"
			} else {
				copy.Spec.Template.Annotations[k3ktypes.W7_CREATE_POD] = "true"
				copy.Annotations[k3ktypes.W7_CREATE_POD] = "true"
			}
			return nil
		})
		if err != nil {
			slog.Error("serviceaccount 修改 statefulset pause 失败", "err", err)
			return err
		}
		if currentUser.IsPause() {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      statefulSetName + "-0",
					Namespace: namespace,
				},
			}
			err := m.client.Delete(context.Background(), pod)
			if err != nil {
				slog.Error("serviceaccount delete pod  失败", "err", err)
				return err
			}
			// time.Sleep(time.Second * 3)
			// m.restartClusterServer(ctx, statefulSetName, namespace)
		}
		return err
	}
	return nil
}

func (m *ResourceMutator) restartClusterServer(ctx context.Context, statefulSetName string, namespace string, bandwidth resource.Quantity) error {
	time.AfterFunc(time.Second*3, func() {
		statuset := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      statefulSetName,
				Namespace: namespace,
			},
		}
		err := m.client.Get(context.Background(), types.NamespacedName{Name: statefulSetName, Namespace: namespace}, statuset)
		if err != nil {
			slog.Error("重启集群statefulset 失败", "err", err)
			return
		}
		// m.client.Patch(ctx, statuset, client.Patch{})
		_, err = controllerutil.CreateOrPatch(context.Background(), m.client, statuset, func() error {
			if statuset.Annotations == nil {
				statuset.Annotations = make(map[string]string)
			}
			statuset.Annotations["restart"] = time.Now().String()
			if statuset.Spec.Template.Annotations == nil {
				statuset.Spec.Template.Annotations = make(map[string]string)
			}
			statuset.Spec.Template.Annotations["restart"] = time.Now().String()
			if !bandwidth.IsZero() {
				quantitystr := bandwidth.String()
				slog.Info("Pod 带宽限制", slog.String("bandwidth", quantitystr))
				// quantitystr = strings.ReplaceAll(quantitystr, "Mi", "Mbps")

				statuset.Spec.Template.Annotations["kubernetes.io/egress-bandwidth"] = quantitystr
				statuset.Spec.Template.Annotations["kubernetes.io/ingress-bandwidth"] = quantitystr
			}
			return nil
		})
		if err != nil {
			slog.Error("重启集群statefulset 失败 error patch", "err", err)
		}
	})
	return nil
}
