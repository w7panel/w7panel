package pid

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/terminal"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// webhook 入口获取pid
func LoadPid(pod *corev1.Pod) (int, error) {
	//如果是主集群 转发请求到agent节点获取pid
	sdk := k8s.NewK8sClient()
	sigClient, err := sdk.ToSigClient()
	if err != nil {
		return 0, err
	}

	if len(pod.Status.ContainerStatuses) == 0 {
		return 0, fmt.Errorf("not found pod containerId")
	}
	if pod.Status.ContainerStatuses[0].Ready == false {
		return 0, fmt.Errorf("cluster pod is not running")
	}
	containerId := pod.Status.ContainerStatuses[0].ContainerID
	if helper.IsChildAgent() {
		if helper.IsK3kVirtual() {
			//os 执行命令
			cmd := []string{"inspect", "--output", "go-template", fmt.Sprintf("--template='{{.info.pid}}'"), containerId}
			output, err := exec.Command("crictl", cmd...).Output()
			if err != nil {
				slog.Error("run cmd err", "err", err)
				return 0, err
			}
			pid, err := bytesToPid(output)
			if err != nil {
				slog.Error("bytesToPid", "err", err)
				return 0, err
			}

			controllerutil.CreateOrPatch(sdk.Ctx, sigClient, pod, func() error {
				if pod.Annotations == nil {
					pod.Annotations = make(map[string]string)
				}
				pod.Annotations["w7.cc/pid"] = strconv.Itoa(pid)
				pod.Annotations["w7.cc/container-id"] = containerId
				return nil
			})
			return pid, nil
		}
		if helper.IsK3kShared() {
			// 使用的主集群pod 不需要处理
		}
	} else {
		daemonsetPod, err := sdk.GetDaemonsetAgentPod(sdk.GetNamespace(), pod.Status.HostIP)
		if err != nil {
			slog.Error("get  daemonsetPod err", "err", err)
			return 0, err
		}
		if (len(pod.Status.ContainerStatuses) > 0) && containerId != "" {
			pid, err := GetPid(daemonsetPod, containerId, true, sdk.Sdk)
			if err != nil {
				return 0, err
			}
			controllerutil.CreateOrPatch(sdk.Ctx, sigClient, pod, func() error {
				if pod.Annotations == nil {
					pod.Annotations = make(map[string]string)
				}
				pod.Annotations["w7.cc/pid"] = strconv.Itoa(pid)
				pod.Annotations["w7.cc/container-id"] = containerId
				return nil
			})
			return pid, nil
		}
	}
	return 0, errors.New("not found pid")
	//如果是子集群 直接通过当前shell获取
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

func GetContainerPid(agentPod *corev1.Pod, pod *corev1.Pod, containerId string, nscener bool, sdk *k8s.Sdk) (int, error) {

	pid, err := getAnnotationPodPid(pod)
	if err != nil {
		slog.Error("getAnnotationPodPid", "err", err)
	}
	if err == nil && pid != 0 {
		return pid, nil
	}
	return GetPid(agentPod, containerId, nscener, sdk)
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

func checkPodRunning(pod *corev1.Pod) error {
	if len(pod.Status.ContainerStatuses) > 0 {
		if (pod.Status.ContainerStatuses[0].State.Running == nil) && (pod.Status.ContainerStatuses[0].State.Terminated != nil) {
			return errors.New("cluster pod is not running")
		}
	}
	return nil
}
