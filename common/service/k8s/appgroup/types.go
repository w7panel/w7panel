package appgroup

import (
	"encoding/json"
	"time"

	"gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	appv1 "gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
)

const HELM_SECRET_TYPE = "helm.sh/release.v1"

type K8sResourceEvent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	EventType         string `json:"eventType"`
	IsInit            bool   `json:"isInit"`
}

//	func NewK8sResourceEvent(apiVerion, kind string, name, namespace, eventType string, isInit bool) *K8sResourceEvent {
//		return &K8sResourceEvent{
//			TypeMeta: metav1.TypeMeta{
//				Kind:       kind,
//				APIVersion: apiVerion,
//			},
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      name,
//				Namespace: namespace,
//			},
//			EventType: eventType,
//			IsInit:    isInit,
//		}
//	}
func NewK8sResourceEvent(typemeta metav1.TypeMeta, objMeta metav1.ObjectMeta, eventType string, isInit bool) *K8sResourceEvent {
	return &K8sResourceEvent{
		TypeMeta:   typemeta,
		ObjectMeta: objMeta,
		EventType:  eventType,
		IsInit:     isInit,
	}
}

func NewWorkloadFromEvent(event *K8sResourceEvent) WorkloadInterface {
	// annotations := map[string]string{"w7.cc/release-name": groupname}
	// labels := map[string]string{"w7.cc/release-name": groupname}
	return &WorkloadMock{
		apiversion: event.APIVersion,
		kind:       event.Kind,
		typeMeta:   event.TypeMeta,
		metaData:   event.ObjectMeta,
		// CreateTime:  obj.Metadata().CreationTimestamp,
	}
}
func NewK8sResourceEventFromJson(data []byte) (*K8sResourceEvent, error) {
	event := &K8sResourceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}
func (event *K8sResourceEvent) ToJson() ([]byte, error) {
	return json.Marshal(event)
}
func (event *K8sResourceEvent) ToString() string {
	b, _ := event.ToJson()
	return string(b)
}

type GvkObject struct {
	ApiVerion    string
	Kind         string
	Name         string
	Ready        bool
	Title        string
	IsMaster     bool
	DisployOrder int
	CreateTime   metav1.Time
}

type WorkloadInterface interface {
	ApiVersion() string
	Kind() string
	Metadata() metav1.ObjectMeta
	Labels() map[string]string
	Name() string
	Namespace() string
	Annotations() map[string]string
	Ready() bool
	DeployStatus() string
	IsZeroReplicas() bool
	Containers() []corev1.Container
	PodTemplate() *corev1.PodTemplateSpec
}
type WorkloadWrapperInterface interface {
	WorkloadInterface
	IsHelm() bool
	IsMaster() bool
	Title() string
	ParentName() string
	Identifie() string
	GvkObject() GvkObject
	ToItemStatus() appv1.AppGroupItemStatus
	ReleaseName() string
	ResourceVersion() string
	IsDelete() bool
	queueKey() string
	ContainerPorts() []corev1.ContainerPort
}
type WorkloadWrapper struct {
	WorkloadInterface
	isDelete bool
}

func NewWorkloadWrapper(obj interface{}) WorkloadWrapperInterface {

	switch v := obj.(type) {
	case *appsv1.Deployment:
		return &WorkloadWrapper{&DeploymentWrapper{v}, false}
	case *appsv1.StatefulSet:
		return &WorkloadWrapper{&StatefulSetWrapper{v}, false}
	case *appsv1.DaemonSet:
		return &WorkloadWrapper{&DaemonsetWrapper{v}, false}
	case *batchv1.Job:
		return &WorkloadWrapper{&JobWrapper{v}, false}
	case *corev1.Secret:
		secret := obj.(*corev1.Secret)
		wi := NewWorkloadMock("v1", "Secret", secret.Name, secret.Namespace, secret.Labels["name"])
		return &WorkloadWrapper{wi, false}
	case *WorkloadMock:
		return &WorkloadWrapper{v, false}
	default:
		return nil
	}
}

