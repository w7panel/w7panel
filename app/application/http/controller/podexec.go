package controller

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/remotecommand"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/shell"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/terminal"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	remotecommand2 "k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/cmd/cp"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const LS = `ls -l -AF /proc/76008/root/w7panel | awk -v passwd="/proc/76008/root/etc/passwd" -v group="/proc/76008/root/etc/group" '
BEGIN {
    while ((getline < passwd) > 0) {
        split($0, fields, ":");
        uid_to_user[fields[3]] = fields[1];
    }
    close(passwd);
    while ((getline < group) > 0) {
        split($0, fields, ":");
        gid_to_group[fields[3]] = fields[1];
    }
    close(group);
}
{
    uid = $3;
    gid = $4;
    user = (uid in uid_to_user)? uid_to_user[uid] : uid;
    group = (gid in gid_to_group)? gid_to_group[gid] : gid;
    $3 = user;
    $4 = group;
    print;
}'`

func lsProxy(pid string) string {
	result := strings.ReplaceAll(LS, "76008", pid)
	result = strings.ReplaceAll(result, "ls -l -AF /proc/76008/root/", "")
	return result
}

type PodExec struct {
	controller.Abstract
}

func (self PodExec) Exec(http *gin.Context) {
	type ParamsValidate struct {
		Namespace     string   `form:"namespace" binding:"required"`
		PodName       string   `form:"podName" binding:"required"`
		ContainerName string   `form:"containerName" binding:"required"`
		Command       []string `form:"command" binding:"required"`
		Tty           bool     `form:"tty"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	r := http.Request
	w := http.Writer

	var conn *websocket.Conn
	var err error
	if websocket.IsWebSocketUpgrade(r) {
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			self.JsonResponseWithServerError(http, err)
			return
		}
	}

	session := terminal.NewTerminalSession(conn)
	defer session.Close()

	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	rootsdk := k8s.NewK8sClient().Sdk
	if strings.Contains(params.PodName, "w7panel-agent") {
		client = rootsdk
	}
	cmd := params.Command
	// if params.Pid != "" && len(cmd) > 0 && cmd[0] == "ls" {
	// 	cmd = []string{"/bin/sh", "-c", lsProxy(params.Pid)}
	// }
	err = client.RunExec(session, params.Namespace, params.PodName, params.ContainerName, cmd, params.Tty)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	if conn == nil {
		http.Writer.Write(session.GetWriterBytes())
	}
}

func (p PodExec) NodeTty(http *gin.Context) {
	type ParamsValidate struct {
		Shell  string `form:"shell,default=/bin/bash" binding:"oneof=/bin/sh /bin/bash"`
		HostIp string `form:"hostIp" binding:"required"`
	}
	params := ParamsValidate{}
	if !p.Validate(http, &params) {
		return
	}

	conn, err := upgrader.Upgrade(http.Writer, http.Request, nil)
	if err != nil {
		p.JsonResponseWithServerError(http, err)
		return
	}

	token := http.MustGet("k8s_token").(string)
	k8sToken := k8s.NewK8sToken(token)
	session := terminal.NewTerminalSession(conn)
	defer session.Close()
	rootsdk := k8s.NewK8sClient().Sdk
	var findPod *corev1.Pod
	shells := []string{params.Shell}
	if k8sToken.IsK3kCluster() {
		// client, err := k8s.NewK8sClient().ChannelLocal(http.MustGet("k8s_token").(string), true)
		k3kConfig, err := k8sToken.GetK3kConfig()
		if err != nil {
			p.JsonResponseWithServerError(http, err)
			return
		}
		pods, err := rootsdk.ClientSet.CoreV1().Pods(k3kConfig.Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "cluster"})
		if err != nil {
			p.JsonResponseWithServerError(http, err)
			return
		}
		for _, pod := range pods.Items {
			if pod.Status.PodIP == params.HostIp {
				findPod = &pod
			}
		}
	} else {
		findPod, err = rootsdk.GetDaemonsetAgentPod(rootsdk.GetNamespace(), params.HostIp)
		if err != nil {
			p.JsonResponseWithServerError(http, err)
			return
		}
		shells = []string{"nsenter", "-t", "1", "--mount", "--uts", "--ipc", "--net", "--pid", "--", params.Shell}
	}
	err = rootsdk.RunExec(session, findPod.Namespace, findPod.Name, findPod.Spec.Containers[0].Name, shells, true)
	if err != nil {
		p.JsonResponseWithServerError(http, err)
		return
	}
}

func (self PodExec) Tty(http *gin.Context) {
	type ParamsValidate struct {
		Shell string `form:"shell,default=/bin/bash" binding:"oneof=/bin/sh /bin/bash"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	ttyChrootDir := facade.GetConfig().GetString("k8s.tty_chroot_dir")
	cmd := exec.Command(params.Shell)
	// 获取当前进程的所有环境变量
	cmd.Env = os.Environ()
	// 设置新的环境变量
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")

	if ttyChrootDir != "" {
		// cmd.Dir = "ttytmp" //*ttyChrootDir 不能用 /
		// cmd.Env = []string{
		// 	"TERM=xterm-256color",
		// 	"KUBERNETES_TOKEN=" + config.BearerToken,
		// 	"KUBERNETES_SERVICE_HOST=" + os.Getenv("KUBERNETES_SERVICE_HOST"),
		// 	"KUBERNETES_SERVICE_PORT=" + os.Getenv("KUBERNETES_SERVICE_PORT"),
		// 	"KUBERNETES_CAFILE=" + "/.kube/ca.crt",
		// 	"HOME=" + os.Getenv("HOME"),
		// }
	}
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Chroot: ttyChrootDir,
	// }

	conn, err := upgrader.Upgrade(http.Writer, http.Request, nil)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	session := terminal.NewTerminalSession(conn)
	defer session.Close()
	token := http.MustGet("k8s_token").(string)
	k8sToken := k8s.NewK8sToken(token)
	if k8sToken.IsK3kCluster() {
		// client, err := k8s.NewK8sClient().ChannelLocal(http.MustGet("k8s_token").(string), true)
		client := k8s.NewK8sClient()

		k3kConfig, err := k8sToken.GetK3kConfig()
		if err != nil {
			self.JsonResponseWithServerError(http, err)
			return
		}
		// clientsdk, err := client.Channel(token)
		// if err != nil {
		// 	self.JsonResponseWithServerError(http, err)
		// 	return
		// }
		params.Shell = "/bin/sh" //k3k pod 只支持 /bin/sh
		err = client.RunExec(session, k3kConfig.Namespace, k3kConfig.GetK3kServer0Name(), k3kConfig.GetK3kServer0ContainerName(), []string{params.Shell}, true)
		if err != nil {
			self.JsonResponseWithServerError(http, err)
			return
		}
	} else {
		err = remotecommand.NewLocalExecutor(cmd).StreamWithContext(session.Context(), remotecommand2.StreamOptions{
			Stdin:             session,
			Stdout:            session,
			Stderr:            session,
			Tty:               true,
			TerminalSizeQueue: session,
		})
		if err != nil {
			slog.Error("tty error", "err", err)
			return
		}
	}

}

// no test pass
func (self PodExec) KubectlCp(http *gin.Context) {
	baseDir := facade.Config.GetString("s3.base_dir")
	// token := http.MustGet("k8s_token").(string)
	type ParamsValidate struct {
		From      string `form:"from"      binding:"required"`
		To        string `form:"to"        binding:"required"`
		Namespace string `form:"namespace" binding:"required"`
		Upload    string `form:"upload"  binding:"required"`
		Podname   string `form:"podName" binding:"required"`
	}

	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}

	if params.Upload == "1" {
		params.From = filepath.Join(baseDir, params.From)
		params.To = params.Podname + ":" + params.To
	} else {
		params.To = filepath.Join(baseDir, params.To)
		params.From = params.Podname + ":" + params.From
	}

	rootSdk := k8s.NewK8sClient().Sdk
	token := http.MustGet("k8s_token").(string)
	client, err := k8s.NewK8sClient().Channel(http.MustGet("k8s_token").(string))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	k8stoken := k8s.NewK8sToken(token)
	if k8stoken.IsK3kCluster() {
		client = rootSdk
	}
	cmdutil.BehaviorOnFatal(func(errstr string, code int) {

		self.JsonResponseWithServerError(http, fmt.Errorf("%v", errstr))

		// if err := recover(); err != nil {
		// 	self.JsonResponseWithServerError(http, fmt.Errorf("%v", err))
		// }
		http.Abort()
	})
	factory := cmdutil.NewFactory(client.PodExecClient())
	cmd := cp.NewCmdCp(factory, genericiooptions.NewTestIOStreamsDiscard())
	cmd.SetArgs([]string{params.From, params.To})
	err = cmd.Execute()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonSuccessResponse(http)

}

