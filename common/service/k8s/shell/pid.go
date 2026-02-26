package shell

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/terminal"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func WebHookPid(pod *corev1.Pod) {
	time.AfterFunc(time.Second*5, func() {
		if len(pod.Status.ContainerStatuses) == 0 {
			return
		}
		if pod.Annotations == nil {
			pod.Annotations = map[string]string{}
		}
		//如果已经设置 直接返回
		containerId := pod.Status.ContainerStatuses[0].ContainerID
		annoContainerId, ok := pod.Annotations["w7.cc/container-id"]
		if ok && annoContainerId == containerId {
			return
		}
		err := LoadPid(pod)
		if err != nil {
			slog.Error("load pid error", "error", err)
		}
	})

}

func getAnnotationPodPid(pod *corev1.Pod) (int, error) {
	if len(pod.Status.ContainerStatuses) == 0 {
		return 0, errors.New("not found pod containerId")
	}
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	annoContainerId := pod.Annotations["w7.cc/container-id"]
	if annoContainerId == "" {
		return 0, errors.New("not found pod containerId")
	}
	containerId := pod.Status.ContainerStatuses[0].ContainerID
	if containerId != annoContainerId {
		return 0, errors.New("containerId is not equal")
	}

	pid, ok := pod.Annotations["w7.cc/pid"]
	if ok {
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			return 0, err
		}
		return pidInt, nil
	}
	return 0, errors.New("error found annotion pid")
}
func LoadPid(pod *corev1.Pod) error {
	//如果是主集群 转发请求到agent节点获取pid
	sdk := k8s.NewK8sClient()
	sigClient, err := sdk.ToSigClient()
	if err != nil {
		return err
	}

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("not found pod containerId")
	}
	if pod.Status.ContainerStatuses[0].Ready == false {
		return fmt.Errorf("cluster pod is not running")
	}
	containerId := pod.Status.ContainerStatuses[0].ContainerID
	if helper.IsChildAgent() {
		if helper.IsK3kVirtual() {
			//os 执行命令
			cmd := []string{"inspect", "--output", "go-template", fmt.Sprintf("--template='{{.info.pid}}'"), containerId}
			output, err := exec.Command("crictl", cmd...).Output()
			if err != nil {
				slog.Error("run cmd err", "err", err)
				return err
			}
			pid, err := bytesToPid(output)
			if err != nil {
				slog.Error("bytesToPid", "err", err)
				return err
			}

			controllerutil.CreateOrPatch(sdk.Ctx, sigClient, pod, func() error {
				if pod.Annotations == nil {
					pod.Annotations = make(map[string]string)
				}
				pod.Annotations["w7.cc/pid"] = strconv.Itoa(pid)
				pod.Annotations["w7.cc/container-id"] = containerId
				return nil
			})
			return nil
		}
		if helper.IsK3kShared() {
			// 使用的主集群pod 不需要处理
		}
	} else {
		daemonsetPod, err := sdk.GetDaemonsetAgentPod(sdk.GetNamespace(), pod.Status.HostIP)
		if err != nil {
			slog.Error("get  daemonsetPod err", "err", err)
			return err
		}
		if (len(pod.Status.ContainerStatuses) > 0) && containerId != "" {
			pid, err := GetPid(daemonsetPod, containerId, true, sdk.Sdk)
			if err != nil {
				return err
			}
			controllerutil.CreateOrPatch(sdk.Ctx, sigClient, pod, func() error {
				if pod.Annotations == nil {
					pod.Annotations = make(map[string]string)
				}
				pod.Annotations["w7.cc/pid"] = strconv.Itoa(pid)
				pod.Annotations["w7.cc/container-id"] = containerId
				return nil
			})
		}
	}
	return nil
	//如果是子集群 直接通过当前shell获取
}

func bytesToPid(data []byte) (int, error) {
	pid := string(data)
	pid = strings.Replace(pid, "\n", "", -1)
	pid = strings.Replace(pid, "'", "", -1)
	//pid string to int
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return 0, err
	}
	return pidInt, nil
}