func NewWrapper(obj WorkloadInterface) *WorkloadWrapper {
	return &WorkloadWrapper{
		WorkloadInterface: obj,
	}
}
func (w *WorkloadWrapper) GvkObject() GvkObject {
	return GvkObject{
		ApiVerion:    w.WorkloadInterface.ApiVersion(),
		Kind:         w.WorkloadInterface.Kind(),
		Name:         w.WorkloadInterface.Name(),
		Ready:        w.Ready(),
		Title:        w.Title(),
		IsMaster:     w.IsMaster(),
		DisployOrder: 0,
		CreateTime:   w.Metadata().CreationTimestamp,
	}
}

func (d *WorkloadWrapper) ToItemStatus() appv1.AppGroupItemStatus {
	gvk := d.GvkObject()
	return appv1.AppGroupItemStatus{
		ApiVersion:        gvk.ApiVerion,
		Kind:              gvk.Kind,
		Name:              gvk.Name,
		Ready:             gvk.Ready,
		Title:             gvk.Title,
		CreationTimestamp: gvk.CreateTime,
		IsHelmWorkLoad:    d.IsHelm(),
		DeployStatus:      d.DeployStatus(),
		IsZeroReplicas:    d.IsZeroReplicas(),
	}
}
func (w *WorkloadWrapper) ResourceVersion() string {
	return w.WorkloadInterface.Metadata().ResourceVersion
}

func (w *WorkloadWrapper) IsDelete() bool {
	return !w.isDelete
}

func (w *WorkloadWrapper) queueKey() string {
	name := w.Name()
	releaseName := w.ReleaseName()
	if releaseName != "" {
		name = releaseName
	}
	return w.ApiVersion() + "." + w.Kind() + "." + w.Name() + "." + w.Namespace() + "." + name
}

func (w *WorkloadWrapper) IsHelm() bool {
	labels := w.Labels()
	if labels == nil {
		return false
	}
	val, ok := labels["app.kubernetes.io/managed-by"]
	if ok {
		return val == "Helm"
	}
	return ok
}
func (d *WorkloadWrapper) IsMaster() bool {
	labels := d.Labels()
	if labels == nil {
		return false
	}
	_, ok := labels["parent"]
	return !ok
}
func (d *WorkloadWrapper) ParentName() string {
	labels := d.Labels()
	if labels == nil {
		return ""
	}
	_, ok := labels["parent"]
	if ok {
		return labels["parent"]
	}
	return ""
	//isHelm
}

// 后缀标识
func (d *WorkloadWrapper) ReleaseName() string {
	labels := d.Labels()
	if labels == nil {
		return ""
	}
	// 多helm 应用使用group-name 来分组
	_, ok3 := labels["w7.cc/group-name"]
	if ok3 {
		return labels["w7.cc/group-name"]
	}
	_, ok2 := labels["w7.cc/release-name"]
	if ok2 {
		return labels["w7.cc/release-name"]
	}
	_, ok1 := labels["app.kubernetes.io/instance"]
	if ok1 {
		return labels["app.kubernetes.io/instance"]
	}
	_, ok := labels["w7.cc/suffix"]
	if ok {
		return labels["w7.cc/suffix"]
	}

	return ""
	//isHelm
}
func (d *WorkloadWrapper) Identifie() string {
	return d.ReleaseName()
	//isHelm
}

func (d *WorkloadWrapper) Title() string {
	ann := d.Annotations()
	if ann == nil {
		return d.Name()
	}
	_, ok := ann["title"]
	if ok {
		return ann["title"]
	}
	return ""
	//isHelm
}

func (d *WorkloadWrapper) ContainerPorts() []corev1.ContainerPort {
	containers := d.Containers()
	ports := make([]corev1.ContainerPort, 0)
	for _, container := range containers {
		// return container.Ports
		for _, port := range container.Ports {
			ports = append(ports, port)
		}
	}
	return ports
	//isHelm
}

type DeploymentWrapper struct {
	*appsv1.Deployment
}

