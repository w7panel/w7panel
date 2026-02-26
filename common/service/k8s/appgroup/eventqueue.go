package appgroup

import (
	"fmt"
	"log/slog"
	"time"

	"gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	// maxRetries is the number of times a deployment will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms
	maxRetries = 3
)

type EventQueue struct {
	queue  workqueue.TypedRateLimitingInterface[any]
	handle func(key interface{}) error
}

func NewDefaultEventQueue(handle func(key interface{}) error) *EventQueue {
	name := "groupcontroller"
	nameConfig := workqueue.TypedRateLimitingQueueConfig[any]{Name: name}
	queue := workqueue.NewTypedRateLimitingQueueWithConfig[any](EnhancedDefaultControllerRateLimiter(), nameConfig)
	return NewEventQueue(queue, handle)
}

func NewEventQueue(queue workqueue.TypedRateLimitingInterface[any], handle func(key interface{}) error) *EventQueue {
	return &EventQueue{
		queue:  queue,
		handle: handle,
	}
}

func (c *EventQueue) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	slog.Error("Starting groupcontroller queue")
	defer slog.Error("Shut down groupcontroller queue")

	// if !cache.WaitForNamedCacheSync("longhorn engines", stopCh, c.cacheSyncs...) {
	// 	return
	// }

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
}

func (c *EventQueue) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *EventQueue) processNextWorkItem() bool {
	// slog.Debug("Working next item from groupcontroller queue")
	key, quit := c.queue.Get()

	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.handle(key)
	c.handleErr(err, key)

	return true
}

func (c *EventQueue) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetries {
		// slog.Info("Failed to sync groupcontroller event, requeuing")
		c.queue.AddRateLimited(key)
		return
	}

	utilruntime.HandleError(err)
	// slog.Error("Dropping groupcontroller event out of the queue")
	c.queue.Forget(key)
}

func (c *EventQueue) OnAdd(obj interface{}, isInit bool) {
	c.OnEvent(obj, "add", isInit)
}

func (c *EventQueue) OnUpdate(obj interface{}, new interface{}) {
	c.OnEvent(new, "update", false)
}

func (c *EventQueue) OnDelete(obj interface{}) {
	if deleted, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		// Extract the actual object from the DeletedFinalStateUnknown struct
		obj = deleted.Obj
	}
	c.OnEvent(obj, "delete", false)

}

func (c *EventQueue) Push(key string) {
	c.queue.Add(key)
}

func (c *EventQueue) OnEvent(obj interface{}, eventType string, isInit bool) {
	obj2 := obj.(runtime.Object)
	typeMeta := metav1.TypeMeta{
		Kind:       obj2.GetObjectKind().GroupVersionKind().Kind,
		APIVersion: obj2.GetObjectKind().GroupVersionKind().GroupVersion().String(),
	}
	switch v := obj.(type) {
	case *appsv1.Deployment:
		typeMeta = metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		}
		c.queue.Add(NewK8sResourceEvent(typeMeta, v.ObjectMeta, eventType, isInit).ToString())
	case *appsv1.StatefulSet:
		typeMeta = metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		}
		c.queue.Add(NewK8sResourceEvent(typeMeta, v.ObjectMeta, eventType, isInit).ToString())
	case *appsv1.DaemonSet:
		typeMeta = metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		}
		c.queue.Add(NewK8sResourceEvent(typeMeta, v.ObjectMeta, eventType, isInit).ToString())
	case *batchv1.Job:
		typeMeta = metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		}
		c.queue.Add(NewK8sResourceEvent(typeMeta, v.ObjectMeta, eventType, isInit).ToString())
	case *corev1.Secret:
		typeMeta = metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		}
		c.queue.Add(NewK8sResourceEvent(typeMeta, v.ObjectMeta, eventType, isInit).ToString())
	case *v1alpha1.AppGroup:
		typeMeta = metav1.TypeMeta{
			Kind:       "AppGroup",
			APIVersion: "appgroup.w7.cc/v1alpha1",
		}
		c.queue.Add(NewK8sResourceEvent(typeMeta, v.ObjectMeta, eventType, isInit).ToString())
	case *corev1.Event:
		typeMeta = metav1.TypeMeta{
			Kind:       "Event",
			APIVersion: "v1",
		}
		c.queue.Add(NewK8sResourceEvent(typeMeta, v.ObjectMeta, eventType, isInit).ToString())
	default:
		return
	}
}
func (c *EventQueue) AddEvent(event *K8sResourceEvent) {
	c.queue.Add(event.ToString())

}
func (c *EventQueue) AddAfter(event *K8sResourceEvent, duration time.Duration) {
	c.queue.AddAfter(event.ToString(), duration)
}

func ParseEvent(obj interface{}) (*K8sResourceEvent, error) {
	keystr, ok := obj.(string)
	if !ok {
		return nil, fmt.Errorf("key is not a string")
	}
	event, err := NewK8sResourceEventFromJson([]byte(keystr))
	return event, err

}

func ExtractMetadata(obj runtime.Object) (metav1.TypeMeta, metav1.ObjectMeta, error) {
	// 通过反射获取 TypeMeta
	typeMeta := metav1.TypeMeta{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.(runtime.Unstructured).UnstructuredContent(), &typeMeta); err != nil {
		return metav1.TypeMeta{}, metav1.ObjectMeta{}, fmt.Errorf("failed to extract TypeMeta: %v", err)
	}

	// 通过反射获取 ObjectMeta
	objectMeta := metav1.ObjectMeta{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.(runtime.Unstructured).UnstructuredContent(), &objectMeta); err != nil {
		return metav1.TypeMeta{}, metav1.ObjectMeta{}, fmt.Errorf("failed to extract ObjectMeta: %v", err)
	}

	return typeMeta, objectMeta, nil
}
