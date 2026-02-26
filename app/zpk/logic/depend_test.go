package logic

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
)

func TestDependEnv_LoadHelmEnv(t *testing.T) {
	client := k8s.NewK8sClientInner()
	d := NewDependEnv(client)
	identifie := "gpu_hami"
	namespace := "default"
	result, err := d.LoadHelmEnv(identifie, namespace)
	if err != nil {
		t.Errorf("LoadHelmEnv failed: %v", err)
	}
	if result.Installed {
		t.Errorf("LoadHelmEnv result.Installed should be false")
	}
	if len(result.Envs) != 0 {
		t.Errorf("LoadHelmEnv result.Envs should be empty")
	}
}

func TestDependEnv_LoadLastVersionEnv_MaybeNames(t *testing.T) {
	client := k8s.NewK8sClientInner()
	d := NewDependEnv(client)
	name := "w7-pros-28694-jyvtanqm9x"
	namespace := "default"
	result, err := d.LoadLastVersionEnv(name, namespace)
	if err != nil {
		t.Errorf("LoadLastVersionEnv failed: %v", err)
	}
	if !result.Installed {
		t.Errorf("LoadLastVersionEnv result.Installed should be true for maybe names case")
	}
}

func TestDependEnv_LoadLastVersionEnv_NotFound(t *testing.T) {
	client := k8s.NewK8sClientInner()
	d := NewDependEnv(client)
	name := "non-existent"
	namespace := "default"
	_, err := d.LoadLastVersionEnv(name, namespace)
	if err == nil {
		t.Errorf("LoadLastVersionEnv should return error for non-existent resource")
	}
}

func TestDependEnv_LoadLastVersionEnv_Deployment(t *testing.T) {
	client := k8s.NewK8sClientInner()
	d := NewDependEnv(client)
	name := "test-deployment"
	namespace := "default"
	result, err := d.LoadLastVersionEnv(name, namespace)
	if err != nil {
		t.Errorf("LoadLastVersionEnv failed: %v", err)
	}
	if !result.Installed {
		t.Errorf("LoadLastVersionEnv result.Installed should be true for deployment")
	}
}

func TestDependEnv_LoadLastVersionEnv_Helm(t *testing.T) {
	client := k8s.NewK8sClientInner()
	d := NewDependEnv(client)
	name := "test-helm"
	namespace := "default"
	result, err := d.LoadLastVersionEnv(name, namespace)
	if err != nil {
		t.Errorf("LoadLastVersionEnv failed: %v", err)
	}
	if result.Installed {
		t.Errorf("LoadLastVersionEnv result.Installed should be false for helm")
	}
}