func (d *DeploymentWrapper) ApiVersion() string {
	return "apps/v1"
}
func (d *DeploymentWrapper) Kind() string {
	return "Deployment"
}
func (d *DeploymentWrapper) Metadata() metav1.ObjectMeta {
	return d.Deployment.ObjectMeta
}
func (d *DeploymentWrapper) Labels() map[string]string {
	return d.Deployment.Labels
}
func (d *DeploymentWrapper) Name() string {
	return d.Deployment.Name
}
func (d *DeploymentWrapper) Namespace() string {
	return d.Deployment.Namespace
}

func (d *DeploymentWrapper) Annotations() map[string]string {
	return d.Deployment.Annotations
}

func (d *DeploymentWrapper) Ready() bool {
	return d.Status.ReadyReplicas == d.Status.Replicas
}
func (d *DeploymentWrapper) IsZeroReplicas() bool {
	return d.Status.Replicas == 0
}

func (d *DeploymentWrapper) DeployStatus() string {
	if d.Ready() {
		return v1alpha1.StatusDeployed
	}
	return v1alpha1.StatusPendingInstall
}
func (d *DeploymentWrapper) IsDelete() bool {
	return false
}

func (d *DeploymentWrapper) Containers() []corev1.Container {
	return d.Spec.Template.Spec.Containers
}

func (d *DeploymentWrapper) PodTemplate() *corev1.PodTemplateSpec {
	return &d.Spec.Template
}

type DaemonsetWrapper struct {
	*appsv1.DaemonSet
}

func (d *DaemonsetWrapper) ApiVersion() string {
	return "apps/v1"
}
func (d *DaemonsetWrapper) Kind() string {
	return "DaemonSet"
}
func (d *DaemonsetWrapper) Metadata() metav1.ObjectMeta {
	return d.DaemonSet.ObjectMeta
}
func (d *DaemonsetWrapper) Labels() map[string]string {
	return d.DaemonSet.Labels
}
func (d *DaemonsetWrapper) Name() string {
	return d.DaemonSet.Name
}
func (d *DaemonsetWrapper) Namespace() string {
	return d.DaemonSet.Namespace
}
func (d *DaemonsetWrapper) Annotations() map[string]string {
	return d.DaemonSet.Annotations
}
func (d *DaemonsetWrapper) Ready() bool {
	return d.Status.NumberReady == d.Status.NumberAvailable && d.Status.NumberReady > 0
}

func (d *DaemonsetWrapper) IsZeroReplicas() bool {
	return d.Status.DesiredNumberScheduled == 0
}

func (d *DaemonsetWrapper) Containers() []corev1.Container {
	return d.Spec.Template.Spec.Containers
}

func (d *DaemonsetWrapper) PodTemplate() *corev1.PodTemplateSpec {
	return &d.Spec.Template
}

func (d *DaemonsetWrapper) DeployStatus() string {
	if d.Ready() {
		return v1alpha1.StatusDeployed
	}
	return v1alpha1.StatusPendingInstall
}

type StatefulSetWrapper struct {
	*appsv1.StatefulSet
}

func (d *StatefulSetWrapper) ApiVersion() string {
	return "apps/v1"
}
func (d *StatefulSetWrapper) Kind() string {
	return "StatefulSet"
}
func (d *StatefulSetWrapper) Metadata() metav1.ObjectMeta {
	return d.StatefulSet.ObjectMeta
}
func (d *StatefulSetWrapper) Labels() map[string]string {
	return d.StatefulSet.Labels
}
func (d *StatefulSetWrapper) Name() string {
	return d.StatefulSet.Name
}
func (d *StatefulSetWrapper) Namespace() string {
	return d.StatefulSet.Namespace
}
func (d *StatefulSetWrapper) Annotations() map[string]string {
	return d.StatefulSet.Annotations
}

func (d *StatefulSetWrapper) Ready() bool {
	return d.Status.ReadyReplicas == d.Status.Replicas
}

func (d *StatefulSetWrapper) DeployStatus() string {
	if d.Ready() {
		return v1alpha1.StatusDeployed
	}
	return v1alpha1.StatusPendingInstall
}
func (d *StatefulSetWrapper) IsZeroReplicas() bool {
	return d.Status.Replicas == 0
}

