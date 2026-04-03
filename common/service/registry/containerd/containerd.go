package containerd

import (
	client "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/core/images"
)

type containerdClient struct {
	*client.Client
	store    content.Store
	imageSvc images.Store
}

func NewContainerDApi(client *client.Client) {

}