// 根据pod获取pid
func (self PodExec) GetAgentPodAndPid(http *gin.Context) {

	params := shell.PidParam{}
	if !self.Validate(http, &params) {
		return
	}
	token := http.MustGet("k8s_token").(string)
	pidobj, err := shell.NewPid(token)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	pid := 1
	subPid := 0
	var pod *corev1.Pod
	fromPod, fromPodErr := pidobj.GetFromPod(params.FromPodName, params.Namespace)
	slog.Info("GetFromPod result", "fromPodName", params.FromPodName, "namespace", params.Namespace, "fromPodErr", fromPodErr, "fromPodIsNil", fromPod == nil)
	if fromPodErr != nil {
		slog.Info("fromPod error, setting to nil", "error", fromPodErr)
		fromPod = nil // 确保错误时 fromPod 为 nil
	}
	if fromPod != nil {
		slog.Info("fromPod details", "name", fromPod.Name, "namespace", fromPod.Namespace)
	}

	if pidobj.IsVirtual() {

		clusterPod, err := pidobj.GetVirtualClusterNodePod(params.HostIp)
		if err != nil {
			self.JsonResponseWithServerError(http, err)
			return
		}
		if len(clusterPod.Status.ContainerStatuses) > 0 {
			if (clusterPod.Status.ContainerStatuses[0].State.Running == nil) && (clusterPod.Status.ContainerStatuses[0].State.Terminated != nil) {
				self.JsonResponseWithServerError(http, fmt.Errorf("cluster pod is not running"))
				return
			}
		}
		pod, err = pidobj.GetPanelAgentPod(clusterPod.Status.HostIP)
		if err != nil {
			self.JsonResponseWithServerError(http, err)
			return
		}

		pid, err = pidobj.GetContainerPid2(pod, clusterPod, clusterPod.Status.ContainerStatuses[0].ContainerID, true)
		if err != nil {
			self.JsonResponseWithServerError(http, err)
			return
		}
		if params.ContainerId != "" {
			subPid, err = pidobj.GetContainerPid2(clusterPod, fromPod, params.ContainerId, false)
			if err != nil {
				self.JsonResponseWithServerError(http, err)
				return
			}
		}

	} else {
		pod, err = pidobj.GetPanelAgentPod(params.HostIp)
		if err != nil {
			self.JsonResponseWithServerError(http, err)
			return
		}
		if params.ContainerId != "" {
			pid, err = pidobj.GetContainerPid2(pod, fromPod, params.ContainerId, true)
			if err != nil {
				self.JsonResponseWithServerError(http, err)
				return
			}
		}

	}
	pwd := "/"
	if params.FromPodName != "" && params.FromPodContainerName != "" {
		pwd, err = pidobj.GetPwd(params)
		if err != nil {
			pwd = "/"
		}
	}
	///v1/:name/proxy/*path
	podIp := pod.Status.PodIP
	pidstr := strconv.Itoa(pid)

	// debug/local_mock 模式: 使用当前开发环境替代 agent 服务
	// 生产环境: webdavUrl 通过 {podIp}:8000 代理访问独立的 agent pod
	// 开发环境: 当前服务(8080端口)直接处理 /k8s/webdav-agent 路由
	localMock := facade.Config.GetBool("app.local_mock")

	var webdavUrl, webdavBasePath, compressUrl, permissionUrl string
	if localMock {
		// 本地模拟模式: 直接使用本地路由，不通过 agent 代理
		webdavUrl = "/k8s/webdav-agent/" + pidstr + "/agent"
		webdavBasePath = "k8s/webdav-agent/" + pidstr + "/agent"
		compressUrl = "/k8s/compress-agent/" + pidstr
		permissionUrl = "/k8s/permission-agent/" + pidstr
		if subPid > 0 {
			subpidstr := strconv.Itoa(subPid)
			webdavUrl = "/k8s/webdav-agent/" + pidstr + "/subagent/" + subpidstr + "/agent"
			webdavBasePath = "k8s/webdav-agent/" + pidstr + "/subagent/" + subpidstr + "/agent"
			compressUrl = "/k8s/compress-agent/" + pidstr + "/subagent/" + subpidstr
			permissionUrl = "/k8s/permission-agent/" + pidstr + "/subagent/" + subpidstr
		}
		slog.Info("local_mock mode enabled", "webdavUrl", webdavUrl, "compressUrl", compressUrl, "permissionUrl", permissionUrl)
	} else {
		// 生产模式: 通过 agent pod 代理访问
		webdavUrl = "/k8s/v1/" + podIp + ":8000/proxy/k8s/webdav-agent/" + pidstr + "/agent"
		webdavBasePath = "k8s/webdav-agent/" + pidstr + "/agent"
		compressUrl = "/k8s/v1/" + podIp + ":8000/proxy/k8s/compress-agent/" + pidstr
		permissionUrl = "/k8s/v1/" + podIp + ":8000/proxy/k8s/permission-agent/" + pidstr
		if subPid > 0 {
			subpidstr := strconv.Itoa(subPid)
			webdavUrl = "/k8s/v1/" + podIp + ":8000/proxy/k8s/webdav-agent/" + pidstr + "/subagent/" + subpidstr + "/agent"
			webdavBasePath = "k8s/webdav-agent/" + pidstr + "/subagent/" + subpidstr + "/agent"
			compressUrl = "/k8s/v1/" + podIp + ":8000/proxy/k8s/compress-agent/" + pidstr + "/subagent/" + subpidstr
			permissionUrl = "/k8s/v1/" + podIp + ":8000/proxy/k8s/permission-agent/" + pidstr + "/subagent/" + subpidstr
		}
	}

	// 同步获取用户列表（确保首次请求有值，避免前端权限设置页面错误）
	users := self.getUsersWithCache(fromPod, pod, params.Namespace, params.FromPodName, params.FromPodContainerName)

	self.JsonResponse(http, gin.H{
		"podName":        pod.Name,
		"pid":            pid,
		"subPid":         subPid,
		"namespace":      pod.Namespace,
		"containerName":  pod.Spec.Containers[0].Name,
		"podIp":          podIp,
		"pwd":            pwd,
		"webdavUrl":      webdavUrl,
		"webdavToken":    token,
		"webdavBasePath": webdavBasePath,
		"compressUrl":    compressUrl,
		"permissionUrl":  permissionUrl,
		"users":          users,
	}, nil, 200)
}

