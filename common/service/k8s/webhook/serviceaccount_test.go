package webhook

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func TestCreatePatch(t *testing.T) {
	sdk := k8s.NewK8sClient().Sdk
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Fatal(err)
	}
	statefulSetName := "k3k-console-75780-server"
	namespace := "k3k-console-75780"
	statuset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      statefulSetName,
			Namespace: namespace,
		},
	}
	err = client.Get(context.Background(), types.NamespacedName{Name: statefulSetName, Namespace: namespace}, statuset)
	if err != nil {
		slog.Error("重启集群statefulset 失败", "err", err)
		return
	}
	cloneStatefulSet := statuset.DeepCopy()
	// m.client.Patch(ctx, statuset, client.Patch{})
	_, err = controllerutil.CreateOrUpdate(context.Background(), client, cloneStatefulSet, func() error {
		if cloneStatefulSet.Annotations == nil {
			cloneStatefulSet.Annotations = make(map[string]string)
		}
		cloneStatefulSet.Annotations["restart"] = time.Now().Format("2006-01-02 15:04:05")
		if cloneStatefulSet.Spec.Template.Annotations == nil {
			cloneStatefulSet.Spec.Template.Annotations = make(map[string]string)
		}
		cloneStatefulSet.Spec.Template.Annotations["restart"] = time.Now().Format("2006-01-02 15:04:05")
		cloneStatefulSet.Spec.Template.Annotations["kubernetes.io/egress-bandwidth"] = "101M"
		cloneStatefulSet.Spec.Template.Annotations["kubernetes.io/ingress-bandwidth"] = "101M"

		return nil
	})
	if err != nil {
		slog.Error("重启集群statefulset 失败 error patch", "err", err)
	}
}
