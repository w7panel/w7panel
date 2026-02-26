// nolint
package zpk

import (
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/zpk/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockBuildImageOption struct {
	types.BuildImageInterface
}

func TestToZpkBuildJob(t *testing.T) {
	params := &types.BuildImageParams{}
	params.DockerRegistry = types.DockerRegistry{
		Host:      "ccr.ccs.tencentyun.com",
		Username:  "446897682",
		Password:  "Fykdmx521",
		Namespace: "afan-public",
	}
	params.DockerfilePath = "Dockerfile"
	params.BuildContext = "/workspace/"
	params.PushImage = "ccr.ccs.tencentyun.com/afan-public/test:latest"
	params.ZipUrl = "http://218.23.2.55:9090/k8s/download/Dockerfile.zip?api-token=eyJhbGciOiJSUzI1NiIsImtpZCI6IklZd3JvWW5mQWM5Uk1VaWlhVmVLOWQzcE1SZFoteloyUXhCOC10YlQxakEifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzU2Nzk3NDEyLCJpYXQiOjE3NTY3OTM4MTIsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiMjQ2MDQxMDgtMTdiNy00ODAyLTkzOGUtMmU2NjIxNTEyZTBhIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImFkbWluIiwidWlkIjoiOWE3YWVkYWItYTU2Yi00NzQyLWFmNTEtZmYyMmIyZGM4ZDZkIn19LCJuYmYiOjE3NTY3OTM4MTIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmFkbWluIn0.Lpn1s-4LmBMELhwQOehowi-XJ17auNkTKW-Jqf6rlwJYSDSIXG_zVtzzBkxLeeBG6PPOps0hya8y5DahTdSW1zUEWUlFmtpj0mm3LcBulQerIBaLCzBnJvNL8dw2tL7nl_JsLqJK43MKdtyE0ZaykcQ9kY-QJmRzzzqCryAzC4APOH6AsLBfestsKSOQxGANmXq5NRhgpUNAViKMNYqQMzfWUWuIDMPbawNouQ5rnq9-RPKrS3uNpzlGtNJEnrcbZM3_q4HzihSC5JfcKKZyHUgPz1I2QlvrSeKJAZJz_AfTPIZ1rR-ye95uTZqIUVVyWdIarmtaPPQnIb6C60tlgA"
	params.Identifie = "test"
	params.NotifyCompletionUrl = "http://192.168.0.1:5000/test"
	// params.HostNetwork = false
	// params.HostAliases = []types.HostAlias{
	// 	{
	// 		IP:        "192.168.0.1",
	// 		Hostnames: []string{"192.168.0.1"},
	// 	},
	// }
	params.Title = "test1"
	params.Labels = map[string]string{
		"test": "test",
	}
	params.BuildJobName = "test1"

	sdk := k8s.NewK8sClient()
	// job := ToZpkBuildJob(params)
	// sdk.ClientSet.BatchV1().Jobs("default").Create(sdk.Ctx, job, metav1.CreateOptions{})

	cronjob := ToZpkBuildCronJob(params, "*/1 * * * *")
	sdk.ClientSet.BatchV1().CronJobs("default").Create(sdk.Ctx, cronjob, metav1.CreateOptions{})

}
