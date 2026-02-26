package appgroup

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (d *WorkloadManager) createOrUpdateSvc(svcName, svcNamespace string, svcPorts []corev1.ServicePort, selector map[string]string, svcType corev1.ServiceType, headless bool) {
	svc, err := d.sdk.ClientSet.CoreV1().Services(svcNamespace).Get(context.TODO(), svcName, metav1.GetOptions{})
	isLb := false
	if svcType == corev1.ServiceTypeLoadBalancer {
		isLb = true
	}
	if err != nil {
		if errors.IsNotFound(err) {
			svc = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      svcName,
					Namespace: svcNamespace,
				},
				Spec: corev1.ServiceSpec{
					Ports:    svcPorts,
					Selector: selector,
					Type:     svcType,
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
			}
			if headless {
				svc.Spec.ClusterIP = "None"
			}
			_, err := d.sdk.ClientSet.CoreV1().Services(svcNamespace).Create(context.TODO(), svc, metav1.CreateOptions{})
			if err != nil {
				slog.Error("failed to create service", slog.String("error", err.Error()))
			}
			return
		}
	}
	if svc != nil {
		if headless {
			svc.Spec.ClusterIP = "None"
		}
		if !isLb && svc.Spec.Type != corev1.ServiceTypeClusterIP { //
			// 如果不是是负载均衡，并且原始类型不是ClusterIP，不做更新操作, 因为其他应用如helm创建了同名service
			return
		}
		svc.Spec.Ports = svcPorts
		_, err := d.sdk.ClientSet.CoreV1().Services(svcNamespace).Update(context.TODO(), svc, metav1.UpdateOptions{})
		if err != nil {
			slog.Error("failed to update service", slog.String("error", err.Error()))
		}
	}
}

func (d *WorkloadManager) delSvc(svcName, svcNamespace string) {
	err := d.sdk.ClientSet.CoreV1().Services(svcNamespace).Delete(context.TODO(), svcName, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("failed to delete service", slog.String("error", err.Error()))
	}
}

func (d *WorkloadManager) fixSvc(ds WorkloadWrapperInterface, delete bool) {
	if delete {
		d.delSvc(ds.Name(), ds.Namespace())
		d.delSvc(ds.Name()+"-lb", ds.Namespace())
		d.delSvc(ds.Name()+"-headless", ds.Namespace())
		return
	}
	d.fixNoramlSvc(ds, delete)
	d.fixLbSvc(ds, delete)
}

func (d *WorkloadManager) fixLbSvc(ds WorkloadWrapperInterface, delete bool) {
	lbPortstr, ok := ds.Annotations()["w7.cc.app/ports"]
	if !ok {
		return
	}
	var lbPorts map[string]int32
	if err := json.Unmarshal([]byte(lbPortstr), &lbPorts); err != nil {
		slog.Error("failed to unmarshal ports", slog.String("error", err.Error()))
		return
	}
	ports := ds.ContainerPorts()
	if len(ports) == 0 {
		return
	}
	template := ds.PodTemplate()
	if template == nil {
		return
	}
	servicePorts := make([]corev1.ServicePort, 0)

	for _, port := range ports {

		p := strconv.Itoa(int(port.ContainerPort))
		lbPort, ok := lbPorts[p]
		if !ok {
			continue
		}
		if lbPort == 0 {
			continue
		}
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:     port.Name,
			Protocol: port.Protocol,
			Port:     lbPort,
			TargetPort: intstr.IntOrString{
				IntVal: port.ContainerPort,
			},
		})
	}
	if len(servicePorts) == 0 {
		err := d.sdk.ClientSet.CoreV1().Services(ds.Namespace()).Delete(context.TODO(), ds.Name()+"-lb", metav1.DeleteOptions{})
		if err != nil {
			slog.Error("failed to delete service lb", slog.String("error", err.Error()))
		}
		return
	}
	d.createOrUpdateSvc(ds.Name()+"-lb", ds.Namespace(), servicePorts, template.Labels, corev1.ServiceTypeLoadBalancer, false) //更新svc-lb
}

/**/
func (d *WorkloadManager) fixNoramlSvc(ds WorkloadWrapperInterface, delete bool) {

	createHdSvc, headlessOk := ds.Annotations()["w7.cc/create-headless-svc"]
	if !headlessOk || createHdSvc != "true" {
		d.delSvc(ds.Name()+"-headless", ds.Namespace())
	}
	createSvc, ok := ds.Annotations()["w7.cc/create-svc"]
	if !ok {
		return
	}
	if ok && createSvc == "false" {
		return
	}

	ports := ds.ContainerPorts()
	if len(ports) == 0 {
		return
	}
	template := ds.PodTemplate()
	if template == nil {
		return
	}
	servicePorts := make([]corev1.ServicePort, 0)
	for _, port := range ports {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:     port.Name,
			Protocol: port.Protocol,
			Port:     port.ContainerPort,
			TargetPort: intstr.IntOrString{
				IntVal: port.ContainerPort,
			},
		})
	}
	if len(servicePorts) == 0 {
		d.delSvc(ds.Name(), ds.Namespace())
		d.delSvc(ds.Name()+"-lb", ds.Namespace())
		d.delSvc(ds.Name()+"-headless", ds.Namespace())
		return
	}
	d.createOrUpdateSvc(ds.Name(), ds.Namespace(), servicePorts, template.Labels, corev1.ServiceTypeClusterIP, false) //更新svc
	if headlessOk && createHdSvc == "true" {
		d.createOrUpdateSvc(ds.Name()+"-headless", ds.Namespace(), servicePorts, template.Labels, corev1.ServiceTypeClusterIP, true) //更新svc
	}
	// 更新svc-lb

}
