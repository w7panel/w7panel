package appgroup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/w7panel/w7panel/common/service/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOnAdd(t *testing.T) {
	sdk := k8s.NewK8sClient()
	ingress, err := sdk.ClientSet.NetworkingV1().Ingresses("default").Get(sdk.Ctx, "ing-kubhutmv1", metav1.GetOptions{})
	assert.Nil(t, err)
	sigClient, err := sdk.ToSigClient()
	assert.Nil(t, err)
	OnAddIngress(sigClient, ingress)
}

func TestOnDel(t *testing.T) {
	sdk := k8s.NewK8sClient()
	ingress, err := sdk.ClientSet.NetworkingV1().Ingresses("default").Get(sdk.Ctx, "ing-alqzvbhs", metav1.GetOptions{})
	assert.Nil(t, err)
	sigClient, err := sdk.ToSigClient()
	assert.Nil(t, err)
	OnDeleteIngress(sigClient, ingress)
}
