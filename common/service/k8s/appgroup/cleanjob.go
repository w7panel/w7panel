package appgroup

import (
	"encoding/json"
	"strings"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
	"gitee.com/we7coreteam/k8s-offline/k8s/pkg/apis/appgroup/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ToUninstallTmpJob(group1 *v1alpha1.AppGroup, shellType string) *batchv1.Job {
	jobName := group1.Name + "-uninstall-" + strings.ToLower(helper.RandomString(12))
	afterSeconds := int32(300)
	// matchlabels := map[string]string{
	// 	"job": jobName,
	// }
	shellstr, ok := group1.Annotations["w7.cc/shells"]
	if !ok {
		return nil

	}

	var shells []types.Shell
	err := json.Unmarshal([]byte(shellstr), &shells)
	if err != nil {
		return nil
	}
	shell := types.GetShellByType(shells, shellType)
	if (shell == nil) || (len(shell.GetShell()) == 0) {
		return nil
	}
	cmd := []string{"sh", "-c", shell.GetShell()}
	image := shell.GetImage()
	if len(image) == 0 {
		image = helper.SelfImage()
	}
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: group1.Namespace,
		},
		Spec: corev1.PodSpec{
			DNSPolicy:     corev1.DNSClusterFirstWithHostNet,
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Command:         cmd,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Image:           image,
					Name:            jobName,
				},
			},
			ServiceAccountName: helper.ServiceAccountName(),
		},
	}
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Template:                podTemplate,
			TTLSecondsAfterFinished: &afterSeconds,
		},
	}
	return job
}
