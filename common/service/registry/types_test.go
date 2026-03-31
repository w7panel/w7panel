package registry

import (
	"testing"
)

func TestParse(t *testing.T) {
	ref, err := parseRef("docker.io/library/nginx:latest")
	if err != nil {
		t.Error(err)
	}
	t.Log(ref)
	//ccr.ccs.tencentyun.com/afan-public/mysql@sha256:23154f427abcf32b79634e44ff0f0d6da9d4dfb3f5437c110d602b6e61207255
}

func TestParse2(t *testing.T) {
	ref, err := parseRef("ccr.ccs.tencentyun.com/afan-public/mysql@sha256:23154f427abcf32b79634e44ff0f0d6da9d4dfb3f5437c110d602b6e61207255")
	if err != nil {
		t.Error(err)
	}
	t.Log(ref)
	//
}
