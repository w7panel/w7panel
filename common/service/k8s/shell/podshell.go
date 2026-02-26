package shell

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
)

type PodShell struct {
	cmd     string
	srcPath string
	toPath  string
	pid     string
	args    []string
	subpid  string
}

func NewPodShell(cmd, src string, toPath string, args []string, pid string, subpid string) *PodShell {

	if strings.HasPrefix(src, "/") {
		//移除 /
		src = strings.TrimPrefix(src, "/")
	}
	if strings.HasPrefix(toPath, "/") {
		//移除 /
		toPath = strings.TrimPrefix(toPath, "/")
	}
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "/") {
			//移除 /
			args[i] = filepath.Clean(args[i][1:])
		}
	}
	return &PodShell{
		cmd:     cmd,
		srcPath: filepath.Clean(src),
		toPath:  filepath.Clean(toPath),
		args:    args,
		pid:     pid,
		subpid:  subpid,
	}
}

func (s *PodShell) procRoot() string {
	root := "/proc/" + s.pid + "/root"
	_, err := strconv.Atoi(s.subpid)
	if s.subpid != "" && s.subpid != "0" && s.subpid != "undefined" && err == nil {
		root = "/proc/" + s.pid + "/root/proc/" + s.subpid + "/root"
	}
	return root
}

func (s *PodShell) Run() error {
	// slog.Info("pod shell run", "cmd", s.cmd, "srcPath", s.srcPath, "toPath", s.toPath, "args", s.args)
	err := os.Chdir(s.procRoot())
	// err := os.Chdir("/tmp")
	if err != nil {
		slog.Error("chdir error", "error", err)
		return err
	}
	if s.cmd == "ls" {
		return s.RunLs()
	}

	if s.cmd == "pwd" {
		return s.Runsh("pwd")
	}

	if s.cmd == "cat" {
		return s.Runsh("cat", s.srcPath)
	}

	if s.cmd == "cp" {
		// args := append([]string{"-r"}, s.args...)
		return s.Runsh("cp", "-r", s.srcPath, s.toPath)
	}

	if s.cmd == "rm" {
		args := append([]string{"-rf"}, s.args...)
		return s.Runsh("rm", args...)
	}

	if s.cmd == "kill" {
		return s.Runsh("kill", "-9", s.pid)
	}
	if s.cmd == "mkdir" {
		return s.Runsh("mkdir", "-p", s.srcPath)
	}
	if s.cmd == "touch" {
		return s.Runsh("touch", s.srcPath)
	}
	if s.cmd == "mv" {
		return s.Runsh("mv", s.srcPath, s.toPath)
	}
	if s.cmd == "chmod" {
		args := append([]string{"-R"}, s.args...)
		return s.Runsh("chmod", args...)
	}
	if s.cmd == "chown" {
		args := append([]string{"-R"}, s.args...)
		return s.Runsh("chown", args...)
	}
	if s.cmd == "du" {
		return s.Runsh("du", "-b", s.srcPath)
	}
	if s.cmd == "zip" {
		manyArgs := []string{"-rj", s.srcPath}
		manyArgs = append(manyArgs, s.args...)
		return s.Runsh("zip", manyArgs...)
	}
	if s.cmd == "unzip" {
		return helper.Unzip(s.srcPath, s.toPath, false)
	}
	if s.cmd == "untar" {
		if strings.HasSuffix(s.srcPath, ".tar") {
			return s.Runsh("tar", "-xvf", s.srcPath, "-C", s.toPath)
		}
		if strings.HasSuffix(s.srcPath, ".tar.gz") || strings.HasSuffix(s.srcPath, ".tgz") {
			return s.Runsh("tar", "-xzvf", s.srcPath, "-C", s.toPath)
		}
		if strings.HasSuffix(s.srcPath, ".tar.bz2") || strings.HasSuffix(s.srcPath, ".tbz2") {
			return s.Runsh("tar", "-xjvf", s.srcPath, "-C", s.toPath)
		}
		if strings.HasSuffix(s.srcPath, ".tar.xz") || strings.HasSuffix(s.srcPath, ".txz") {
			return s.Runsh("tar", "-xJvf", s.srcPath, "-C", s.toPath)
		}
	}
	if s.cmd == "tar" {
		if strings.HasSuffix(s.toPath, ".tar") {
			return s.Runsh("tar", "-cvf", s.toPath, s.srcPath)
		}
		if strings.HasSuffix(s.toPath, ".tar.gz") {
			return s.Runsh("tar", "-czvf", s.toPath, s.srcPath)
		}
		if strings.HasSuffix(s.toPath, ".tar.bz2") {
			return s.Runsh("tar", "-cjvf", s.toPath, s.srcPath)
		}
		if strings.HasSuffix(s.toPath, ".tar.xz") {
			return s.Runsh("tar", "-cJvf", s.toPath, s.srcPath)
		}
	}

	return nil

}

/*
*
ls -l -AF /home/wwwroot | awk -v passwd="/etc/passwd" -v group="/etc/group" '

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
	}'
*/
func (s *PodShell) RunLs() error {
	shellstr := `
ls -ln -AF --full-time ` + s.procRoot() + "/" + s.srcPath + ` | awk -v passwd="` + s.procRoot() + `/etc/passwd" -v group="` + s.procRoot() + `/etc/group" '
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
}'
	`
	// print(shellstr)
	return s.Runsh("sh", "-c", shellstr)
}

func (s *PodShell) ChrootSh(chroot string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	// 获取当前进程的所有环境变量
	cmd.Env = os.Environ()
	// 设置新的环境变量
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")

	if chroot != "" {
		// cmd.Dir = "ttytmp" //*ttyChrootDir 不能用 /
		// cmd.Env = []string{
		// 	"TERM=xterm-256color",
		// 	"KUBERNETES_TOKEN=" + config.BearerToken,
		// 	"KUBERNETES_SERVICE_HOST=" + os.Getenv("KUBERNETES_SERVICE_HOST"),
		// 	"KUBERNETES_SERVICE_PORT=" + os.Getenv("KUBERNETES_SERVICE_PORT"),
		// 	"KUBERNETES_CAFILE=" + "/.kube/ca.crt",
		// 	"HOME=" + os.Getenv("HOME"),
		// }
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Chroot: chroot,
		}
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Chroot: chroot,
	}
	var out bytes.Buffer
	var errOut bytes.Buffer

	// 设置命令的标准输出和标准错误输出
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	// 执行命令
	err := cmd.Run()
	if err != nil {
		// slog.Info("Command failed with error: %s\n", err)
		print(errOut.String())
		// fmt.Print(errOut.String())
		// fmt.Printf("Command failed with error: %s\n", err)
		// fmt.Printf("Error output:\n%s\n", errOut.String())
		return err
	}
	// fmt.Print(out.String())
	// slog.Info("Command failed with error: %s\n", out.String())
	print(out.String())
	return nil

}

func (s *PodShell) Runsh(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	// 创建一个 bytes.Buffer 用于存储命令的输出
	var out bytes.Buffer
	var errOut bytes.Buffer

	// 设置命令的标准输出和标准错误输出
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	// 执行命令
	err := cmd.Run()
	if err != nil {
		// slog.Info("Command failed with error: %s\n", err)
		print(errOut.String())
		// fmt.Print(errOut.String())
		// fmt.Printf("Command failed with error: %s\n", err)
		// fmt.Printf("Error output:\n%s\n", errOut.String())
		return err
	}
	// fmt.Print(out.String())
	// slog.Info("Command failed with error: %s\n", out.String())
	print(out.String())
	return nil
}