// getUsersWithCache 获取用户列表（带缓存）
// 参考 pid 的缓存方案，使用 Pod Annotation 存储
// fromPod: 应用 Pod（可能为 nil），pod: Agent Pod
// fromPodName/fromContainerName: 用于重新获取 fromPod
func (self PodExec) getUsersWithCache(fromPod *corev1.Pod, pod *corev1.Pod, namespace string, fromPodName string, fromContainerName string) []map[string]interface{} {
	sdk := k8s.NewK8sClient().Sdk
	var targetPod *corev1.Pod
	var containerId string

	// 确定目标 Pod
	if fromPod != nil && fromPod.Name != "" {
		targetPod = fromPod
		slog.Info("Using fromPod", "podName", fromPod.Name, "containerName", fromContainerName)
	} else if fromPodName != "" {
		// fromPod 获取失败，重新尝试获取
		var err error
		targetPod, err = sdk.ClientSet.CoreV1().Pods(namespace).Get(context.TODO(), fromPodName, metav1.GetOptions{})
		if err != nil {
			slog.Error("Failed to get fromPod", "namespace", namespace, "podName", fromPodName, "error", err)
			// 获取失败，从 Agent Pod 本地读取
			return self.getLocalUsers()
		}
		slog.Info("Got fromPod by name", "podName", targetPod.Name)
	} else {
		slog.Info("No fromPod info, using agentPod")
		// 没有应用 Pod 信息，从 Agent Pod 本地读取
		return self.getLocalUsers()
	}

	if len(targetPod.Status.ContainerStatuses) == 0 {
		slog.Warn("No container statuses", "podName", targetPod.Name)
		return []map[string]interface{}{}
	}
	containerId = targetPod.Status.ContainerStatuses[0].ContainerID
	slog.Info("Target pod containerId", "containerId", containerId)

	// 尝试从缓存读取
	users, err := self.getUsersFromCache(targetPod, containerId)
	if err == nil && len(users) > 0 {
		slog.Info("Found users in cache", "count", len(users))
		return users
	}

	slog.Info("Cache miss or empty, reading from container")

	// 缓存未命中，从容器读取
	users = self.getUsersFromContainer(targetPod, namespace, fromContainerName)
	slog.Info("Users from container", "count", len(users))

	// 保存到缓存
	if len(users) > 0 {
		slog.Info("Saving users to cache")
		_ = self.saveUsersToCache(targetPod, containerId, users)
	}

	return users
}

