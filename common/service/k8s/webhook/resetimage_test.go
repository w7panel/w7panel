// nolint
package webhook

import (
	"testing"
)

// ccr.ccs.tencentyun.com/afan-public/nginx@sha256:05cb367d77e6e84ac49023e4ec4cd8901596e0f1b4cbb174aa092bcb99a2aaa6
func TestReSetDeploymentImage(t *testing.T) {
	ReSetDeploymentImage("default", "nginx-wlyonjug")
}

func TestResetImage(t *testing.T) {
	ResetImageNow("default", "nginxstate-wlyonjug", "StatefulSet", map[string]string{"w7.cc/image-to-sha256": "true"})
}
