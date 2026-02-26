package controller

import (
	"os"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/cp"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func TestPodExec_KubectlCp(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := k8s.NewK8sClient().PodExecClient()

			factory := cmdutil.NewFactory(client)
			// inb := []byte("")
			// out := os.Stdout
			// in := os.Stdin
			// stream := genericclioptions.IOStreams{
			// 	In:     in,
			// 	Out:    out,
			// 	ErrOut: out,
			// }
			// reader, writer := io.Pipe()
			stream := genericclioptions.IOStreams{
				In:     os.Stdin,
				Out:    os.Stdout,
				ErrOut: os.Stderr,
			}

			cmd := cp.NewCmdCp(factory, stream)
			// cmd.SetArgs([]string{"k8s-offline-8484474cb4-8mrqq:/tmp/test.txt", "/tmp/test.txt"})
			// // cmd.SetOutput(out)
			cmd.SetArgs([]string{"/tmp/test.txt", "k8s-offline-8484474cb4-8mrqq:/tmp/test.txt"})

			err := cmd.Execute()
			// restConfig, _ := client.ToRESTConfig()

			// copyOptions := cp.NewCopyOptions(stream)
			// copyOptions.Complete(factory, cmd, args)
			// copyOptions.Clientset = client.ClientSet
			// copyOptions.ClientConfig = restConfig
			// copyOptions.Container = "k8s-offline"

			// err := copyOptions.Run()

			if err != nil {
				// print(err.Error() + "--------s")
				panic(err)
				t.Error(err)
			}
			// print(out.String() + "\n")
		})
	}
}
