package ingress

import (
	"encoding/json"
	"strings"
	"testing"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_syncRootDomain(t *testing.T) {
	sdk := k8s.NewK8sClient()

	ing, err := sdk.ClientSet.NetworkingV1().Ingresses("default").Get(sdk.Ctx, "ing-gverlhgv", metav1.GetOptions{})
	domains := []domain{
		{
			Name:        "a",
			Host:        "a.com",
			AutoSsl:     false,
			SslRedirect: false,
		},
		{
			Name:        "b",
			Host:        "b.com",
			AutoSsl:     true,
			SslRedirect: true,
		},
	}
	data, err := json.Marshal(domains)
	if err != nil {
		t.Error(err)
	}
	ing.Annotations["w7.cc/child-hosts"] = string(data)
	sdk.ClientSet.NetworkingV1().Ingresses("default").Update(sdk.Ctx, ing, metav1.UpdateOptions{})
}

func Test_syncRoot(t *testing.T) {
	sdk := k8s.NewK8sClient()
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
	}
	ing, err := sdk.ClientSet.NetworkingV1().Ingresses("default").Get(sdk.Ctx, "ing-cggfcihm", metav1.GetOptions{})
	sync := &ingressSync{client: client, ingress: &ingress{ing}}
	sync.syncRoot()
}

func Test_syncChild(t *testing.T) {
	sdk := k8s.NewK8sClient()
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
	}
	ing, err := sdk.ClientSet.NetworkingV1().Ingresses("default").Get(sdk.Ctx, "ing-hlrntmgb", metav1.GetOptions{})
	sync := &ingressSync{client: client, ingress: &ingress{ing}}
	sync.syncChild()
}

func Test_syncRootDelete(t *testing.T) {
	sdk := k8s.NewK8sClient()
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
	}
	ing, err := sdk.ClientSet.NetworkingV1().Ingresses("default").Get(sdk.Ctx, "ing-drttficj", metav1.GetOptions{})
	sync := &ingressSync{client: client, ingress: &ingress{ing}}
	sync.syncRootDelete()

}

func Test_syncChildDelete(t *testing.T) {
	sdk := k8s.NewK8sClient()
	client, err := sdk.ToSigClient()
	if err != nil {
		t.Error(err)
	}
	ing, err := sdk.ClientSet.NetworkingV1().Ingresses("default").Get(sdk.Ctx, "ing-byndqqar", metav1.GetOptions{})
	sync := &ingressSync{client: client, ingress: &ingress{ing}}
	sync.syncChildDelete()
}

func TestMd5(t *testing.T) {
	md5str := helper.StringToMD5("/ccc")
	t.Log(strings.ToLower(md5str))
}