// getUsersFromContainer 从容器读取用户列表
func (self PodExec) getUsersFromContainer(pod *corev1.Pod, namespace string, containerName string) []map[string]interface{} {
	sdk := k8s.NewK8sClient().Sdk
	if containerName == "" && len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	req := sdk.ClientSet.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "false").
		Param("command", "cat").
		Param("command", "/etc/passwd")

	restConfig, err := sdk.ToRESTConfig()
	if err != nil {
		slog.Error("Failed to get rest config", "error", err)
		return []map[string]interface{}{}
	}

	exec, err := remotecommand2.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		slog.Error("Failed to create executor", "error", err)
		return []map[string]interface{}{}
	}

	var stdout, stderr bytes.Buffer
	var stdin bytes.Buffer
	ctx := context.Background()
	err = exec.StreamWithContext(ctx, remotecommand2.StreamOptions{
		Stdin:  &stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		slog.Error("Exec stream failed", "error", err, "stderr", stderr.String())
		return []map[string]interface{}{}
	}

	slog.Info("Passwd output", "stdout", stdout.String(), "stderr", stderr.String())
	return self.parsePasswdOutput(stdout.String())
}

// parsePasswdOutput 解析 /etc/passwd 输出
func (self PodExec) parsePasswdOutput(content string) []map[string]interface{} {
	var users []map[string]interface{}
	lines := strings.Split(content, "\n")

	slog.Info("Parse passwd output", "lines", len(lines), "content", content)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			uid, err := strconv.Atoi(parts[2])
			if err == nil {
				users = append(users, map[string]interface{}{
					"id":   uid,
					"name": parts[0],
				})
				slog.Info("Parsed user", "name", parts[0], "uid", uid)
			} else {
				slog.Warn("Failed to parse uid", "line", line, "error", err)
			}
		}
	}

	slog.Info("Total users parsed", "count", len(users))
	return users
}