type PidParam struct {
	Namespace            string `form:"namespace" binding:"required"`
	HostIp               string `form:"HostIp" binding:"required"`
	ContainerId          string `form:"containerId"`
	FromPodName          string `form:"podName"`       //原始pod名
	FromPodContainerName string `form:"containerName"` //原始pod container名
}

type Pid struct {
	token             string
	rootSdk           *k8s.Sdk
	clientSdk         *k8s.Sdk
	k8stoken          *k8s.K8sToken
	domonsetAgentPods *corev1.PodList
}

func NewPid(token string) (*Pid, error) {
	clientsdk, err := k8s.NewK8sClient().Channel(token)
	if err != nil {
		return nil, err
	}
	rootsdk := k8s.NewK8sClient().Sdk
	daemonsetPods, err := rootsdk.GetDaemonsetAgentPods(rootsdk.GetNamespace())
	if err != nil {
		return nil, err
	}
	return &Pid{
		k8stoken:          k8s.NewK8sToken(token),
		rootSdk:           rootsdk,
		clientSdk:         clientsdk,
		token:             token,
		domonsetAgentPods: daemonsetPods,
	}, nil
}

func (self *Pid) IsVirtual() bool {
	return self.k8stoken.IsVirtual()
}

func (self *Pid) IsShared() bool {
	return self.k8stoken.IsShared()
}

func (self *Pid) IsRoot() bool {
	k8stoken := k8s.NewK8sToken(self.token)
	return !k8stoken.IsK3kCluster()
}

