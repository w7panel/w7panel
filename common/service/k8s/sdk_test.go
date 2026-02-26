package k8s

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdapply "k8s.io/kubectl/pkg/cmd/apply"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func TestSdk_ApplyYaml(t *testing.T) {
	type fields struct {
		restConfig         *rest.Config
		ClientSet          *kubernetes.Clientset
		Ctx                context.Context
		Namespace          string
		serviceAccountName string
		DynamicClient      *dynamic.DynamicClient
		RestMapper         meta.RESTMapper
	}
	type args struct {
		Yamlbytes []byte
		options   ApplyOptions
	}
	file, _ := os.MkdirTemp("", "kompose")
	print(file)

	// file.WriteString("test")
	bytes, err := os.ReadFile("/cloudide/workspace/k8s-offline/test.txt")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		args    args
		want    *unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "nginx",
			args: args{Yamlbytes: bytes, options: ApplyOptions{Namespace: "ns-mtmhshtr"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := NewK8sClient()
			_, err := self.ApplyYaml(tt.args.Yamlbytes, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sdk.ApplyYaml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Sdk.ApplyYaml() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestSdk_GetRestMapping(t *testing.T) {

	type args struct {
		apiVersion string
		kind       string
	}
	tests := []struct {
		name    string
		args    args
		want    *meta.RESTMapping
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "test",
			args:    args{apiVersion: "v1", kind: "Pod"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := NewK8sClientInner()
			// token := "..LG1u7sm7XwoUdFAgupVPML82--E8UomZ7uJdKnHh9mdJT26Azald2uQUbwCl3vLvirDwwFFJUla2QNUar3ijSoxm4W9Tr4xW_dL9O6jY9rOpZVYsY1Sl_cfQDqSvKHQXbop7CsIYLSYTTBUAJd9_RrxrYv4Uuz9_dhcRGSmhwazTgg28YdqYZPgEKP4uTTd62BEYfZoHFwjXMM5k_3CmRV7E4RzLmAQJIJbQzitHQ"
			var token = "eyJhbGciOiJSUzI1NiIsImtpZCI6Iml3WktfcGw1VkMzdHdMei1SMjNpMzhWelQ2V0Iyb3FLSUd4RlFaWlMxbncifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzI0MTQ4NjA3LCJpYXQiOjE3MjQxNDUwMDcsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiYzc2MDZmZWEtYjg4Mi00NWU1LWI1ZGUtM2Q5NDZiZmRlNDc2Iiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImRlYnVnIiwidWlkIjoiMzBiNjg4ZTgtOGRiOS00YzNiLThjMTAtYjQzMTU1MDVjZGM2In19LCJuYmYiOjE3MjQxNDUwMDcsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlYnVnIn0.RY0HfYFQNljaVOXSkFbc59xI-YtGd44GLNJq1hgYRjZt7vViKuvkFfq59bLIz4cYOXS_ldTvkvZo90rY-MBzP9r7a-D_clTr6KSXVqu-_Qn3OJ8HWwaZQvOMjX32zItGSyH-Wdu-liaQjMQAE46NIsPd25ReaIPsVtiRzncrXcuFNCSUd7rmWdzE8hlgEkdir3fH7I5Ct-DHVOzFTFP9QQHJyVLCWKs9iKTGJkG-QPQoEILEKGsubrHuLNvnNI_nk46of3kQYV3exPIGCokrAwKzCdCiDmtJEzUq-clpVVzde0ocllqSB_T-0UINn1MimXg-3F4XmgCjssWLSU73ZQ"
			self, err := self.Channel(token)
			if err != nil {
				t.Error(err)
			}
			// _, err := self.ClientSet.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
			// if err != nil {
			// 	t.Error(err)
			// 	// return
			// }

			_, _, err = self.ClientSet.Discovery().ServerGroupsAndResources()
			if err != nil {
				t.Error(err)
			}
			got, err := self.GetRestMapping(tt.args.apiVersion, tt.args.kind)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sdk.GetRestMapping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				//t.Errorf("Sdk.GetRestMapping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSdk_Register(t *testing.T) {
	self := NewK8sClient()
	if err := self.Register("admin8", "123456", "default", "cluster-admin", true, "normal"); err != nil {
		t.Log(err)
	}
}

func TestSdk_ApplyRaw(t *testing.T) {
	type fields struct {
		restConfig         *rest.Config
		ClientSet          *kubernetes.Clientset
		Ctx                context.Context
		Namespace          string
		serviceAccountName string
		DynamicClient      *dynamic.DynamicClient
		RestMapper         meta.RESTMapper
	}
	type args struct {
		data    []byte
		options ApplyOptions
		wantErr bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "invalid yaml",
			args: args{data: []byte("invalid:"), options: ApplyOptions{Namespace: "default"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("KUBERNETES_SERVICE_HOST", "172.16.1.13")
			os.Setenv("KUBERNETES_SERVICE_PORT", "6443")
			token := "eyJhbGciOiJSUzI1NiIsImtpZCI6Iml3WktfcGw1VkMzdHdMei1SMjNpMzhWelQ2V0Iyb3FLSUd4RlFaWlMxbncifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzI2MTA4NjkzLCJpYXQiOjE3MjYxMDUwOTMsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiOTBiZThjNGMtYTc4OS00ZmQ5LWFkZWYtYzIyMDljM2JhNDBlIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImFkbWluIiwidWlkIjoiNWJlYmIzYjUtOGVjNS00NWJhLTgwYmMtOTg4OWRjZmU0ZGYwIn19LCJuYmYiOjE3MjYxMDUwOTMsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmFkbWluIn0.lWyApFz9CvIIuEdPJCsb-xVR51rc5GliKDUa9gvSV1eB2pVXQVfEneY8IqxWOeFiNw70n0e6jZUxJFbFHEVTVf6BiHnTJkH9yn_9pr4lwI4CG_JxwW-6Q-i3PkdaBA6H-4U_wyXJFoYpTq4IY5qftcfP-5ann9wakQruOxl9GwPyyRVVzcGGdxOmtxLEvYGfa7jN3j4Pc6u2PqfMeVxmBN-Poc0S5rczxIGzRUyooEhKUogJZduCcFTEUboqfftml83-QuXpFdSaLNzMjJpSUXU3QF6AIHqNqc4OGvi6ZTcYgVQjabkdvssuXWUhOxkAFrxM6rs1cCbZ1JfimIxKSQ"
			self, err := NewK8sClient().Channel(token)
			if err != nil {
				t.Error(err)
			}
			if err := self.ApplyBytes(tt.args.data, tt.args.options); (err != nil) != tt.args.wantErr {
				panic(err)
				t.Errorf("Sdk.ApplyRaw() error = %v, wantErr %v", err, tt.args.wantErr)
			}
		})
	}
}

func TestSdk_ToKubeconfig(t *testing.T) {
	// cluster := &v1.NamedCluster{}
	sdk := NewK8sClientInner()
	sdk.GetNamespaces()
	// sdk2, _ := sdk.Channel("token")
	tests := []struct {
		name string
		self *Sdk
		// want    *clientcmdapi.Config
		wantErr bool
	}{
		{
			name:    "test",
			self:    sdk,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.self.ToKubeconfig("")
			if (err != nil) != tt.wantErr {
				t.Errorf("ToKubeconfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			yaml, err := helper.K8sObjToYaml(got)
			print(string(yaml))
		})
	}
}

func TestSdk_GetApiServerUrl(t *testing.T) {
	sdk := NewK8sClient()
	api, err := sdk.GetApiServerUrl()
	if err != nil {
		t.Error(err)
	}
	t.Log(api)
}

func TestSdk_GetApiConfigmap(t *testing.T) {
	sdk := NewK8sClient()
	api, err := sdk.ClientSet.CoreV1().ConfigMaps("default").Get(context.TODO(), "registries123", metav1.GetOptions{})
	os.WriteFile("/tmp/test.yaml", []byte(api.Data["default.cnf"]), 0644)
	print(api.Data["default-cnf"])
	if err != nil {
		t.Error(err)
	}
	t.Log(api)
}

func TestSdk_GetContainerPid(t *testing.T) {
	sdk := NewK8sClient()
	pod, err := sdk.GetDaemonsetAgentPod("default", "10.0.72.46")
	if err != nil {
		t.Error(err)
	}
	pid, err := sdk.GetContainerPid(pod, "containerd://d8805359e736a93e4022d941cc7e3989058680459dae93eb26af90b21f3fa874")
	if err != nil {
		t.Error(err)
	}

	t.Log(pid)
}

func TestGetTokenSaName(t *testing.T) {
	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImxvTHVFYTh4NTZHY1BTVXZWUmFxNkM3TUQ2aFJPamlfMmtjZjBxWjJjYm8ifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzM5MjQyOTU0LCJpYXQiOjE3MzkyMzkzNTQsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiMjc2ZDlkYmEtYWViMC00NjQ4LTgwY2ItZGJjOGE4ZjA4NzBhIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImFkbWluIiwidWlkIjoiODcxOTEyMDUtMDUwMC00YjQ4LWExMjktNzUzZDIzMjQ5YTdmIn19LCJuYmYiOjE3MzkyMzkzNTQsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmFkbWluIn0.D3r76qBtlLIMtWfFsg0jAZXgxDOVlw_Pexna17BixvbehIsVlAe7EM-Is6wRMkbkPccaoi_J0AHDf-93ZxfhJ5o0Fo4mjVh0TcbMf6ua0ppDvgs0wisl0Mzg9rOCC5oBEJtGHN_ogiIYMPw6C_KEDuVAEliZ4xaz9oYLA6d8bqFgSy2ti9IKSDy2t3NhDK_Dy_Sj_8uOzz_84SMvclva9EGxmqP2fxkWcU54tU_-HtJTA4VS_xcV37ckJOF1cw9vBp7MnAsz97HTp68UJi_6rUjpHBmwkRYN3GcAQ1URCKg2K8JEf7UnQ67lgLVMAzZwABLn3BxH3OOsJv1jWsIbng"
	saName, expireData := GetTokenSaName(token)
	t.Log(saName)
	// 判断date 是否10分钟内过期
	expireData.After(time.Now().Add(10 * time.Minute))
}

func TestSdk_CreateTokenRequest(t *testing.T) {
	self := NewK8sClientInner()
	token, err := self.CreateTokenRequest("admin", 600, []string{})
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
}

func GetK8sClientConfig() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig
}

func TestSdk_GetDeploymentAppByIdentifie(t *testing.T) {
	sdk := NewK8sClient()
	deploymentApps, err := sdk.GetDeploymentAppByIdentifie(Namespace, "w7-mysql")
	if err != nil {
		t.Error(err)
	}
	t.Log(deploymentApps)
}

func TestApplyCmd(t *testing.T) {
	sdk := NewK8sClient()
	factory := cmdutil.NewFactory(sdk.Sdk)
	stream := genericiooptions.NewTestIOStreamsDiscard()
	applyCmd := cmdapply.NewCmdApply("kubectl", factory, stream)
	// rootcmd :=
	args2 := []string{"-f", "/tmp/test.yaml"}
	stream.In = bytes.NewReader([]byte("apply -f /tmp/test.yaml"))
	flags := cmdapply.NewApplyFlags(stream)
	// file := "/tmp/test.yaml"
	// flags.DeleteFlags.FileNameFlags.Filenames = &[]string{file}
	applyCmd.Run = nil
	applyCmd.RunE = func(cmd *cobra.Command, args []string) error {

		o, err := flags.ToOptions(factory, cmd, "kubectl", args)
		if err != nil {
			return err
		}
		err = o.Validate()
		if err != nil {
			slog.Error("err", "err", err)
			return err
		}
		err = o.Run()
		if err != nil {
			slog.Error("err", "err", err)
			return err
		}
		return nil
	}
	applyCmd.SetIn(stream.In)
	applyCmd.SetArgs(args2)
	err := applyCmd.Execute()
	if err != nil {
		t.Error(err)
	}

}

func TestFactory(t *testing.T) {
	os.Setenv("SDK_DEBUG", "true")
	// token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IlRuSTlhci1sQ3lRcXBJNVVTSmdvOUlGY0NhM3lIOTRxNmN5TWxnVTlNeWsifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzUwMjQ1MTE4LCJpYXQiOjE3NTAyNDQ1MTgsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiMDllMDg2ZTAtODE1NC00YzRiLTgwZmItYjgzZmQwZDQ0N2VmIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6Ims4IiwidWlkIjoiMWEzMjk5ZjAtYzliOC00YzI0LTliYjUtNTllNTcwMzU1MDM5In19LCJuYmYiOjE3NTAyNDQ1MTgsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0Oms4In0.m47xE157b_Hw9W982fnP3xrb-GYwtNsTCuyk4DTO5o5YyE7JCHqMgJUgLmZfdFd1g_GEGsNOJIzd8M5U6H8LncGvwiTNZlN7xhlQuAHArPy-lQ1R70vSjYnXlNVB9-Wprv_jQ7BPa7SYng_GnueWeFDCNN9uTiGr7CRjtSzD38eKg9orWttWNDpVmQOyN_o5reWcflPXOwUJphL4Vdxh8k1IMO3klu0CAQ_pe_etF4GIm_nUusIfyWGp-0mObSXZ41_VxRqC4ayuEdJHYklYqMIRs2kqxT2rRBrZKXi42kHuQx26OUZX_pqcseQpP3DKAk-nkxiReA0uACUaFx9nhQ"
	sdkroot := NewK8sClient()

	k3kconfig := K3kConfig{
		Name:      "k8",
		Namespace: "k3k-k8",
		ApiServer: "test",
	}
	sdk, err := sdkroot.GetK3kClusterSdkByConfig(&k3kconfig)
	if err != nil {
		t.Error(err)
	}
	sa, err := sdk.GetNamespaces()
	if err != nil {
		t.Error(err)
	}
	t.Log(sa)
}

func TestApplyBytes(t *testing.T) {
	data, err := os.ReadFile("/home/workspace/k8s-offline/kodata/test/mcp/mcpserver-mysql.yaml")
	if err != nil {
		t.Error(err)
	}
	sdk := NewK8sClient()
	err = sdk.ApplyBytes(data, ApplyOptions{Namespace: "default"})
	if err != nil {
		t.Error(err)
	}
}
