package shell

import (
	"log/slog"
	"os"
	"strconv"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	k3sSwapAnnotation           = "w7.cc.swap"
	k3sSwapLastModifyAnnotation = "w7.cc.swap/last-modify" //k3s 非k3s 说明人工设置
)

func (s *k3sConfigController) initK3sNodeSwap(node *v1.Node) error {
	if !isCurrentDaemonsetNode(node) {
		return nil
	}
	openstr := "false"
	if isEnabledSwap() {
		openstr = "true"
	}
	node.Annotations[k3sSwapAnnotation] = openstr
	node.Annotations[k3sSwapLastModifyAnnotation] = "k3s"
	_, err := s.Sdk.ClientSet.CoreV1().Nodes().Update(s.Ctx, node, metav1.UpdateOptions{})
	if err != nil {
		slog.Error("update node error", "error", err)
		return err
	}
	return nil
}
func (s *k3sConfigController) updateK3sNodeSwap(node *v1.Node) error {
	if !isCurrentDaemonsetNode(node) {
		return nil
	}
	role, ok := node.Annotations[k3sSwapLastModifyAnnotation]
	if !ok {
		return nil
	}
	if role == "k3s" {
		return nil
	}
	isOpen, ok := node.Annotations[k3sSwapAnnotation]
	if !ok {
		return nil
	}
	opend := isOpen == "true"
	err := setUpSwap(opend)
	if err != nil {
		slog.Error("swap error", "error", err, "open", opend)
		return err
	}
	node.Annotations[k3sSwapAnnotation] = strconv.FormatBool(opend)
	node.Annotations[k3sSwapLastModifyAnnotation] = "k3s"
	_, err = s.Sdk.ClientSet.CoreV1().Nodes().Update(s.Ctx, node, metav1.UpdateOptions{})
	if err != nil {
		slog.Error("update node error", "error", err)
		return err
	}
	os.Exit(0)
	// s.restartNode(node)

	return nil
}

func setUpSwap(open bool) error {
	args := "clean"
	if open {
		args = "setup"
	}
	file := "swap.sh"
	err := runFile(file, args)
	if err != nil {
		slog.Error("run swap error", "error", err)
		return err
	}
	return nil
}
func isEnabledSwap() bool {
	err := runFile("swap.sh", "check")
	return err == nil
}

func runFile(fileName string, args string) error {
	checkSwapShell, err := os.ReadFile(os.Getenv("KO_DATA_PATH") + "/shell/" + fileName)
	if err != nil {
		slog.Error("read swap file error", "error", err, "fileName", fileName)
		return err
	}
	err = helper.WriteFileAtomic("/host/tmp/"+fileName, checkSwapShell)
	if err != nil {
		slog.Error("write swap error", "error", err, "fileName", fileName)
		return err
	}
	shell := "chmod +x /tmp/" + fileName + " && /tmp/" + fileName
	if args != "" {
		shell += " " + args
	}
	errstr, err := helper.RunNcenterBinsh(shell)
	if err != nil {
		slog.Error("run swap error", "error", err, "fileName", fileName, "errstr", errstr)
		return err
	}
	return nil
}