func (self *Pid) GetPanelAgentPod(hostIp string) (*corev1.Pod, error) {
	for _, pod := range self.domonsetAgentPods.Items {
		if pod.Status.HostIP == hostIp {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("not found pod")
}

func (self *Pid) GetFromPod(fromPodName, namespace string) (*corev1.Pod, error) {
	if fromPodName == "" {
		return nil, fmt.Errorf("fromPodName is empty")
	}
	if namespace == "" {
		namespace = "default"
	}
	if self.IsVirtual() {
		return self.clientSdk.ClientSet.CoreV1().Pods(namespace).Get(self.clientSdk.Ctx, fromPodName, metav1.GetOptions{})
	}
	return self.rootSdk.ClientSet.CoreV1().Pods(namespace).Get(self.clientSdk.Ctx, fromPodName, metav1.GetOptions{})
}

// k3k-console-server-1 pod
func (self *Pid) GetVirtualClusterNodePod(hostIp string) (*corev1.Pod, error) {
	nodes, err := self.clientSdk.ClientSet.CoreV1().Nodes().List(self.clientSdk.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	podName := ""
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Address == hostIp {
				podName = node.Name
				break
			}
		}
	}
	if podName == "" {
		return nil, fmt.Errorf("not found node")
	}
	ns := self.k8stoken.GetNamespace()
	pod, err := self.rootSdk.ClientSet.CoreV1().Pods(ns).Get(self.clientSdk.Ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func (self *Pid) GetVirtualClusterNodePodByName(podName string) (*corev1.Pod, error) {
	ns := self.k8stoken.GetNamespace()
	pod, err := self.rootSdk.ClientSet.CoreV1().Pods(ns).Get(self.clientSdk.Ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

func (self *Pid) GetPidFromPod(pod *corev1.Pod, containerId string) (int, error) {
	if self.IsVirtual() {
		return self.GetContainerPid(pod, containerId, false)
	}
	return self.GetContainerPid(pod, containerId, true)
}

func (self *Pid) GetContainerPid(findPod *corev1.Pod, containerId string, nscener bool) (int, error) {

	return GetPid(findPod, containerId, nscener, self.rootSdk)
}

func (self *Pid) GetContainerPid2(findPod *corev1.Pod, fromPod *corev1.Pod, containerId string, nscener bool) (int, error) {
	if fromPod == nil {
		return GetPid(findPod, containerId, nscener, self.rootSdk)
	}
	pid, err := getAnnotationPodPid(fromPod)
	if err != nil {
		slog.Error("getAnnotationPodPid", "err", err)
	}
	if err == nil && pid != 0 {
		return pid, nil
	}
	return GetPid(findPod, containerId, nscener, self.rootSdk)
}

func GetPid(findPod *corev1.Pod, containerId string, nscener bool, sdk *k8s.Sdk) (int, error) {
	session := terminal.NewTerminalSession(nil)
	defer session.Close()
	containerName := findPod.Spec.Containers[0].Name

	if strings.HasPrefix(containerId, "containerd://") {
		containerId = containerId[len("containerd://"):]
	}
	cmd := []string{"nsenter", "-t", "1", "--mount", "--pid", "--", "crictl", "inspect", "--output", "go-template", fmt.Sprintf("--template='{{.info.pid}}'"), containerId}
	if !nscener {
		cmd = []string{"crictl", "inspect", "--output", "go-template", fmt.Sprintf("--template='{{.info.pid}}'"), containerId}
	}

	err := sdk.RunExec(session, findPod.Namespace, findPod.Name, containerName, cmd, false)
	if err != nil {
		return 0, err
	}
	pid := string(session.GetWriterBytes())
	pid = strings.Replace(pid, "\n", "", -1)
	pid = strings.Replace(pid, "'", "", -1)
	//pid string to int
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return 0, err
	}
	return pidInt, nil
}

func (self *Pid) GetPwd(params PidParam) (string, error) {
	session := terminal.NewTerminalSession(nil)
	defer session.Close()
	sdk := self.rootSdk
	if !self.IsRoot() {
		sdk = self.clientSdk
	}
	err := sdk.RunExec(session, params.Namespace, params.FromPodName, params.FromPodContainerName, []string{"pwd"}, false)
	if err != nil {
		return "", err
	}
	return string(session.GetWriterBytes()), nil
}

// params := ParamsValidate{}
// if !self.Validate(http, &params) {
// 	return
// }
// token := http.MustGet("k8s_token").(string)
// // k3sToken := k8s.NewK8sToken(token)
// rootSdk := k8s.NewK8sClient()
// // client, err := rootSdk.Channel(token)
// // if err != nil {
// // 	self.JsonResponseWithServerError(http, err)
// // 	return
// // }
// pod, err := rootSdk.GetDaemonsetAgentPod(params.Namespace, params.HostIp)
// // pod, err := client.GetDaemonsetAgentPod("default", params.HostIp)
// if err != nil {
// 	self.JsonResponseWithServerError(http, err)
// 	return
// }
// // pid, err := client.GetContainerPid(pod, params.ContainerId)
// pid, err := rootSdk.GetContainerPid(pod, params.ContainerId)
// if err != nil {
// 	self.JsonResponseWithServerError(http, err)
// 	return
// }

// cmd := []string{"pwd"}
// // if params.Pid != "" && len(cmd) > 0 && cmd[0] == "ls" {
// // 	cmd = []string{"/bin/sh", "-c", lsProxy(params.Pid)}
// // }
// session := terminal.NewTerminalSession(nil)
// defer session.Close()

// pwd := "/"
// if params.FromPodName != "" && params.FromPodContainerName != "" {
// 	// err = client.RunExec(session, params.Namespace, params.FromPodName, params.FromPodContainerName, cmd, false)
// 	err = rootSdk.RunExec(session, params.Namespace, params.FromPodName, params.FromPodContainerName, cmd, false)
// 	if err == nil {
// 		pwd = string(session.GetWriterBytes())
// 	}
// }
// ///v1/:name/proxy/*path
// podIp := pod.Status.PodIP
// pidstr := strconv.Itoa(pid)
// webdavUrl := "/k8s/v1/" + podIp + ":8000/proxy/k8s/webdav-agent/" + pidstr + "/agent"
// self.JsonResponse(http, gin.H{
// 	"podName":        pod.Name,
// 	"pid":            pid,
// 	"namespace":      pod.Namespace,
// 	"containerName":  pod.Spec.Containers[0].Name,
// 	"podIp":          pod.Status.PodIP,
// 	"pwd":            pwd,
// 	"webdavUrl":      webdavUrl,
// 	"webdavToken":    token,
// 	"webdavFullUrl":  "?api-url=" + webdavUrl + "&api-token=" + token,
// 	"webdavBasePath": "k8s/webdav-agent/" + pidstr + "/agent",
// }, nil, 200)