func (d *StatefulSetWrapper) Containers() []corev1.Container {
	return d.Spec.Template.Spec.Containers
}

func (d *StatefulSetWrapper) PodTemplate() *corev1.PodTemplateSpec {
	return &d.Spec.Template
}

type JobWrapper struct {
	*batchv1.Job
}

func (d *JobWrapper) ApiVersion() string {
	return "batch/v1"
}
func (d *JobWrapper) Kind() string {
	return "Job"
}
func (d *JobWrapper) Metadata() metav1.ObjectMeta {
	return d.Job.ObjectMeta
}
func (d *JobWrapper) Labels() map[string]string {
	return d.Job.Labels
}
func (d *JobWrapper) Name() string {
	return d.Job.Name
}
func (d *JobWrapper) Namespace() string {
	return d.Job.Namespace
}
func (d *JobWrapper) Annotations() map[string]string {
	return d.Job.Annotations
}
func (d *JobWrapper) Ready() bool {
	return d.Status.Succeeded >= 1 //&& d.Status.Failed == 0
}
func (d *JobWrapper) IsZeroReplicas() bool {
	return false
}

func (d *JobWrapper) DeployStatus() string {
	if d.Ready() {
		return v1alpha1.StatusDeployed
	}

	if d.Job.Status.Failed >= *d.Job.Spec.BackoffLimit {
		return v1alpha1.StatusFailed
	}

	return v1alpha1.StatusPendingInstall
}
func (d *JobWrapper) Containers() []corev1.Container {
	return d.Spec.Template.Spec.Containers
}

func (d *JobWrapper) PodTemplate() *corev1.PodTemplateSpec {
	return &d.Spec.Template
}

func EnhancedDefaultControllerRateLimiter() workqueue.TypedRateLimiter[any] {
	return workqueue.NewTypedMaxOfRateLimiter[any](
		workqueue.NewTypedItemExponentialFailureRateLimiter[any](5*time.Millisecond, 1000*time.Second),
		// 100 qps, 1000 bucket size
		&workqueue.TypedBucketRateLimiter[any]{Limiter: rate.NewLimiter(rate.Limit(100), 1000)},
	)
}

type WorkloadMock struct {
	apiversion string
	kind       string
	typeMeta   metav1.TypeMeta
	metaData   metav1.ObjectMeta
}

func NewWorkloadMock(apiversion, kind, name, namespace, groupname string) WorkloadInterface {
	annotations := map[string]string{"w7.cc/release-name": groupname}
	labels := map[string]string{"w7.cc/release-name": groupname}
	return &WorkloadMock{
		apiversion: apiversion,
		kind:       kind,
		typeMeta: metav1.TypeMeta{
			APIVersion: apiversion,
			Kind:       kind,
		},
		metaData: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		// CreateTime:  obj.Metadata().CreationTimestamp,
	}
}

func (d *WorkloadMock) ApiVersion() string {
	return d.apiversion
}
func (d *WorkloadMock) Kind() string {
	return d.kind
}
func (d *WorkloadMock) Metadata() metav1.ObjectMeta {
	return d.metaData
}
func (d *WorkloadMock) Labels() map[string]string {
	return d.Metadata().Labels
}
func (d *WorkloadMock) Name() string {
	return d.Metadata().Name
}
func (d *WorkloadMock) Namespace() string {
	return d.Metadata().Namespace
}

func (d *WorkloadMock) Annotations() map[string]string {
	return d.Metadata().Annotations
}

func (d *WorkloadMock) Ready() bool {
	return false
}
func (d *WorkloadMock) IsZeroReplicas() bool {
	return false
}

func (d *WorkloadMock) DeployStatus() string {
	if d.Ready() {
		return v1alpha1.StatusDeployed
	}
	return v1alpha1.StatusPendingInstall
}

func (d *WorkloadMock) Containers() []corev1.Container {
	return []corev1.Container{}
}

func (d *WorkloadMock) PodTemplate() *corev1.PodTemplateSpec {
	return nil
}