// getLocalUsers 读取宿主机上的用户列表（用于 Agent Pod）
func (self PodExec) getLocalUsers() []map[string]interface{} {
	var users []map[string]interface{}
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return users
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			uid, err := strconv.Atoi(parts[2])
			if err == nil {
				users = append(users, map[string]interface{}{
					"id":   uid,
					"name": parts[0],
				})
			}
		}
	}
	return users
}

func (self PodExec) GetNodePid(http *gin.Context) {

	token := http.MustGet("k8s_token").(string)
	k8sToken := k8s.NewK8sToken(token)
	if !k8sToken.IsK3kCluster() {
		self.JsonResponse(http, gin.H{
			"pid": 1,
		}, nil, 200)
		return
	}
	type VParam struct {
		Namespace string `form:"namespace" binding:"required"`
		PodName   string `form:"podName" binding:"required"`
	}
	params := VParam{}
	if !self.Validate(http, &params) {
		return
	}

	sdk := k8s.NewK8sClient().Sdk
	pod, err := sdk.ClientSet.CoreV1().Pods(k8sToken.GetNamespace()).Get(context.TODO(), params.PodName, metav1.GetOptions{})
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	agentPod, err := sdk.GetDaemonsetAgentPod("default", pod.Status.HostIP)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	pid, err := shell.GetPid(agentPod, pod.Status.ContainerStatuses[0].ContainerID, true, sdk)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponse(http, gin.H{
		"pid": pid,
	}, nil, 200)
}

// getUsersFromCache 从 Pod Annotation 缓存读取用户列表
// 使用 w7.cc/container-id 判断缓存有效性
func (self PodExec) getUsersFromCache(pod *corev1.Pod, currentContainerId string) ([]map[string]interface{}, error) {
	if pod.Annotations == nil {
		return nil, fmt.Errorf("no annotations")
	}

	// 检查 container-id 是否匹配（使用 w7.cc/container-id）
	cachedContainerId := pod.Annotations["w7.cc/container-id"]
	if cachedContainerId != currentContainerId {
		return nil, fmt.Errorf("container id mismatch")
	}

	// 读取缓存的用户列表
	usersJson := pod.Annotations["w7.cc/users-list"]
	if usersJson == "" {
		return nil, fmt.Errorf("no cached users")
	}

	var users []map[string]interface{}
	err := json.Unmarshal([]byte(usersJson), &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// saveUsersToCache 将用户列表保存到 Pod Annotation
// 使用 w7.cc/container-id 判断缓存有效性
func (self PodExec) saveUsersToCache(pod *corev1.Pod, containerId string, users []map[string]interface{}) error {
	usersJson, err := json.Marshal(users)
	if err != nil {
		return err
	}

	sigClient, err := k8s.NewK8sClient().ToSigClient()
	if err != nil {
		return err
	}

	_, err = controllerutil.CreateOrPatch(context.TODO(), sigClient, pod, func() error {
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		pod.Annotations["w7.cc/users-list"] = string(usersJson)
		return nil
	})

	return err
}
